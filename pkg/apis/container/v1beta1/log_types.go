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

package v1beta1

import (
	"context"
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource/resourcestrategy"
)

// +genclient

// Log
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Log struct {
	metav1.TypeMeta `json:",inline"`
	// ObjectMeta is only included to fullfil metav1.Object interface,
	// it will be omitted from any json de and encoding. It is required for storage.ConvertToTable()
	metav1.ObjectMeta `json:"-"`
	Metadata          LogMetadata     `json:"metadata"`
	Container         string          `json:"container"`
	Pod               string          `json:"pod"`
	Unstructured      json.RawMessage `json:"unstructured"`
	Env               json.RawMessage `json:"env"`
}

type LogMetadata struct {
	Namespace         string      `json:"namespace"`
	CreationTimestamp metav1.Time `json:"creationTimestamp"`
}

// LogList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type LogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Log `json:"items"`
}

var _ resource.Object = &Log{}
var _ resourcestrategy.Validater = &Log{}

func (in *Log) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *Log) NamespaceScoped() bool {
	return true
}

func (in *Log) New() runtime.Object {
	return &Log{}
}

func (in *Log) NewList() runtime.Object {
	return &LogList{}
}

func (in *Log) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "container.kjournal",
		Version:  "v1beta1",
		Resource: "logs",
	}
}

func (in *Log) IsStorageVersion() bool {
	return true
}

func (in *Log) Validate(ctx context.Context) field.ErrorList {
	return nil
}

var _ resource.ObjectList = &LogList{}

func (in *LogList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}
