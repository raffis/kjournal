package elasticsearch

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +genclient

// Dummy
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Dummy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Payload           json.RawMessage `json:"payload"`
}

// DummyList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DummyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Dummy `json:"items"`
}

var _ resource.Object = &Dummy{}

func (in *Dummy) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *Dummy) NamespaceScoped() bool {
	return true
}

func (in *Dummy) New() runtime.Object {
	return &Dummy{}
}

func (in *Dummy) NewList() runtime.Object {
	return &DummyList{}
}

func (in *Dummy) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "testing",
		Version:  "v1alpha1",
		Resource: "Dummy",
	}
}

func (in *Dummy) IsStorageVersion() bool {
	return true
}

var _ resource.ObjectList = &DummyList{}

func (in *DummyList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}
