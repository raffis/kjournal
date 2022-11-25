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
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"

	"github.com/Masterminds/semver"
	"github.com/pyroscope-io/client/pyroscope"
	adapterv1alpha1 "github.com/raffis/kjournal/internal/apis/core/v1alpha1"
	"github.com/raffis/kjournal/pkg/apis/core/v1alpha1"
	"github.com/raffis/kjournal/pkg/apiserver"
	"github.com/spf13/cobra"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sversion "k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/klog/v2"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
)

const (
	version = "0.0.0-dev"
	commit  = "none"
	date    = "unknown"
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
	q := r.URL.Query()
	fieldSelector := q.Get("fieldSelector")
	q.Set("labelSelector", fieldSelector)
	q.Del("fieldSelector")
	r.URL.RawQuery = q.Encode()
	m.w.ServeHTTP(w, r)
}

func storageMapper(obj resource.Object) apiserver.StorageProvider {
	return func(scheme *k8sruntime.Scheme, getter generic.RESTOptionsGetter) (rest.Storage, error) {
		return provider.Provide(obj, scheme, getter)
	}
}

func main() {
	if profilingEnabled, ok := os.LookupEnv("PYROSCOPE_START_PROFILING"); ok && profilingEnabled == "true" {
		runtime.SetMutexProfileFraction(5)
		runtime.SetBlockProfileRate(5)
		_, err := pyroscope.Start(getProfilerConfig())
		if err != nil {
			panic(err)
		}
	}

	withResourceAndHandler(&v1alpha1.Log{}, storageMapper(&v1alpha1.Log{}))
	withResourceAndHandler(&v1alpha1.ContainerLog{}, storageMapper(&v1alpha1.ContainerLog{}))
	withResourceAndHandler(&adapterv1alpha1.AuditEvent{}, storageMapper(&adapterv1alpha1.AuditEvent{}))
	withResourceAndHandler(&adapterv1alpha1.Event{}, storageMapper(&adapterv1alpha1.Event{}))

	/*
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
			WithServerFns(func(server *builder.GenericAPIServer) *builder.GenericAPIServer {
				server.Version = getVersion()
				return server
			}).
			Build()*/

	o := NewServerOptions(os.Stdout, os.Stderr) //, a.orderedGroupVersions...)
	cmd := NewCommandStartServer(o, genericapiserver.SetupSignalHandler())
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	/*	cmd.Flags().StringVar(&apiServerArgs.configFile, "config", "", "Path to kjournal config")


		cmd.AddCommand(cmdMan)
		cmd.AddCommand(cmdRef)
	*/
	err := cmd.Execute()
	if err != nil {
		klog.Fatal(err)
	}
}

func getVersion() *k8sversion.Info {
	v := &k8sversion.Info{
		GitVersion:   version,
		GitCommit:    commit,
		GitTreeState: "clean",
		BuildDate:    date,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	srvSemantic, err := semver.NewVersion(version)
	if err != nil {
		return v
	}

	v.Major = strconv.Itoa(int(srvSemantic.Major()))
	v.Minor = strconv.Itoa(int(srvSemantic.Minor()))
	return v
}

func getProfilerConfig() pyroscope.Config {
	cfg := pyroscope.Config{
		ApplicationName: "kjournal",
		ServerAddress:   "http://pyroscope:4040",
		Logger:          pyroscope.StandardLogger,
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	}

	if appName, ok := os.LookupEnv("PYROSCOPE_APPLICATION_NAME"); ok {
		cfg.ApplicationName = appName
	}

	if address, ok := os.LookupEnv("PYROSCOPE_SERVER_ADDRESS"); ok {
		cfg.ServerAddress = address
	}

	if token, ok := os.LookupEnv("PYROSCOPE_AUTH_TOKEN"); ok {
		cfg.AuthToken = token
	}

	return cfg
}
