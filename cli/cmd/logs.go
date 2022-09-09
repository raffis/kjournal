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

type logsFlags struct {
	log       string
	noColor   bool
	timestamp bool
}

var logsArgs logsFlags

var logCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get generic logs",
	Long:  "The log command prints generic logs",
	Example: `  # Print logs from all pods in the same namespace
  kjoural log -n mynamespace`,
	//ValidArgsFunction: resourceNamesCompletionFunc(logsv1beta1.GroupVersion.WithKind(logsv1beta1.LogKind)),
	RunE: func(cmd *cobra.Command, args []string) error {
		get := getCommand{
			apiType: logAdapterType,
			list:    &logListAdapter{&corev1alpha1.LogList{}},
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

					fieldSelector = append(fieldSelector, fmt.Sprintf("metadata.creationTimestamp>%d", time.Now().Unix()*1000-ts.Milliseconds()))
				}

				if logsArgs.log != "" {
					fieldSelector = append(fieldSelector, fmt.Sprintf("log=%s", logsArgs.log))
				}

				opts.FieldSelector = strings.Join(fieldSelector, ",")
				return nil
			},
			defaultPrinter: func(obj runtime.Object) error {
				var list corev1alpha1.LogList
				log, ok := obj.(*corev1alpha1.Log)
				if ok {
					list.Items = append(list.Items, *log)
				}

				for _, item := range list.Items {
					printLog(item)
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

func init() {
	logCmd.PersistentFlags().BoolVarP(&logsArgs.timestamp, "timestamp", "t", false, "Print creationTime timestamp in the default output.")

	addGetFlags(logCmd)
	rootCmd.AddCommand(logCmd)
}

// Print prints a color coded log message with the pod and container names
func printLog(log corev1alpha1.Log) {
	vm := Log{
		Message: string(log.Payload),
	}

	t := "{{.Message}}\n"

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

var logAdapterType = apiType{
	kind:      "Log",
	humanKind: "log",
	resource:  "logs",
	groupVersion: schema.GroupVersion{
		Group:   "core.kjournal",
		Version: "v1alpha1",
	},
}

type logListAdapter struct {
	*corev1alpha1.LogList
}

func (h logListAdapter) asClientList() ObjectList {
	return h.LogList
}

func (h logListAdapter) len() int {
	return len(h.LogList.Items)
}
