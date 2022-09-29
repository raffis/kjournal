/*
Copyright 2022.

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
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"k8s.io/klog/v2"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"

	// +kubebuilder:scaffold:resource-imports

	"github.com/pyroscope-io/client/pyroscope"
	adapterv1alpha1 "github.com/raffis/kjournal/internal/apis/core/v1alpha1"
	"github.com/raffis/kjournal/pkg/apis/core/v1alpha1"
	"github.com/spf13/cobra"
)

type apiServerFlags struct {
	configFile string
}

type httpWrap struct {
	w http.Handler
}

var (
	apiServerArgs apiServerFlags
)

var (
	rootCmd *cobra.Command
)

func (m *httpWrap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("req %#v\n", r.RequestURI)
	q := r.URL.Query()
	fieldSelector := q.Get("fieldSelector")
	q.Set("labelSelector", fieldSelector)
	q.Del("fieldSelector")
	r.URL.RawQuery = q.Encode()
	m.w.ServeHTTP(w, r)
}

func main() {
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	pyroscope.Start(pyroscope.Config{
		ApplicationName: "simple.golang.app",

		// replace this with the address of pyroscope server
		ServerAddress: "http://pyroscope:4040",

		// you can disable logging by setting this to nil
		Logger: pyroscope.StandardLogger,

		// optionally, if authentication is enabled, specify the API key:
		// AuthToken: os.Getenv("PYROSCOPE_AUTH_TOKEN"),

		ProfileTypes: []pyroscope.ProfileType{
			// these profile types are enabled by default:
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,

			// these profile types are optional:
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})

	cmd, err := builder.APIServer.
		// +kubebuilder:scaffold:resource-register
		WithResourceAndHandler(&v1alpha1.Log{}, storageMapper(&v1alpha1.Log{})).
		WithResourceAndHandler(&v1alpha1.ContainerLog{}, storageMapper(&v1alpha1.ContainerLog{})).
		WithResourceAndHandler(&adapterv1alpha1.AuditEvent{}, storageMapper(&adapterv1alpha1.AuditEvent{})).
		WithResourceAndHandler(&adapterv1alpha1.Event{}, storageMapper(&adapterv1alpha1.Event{})).
		WithLocalDebugExtension().
		WithoutEtcd().
		WithServerFns(func(server *builder.GenericAPIServer) *builder.GenericAPIServer {
			wrap := server.Handler.FullHandlerChain

			server.Handler.FullHandlerChain = &httpWrap{
				w: wrap,
			}

			return server
		}).
		Build()
	if err != nil {
		klog.Fatal(err)
	}

	cmd.Flags().StringVar(&apiServerArgs.configFile, "config", "", "Path to kjournal config")

	rootCmd = cmd
	rootCmd.Use = "kjournal-apiserver"
	rootCmd.Short = "Launches the kjournal kubernetes apiserver"

	// TODO: workaorund for removing etcd related flags. Apparantly WithoutEtcd() does not work.
	rootCmd.Flags().MarkHidden("etcd-cafile")
	rootCmd.Flags().MarkHidden("etcd-certfile")
	rootCmd.Flags().MarkHidden("etcd-compaction-interval")
	rootCmd.Flags().MarkHidden("etcd-count-metric-poll-period")
	rootCmd.Flags().MarkHidden("etcd-db-metric-poll-interval")
	rootCmd.Flags().MarkHidden("etcd-healthcheck-timeout")
	rootCmd.Flags().MarkHidden("etcd-keyfile")
	rootCmd.Flags().MarkHidden("etcd-prefix")
	rootCmd.Flags().MarkHidden("etcd-servers")
	rootCmd.Flags().MarkHidden("etcd-servers-overrides")

	cmd.AddCommand(cmdMan)
	cmd.AddCommand(cmdRef)

	err = cmd.Execute()
	if err != nil {
		klog.Fatal(err)
	}
}
