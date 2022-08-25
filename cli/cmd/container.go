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

	logsv1beta1 "github.com/raffis/kjournal/pkg/apis/container/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type diaryFlags struct {
	container string
	noColor   bool
	timestamp bool
}

var diaryArgs diaryFlags

var diaryCmd = &cobra.Command{
	Use:   "diary",
	Short: "Get container logs",
	Long:  "The diary command prints logs from containers",
	Example: `  # Print logs from all pods in the same namespace
  kjoural diary -n mynamespace`,
	//ValidArgsFunction: resourceNamesCompletionFunc(logsv1beta1.GroupVersion.WithKind(logsv1beta1.LogKind)),
	RunE: func(cmd *cobra.Command, args []string) error {
		get := getCommand{
			apiType: logAdapterType,
			list:    &logListAdapter{&logsv1beta1.LogList{}},
			filter: func(args []string, opts *metav1.ListOptions) error {
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

					fieldSelector = append(fieldSelector, fmt.Sprintf("creationTimestamp>%d", time.Now().Unix()*1000-ts.Milliseconds()))
				}

				if diaryArgs.container != "" {
					fieldSelector = append(fieldSelector, fmt.Sprintf("container=%s", diaryArgs.container))
				}

				opts.FieldSelector = strings.Join(fieldSelector, ",")
				return nil
			},
			defaultPrinter: func(obj runtime.Object) error {
				var list logsv1beta1.LogList
				log, ok := obj.(*logsv1beta1.Log)
				if ok {
					list.Items = append(list.Items, *log)
				}

				for _, item := range list.Items {
					print(item)
				}
				return nil
			},
		}

		if err := get.run(cmd, args); err != nil {
			return err
		}

		return nil
	},
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
func print(log logsv1beta1.Log) {
	podColor, containerColor := determineColor(log.Pod)

	vm := Log{
		Message:        string(log.Unstructured),
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
	diaryCmd.PersistentFlags().StringVarP(&diaryArgs.container, "container", "c", "", "Only dump logs from container names matching. (This is the same as --field-selector container=name)")
	diaryCmd.PersistentFlags().BoolVarP(&diaryArgs.noColor, "no-color", "", false, "Don't use colors in the default output")
	diaryCmd.PersistentFlags().BoolVarP(&diaryArgs.timestamp, "timestamp", "t", false, "Print creationTime timestamp in the default output.")

	addGetFlags(diaryCmd)
	rootCmd.AddCommand(diaryCmd)
}

var logAdapterType = apiType{
	kind:      "Log",
	humanKind: "log",
	resource:  "logs",
	groupVersion: schema.GroupVersion{
		Group:   "container.kjournal",
		Version: "v1beta1",
	},
}

type logListAdapter struct {
	*logsv1beta1.LogList
}

func (h logListAdapter) asClientList() ObjectList {
	return h.LogList
}

func (h logListAdapter) len() int {
	return len(h.LogList.Items)
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
