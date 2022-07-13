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

package v1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource/resourcestrategy"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterEvent
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterEvent struct {
	// ObjectMeta is only included to fullfil metav1.Object interface,
	// it will be omitted from any json de and encoding. It is required for storage.ConvertToTable()
	metav1.ObjectMeta `json:"-"`

	auditv1.Event `json:",inline"`
}

// ClusterEventList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterEventList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ClusterEvent `json:"items"`
}

var _ resource.Object = &ClusterEvent{}
var _ resourcestrategy.Validater = &ClusterEvent{}

func (in *ClusterEvent) GetObjectMeta() *metav1.ObjectMeta {
	return nil
	//return &in.ObjectMeta
}

func (in *ClusterEvent) NamespaceScoped() bool {
	return false
}

func (in *ClusterEvent) New() runtime.Object {
	return &ClusterEvent{
		//Force set Event kind as we ditch it while fetching from the storage
		Event: auditv1.Event{
			TypeMeta: metav1.TypeMeta{
				Kind: "Event",
			},
		},
	}
}

func (in *ClusterEvent) NewList() runtime.Object {
	return &ClusterEventList{}
}

func (in *ClusterEvent) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "audit.kjournal",
		Version:  "v1",
		Resource: "clusterevents",
	}
}

func (in *ClusterEvent) IsStorageVersion() bool {
	return true
}

func (in *ClusterEvent) Validate(ctx context.Context) field.ErrorList {
	// TODO(user): Modify it, adding your API validation here.
	return nil
}

var _ resource.ObjectList = &ClusterEventList{}

func (in *ClusterEventList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}
