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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/raffis/kjournal/cli/pkg/printers"
	auditv1 "github.com/raffis/kjournal/pkg/apis/audit/v1"
)

var clusterAuditCmd = &cobra.Command{
	Use:   "clusteraudit",
	Short: "Get cluster audit events",
	Long:  "The clusteraudit command fetchs events from alle resources without a namespace",
	Example: `  # Stream cluster events from the last 48h
  kjournal clusteraudit --since 48h
  
  # Stream events for all clusterroles
  kjournal clusteraudit clusterroles
  
  # Stream events for the namespace abc
  kjournal clusteraudit namespaces/abc`,
	//ValidArgsFunction: resourceNamesCompletionFunc(auditv1.GroupVersion.WithKind(auditv1.EventKind)),
	RunE: func(cmd *cobra.Command, args []string) error {
		firstIteration := true

		get := getCommand{
			apiType: eventAdapterType,
			list:    &clustereventListAdapter{&auditv1.ClusterEventList{}},
			filter: func(args []string, opts *metav1.ListOptions) error {
				var fieldSelector []string
				if opts.FieldSelector != "" {
					fieldSelector = strings.Split(opts.FieldSelector, ",")
				}

				if len(args) == 1 {
					kn := strings.Split(args[0], "/")
					if len(kn) > 2 {
						return errors.New("expects either resource/name or resource. Invalid number of parts given")
					}

					if len(kn) > 0 {
						fieldSelector = append(fieldSelector, fmt.Sprintf("objectRef.resource=%s", kn[0]))
					}

					if len(kn) == 2 {
						fieldSelector = append(fieldSelector, fmt.Sprintf("objectRef.name=%s", kn[1]))
					}
				}

				if getArgs.since != "" {
					ts, err := time.ParseDuration(getArgs.since)
					if err != nil {
						return err
					}

					fieldSelector = append(fieldSelector, fmt.Sprintf("requestReceivedTimestamp>%d", time.Now().Unix()*1000-ts.Milliseconds()))
				}

				opts.FieldSelector = strings.Join(fieldSelector, ",")
				return nil
			},
			defaultPrinter: func(obj runtime.Object) error {
				var list auditv1.ClusterEventList
				log, ok := obj.(*auditv1.ClusterEvent)
				if ok {
					list.Items = append(list.Items, *log)
				}

				for _, item := range list.Items {
					var headers []string

					if firstIteration {
						headers = []string{"Received", "Verb", "Status", "Level", "Username"}
						firstIteration = false
					}

					err := printers.TablePrinter(headers).Print(cmd.OutOrStdout(), [][]string{[]string{
						item.RequestReceivedTimestamp.String(),
						item.Verb,
						fmt.Sprintf("%d", item.ResponseStatus.Code),
						string(item.Level),
						item.User.Username,
					}})

					if err != nil {
						return err
					}
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
	addGetFlags(clusterAuditCmd)
	clusterAuditCmd.PersistentFlags().BoolVarP(&auditArgs.noHeader, "no-header", "", false, "skip the header when printing the results")

	rootCmd.AddCommand(clusterAuditCmd)
}

var clusterEventAdapterType = apiType{
	kind:      "ClusterEvent",
	humanKind: "clusterevent",
	resource:  "clusterevents",
	groupVersion: schema.GroupVersion{
		Group:   "audit.kjournal",
		Version: "v1beta1",
	},
}

type clustereventListAdapter struct {
	*auditv1.ClusterEventList
}

func (h clustereventListAdapter) asClientList() ObjectList {
	return h.ClusterEventList
}

func (h clustereventListAdapter) len() int {
	return len(h.ClusterEventList.Items)
}
