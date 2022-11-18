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

type podsFlags struct {
	pods      string
	noColor   bool
	timestamp bool
}

var podsArgs podsFlags

var podsCmd = &cobra.Command{
	Use:   "pods",
	Short: "Get pod logs",
	Long:  "The pods command prints logs from pods",
	Example: `  # Print logs from all pods in the same namespace
  kjoural pods -n mynamespace`,
	//ValidArgsFunction: resourceNamesCompletionFunc(logsv1beta1.GroupVersion.WithKind(logsv1beta1.LogKind)),
	RunE: func(cmd *cobra.Command, args []string) error {
		get := getCommand{
			command: &podsCommand{},
			apiType: podsLogAdapterType,
			list:    &podsLogListAdapter{&corev1alpha1.ContainerLogList{}},
		}
		return get.run(cmd, args)
	},
}

func init() {
	podsCmd.PersistentFlags().StringVarP(&podsArgs.pods, "pods", "c", "", "Only dump logs from pods names matching. (This is the same as --field-selector pods=name)")
	podsCmd.PersistentFlags().BoolVarP(&podsArgs.noColor, "no-color", "", false, "Don't use colors in the default output")
	podsCmd.PersistentFlags().BoolVarP(&podsArgs.timestamp, "timestamp", "t", false, "Print creationTime timestamp in the default output.")

	addGetFlags(podsCmd)
	rootCmd.AddCommand(podsCmd)
}

type podsCommand struct {
}

func (cmd *podsCommand) filter(args []string, opts *metav1.ListOptions) error {
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

	if podsArgs.pods != "" {
		fieldSelector = append(fieldSelector, fmt.Sprintf("pods=%s", podsArgs.pods))
	}

	opts.FieldSelector = strings.Join(fieldSelector, ",")
	return nil
}

func (cmd *podsCommand) defaultPrinter(obj runtime.Object) error {
	list := &corev1alpha1.ContainerLogList{}

	if log, ok := obj.(*corev1alpha1.ContainerLog); ok {
		list.Items = append(list.Items, *log)
	} else if obj, ok := obj.(*corev1alpha1.ContainerLogList); ok {
		list = obj
	}

	for _, item := range list.Items {
		if err := printContainerLog(item); err != nil {
			return err
		}
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

	// ContainerName of the pods
	ContainerName string `json:"podsName"`

	PodColor       *color.Color `json:"-"`
	ContainerColor *color.Color `json:"-"`
}

// Print prints a color coded log message with the pod and pods names
func printContainerLog(log corev1alpha1.ContainerLog) error {
	podColor, podsColor := determineColor(log.Pod)
	vm := Log{
		Message:        string(log.Payload),
		PodName:        log.Pod,
		ContainerName:  log.Container,
		PodColor:       podColor,
		ContainerColor: podsColor,
	}

	t := "{{color .PodColor .PodName}} {{color .ContainerColor .ContainerName}} {{.Message}}"

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

var podsLogAdapterType = apiType{
	kind:      "ContainerLog",
	humanKind: "containerlog",
	resource:  "containerlogs",
	groupVersion: schema.GroupVersion{
		Group:   "core.kjournal",
		Version: "v1alpha1",
	},
	namespaced: true,
}

type podsLogListAdapter struct {
	*corev1alpha1.ContainerLogList
}

func (h podsLogListAdapter) asClientList() ObjectList {
	return h.ContainerLogList
}

func (h podsLogListAdapter) len() int {
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

func determineColor(podName string) (podColor, podsColor *color.Color) {
	hash := fnv.New32()
	_, _ = hash.Write([]byte(podName))
	idx := hash.Sum32() % uint32(len(colorList))

	colors := colorList[idx]
	return colors[0], colors[1]
}
