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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	auditv1 "k8s.io/apiserver/pkg/apis/audit"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
)

// +genclient

// AuditEvent
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AuditEvent struct {
	// ObjectMeta is only included to fullfil metav1.Object interface,
	// it will be omitted from any json de and encoding. It is required for storage.ConvertToTable()
	metav1.ObjectMeta `json:"-"`
	metav1.TypeMeta   `json:",inline"`

	auditv1.Event `json:",inline"`
}

// AuditEventList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AuditEventList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []AuditEvent `json:"items"`
}

var _ resource.Object = &AuditEvent{}

func (in *AuditEvent) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AuditEvent) NamespaceScoped() bool {
	return false
}

func (in *AuditEvent) New() runtime.Object {
	return &AuditEvent{
		//Force set Event kind as we ditch it while fetching from the storage
		Event: auditv1.Event{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AuditEvent",
				APIVersion: "core.kjournal/v1alpha1",
			},
		},
	}
}

func (in *AuditEvent) NewList() runtime.Object {
	return &AuditEventList{}
}

func (in *AuditEvent) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "core.kjournal",
		Version:  "v1alpha1",
		Resource: "auditevents",
	}
}

func (in *AuditEvent) IsStorageVersion() bool {
	return true
}

var _ resource.ObjectList = &AuditEventList{}

func (in *AuditEventList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}
