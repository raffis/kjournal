/*
Copyright 2020 The Flux authors

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

package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Most commands need one or both of the kind (e.g.,
// `"ImageRepository"`) and a human-palatable name for the kind (e.g.,
// `"image repository"`), to be interpolated into output. It's
// convenient to package these up ahead of time, then the command
// implementation can pick whichever it wants to use.
type apiType struct {
	kind, humanKind, resource string
	groupVersion              schema.GroupVersion
}

// listAdapater is the analogue to adapter, but for lists; the
// controller runtime distinguishes between methods dealing with
// objects and lists.
type listAdapter interface {
	asClientList() ObjectList
	len() int
}

type ObjectList interface {
	metav1.ListInterface
	runtime.Object
}
