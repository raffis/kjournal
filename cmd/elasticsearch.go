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
	"crypto/tls"
	"crypto/x509"
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
	builderrest "sigs.k8s.io/apiserver-runtime/pkg/builder/rest"

	// +kubebuilder:scaffold:resource-imports
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/raffis/kjournal/pkg/config"
	"github.com/raffis/kjournal/pkg/storage"
	"github.com/spf13/cobra"
)

type elasticsearchStorageFlags struct {
	url                     []string
	allowInsecureTLS        bool
	caCert                  string
	auditIndex              string
	auditTimestampField     string
	containerIndex          string
	containerTimestampField string
	containerNamespaceField string
	refreshRate             time.Duration
}

var (
	elasticsearchStorageArgs elasticsearchStorageFlags
	client                   *elasticsearch.Client
)

func getESClient() (*elasticsearch.Client, error) {
	if client != nil {
		return client, nil
	}

	var cert []byte
	if elasticsearchStorageArgs.caCert != "" {
		c, err := os.ReadFile(elasticsearchStorageArgs.caCert)
		if err != nil {
			return nil, err
		}

		cert = c
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	if len(cert) > 0 {
		pool.AppendCertsFromPEM(cert)
	}

	cfg := elasticsearch.Config{
		Addresses: elasticsearchStorageArgs.url,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: elasticsearchStorageArgs.allowInsecureTLS,
				ClientCAs:          pool,
			},
		},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err == nil {
		client = es
	}

	return es, err
}

func elasticsearchFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&elasticsearchStorageArgs.url, "es-url", []string{"http://localhost:9200"}, "Elasticsearch URL, you may add multiple ones comma separated")
	cmd.Flags().BoolVarP(&elasticsearchStorageArgs.allowInsecureTLS, "es-allow-insecure-tls", "", false, "Allow insecure TLS connections. Do not verify the certificate")
	cmd.Flags().StringVarP(&elasticsearchStorageArgs.caCert, "es-cacert", "", "", "Path to the CA (PEM) used to verify the server tls certificate")
	cmd.Flags().StringVar(&elasticsearchStorageArgs.auditIndex, "es-audit-index", "", "The index pattern where the kubernetes audit documents are stored. (For example: `audit-*`). You may specify multiple ones comma separated")
	cmd.Flags().StringVar(&elasticsearchStorageArgs.auditTimestampField, "es-audit-timestamp-field", "@timestamp", "The index field which is used as timestamop field for the audit documents")
	cmd.Flags().StringVar(&elasticsearchStorageArgs.containerIndex, "es-container-index", "", "The index pattern where the kubernetes container logs are stored. (For example: `logstash-*`). You may specify multiple ones comma separated")
	cmd.Flags().StringVar(&elasticsearchStorageArgs.containerTimestampField, "es-container-timestamp-field", "@timestamp", "The index field which is used as timestamop field for the audit documents")
	cmd.Flags().StringVar(&elasticsearchStorageArgs.containerNamespaceField, "es-container-namespace-field", "kubernetes.namespace_name.keyword", "The field which holds the kubernetes namespace. This field must not be indexed using any analyzers! Usually a .keyword field is wanted here")
	cmd.Flags().DurationVar(&elasticsearchStorageArgs.refreshRate, "es-refresh-rate", 500*time.Millisecond, "The refresh rate to poll from elasticsearch while checking for new documents during watch requests.")
}

func newElasticsearchLogStorageProvider(obj resource.Object) builderrest.ResourceHandlerProvider {
	return func(scheme *runtime.Scheme, getter generic.RESTOptionsGetter) (rest.Storage, error) {
		opts := storage.ElasticsearchOptions{
			Index:          elasticsearchStorageArgs.containerIndex,
			TimestampField: elasticsearchStorageArgs.containerTimestampField,
			NamespaceField: elasticsearchStorageArgs.containerNamespaceField,
			RefreshRate:    elasticsearchStorageArgs.refreshRate,
		}

		return newElasticsearchStorageProvider(obj, scheme, getter, opts)
	}
}

func newElasticsearchAuditStorageProvider(obj resource.Object) builderrest.ResourceHandlerProvider {
	return func(scheme *runtime.Scheme, getter generic.RESTOptionsGetter) (rest.Storage, error) {
		opts := storage.ElasticsearchOptions{
			Index:          elasticsearchStorageArgs.auditIndex,
			TimestampField: elasticsearchStorageArgs.auditTimestampField,
			NamespaceField: "objectRef.namespace.keyword",
			RefreshRate:    elasticsearchStorageArgs.refreshRate,
		}

		return newElasticsearchStorageProvider(obj, scheme, getter, opts)
	}
}

func newElasticsearchStorageProvider(obj resource.Object, scheme *runtime.Scheme, getter generic.RESTOptionsGetter, bucket config.Bucket) (rest.Storage, error) {
	gr := obj.GetGroupVersionResource().GroupResource()
	codec, _, err := srvstorage.NewStorageCodec(srvstorage.StorageCodecConfig{
		StorageMediaType:  runtime.ContentTypeJSON,
		StorageSerializer: serializer.NewCodecFactory(scheme),
		StorageVersion:    scheme.PrioritizedVersionsForGroup(obj.GetGroupVersionResource().Group)[0],
		MemoryVersion:     scheme.PrioritizedVersionsForGroup(obj.GetGroupVersionResource().Group)[0],
		Config:            storagebackend.Config{}, // useless fields..
	})

	if err != nil {
		return nil, err
	}

	client, err := getESClient()
	if err != nil {
		return nil, err
	}

	return storage.NewElasticsearchREST(
		gr,
		codec,
		client,
		bucket,
		obj.NamespaceScoped(),
		obj.New,
		obj.NewList,
	), nil
}
