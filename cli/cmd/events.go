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
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/printers"

	corev1alpha1 "github.com/raffis/kjournal/pkg/apis/core/v1alpha1"
)

type eventsFlags struct {
	noHeader bool
}

var eventsArgs eventsFlags

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Get events events",
	Long:  "The events command fetchs events from namespaced resources",
	Example: `  # Stream all events events from the namespace mynamespace
  kjournal events -n mynamespace
  
  # Stream events from the last 48 hours
  kjournal events -n mynamespace --since 48h
  
  # Stream events for all deployments
  kjournal events -n mynamespace deployments
  
  # Stream events for a pod named abc
  kjournal events -n mynamespace pods/abc`,
	//ValidArgsFunction: resourceNamesCompletionFunc(eventsv1.GroupVersion.WithKind(corev1alpha1.EventKind)),
	RunE: func(cmd *cobra.Command, args []string) error {
		get := getCommand{
			command: &eventsCommand{
				cmd: cmd,
			},
			apiType: eventAdapterType,
			list:    &eventListAdapter{&corev1alpha1.EventList{}},
		}

		return get.run(cmd, args)
	},
}

func init() {
	addGetFlags(eventsCmd)
	eventsCmd.PersistentFlags().BoolVarP(&eventsArgs.noHeader, "no-header", "", false, "skip the header when printing the results")

	rootCmd.AddCommand(eventsCmd)
}

type eventsCommand struct {
	printer *tabwriter.Writer
	cmd     *cobra.Command
}

func (cmd *eventsCommand) filter(args []string, opts *metav1.ListOptions) error {
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
			fieldSelector = append(fieldSelector, fmt.Sprintf("regarding.kind.=%s", kn[0]))
		}

		if len(kn) == 2 {
			fieldSelector = append(fieldSelector, fmt.Sprintf("regarding.name=%s", kn[1]))
		}
	}

	timeSelectors, err := timeRange(getArgs)
	if err != nil {
		return err
	}

	fieldSelector = append(fieldSelector, timeSelectors...)

	opts.FieldSelector = strings.Join(fieldSelector, ",")
	return nil
}

func (cmd *eventsCommand) defaultPrinter(obj runtime.Object) error {
	list := &corev1alpha1.EventList{}

	if log, ok := obj.(*corev1alpha1.Event); ok {
		list.Items = append(list.Items, *log)
	} else if obj, ok := obj.(*corev1alpha1.EventList); ok {
		list = obj
	}

	for _, item := range list.Items {
		if cmd.printer == nil {
			cmd.printer = printers.GetNewTabWriter(cmd.cmd.OutOrStdout())
			fmt.Fprintln(cmd.printer, strings.Join([]string{"LAST SEEN", "TYPE", "REASON", "OBJECT", "MESSAGE"}, "\t"))
		}

		fmt.Fprintf(cmd.printer, "%s\t%s\t%s\t%s/%s\t%s\n",
			cmd.getInterval(item),
			item.Type,
			item.Reason,
			item.Regarding.Kind,
			item.Regarding.Name,
			item.Note,
		)
	}

	if cmd.printer != nil {
		cmd.printer.Flush()
	}

	return nil
}

var eventAdapterType = apiType{
	kind:      "Event",
	humanKind: "event",
	resource:  "events",
	groupVersion: schema.GroupVersion{
		Group:   "core.kjournal",
		Version: "v1alpha1",
	},
	namespaced: true,
}

type eventListAdapter struct {
	*corev1alpha1.EventList
}

func (h eventListAdapter) asClientList() ObjectList {
	return h.EventList
}

func (h eventListAdapter) len() int {
	return len(h.EventList.Items)
}

func (cmd *eventsCommand) getInterval(e corev1alpha1.Event) string {
	var interval string
	firstTimestampSince := translateMicroTimestampSince(e.EventTime)
	if e.EventTime.IsZero() {
		firstTimestampSince = translateTimestampSince(e.DeprecatedFirstTimestamp)
	}
	if e.Series == nil {
		interval = firstTimestampSince
	} else if e.Series.Count > 1 {
		interval = fmt.Sprintf("%s (x%d over %s)", translateTimestampSince(e.DeprecatedLastTimestamp), e.Series.Count, firstTimestampSince)
	} else {
		interval = fmt.Sprintf("%s (x%d over %s)", translateMicroTimestampSince(e.Series.LastObservedTime), e.Series.Count, firstTimestampSince)
	}

	return interval
}

// translateMicroTimestampSince returns the elapsed time since timestamp in
// human-readable approximation.
func translateMicroTimestampSince(timestamp metav1.MicroTime) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}

// translateTimestampSince returns the elapsed time since timestamp in
// human-readable approximation.
func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}
