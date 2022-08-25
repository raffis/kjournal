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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource/resourcestrategy"
)

// +genclient

// Bucket
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Bucket struct {
	Namespace string `json:"namespace"`
}

// BucketList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Bucket `json:"items"`
}

var _ resource.Object = &Bucket{}
var _ resourcestrategy.Validater = &Bucket{}

func (in *Bucket) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *Bucket) NamespaceScoped() bool {
	return false
}

func (in *Bucket) New() runtime.Object {
	return &Bucket{}
}

func (in *Bucket) NewList() runtime.Object {
	return &BucketList{}
}

func (in *Bucket) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "core.kjournal",
		Version:  "v1alpha1",
		Resource: "buckets",
	}
}

func (in *Bucket) IsStorageVersion() bool {
	return true
}

func (in *Bucket) Validate(ctx context.Context) field.ErrorList {
	return nil
}

var _ resource.ObjectList = &BucketList{}

func (in *BucketList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}
