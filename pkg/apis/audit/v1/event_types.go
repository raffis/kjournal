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

var eventTableColums = []metav1.TableColumnDefinition{
	{Name: "Received", Type: "", Format: "name", Description: "s"},
	{Name: "Verb", Type: "string", Format: "name", Description: "s"},
	{Name: "Status", Type: "string", Description: "The reference to the service that hosts this API endpoint."},
	{Name: "Username", Type: "sstringtring", Description: "Whether this service is available."},
}

// Event
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Event struct {
	// ObjectMeta is only included to fullfil metav1.Object interface,
	// it will be omitted from any json de and encoding. It is required for storage.ConvertToTable()
	metav1.ObjectMeta `json:"-"`

	auditv1.Event `json:",inline"`
}

// EventList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EventList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Event `json:"items"`
}

var _ resource.Object = &Event{}
var _ resourcestrategy.Validater = &Event{}

func (in *Event) GetObjectMeta() *metav1.ObjectMeta {
	return nil
	//return &in.ObjectMeta
}

func (in *Event) GetNamespace() string {
	return in.ObjectRef.Namespace
}

func (in *Event) NamespaceScoped() bool {
	return true
}

func (in *Event) New() runtime.Object {
	return &Event{
		//Force set Event kind as we ditch it while fetching from the storage
		Event: auditv1.Event{
			TypeMeta: metav1.TypeMeta{
				Kind: "Event",
			},
		},
	}
}

func (in *Event) NewList() runtime.Object {
	return &EventList{}
}

func (in *Event) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "audit.kjournal",
		Version:  "v1",
		Resource: "events",
	}
}

func (in *Event) IsStorageVersion() bool {
	return true
}

func (in *Event) Validate(ctx context.Context) field.ErrorList {
	// TODO(user): Modify it, adding your API validation here.
	return nil
}

var _ resource.ObjectList = &EventList{}

func (in *EventList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}

// ConvertToTable implements the TableConvertor interface for REST.
func (in *Event) ConvertToTable(ctx context.Context, tableOptions runtime.Object) (*metav1.Table, error) {
	table := &metav1.Table{
		ColumnDefinitions: eventTableColums,
		TypeMeta:          in.TypeMeta,
	}

	rows := make([]metav1.TableRow, 0, 1)
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: in},
		Cells:  []interface{}{in.RequestReceivedTimestamp, in.Verb, in.ResponseStatus.Code, in.User.Username},
	}

	rows = append(rows, row)
	table.Rows = rows
	return table, nil
}

// ConvertToTable implements the TableConvertor interface for REST.
func (in *EventList) ConvertToTable(ctx context.Context, tableOptions runtime.Object) (*metav1.Table, error) {
	table := &metav1.Table{
		ColumnDefinitions: eventTableColums,
		TypeMeta:          in.TypeMeta,
	}

	rows := make([]metav1.TableRow, 0, 1)

	for _, v := range in.Items {
		row := metav1.TableRow{
			Object: runtime.RawExtension{Object: &v},
			Cells:  []interface{}{v.RequestReceivedTimestamp, v.Verb, v.ResponseStatus.Code, v.User.Username},
		}
		rows = append(rows, row)
	}

	table.Rows = rows
	return table, nil
}
