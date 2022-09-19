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

	"github.com/liggitt/tabwriter"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/printers"

	corev1alpha1 "github.com/raffis/kjournal/pkg/apis/core/v1alpha1"
)

type auditFlags struct {
	noHeader bool
}

var auditArgs auditFlags

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Get audit events",
	Long:  "The audit command fetchs events from namespaced resources",
	Example: `  # Stream all audit events from the namespace mynamespace
  kjournal audit -n mynamespace
  
  # Stream events from the last 48 hours
  kjournal audit -n mynamespace --since 48h
  
  # Stream events for all deployments
  kjournal audit -n mynamespace deployments
  
  # Stream events for a pod named abc
  kjournal audit -n mynamespace pods/abc`,
	//ValidArgsFunction: resourceNamesCompletionFunc(auditv1.GroupVersion.WithKind(corev1alpha1.AuditEventKind)),
	RunE: func(cmd *cobra.Command, args []string) error {
		get := getCommand{
			command: &auditCommand{
				cmd: cmd,
			},
			apiType: auditEventAdapterType,
			list:    &auditEventListAdapter{&corev1alpha1.AuditEventList{}},
		}
		return get.run(cmd, args)
	},
}

func init() {
	addGetFlags(auditCmd)
	auditCmd.PersistentFlags().BoolVarP(&auditArgs.noHeader, "no-header", "", false, "skip the header when printing the results")

	rootCmd.AddCommand(auditCmd)
}

type auditCommand struct {
	printer *tabwriter.Writer
	cmd     *cobra.Command
}

func (cmd *auditCommand) filter(args []string, opts *metav1.ListOptions) error {
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
	} else if getArgs.timeRange != "" {
		parts := strings.Split(getArgs.timeRange, "-")

		fromTimestamp, err := time.ParseDuration(parts[0])
		if err != nil {
			return err
		}
		toTimestamp, err := time.ParseDuration(parts[1])
		if err != nil {
			return err
		}

		fieldSelector = append(
			fieldSelector,
			fmt.Sprintf("metadata.creationTimestamp<%d", time.Now().Unix()*1000-fromTimestamp.Milliseconds()),
			fmt.Sprintf("metadata.creationTimestamp>%d", time.Now().Unix()*1000-toTimestamp.Milliseconds()),
		)
	}

	opts.FieldSelector = strings.Join(fieldSelector, ",")
	return nil
}

func (cmd *auditCommand) defaultPrinter(obj runtime.Object) error {
	list := &corev1alpha1.AuditEventList{}

	if log, ok := obj.(*corev1alpha1.AuditEvent); ok {
		list.Items = append(list.Items, *log)
	} else if obj, ok := obj.(*corev1alpha1.AuditEventList); ok {
		list = obj
	}

	for _, item := range list.Items {
		if cmd.printer == nil {
			cmd.printer = printers.GetNewTabWriter(cmd.cmd.OutOrStdout())
			fmt.Fprintln(cmd.printer, strings.Join([]string{"RECEIVED", "VERB", "STATUS", "LEVEL", "USERNAME"}, "\t"))
		}

		var code int32
		if item.ResponseStatus != nil {
			code = item.ResponseStatus.Code
		}

		fmt.Fprintf(cmd.printer, "%s\t%s\t%d\t%s\t%s\n",
			item.RequestReceivedTimestamp.String(),
			item.Verb,
			code,
			string(item.Level),
			item.User.Username,
		)
	}

	if cmd.printer != nil {
		cmd.printer.Flush()
	}

	return nil
}

var auditEventAdapterType = apiType{
	kind:      "AuditEvent",
	humanKind: "auditevent",
	resource:  "auditevents",
	groupVersion: schema.GroupVersion{
		Group:   "core.kjournal",
		Version: "v1alpha1",
	},
}

type auditEventListAdapter struct {
	*corev1alpha1.AuditEventList
}

func (h auditEventListAdapter) asClientList() ObjectList {
	return h.AuditEventList
}

func (h auditEventListAdapter) len() int {
	return len(h.AuditEventList.Items)
}
