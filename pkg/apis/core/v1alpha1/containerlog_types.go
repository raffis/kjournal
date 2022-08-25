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

package v1alpha1

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

// ContainerLog
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ContainerLog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Container         string          `json:"container"`
	Pod               string          `json:"pod"`
	Payload           json.RawMessage `json:"payload"`
}

// ContainerLogList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ContainerLogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ContainerLog `json:"items"`
}

var _ resource.Object = &ContainerLog{}
var _ resourcestrategy.Validater = &ContainerLog{}

func (in *ContainerLog) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *ContainerLog) NamespaceScoped() bool {
	return true
}

func (in *ContainerLog) New() runtime.Object {
	return &ContainerLog{}
}

func (in *ContainerLog) NewList() runtime.Object {
	return &ContainerLogList{}
}

func (in *ContainerLog) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "core.kjournal",
		Version:  "v1alpha1",
		Resource: "containerlogs",
	}
}

func (in *ContainerLog) IsStorageVersion() bool {
	return true
}

func (in *ContainerLog) Validate(ctx context.Context) field.ErrorList {
	return nil
}

var _ resource.ObjectList = &ContainerLogList{}

func (in *ContainerLogList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}
