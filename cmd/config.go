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
	"io/ioutil"
	"os"
	"sync"

	configv1alpha1 "github.com/raffis/kjournal/pkg/apis/config/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	builderrest "sigs.k8s.io/apiserver-runtime/pkg/builder/rest"

	// +kubebuilder:scaffold:resource-imports

	"github.com/raffis/kjournal/pkg/storage"
	_ "github.com/raffis/kjournal/pkg/storage/elasticsearch"
)

var (
	provider storage.Provider
	once     sync.Once
	conf     configv1alpha1.APIServerConfig
)

func initConfig() (configv1alpha1.APIServerConfig, error) {
	b, err := ioutil.ReadFile("/config.yaml")
	if err != nil {
		return conf, fmt.Errorf("failed to read apiserver config: %w", err)
	}

	expand := os.ExpandEnv(string(b))

	scheme := runtime.NewScheme()
	_ = configv1alpha1.AddToScheme(scheme)
	codec := serializer.NewCodecFactory(scheme)
	decoder := codec.UniversalDeserializer()

	_, _, err = decoder.Decode([]byte(expand), nil, &conf)
	if err != nil {
		return conf, fmt.Errorf("failed to decode apiserver config: %w", err)
	}

	return conf, nil
}

func storageMapper(obj resource.Object) builderrest.ResourceHandlerProvider {
	return func(scheme *runtime.Scheme, getter generic.RESTOptionsGetter) (rest.Storage, error) {
		var err error
		once.Do(func() {
			conf, e := initConfig()
			if e != nil {
				err = e
				return
			}

			pr, e := storage.NewProvider(conf)
			provider = pr
			err = e
		})

		if err != nil {
			return nil, fmt.Errorf("%w: failed to initialize config", err)
		}

		return provider.Provide(obj, scheme, getter)
	}
}
