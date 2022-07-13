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
	"net/http"

	"k8s.io/klog/v2"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"

	// +kubebuilder:scaffold:resource-imports
	logv1beta1adapter "github.com/raffis/kjournal/internal/container/v1beta1"
	auditv1 "github.com/raffis/kjournal/pkg/apis/audit/v1"
)

type apiServerFlags struct {
	storageBackend string
}

type httpWrap struct {
	w http.Handler
}

var (
	apiServerArgs apiServerFlags
)

func (m *httpWrap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fieldSelector := q.Get("fieldSelector")
	q.Set("labelSelector", fieldSelector)
	q.Del("fieldSelector")
	r.URL.RawQuery = q.Encode()
	m.w.ServeHTTP(w, r)
}

func main() {
	cmd, err := builder.APIServer.
		// +kubebuilder:scaffold:resource-register
		WithResourceAndHandler(&logv1beta1adapter.Log{}, newElasticsearchLogStorageProvider(&logv1beta1adapter.Log{})).
		WithResourceAndHandler(&auditv1.Event{}, newElasticsearchAuditStorageProvider(&auditv1.Event{})).
		WithResourceAndHandler(&auditv1.ClusterEvent{}, newElasticsearchAuditStorageProvider(&auditv1.ClusterEvent{})).
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

	cmd.Flags().StringVar(&apiServerArgs.storageBackend, "log-storage-backend", "elasticsearch", "Storage backend, currently only elasticsearch is supported")
	elasticsearchFlags(cmd)

	err = cmd.Execute()
	if err != nil {
		klog.Fatal(err)
	}
}
