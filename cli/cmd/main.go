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
	"log"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
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

var logger = stderrLogger{stderr: os.Stderr}

type rootFlags struct {
	timeout      time.Duration
	verbose      bool
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

var rootArgs = NewRootFlags()
var kubeconfigArgs = genericclioptions.NewConfigFlags(false)
var kubeclientOptions = new(Options)

func init() {
	rootCmd.PersistentFlags().DurationVar(&rootArgs.timeout, "timeout", 5*time.Minute, "timeout for this operation")
	rootCmd.PersistentFlags().BoolVar(&rootArgs.verbose, "verbose", false, "print generated objects")

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

	rootCmd.RegisterFlagCompletionFunc("context", contextsCompletionFunc)
	rootCmd.RegisterFlagCompletionFunc("namespace", resourceNamesCompletionFunc(corev1.SchemeGroupVersion.WithKind("Namespace")))

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
				logger.Warningf("%v", err)
			} else {
				logger.Failuref("%v", err)
			}

			os.Exit(err.StatusCode)
		}

		logger.Failuref("%v", err)
		os.Exit(1)
	}
}
