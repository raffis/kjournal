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

package elasticsearch

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	srvstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"

	// +kubebuilder:scaffold:resource-imports
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	configv1alpha1 "github.com/raffis/kjournal/pkg/apis/config/v1alpha1"
	"github.com/raffis/kjournal/pkg/storage"
	"github.com/raffis/kjournal/pkg/utils"
)

var esClientRegistry utils.Registry[*elasticsearch.Client]

func init() {
	esClientRegistry = utils.NewRegistry[*elasticsearch.Client]()
	storage.Providers.Add("elasticsearch", newElasticsearchStorageProvider)
}

func getESClient(backend *configv1alpha1.Backend) (*elasticsearch.Client, error) {
	if val, err := esClientRegistry.Get(backend.Name); err == nil {
		return val, nil
	}

	var cert []byte
	if backend.Elasticsearch.CACert != "" {
		c, err := os.ReadFile(backend.Elasticsearch.CACert)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to load cacert", err)
		}

		cert = c
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("%w: failed create cert pool", err)
	}

	if len(cert) > 0 {
		pool.AppendCertsFromPEM(cert)
	}

	cfg := elasticsearch.Config{
		Addresses: backend.Elasticsearch.URL,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: backend.Elasticsearch.AllowInsecureTLS,
				ClientCAs:          pool,
			},
		},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create elasticsearch client", err)
	}

	err = esClientRegistry.Add(backend.Name, es)
	return es, err
}

func newElasticsearchStorageProvider(obj resource.Object, scheme *runtime.Scheme, getter generic.RESTOptionsGetter, backend *configv1alpha1.Backend, apiBinding *configv1alpha1.API) (rest.Storage, error) {
	gr := obj.GetGroupVersionResource().GroupResource()
	codec, _, err := srvstorage.NewStorageCodec(srvstorage.StorageCodecConfig{
		StorageMediaType:  runtime.ContentTypeJSON,
		StorageSerializer: serializer.NewCodecFactory(scheme),
		StorageVersion:    scheme.PrioritizedVersionsForGroup(obj.GetGroupVersionResource().Group)[0],
		MemoryVersion:     scheme.PrioritizedVersionsForGroup(obj.GetGroupVersionResource().Group)[0],
		Config:            storagebackend.Config{}, // useless fields..
	})

	if err != nil {
		return nil, fmt.Errorf("%w: failed to create storage codec", err)
	}

	client, err := getESClient(backend)
	if err != nil {
		return nil, err
	}

	return NewElasticsearchREST(
		gr,
		codec,
		client,
		apiBinding,
		obj.NamespaceScoped(),
		obj.New,
		obj.NewList,
	), nil
}
