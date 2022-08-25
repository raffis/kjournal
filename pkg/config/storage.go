package storage

import (
	"context"
	"fmt"
	"reflect"

	config "github.com/elastic/go-config/v8"
	"k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/klog/v2"
)

var _ rest.Scoper = &configREST{}
var _ rest.Storage = &configREST{}

// NewconfigREST instantiates a new REST storage.
func NewConfigREST(
	groupResource schema.GroupResource,
	codec runtime.Codec,
	es *config.Client,
	isNamespaced bool,
	newFunc func() runtime.Object,
	newListFunc func() runtime.Object,
) rest.Storage {
	return &configREST{
		groupResource: groupResource,
		codec:         codec,
		es:            es,
		metaAccessor:  meta.NewAccessor(),
		isNamespaced:  isNamespaced,
		newFunc:       newFunc,
		newListFunc:   newListFunc,
	}
}

type configREST struct {
	rest.TableConvertor
	groupResource schema.GroupResource
	codec         runtime.Codec
	es            *config.Client
	isNamespaced  bool
	metaAccessor  meta.MetadataAccessor
	newFunc       func() runtime.Object
	newListFunc   func() runtime.Object
}

func (f *configREST) New() runtime.Object {
	return f.newFunc()
}

func (f *configREST) NewList() runtime.Object {
	return f.newListFunc()
}

func (f *configREST) NamespaceScoped() bool {
	return f.isNamespaced
}

func (f *configREST) Get(
	ctx context.Context,
	name string,
	options *metav1.GetOptions,
) (runtime.Object, error) {
	return nil, nil
}

type tableConvertor interface {
	ConvertToTable(ctx context.Context, tableOptions runtime.Object) (*metav1.Table, error)
}

// ConvertToTable implements the TableConvertor interface for REST.
func (f *configREST) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	if convert, ok := obj.(tableConvertor); ok {
		token, err := f.metaAccessor.Continue(obj)
		if err != nil {
			return nil, err
		}
		tbl, err := convert.ConvertToTable(ctx, tableOptions)
		tbl.ListMeta.Continue = token
		return tbl, err
	}

	return nil, nil
}

func (f *configREST) List(
	ctx context.Context,
	options *metainternalversion.ListOptions,
) (runtime.Object, error) {
	klog.InfoS("List request", "options", options)

	newListObj := f.NewList()
	v, err := getListPrt(newListObj)
	if err != nil {
		return nil, err
	}

	//ns, _ := request.NamespaceFrom(ctx)
	for _, hit = range esResults.Hits.Hits {
		newObj := f.newFunc()
		decodedObj, _, err := f.codec.Decode(hit.Source, nil, newObj)
		if err != nil {
			return nil, err
		}

		appendItem(v, decodedObj)
	}

	return newListObj, nil
}

func appendItem(v reflect.Value, obj runtime.Object) {
	v.Set(reflect.Append(v, reflect.ValueOf(obj).Elem()))
}

func getListPrt(listObj runtime.Object) (reflect.Value, error) {
	listPtr, err := meta.GetItemsPtr(listObj)
	if err != nil {
		return reflect.Value{}, err
	}

	v, err := conversion.EnforcePtr(listPtr)
	if err != nil || v.Kind() != reflect.Slice {
		return reflect.Value{}, fmt.Errorf("need ptr to slice: %v", err)
	}

	return v, nil
}
