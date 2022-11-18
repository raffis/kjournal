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
	log string
	//	noColor   bool
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
			command: &logsCommand{},
			apiType: logAdapterType,
			list:    &logListAdapter{&corev1alpha1.LogList{}},
		}
		return get.run(cmd, args)

	},
}

func init() {
	logCmd.PersistentFlags().BoolVarP(&logsArgs.timestamp, "timestamp", "t", false, "Print creationTime timestamp in the default output.")

	addGetFlags(logCmd)
	rootCmd.AddCommand(logCmd)
}

type logsCommand struct {
}

func (cmd *logsCommand) filter(args []string, opts *metav1.ListOptions) error {
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

	if logsArgs.log != "" {
		fieldSelector = append(fieldSelector, fmt.Sprintf("log=%s", logsArgs.log))
	}

	opts.FieldSelector = strings.Join(fieldSelector, ",")
	return nil
}

func (cmd *logsCommand) defaultPrinter(obj runtime.Object) error {
	var list corev1alpha1.LogList
	log, ok := obj.(*corev1alpha1.Log)
	if ok {
		list.Items = append(list.Items, *log)
	}

	for _, item := range list.Items {
		if err := printLog(item); err != nil {
			return err
		}
	}
	return nil
}

// Print prints a color coded log message with the pod and container names
func printLog(log corev1alpha1.Log) error {
	vm := Log{
		Message: string(log.Payload),
	}

	t := "{{.Message}}"

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
		return fmt.Errorf("failed to parse log template: %w", err)
	}

	var buf bytes.Buffer
	if err := template.Execute(&buf, vm); err != nil {
		return fmt.Errorf("failed to execute log template: %w", err)
	}

	fmt.Println(buf.String())
	return nil
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
