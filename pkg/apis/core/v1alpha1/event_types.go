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
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
)

// +genclient

// Event
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Event struct {
	eventsv1.Event `json:",inline"`
}

// EventList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EventList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Event `json:"items"`
}

var _ resource.Object = &Event{}

func (in *Event) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *Event) NamespaceScoped() bool {
	return true
}

func (in *Event) New() runtime.Object {
	return &Event{
		Event: eventsv1.Event{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Event",
				APIVersion: "core.kjournal/v1alpha1",
			},
		},
	}
}

func (in *Event) NewList() runtime.Object {
	return &EventList{}
}

func (in *Event) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "core.kjournal",
		Version:  "v1alpha1",
		Resource: "events",
	}
}

func (in *Event) IsStorageVersion() bool {
	return true
}

var _ resource.ObjectList = &EventList{}

func (in *EventList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}
