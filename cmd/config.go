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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	srvstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	builderrest "sigs.k8s.io/apiserver-runtime/pkg/builder/rest"

	// +kubebuilder:scaffold:resource-imports

	"github.com/raffis/kjournal/pkg/config"
	"github.com/raffis/kjournal/pkg/storage"
)

var cfg config.Config
var buckets config.BucketRegistry

func newBucketConfigStorageProvider(obj resource.Object) builderrest.ResourceHandlerProvider {
	return func(scheme *runtime.Scheme, getter generic.RESTOptionsGetter) (rest.Storage, error) {
		cfg = config.Config{
			Buckets: []config.Bucket{{
				Type: "container",
				Name: "container",
				Backend: config.BucketBackend{
					Elasticsearch: config.BucketBackendElasticsearch{
						Index: "logstash-*",
					},
				},
			}},
		}

		buckets = config.NewBucketRegistry()
		err := buckets.AddBucket(cfg.Buckets[0])
		if err != nil {
			return nil, err
		}

		return BucketStorage(obj, scheme, getter, opts)
	}
}

func newStorageProvider(obj resource.Object, name string) builderrest.ResourceHandlerProvider {
	return func(scheme *runtime.Scheme, getter generic.RESTOptionsGetter) (rest.Storage, error) {
		switch(obj.(type)) {
			case 
		}

	}
}


WithResourceAndHandler(&corev1alpha1.Bucket{}, newBucketConfigStorageProvider(&corev1alpha1.Bucket{})).
//WithResourceAndHandler(&corev1alpha1.Log{}, newStorageProvider(&corev1alpha1.Log{})).
WithResourceAndHandler(&corev1alpha1.ContainerLog{}, newStorageProvider(&logv1beta1.ContainerLog{})).
WithResourceAndHandler(&corev1alpha1.AuditEvent{}, newStorageProvider(&logv1beta1.AuditEvent{})).

