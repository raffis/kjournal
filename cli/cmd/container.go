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
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	corev1alpha1 "github.com/raffis/kjournal/pkg/apis/core/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type containerFlags struct {
	container string
	noColor   bool
	timestamp bool
}

var containerArgs containerFlags

var containerCmd = &cobra.Command{
	Use:   "container",
	Short: "Get container logs",
	Long:  "The container command prints logs from containers",
	Example: `  # Print logs from all pods in the same namespace
  kjoural container -n mynamespace`,
	//ValidArgsFunction: resourceNamesCompletionFunc(logsv1beta1.GroupVersion.WithKind(logsv1beta1.LogKind)),
	RunE: func(cmd *cobra.Command, args []string) error {
		get := getCommand{
			command: &containerCommand{},
			apiType: containerLogAdapterType,
			list:    &containerLogListAdapter{&corev1alpha1.ContainerLogList{}},
		}
		return get.run(cmd, args)
	},
}

type containerCommand struct {
}

func (cmd *containerCommand) filter(args []string, opts *metav1.ListOptions) error {
	var fieldSelector []string
	if opts.FieldSelector != "" {
		fieldSelector = strings.Split(opts.FieldSelector, ",")
	}

	if len(args) == 1 {
		fieldSelector = append(fieldSelector, fmt.Sprintf("pod=%s", args[0]))
	}

	if getArgs.since != "" {
		ts, err := time.ParseDuration(getArgs.since)
		if err != nil {
			return err
		}

		fieldSelector = append(fieldSelector, fmt.Sprintf("metadata.creationTimestamp>%d", time.Now().Unix()*1000-ts.Milliseconds()))
	}

	if containerArgs.container != "" {
		fieldSelector = append(fieldSelector, fmt.Sprintf("container=%s", containerArgs.container))
	}

	opts.FieldSelector = strings.Join(fieldSelector, ",")
	return nil
}

func (cmd *containerCommand) defaultPrinter(obj runtime.Object) error {
	var list corev1alpha1.ContainerLogList
	log, ok := obj.(*corev1alpha1.ContainerLog)
	if ok {
		list.Items = append(list.Items, *log)
	}

	for _, item := range list.Items {
		printContainerLog(item)
	}
	return nil
}

// Log is the object which will be used together with the template to generate
// the output.
type Log struct {
	// Message is the log message itself
	Message string `json:"message"`

	// Node name of the pod
	NodeName string `json:"nodeName"`

	// Namespace of the pod
	Namespace string `json:"namespace"`

	// PodName of the pod
	PodName string `json:"podName"`

	// ContainerName of the container
	ContainerName string `json:"containerName"`

	PodColor       *color.Color `json:"-"`
	ContainerColor *color.Color `json:"-"`
}

// Print prints a color coded log message with the pod and container names
func printContainerLog(log corev1alpha1.ContainerLog) {
	podColor, containerColor := determineColor(log.Pod)
	vm := Log{
		Message:        string(log.Payload),
		PodName:        log.Pod,
		ContainerName:  log.Container,
		PodColor:       podColor,
		ContainerColor: containerColor,
	}

	t := "{{color .PodColor .PodName}} {{color .ContainerColor .ContainerName}} {{.Message}}\n"

	funs := map[string]interface{}{
		"json": func(in interface{}) (string, error) {
			b, err := json.Marshal(in)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
		"parseJSON": func(text string) (map[string]interface{}, error) {
			obj := make(map[string]interface{})
			if err := json.Unmarshal([]byte(text), &obj); err != nil {
				return obj, err
			}
			return obj, nil
		},
		"color": func(color color.Color, text string) string {
			return color.SprintFunc()(text)
		},
	}
	template, err := template.New("log").Funcs(funs).Parse(t)
	if err != nil {
		//return nil, errors.Wrap(err, "unable to parse template")
	}

	var buf bytes.Buffer
	if err := template.Execute(&buf, vm); err != nil {
		//fmt.Fprintf(t.errOut, "expanding template failed: %s\n", err)
		return
	}

	fmt.Printf(buf.String())
}

func init() {
	containerCmd.PersistentFlags().StringVarP(&containerArgs.container, "container", "c", "", "Only dump logs from container names matching. (This is the same as --field-selector container=name)")
	containerCmd.PersistentFlags().BoolVarP(&containerArgs.noColor, "no-color", "", false, "Don't use colors in the default output")
	containerCmd.PersistentFlags().BoolVarP(&containerArgs.timestamp, "timestamp", "t", false, "Print creationTime timestamp in the default output.")

	addGetFlags(containerCmd)
	rootCmd.AddCommand(containerCmd)
}

var containerLogAdapterType = apiType{
	kind:      "ContainerLog",
	humanKind: "containerlog",
	resource:  "containerlogs",
	groupVersion: schema.GroupVersion{
		Group:   "core.kjournal",
		Version: "v1alpha1",
	},
	namespaced: true,
}

type containerLogListAdapter struct {
	*corev1alpha1.ContainerLogList
}

func (h containerLogListAdapter) asClientList() ObjectList {
	return h.ContainerLogList
}

func (h containerLogListAdapter) len() int {
	return len(h.ContainerLogList.Items)
}

var colorList = [][2]*color.Color{
	{color.New(color.FgHiCyan), color.New(color.FgCyan)},
	{color.New(color.FgHiGreen), color.New(color.FgGreen)},
	{color.New(color.FgHiMagenta), color.New(color.FgMagenta)},
	{color.New(color.FgHiYellow), color.New(color.FgYellow)},
	{color.New(color.FgHiBlue), color.New(color.FgBlue)},
	{color.New(color.FgHiRed), color.New(color.FgRed)},
}

func determineColor(podName string) (podColor, containerColor *color.Color) {
	hash := fnv.New32()
	_, _ = hash.Write([]byte(podName))
	idx := hash.Sum32() % uint32(len(colorList))

	colors := colorList[idx]
	return colors[0], colors[1]
}
