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
	"time"

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
)

func init() {
	storage.Providers.Add("elasticsearch", newElasticsearchStorageProvider)
}

var esClient *elasticsearch.Client

func getESClient(backend *configv1alpha1.Backend) (*elasticsearch.Client, error) {
	if esClient != nil {
		return esClient, nil
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
		Logger: &logger{},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create elasticsearch client", err)
	}

	esClient = es
	return es, nil
}

func MakeDefaultOptions() Options {
	return Options{
		Backend: OptionsBackend{
			Index:           "*",
			RefreshRate:     time.Millisecond * 500,
			TimestampFields: []string{"@timestamp"},
			BulkSize:        500,
		},
		DefaultTimeRange: "now-24h",
	}
}

type Options struct {
	FieldMap         map[string][]string
	DropFields       []string
	Filter           map[string]string
	DefaultTimeRange string
	Backend          OptionsBackend
}

type OptionsBackend struct {
	Index           string
	RefreshRate     time.Duration
	TimestampFields []string
	BulkSize        int64
}

func MakeOptionsFromConfig(apiBinding *configv1alpha1.API) Options {
	options := MakeDefaultOptions()
	options.FieldMap = apiBinding.FieldMap
	options.DropFields = apiBinding.DropFields
	options.Filter = apiBinding.Filter

	if apiBinding.Backend.Elasticsearch.Index != "" {
		options.Backend.Index = apiBinding.Backend.Elasticsearch.Index
	}
	if apiBinding.Backend.Elasticsearch.RefreshRate.Duration != 0 {
		options.Backend.RefreshRate = apiBinding.Backend.Elasticsearch.RefreshRate.Duration
	}
	if apiBinding.Backend.Elasticsearch.TimestampFields != nil {
		options.Backend.TimestampFields = apiBinding.Backend.Elasticsearch.TimestampFields
	}
	if apiBinding.Backend.Elasticsearch.BulkSize != 0 {
		options.Backend.BulkSize = apiBinding.Backend.Elasticsearch.BulkSize
	}
	if apiBinding.DefaultTimeRange != "" {
		options.DefaultTimeRange = apiBinding.DefaultTimeRange
	}

	return options
}

func newElasticsearchStorageProvider(obj resource.Object, scheme *runtime.Scheme, getter generic.RESTOptionsGetter, backend *configv1alpha1.Backend, apiBinding *configv1alpha1.API) (rest.Storage, error) {
	opts := MakeOptionsFromConfig(apiBinding)

	gr := obj.GetGroupVersionResource().GroupResource()
	codec, _, err := srvstorage.NewStorageCodec(srvstorage.StorageCodecConfig{
		StorageMediaType:  runtime.ContentTypeJSON,
		StorageSerializer: serializer.NewCodecFactory(scheme),
		StorageVersion:    scheme.PrioritizedVersionsForGroup(obj.GetGroupVersionResource().Group)[0],
		MemoryVersion:     scheme.PrioritizedVersionsForGroup(obj.GetGroupVersionResource().Group)[0],
		Config:            storagebackend.Config{},
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
		opts,
		obj.NamespaceScoped(),
		obj.New,
		obj.NewList,
	), nil
}
