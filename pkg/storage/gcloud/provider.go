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

package gcloud

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/option"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	srvstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"

	// +kubebuilder:scaffold:resource-imports
	"cloud.google.com/go/logging/logadmin"
	configv1alpha1 "github.com/raffis/kjournal/pkg/apis/config/v1alpha1"
	"github.com/raffis/kjournal/pkg/storage"
)

var gcloudClient *logadmin.Client

func init() {
	storage.Providers.Add("gcloud", newGCloudStorageProvider)
}

func getGCCloudClient(ctx context.Context, backend *configv1alpha1.Backend) (*logadmin.Client, error) {
	if gcloudClient != nil {
		return gcloudClient, nil
	}

	j := `{
	"type": "service_account",
	"project_id": "kjournal",
	"private_key_id": "0a9a0540ea48c2455c46387622f3df904d3e50fd",
	"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDXQk1Yr2w4sAi8\n3HOYhVNcBIrc4kfwrLi5ysbcPAUStqo3mXjRgcOVm5xiuvJgkiGeoOjO6rYX6hqM\njqdPBxllvT8hfjwPuVjIS0RPQ/F+r5VzYrWTXUFsFVYn2gJkQQAoavFTd9gOhc4H\nbCIptfUqOSZ9+o5gABceXspYLdqBFcTg4iaDKXJHZDBHDPWwc1KnQZPhpXAeb6nv\naSaHRZHaRGjJTZPds47rX9Sk2oX2NeYI/IlNAqfSUFdk9eS/vUUp+4k80VA41NQX\n06XAjXcuygVY7Elc5HT2yFX7G80H7GSMOXri3ItRnUSlbij8Nv70S3wFeLy57E1+\nLd99e6JjAgMBAAECggEATKXnNLUCLA1Cjz1QS/btf85+Q7ivNRvLixyRQsp8Y/V3\nFuUnCDLUmekW/nDi7VAbeIiDXWpl/I33diU1ngZBHOEOIbb5W//7hRaH9FGVJC8R\nYEy9qwOB0CKo0vfl8hzTGZE67SW3YTRz8GCoqYGJEsfW1PTqzXQ6xy2pj0yEiYG3\nE0aBdZSvhPgPLZqIOG7p+IroZ6m0eN/Cm9ErlDj3TjO3DewF9/2/xk/D86ZbGONc\nnJNnMqGsPaAmaxj7mV7v1D0JxqiZzCFj7x9j1mAjlQ0NjbfLX/onoNIIDbhDGAYy\n7all7qicsY0jtHP46VlWYvezBdRG446V2n9gzdVewQKBgQD/91CZq4gWP8f9NiVZ\nmt3hDElWXEqz3g/gWWjEBcYraoONVa4MLVEHj8p8xn7eoWVXvfQYgYbZFD3I3T7g\nugt4ewBYXAgd6BZKE9mVWRNIEHE0qKcZL4yzSa6xNhOvvEPhRVOisgwa4tbw8ZJA\naXnW7wGGoSxfra1Zwcu0Bm71NwKBgQDXSZsm6A/iBfKC3eWG/TISTWxAICR0FVMc\nkajkr0JeBUQAu3dK6gDaVeJI/9uJJlFqwBtEla0kUUm23WBo1ZLOfd6DxaQ+dlvl\nBLruYJaY5t/wbYNPr5uZtkwCU3cJMe4TQESs1f9EWk5jIZB2bM0hvgwyU0e/hJgw\noWIiMWMSNQKBgC/TML8Vmp61mhNIi5/7XJuQ5R76rYZ/5i1/5yBBB+7SvvOoX5Ws\n3efwyN+ZYtkMBNhpCHOPt/dVXdnq5LWubTg8myrnPyj/VTLQFKZf90dOsygontgI\n11wkVzyLIxCBt5kej+rlI3fejFSGflIEoxwymfFiqdzSoYIUwI/JZ+/vAoGBAMnL\nzsibQTgFhxmv0NPFRUfulodNGZ5N1seyqPMibD0hBmsBTYJE8WO2mRL/8NIPvsUn\nKOgSvGaMY2IrA5GAj8lKJmaxvZBm9SAoXOfQVZkg38vHewwYeOuN+pU7kxplWNlm\npnizZkC1vUAiV/0JYwY708bgVSJpsRX0T73pOQn5AoGBAPFB1XThwgimuKVCsUYL\n/lJXtsAj5gHJ/haUO7BC3NOvQZDrkb7TN7HND/+SE323NnKxyU6Lb8dt4Lb8549V\nEDLaCeKHszBeX7TrlLz364cICvgU/sSRw/gX4WLtjWCSBQYYrfmOqQVaAI76fl6S\nM+es7uF3mrE253OlUswkgkOb\n-----END PRIVATE KEY-----\n",
	"client_email": "kjournal-logs@kjournal.iam.gserviceaccount.com",
	"client_id": "110641057771820055680",
	"auth_uri": "https://accounts.google.com/o/oauth2/auth",
	"token_uri": "https://oauth2.googleapis.com/token",
	"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/kjournal-logs%40kjournal.iam.gserviceaccount.com"
  }  
`
	client, err := logadmin.NewClient(ctx, "kjournal" /*option.WithAPIKey(backend.GCloud.APIKey)*/, option.WithCredentialsJSON([]byte(j)))
	fmt.Printf("\nuse token %s\n\n", backend.GCloud.APIKey)

	gcloudClient = client
	return gcloudClient, err
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

func newGCloudStorageProvider(obj resource.Object, scheme *runtime.Scheme, getter generic.RESTOptionsGetter, backend *configv1alpha1.Backend, apiBinding *configv1alpha1.API) (rest.Storage, error) {
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

	client, err := getGCCloudClient(context.TODO(), backend)
	if err != nil {
		return nil, err
	}

	return NewGCloudREST(
		gr,
		codec,
		client,
		opts,
		obj.NamespaceScoped(),
		obj.New,
		obj.NewList,
	), nil
}
