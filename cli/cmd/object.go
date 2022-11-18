package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type apiType struct {
	kind         string
	humanKind    string
	resource     string
	groupVersion schema.GroupVersion
	namespaced   bool
}

type listAdapter interface {
	asClientList() ObjectList
	len() int
}

type ObjectList interface {
	metav1.ListInterface
	runtime.Object
}
