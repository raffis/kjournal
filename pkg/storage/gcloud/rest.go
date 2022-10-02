package gcloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/logadmin"
	"github.com/Jeffail/gabs"

	"k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/klog/v2"
)

var _ rest.Scoper = &gcloudREST{}
var _ rest.Storage = &gcloudREST{}
var _ rest.TableConvertor = &gcloudREST{}

// NewgcloudREST instantiates a new REST storage.
func NewGCloudREST(
	groupResource schema.GroupResource,
	codec runtime.Codec,
	client *logadmin.Client,
	opts Options,
	isNamespaced bool,
	newFunc func() runtime.Object,
	newListFunc func() runtime.Object,
) rest.Storage {
	return &gcloudREST{
		groupResource: groupResource,
		codec:         codec,
		client:        client,
		opts:          opts,
		metaAccessor:  meta.NewAccessor(),
		isNamespaced:  isNamespaced,
		newFunc:       newFunc,
		newListFunc:   newListFunc,
	}
}

type gcloudREST struct {
	rest.TableConvertor
	groupResource schema.GroupResource
	codec         runtime.Codec
	client        *logadmin.Client
	opts          Options
	isNamespaced  bool
	metaAccessor  meta.MetadataAccessor
	newFunc       func() runtime.Object
	newListFunc   func() runtime.Object
}

func (r *gcloudREST) New() runtime.Object {
	return r.newFunc()
}

func (r *gcloudREST) NewList() runtime.Object {
	return r.newListFunc()
}

func (r *gcloudREST) NamespaceScoped() bool {
	return r.isNamespaced
}

func (r *gcloudREST) Destroy() {
}

func (r *gcloudREST) Get(
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
func (r *gcloudREST) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	if convert, ok := obj.(tableConvertor); ok {
		token, err := r.metaAccessor.Continue(obj)
		if err != nil {
			return nil, err
		}
		tbl, err := convert.ConvertToTable(ctx, tableOptions)
		tbl.ListMeta.Continue = token
		return tbl, err
	}

	return &metav1.Table{}, errors.New("could not convert to table")
}

func (r *gcloudREST) Watch(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	klog.InfoS("Start watch stream", "options", options)

	stream := &stream{
		rest: r,
		ch:   make(chan watch.Event, r.opts.Backend.BulkSize),
	}

	go func() {
		stream.Start(ctx, options)
	}()

	return stream, nil
}

func (r *gcloudREST) List(
	ctx context.Context,
	options *metainternalversion.ListOptions,
) (runtime.Object, error) {
	klog.InfoS("list request", "options", options)

	newListObj := r.NewList()
	/*v, err := getListPrt(newListObj)
	if err != nil {
		return nil, err
	}*/
	/*
		query := QueryFromListOptions(ctx, options, r)

		for _, hit = range esResults.Hits.Hits {
			decodedObj, err := r.decodeFrom(hit)
			if err != nil {
				return nil, err
			}

			appendItem(v, decodedObj)
		}

		// The continue token represents the last sort value from the last hit.
		// Which itself gets used in the next es query as search_after
		// If there is no hit there will be no continue token as this means we reached the end of available results
		if len(hit.Sort) > 0 {
			b, err := json.Marshal(hit.Sort)
			if err != nil {
				return newListObj, err
			}

			klog.InfoS("setting continue token", "token", string(b))
			r.metaAccessor.SetContinue(newListObj, string(b))
		}*/

	return newListObj, nil
}

func (r *gcloudREST) decodeFrom(obj *logging.Entry) (runtime.Object, error) {
	newObj := r.newFunc()
	asJSON, err := json.Marshal(obj)
	if err != nil {
		return newObj, err
	}

	jsonParsed, err := gabs.ParseJSON(asJSON)
	if err != nil {
		return newObj, err
	}

	for k, fields := range r.opts.FieldMap {
		for _, field := range fields {
			if field == "." {
				jsonParsed.SetP(asJSON, k)
			} else {
				if v := jsonParsed.Path(field); v != nil {
					jsonParsed.SetP(v.Data(), k)
					break
				}
			}
		}
	}

	jsonParsed, err = gabs.ParseJSON(jsonParsed.Bytes())
	if err != nil {
		return newObj, err
	}

	for _, field := range r.opts.DropFields {
		jsonParsed.DeleteP(field)
	}

	decodedObj, _, err := r.codec.Decode(jsonParsed.Bytes(), nil, newObj)
	if err != nil {
		return nil, err
	}

	annotations, _ := r.metaAccessor.Annotations(decodedObj)
	if annotations == nil {
		annotations = make(map[string]string)
	}

	if obj.SourceLocation != nil {
		annotations["kjournal/gcloud-location"] = obj.SourceLocation.File
	}

	r.metaAccessor.SetAnnotations(decodedObj, annotations)
	r.metaAccessor.SetUID(decodedObj, types.UID(obj.InsertID))

	return decodedObj, nil

	//return newObj, nil
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
		return reflect.Value{}, fmt.Errorf("need ptr to slice: %w", err)
	}

	return v, nil
}
