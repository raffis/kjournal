/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	watchtools "k8s.io/client-go/tools/watch"
	k8sget "k8s.io/kubectl/pkg/cmd/get"
	"k8s.io/kubectl/pkg/util/interrupt"

	corev1alpha1 "github.com/raffis/kjournal/pkg/apis/core/v1alpha1"
)

type GetFlags struct {
	fieldSelector string
	watch         bool
	noStream      bool
	chunkSize     string
	since         string
	timeRange     string
}

var getArgs GetFlags
var printFlags *k8sget.PrintFlags

func addGetFlags(getCmd *cobra.Command) {
	if printFlags == nil {
		printFlags = k8sget.NewGetPrintFlags()
	}

	getCmd.Flags().StringVarP(printFlags.OutputFormat, "output", "o", *printFlags.OutputFormat, fmt.Sprintf(`Output format. One of: (%s). See custom columns [https://kubernetes.io/docs/reference/kubectl/overview/#custom-columns], golang template [http://golang.org/pkg/text/template/#pkg-overview] and jsonpath template [https://kubernetes.io/docs/reference/kubectl/jsonpath/].`, strings.Join(printFlags.AllowedFormats(), ", ")))
	getCmd.PersistentFlags().StringVarP(&getArgs.since, "since", "", "", "Change the time range from which logs are received. (e.g. `--since=24h`)")
	getCmd.PersistentFlags().StringVarP(&getArgs.timeRange, "range", "", "", "Change the time range from which logs are received. (e.g. `--range=20h-24h`)")
	getCmd.PersistentFlags().BoolVarP(&getArgs.noStream, "no-stream", "", false, "By default all logs are streamed. This behaviour can be disabled. Be mindful that this can lead to an increased memory usage and no output while logs are beeing gathered")
	getCmd.PersistentFlags().BoolVarP(&getArgs.watch, "watch", "w", false, "After dumping all existing logs keep watching for newly added ones")
	getCmd.PersistentFlags().StringVar(&getArgs.fieldSelector, "field-selector", "", "Selector (field query) to filter on, supports '=', '==', '!=', '!=', '>' and '<'. (e.g. --field-selector key1=value1,key2=value2).")
	getCmd.PersistentFlags().StringVarP(&getArgs.chunkSize, "chunk-size", "", "500", "Return large lists in chunks rather than all at once. Pass 0 to disable. This has no impact as long as --no-stream is not set.")
}

// Create the Scheme, methods for serializing and deserializing API objects
// which can be shared by tests.
func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	corev1alpha1.AddToScheme(scheme)

	return scheme
}

func KubeConfig(rcg genericclioptions.RESTClientGetter, opts *Options) (*rest.Config, error) {
	cfg, err := rcg.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("kubernetes configuration load failed: %w", err)
	}

	// avoid throttling request when some Flux CRDs are not registered
	cfg.QPS = opts.QPS
	cfg.Burst = opts.Burst

	return cfg, nil
}

type command interface {
	filter(args []string, opts *metav1.ListOptions) error
	defaultPrinter(obj runtime.Object) error
}

type getCommand struct {
	apiType
	command command
	list    listAdapter
}

func (get getCommand) run(cmd *cobra.Command, args []string) error {
	if getArgs.noStream {
		return get.listObjects(cmd, args)
	}

	return get.streamObjects(cmd, args)
}

func (get getCommand) getClient() (*rest.RESTClient, error) {
	cfg, err := KubeConfig(kubeconfigArgs, kubeclientOptions)
	if err != nil {
		return nil, err
	}

	cfg.GroupVersion = &get.groupVersion
	var Scheme = NewScheme()
	var Codecs = serializer.NewCodecFactory(Scheme)
	cfg.NegotiatedSerializer = Codecs.WithoutConversion()
	cfg.APIPath = "/apis"

	return rest.RESTClientFor(cfg)
}

func (get getCommand) prepareRequest(args []string) (*rest.Request, error) {
	c, err := get.getClient()
	if err != nil {
		return nil, err
	}

	var opts metav1.ListOptions
	selector, err := labels.Parse(getArgs.fieldSelector)
	if err != nil {
		return nil, err
	}

	opts.FieldSelector = selector.String()

	err = get.command.filter(args, &opts)
	if err != nil {
		return nil, err
	}

	r := c.
		Get().
		Resource(get.resource).
		Param("fieldSelector", opts.FieldSelector)

	if get.namespaced {
		r.Namespace(*kubeconfigArgs.Namespace)
	}

	return r, nil
}

func (get getCommand) listObjects(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), rootArgs.timeout)
	defer cancel()
	res := get.list.asClientList()

	r, err := get.prepareRequest(args)
	if err != nil {
		return err
	}

	r.Param("limit", getArgs.chunkSize)
	logger.Debugf("request uri", "uri", r.URL().String())

	response := r.Do(ctx)
	var httpCode int
	response.StatusCode(&httpCode)

	logger.Debugf("apiserver response received", "code", httpCode)

	err = response.Into(res)

	if err != nil {
		return err
	}

	if get.list.len() == 0 {
		logger.Failuref("no %s objects found in %s namespace", get.kind, *kubeconfigArgs.Namespace)
		return nil
	}

	p, err := printFlags.ToPrinter()
	if err != nil {
		return err
	}

	get.list.asClientList().GetObjectKind().SetGroupVersionKind(
		schema.GroupVersionKind{
			Group:   get.groupVersion.Group,
			Version: get.groupVersion.Version,
			Kind:    get.kind,
		},
	)

	if *printFlags.OutputFormat != "" {
		err = p.PrintObj(res, os.Stdout)
		if err != nil {
			return err
		}

		return nil
	}

	return get.command.defaultPrinter(res)
}

func (get getCommand) streamObjects(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), rootArgs.timeout)
	defer cancel()

	r, err := get.prepareRequest(args)
	if err != nil {
		return err
	}

	p, err := printFlags.ToPrinter()
	if err != nil {
		return err
	}

	if getArgs.watch {
		r.Param("watch", "true")
	} else {
		r.Param("limit", "-1")
	}

	w, err := r.Watch(ctx)
	if err != nil {
		return err
	}

	intr := interrupt.New(nil, cancel)
	err = intr.Run(func() error {
		_, err := watchtools.UntilWithoutRetry(ctx, w, func(e watch.Event) (bool, error) {
			objToPrint := e.Object

			if *printFlags.OutputFormat != "" {
				objToPrint.GetObjectKind().SetGroupVersionKind(
					schema.GroupVersionKind{
						Group:   get.groupVersion.Group,
						Version: get.groupVersion.Version,
						Kind:    get.kind,
					},
				)

				if err := p.PrintObj(objToPrint, cmd.OutOrStdout()); err != nil {
					return false, err
				}

				return false, nil
			}

			return false, get.command.defaultPrinter(objToPrint)
		})
		return err
	})

	return err
}
