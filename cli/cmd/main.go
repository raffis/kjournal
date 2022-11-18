package main

import (
	"flag"
	"log"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"
)

var (
	version = "0.0.0-dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:           "kjournal",
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
	Short:         "Command line utility for accessing long-term (kubernetes) logs",
	Long: `
Command line utility for accessing long-term kubernetes logs.`,
}

type rootFlags struct {
	timeout      time.Duration
	pollInterval time.Duration
}

// RequestError is a custom error type that wraps an error returned by the flux api.
type RequestError struct {
	StatusCode int
	Err        error
}

type Options struct {
	// QPS indicates the maximum queries-per-second of requests sent to the Kubernetes API, defaults to 50.
	QPS float32

	// Burst indicates the maximum burst queries-per-second of requests sent to the Kubernetes API, defaults to 100.
	Burst int
}

func (r *RequestError) Error() string {
	return r.Err.Error()
}

var logger = stderrLogger{stderr: os.Stderr}
var rootArgs = NewRootFlags()
var kubeconfigArgs = genericclioptions.NewConfigFlags(false)
var kubeclientOptions = new(Options)

func init() {
	set := &flag.FlagSet{}
	klog.InitFlags(set)
	rootCmd.PersistentFlags().AddGoFlagSet(set)

	rootCmd.PersistentFlags().DurationVar(&rootArgs.timeout, "timeout", 5*time.Minute, "timeout for this operation")

	kubeconfigArgs.APIServer = nil // prevent AddFlags from configuring --server flag
	kubeconfigArgs.Timeout = nil   // prevent AddFlags from configuring --request-timeout flag, we have --timeout instead
	kubeconfigArgs.AddFlags(rootCmd.PersistentFlags())

	// Since some subcommands use the `-s` flag as a short version for `--silent`, we manually configure the server flag
	// without the `-s` short version. While we're no longer on par with kubectl's flags, we maintain backwards compatibility
	// on the CLI interface.
	apiServer := ""
	kubeconfigArgs.APIServer = &apiServer
	rootCmd.PersistentFlags().StringVar(kubeconfigArgs.APIServer, "server", *kubeconfigArgs.APIServer, "The address and port of the Kubernetes API server")

	//kubeclientOptions.BindFlags(rootCmd.PersistentFlags())

	_ = rootCmd.RegisterFlagCompletionFunc("context", contextsCompletionFunc)
	_ = rootCmd.RegisterFlagCompletionFunc("namespace", resourceNamesCompletionFunc(corev1.SchemeGroupVersion.WithKind("Namespace")))

	rootCmd.DisableAutoGenTag = true
	rootCmd.SetOut(os.Stdout)
}

func NewRootFlags() rootFlags {
	rf := rootFlags{
		pollInterval: 2 * time.Second,
	}

	return rf
}

func main() {
	log.SetFlags(0)

	name := path.Base(os.Args[0])
	if len(name) > 8 && name[0:8] == "kubectl-" {
		args := append([]string{name[8:]}, os.Args[1:]...)
		rootCmd.SetArgs(args)
	}

	if err := rootCmd.Execute(); err != nil {
		if err, ok := err.(*RequestError); ok {
			if err.StatusCode == 1 {
				klog.V(1).ErrorS(err, "request failed")
			} else {
				klog.ErrorS(err, "execution failed")
			}

			os.Exit(err.StatusCode)
		}

		klog.ErrorS(err, "execution failed")
		os.Exit(1)
	}
}
