package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/Jeffail/gabs"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
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

var _ rest.Scoper = &elasticsearchREST{}
var _ rest.Storage = &elasticsearchREST{}
var _ rest.TableConvertor = &elasticsearchREST{}

// NewelasticsearchREST instantiates a new REST storage.
func NewElasticsearchREST(
	groupResource schema.GroupResource,
	codec runtime.Codec,
	es *elasticsearch.Client,
	opts Options,
	isNamespaced bool,
	newFunc func() runtime.Object,
	newListFunc func() runtime.Object,
) rest.Storage {
	return &elasticsearchREST{
		groupResource: groupResource,
		codec:         codec,
		es:            es,
		opts:          opts,
		metaAccessor:  meta.NewAccessor(),
		isNamespaced:  isNamespaced,
		newFunc:       newFunc,
		newListFunc:   newListFunc,
	}
}

type elasticsearchREST struct {
	rest.TableConvertor
	groupResource schema.GroupResource
	codec         runtime.Codec
	es            *elasticsearch.Client
	opts          Options
	isNamespaced  bool
	metaAccessor  meta.MetadataAccessor
	newFunc       func() runtime.Object
	newListFunc   func() runtime.Object
}

func (r *elasticsearchREST) New() runtime.Object {
	return r.newFunc()
}

func (r *elasticsearchREST) NewList() runtime.Object {
	return r.newListFunc()
}

func (r *elasticsearchREST) NamespaceScoped() bool {
	return r.isNamespaced
}

func (r *elasticsearchREST) Destroy() {
}

func (r *elasticsearchREST) Get(
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
func (r *elasticsearchREST) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
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

func (r *elasticsearchREST) Watch(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	klog.InfoS("Start watch stream", "options", options)

	stream := &stream{
		refreshRate: r.opts.Backend.RefreshRate,
		rest:        r,
		ch:          make(chan watch.Event, r.opts.Backend.BulkSize),
	}

	go func() {
		options.Limit = r.opts.Backend.BulkSize
		stream.Start(ctx, options)
	}()

	return stream, nil
}

func (r *elasticsearchREST) List(
	ctx context.Context,
	options *metainternalversion.ListOptions,
) (runtime.Object, error) {
	klog.InfoS("list request", "options", options)

	newListObj := r.NewList()
	v, err := getListPrt(newListObj)
	if err != nil {
		return nil, err
	}

	query, err := queryFromListOptions(ctx, options, r)
	if err != nil {
		return newListObj, err
	}

	var hit esHit
	esResults, err := r.fetch(ctx, query, options)
	if err != nil {
		return nil, err
	}

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
		if err := r.metaAccessor.SetContinue(newListObj, string(b)); err != nil {
			return newListObj, err
		}
	}

	return newListObj, nil
}

func (r *elasticsearchREST) fetch(
	ctx context.Context,
	query map[string]interface{},
	options *metainternalversion.ListOptions,
) (esResults, error) {
	var esResults esResults

	// Build the request body.
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		klog.ErrorS(err, "error encoding query")
		return esResults, err
	}

	req := []func(*esapi.SearchRequest){
		r.es.Search.WithContext(ctx),
		r.es.Search.WithBody(&buf),
		//r.es.Search.WithTimeout(time.Duration(int64(time.Second) * int64(*options.TimeoutSeconds))),
		r.es.Search.WithTrackTotalHits(false),
	}

	if _, ok := query["pit"]; !ok {
		req = append(req, r.es.Search.WithIndex(r.opts.Backend.Index))
	}

	if options.Limit != 0 {
		req = append(req, r.es.Search.WithSize(int(options.Limit)))
	}

	res, err := r.es.Search(req...)
	if err != nil {
		klog.ErrorS(err, "error getting response from es")
		return esResults, err
	}

	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			klog.ErrorS(err, "error parsing the response body")
			return esResults, err
		} else {
			klog.ErrorS(err, "error parsing the response body", "status", res.Status(), "body", e)
			return esResults, err
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&esResults); err != nil {
		return esResults, err
	}

	klog.InfoS("elasticsearch query result arrived", "duration", time.Duration(esResults.Took*int64(time.Millisecond)).String(), "timed-out", esResults.TimedOut, "number-of-hits", len(esResults.Hits.Hits), "shards", esResults.Shards)
	return esResults, err
}

func (r *elasticsearchREST) decodeFrom(obj esHit) (runtime.Object, error) {
	newObj := r.newFunc()

	jsonParsed, err := gabs.ParseJSON(obj.Source)
	if err != nil {
		return newObj, err
	}

	for k, fields := range r.opts.FieldMap {
		for _, field := range fields {
			if field == "." {
				if _, err := jsonParsed.SetP(obj.Source, k); err != nil {
					return newObj, err
				}
			} else {
				if v := jsonParsed.Path(field); v != nil {
					if _, err := jsonParsed.SetP(v.Data(), k); err != nil {
						return newObj, err
					}
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
		_ = jsonParsed.DeleteP(field)
	}

	decodedObj, _, err := r.codec.Decode(jsonParsed.Bytes(), nil, newObj)
	if err != nil {
		return nil, err
	}

	annotations, _ := r.metaAccessor.Annotations(decodedObj)
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations["kjournal/es-index"] = obj.Index
	if err := r.metaAccessor.SetAnnotations(decodedObj, annotations); err != nil {
		return decodedObj, err
	}

	if err := r.metaAccessor.SetUID(decodedObj, types.UID(obj.ID)); err != nil {
		return decodedObj, err
	}

	return decodedObj, err
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
