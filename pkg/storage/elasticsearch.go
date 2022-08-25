package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	statuserr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/klog/v2"

	"github.com/raffis/kjournal/pkg/config"
)

var _ rest.Scoper = &elasticsearchREST{}
var _ rest.Storage = &elasticsearchREST{}

// NewelasticsearchREST instantiates a new REST storage.
func NewElasticsearchREST(
	groupResource schema.GroupResource,
	codec runtime.Codec,
	es *elasticsearch.Client,
	bucket config.Bucket,
	isNamespaced bool,
	newFunc func() runtime.Object,
	newListFunc func() runtime.Object,
) rest.Storage {
	return &elasticsearchREST{
		groupResource: groupResource,
		codec:         codec,
		es:            es,
		bucket:        bucket,
		metaAccessor:  meta.NewAccessor(),
		isNamespaced:  isNamespaced,
		newFunc:       newFunc,
		newListFunc:   newListFunc,
	}
}

var operatorMap = map[selection.Operator][]string{
	selection.Equals:       {"must", "match_phrase"},
	selection.DoubleEquals: {"must", "match_phrase"},
	selection.NotEquals:    {"must_not", "match_phrase"},
	selection.GreaterThan:  {"must", "range"},
	selection.LessThan:     {"must", "range"},
	selection.DoesNotExist: {"must_not", "exists"},
	selection.Exists:       {"must", "exists"},
}

type esHit struct {
	Index   string          `json:"_index"`
	DocType string          `json:"_type"`
	ID      string          `json:"_id"`
	Sort    []interface{}   `json:"sort"`
	Score   float64         `json:"_score"`
	Source  json.RawMessage `json:"_source"`
}

type esResults struct {
	Took     int64  `json:"took"`
	TimedOut bool   `json:"timed_out"`
	PitID    string `json:"pit_id"`
	Hits     struct {
		Hits     []esHit `json:"hits"`
		Took     float64 `json:"took"`
		MaxScore float64 `json:"max_score"`
		Total    struct {
			Value    int64  `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
	} `json:"hits"`
}

type elasticsearchREST struct {
	rest.TableConvertor
	groupResource schema.GroupResource
	codec         runtime.Codec
	es            *elasticsearch.Client
	bucket        config.Bucket
	isNamespaced  bool
	metaAccessor  meta.MetadataAccessor
	newFunc       func() runtime.Object
	newListFunc   func() runtime.Object
}

func (f *elasticsearchREST) New() runtime.Object {
	return f.newFunc()
}

func (f *elasticsearchREST) NewList() runtime.Object {
	return f.newListFunc()
}

func (f *elasticsearchREST) NamespaceScoped() bool {
	return f.isNamespaced
}

func (f *elasticsearchREST) Get(
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
func (f *elasticsearchREST) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
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

func (f *elasticsearchREST) Watch(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	klog.InfoS("Start watch stream", "options", options)

	jw := &streamer{
		refreshRate: f.bucket.RefreshRate,
		f:           f,
		ch:          make(chan watch.Event, 500),
	}

	go func() {
		options.Limit = 500
		jw.Start(ctx, options)
	}()

	return jw, nil
}

func (f *elasticsearchREST) List(
	ctx context.Context,
	options *metainternalversion.ListOptions,
) (runtime.Object, error) {
	klog.InfoS("List request", "options", options)

	newListObj := f.NewList()
	v, err := getListPrt(newListObj)
	if err != nil {
		return nil, err
	}

	if options.Limit == -1 {
		streamer := &PITStream{
			f:       f,
			options: options,
			context: ctx,
		}

		return streamer, nil
	}

	query, err := f.buildQuery(ctx, options)
	if err != nil {
		return nil, err
	}

	var hit esHit
	esResults, err := f.fetch(ctx, query, options)
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

		/*if f.isNamespaced {
			meta, err := meta.Accessor(decodedObj)
			if err != nil {
				return newListObj, err
			}

			if meta.GetNamespace() != ns {
				continue
			}
		}*/

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
		f.metaAccessor.SetContinue(newListObj, string(b))
	}

	return newListObj, nil
}

func (f *elasticsearchREST) fetch(
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
		f.es.Search.WithContext(ctx),
		f.es.Search.WithBody(&buf),
		//f.es.Search.WithTimeout(time.Duration(int64(time.Second) * int64(*options.TimeoutSeconds))),
		f.es.Search.WithTrackTotalHits(false),
	}

	if _, ok := query["pit"]; !ok {
		req = append(req, f.es.Search.WithIndex(f.bucket.Index))
	}

	if options.Limit != 0 {
		req = append(req, f.es.Search.WithSize(int(options.Limit)))
	}

	klog.InfoS("executing elasicsearch query...", "query", query)
	res, err := f.es.Search(req...)
	klog.InfoS("done...")

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

	klog.InfoS("elasticsearch query result arrived", "duration", time.Duration(esResults.Took*int64(time.Millisecond)).String(), "timed-out", esResults.TimedOut, "number-of-hits", len(esResults.Hits.Hits))
	return esResults, err
}

func (f *elasticsearchREST) buildQuery(
	ctx context.Context,
	options *metainternalversion.ListOptions,
) (map[string]interface{}, error) {
	query := map[string]interface{}{
		"_source": map[string]interface{}{
			"excludes": []interface{}{"kind"},
		},
		"sort": []map[string]interface{}{
			{
				f.bucket.TimestampField: "asc",
			},
		},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":     []map[string]interface{}{},
				"must_not": []map[string]interface{}{},
			},
		},
	}

	if options.Continue != "" {
		var searchAfter []interface{}
		err := json.Unmarshal([]byte(options.Continue), &searchAfter)
		if err != nil {
			return query, err
		}

		query["search_after"] = searchAfter
	}

	var skipTimestampFilter bool
	requirements, _ := options.LabelSelector.Requirements()
	for _, req := range requirements {
		operator := operatorMap[req.Operator()]

		q := query["query"].(map[string]interface{})["bool"].(map[string]interface{})[operator[0]].([]map[string]interface{})
		field := req.Key()

		switch field {
		case "pod":
			field = "kubernetes.pod_name"
		case "container":
			field = "kubernetes.container_name"
		case "requestReceivedTimestamp":
			field = f.bucket.TimestampField
		case "creationTimestamp":
			field = f.bucket.TimestampField
		}

		var match map[string]interface{}

		switch req.Operator() {
		case selection.LessThan:
			match = map[string]interface{}{
				operator[1]: map[string]interface{}{
					field: map[string]interface{}{
						"lt": req.Values().UnsortedList()[0],
					},
				},
			}
		case selection.GreaterThan:
			match = map[string]interface{}{
				operator[1]: map[string]interface{}{
					field: map[string]interface{}{
						"gt": req.Values().UnsortedList()[0],
					},
				},
			}
		case selection.Exists:
		case selection.DoesNotExist:
			match = map[string]interface{}{
				operator[1]: map[string]interface{}{
					"field": field,
				},
			}
		default:
			match = map[string]interface{}{
				operator[1]: map[string]interface{}{
					field: req.Values().UnsortedList()[0],
				},
			}
		}

		q = append(q, match)
		query["query"].(map[string]interface{})["bool"].(map[string]interface{})[operator[0]] = q

		if !skipTimestampFilter && field == f.bucket.TimestampField {
			skipTimestampFilter = true
		}
	}

	if !skipTimestampFilter {
		q := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})
		match := map[string]interface{}{
			"range": map[string]interface{}{
				f.bucket.TimestampField: map[string]interface{}{
					"gte": "now-5h",
				},
			},
		}
		q = append(q, match)
		query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = q

	}

	/*if f.groupResource.Resource == "events" {
		q := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})
		match := map[string]interface{}{
			"match": map[string]interface{}{
				"kind.keyword": "Event",
			},
		}
		q = append(q, match)
		query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = q
	}*/

	// If resource is namespaced objectRef.namespace will always be set to the current calling context namespace
	if f.isNamespaced {
		ns, _ := request.NamespaceFrom(ctx)
		q := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})
		match := map[string]interface{}{
			"match_phrase": map[string]interface{}{
				f.bucket.NamespaceField: ns,
			},
		}
		q = append(q, match)
		query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = q
	}

	return query, nil
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

type pit struct {
	ID string `json:"id"`
}

type streamer struct {
	usePIT      bool
	f           *elasticsearchREST
	refreshRate time.Duration
	ch          chan watch.Event
	done        chan bool
	pit         pit
}

func (w *streamer) errorAndAbort(err error) {
	status := statuserr.NewBadRequest(err.Error()).Status()
	w.ch <- watch.Event{
		Type:   watch.Error,
		Object: &status,
	}
}

func (w *streamer) Start(ctx context.Context, options *metainternalversion.ListOptions) {
	if w.usePIT {
		res, err := w.f.es.OpenPointInTime([]string{w.f.bucket.Index}, "5m")
		if err != nil {
			w.errorAndAbort(err)
			return
		}

		if err := json.NewDecoder(res.Body).Decode(&w.pit); err != nil {
			w.errorAndAbort(err)
			return
		}
	}

	for {
		query, err := w.f.buildQuery(ctx, options)
		if err != nil {
			w.errorAndAbort(err)
			return
		}

		if w.pit.ID != "" {
			query["pit"] = map[string]interface{}{
				"id":         w.pit.ID,
				"keep_alive": "5m",
			}
		}

		esResults, err := w.f.fetch(ctx, query, options)
		if err != nil {
			w.errorAndAbort(err)
			return
		}

		var hit esHit
		//ns, _ := request.NamespaceFrom(ctx)

		for _, hit = range esResults.Hits.Hits {
			newObj := w.f.newFunc()
			decodedObj, _, err := w.f.codec.Decode(hit.Source, nil, newObj)
			if err != nil {
				return
			}

			/*if w.f.isNamespaced {
				meta, err := meta.Accessor(decodedObj)
				if err != nil {
					w.errorAndAbort(err)
					return
				}

				if meta.GetNamespace() != ns {
					continue
				}
			}*/

			w.ch <- watch.Event{
				Type:   watch.Added,
				Object: decodedObj,
			}
		}

		if len(esResults.Hits.Hits) != 500 && w.refreshRate == 0 {
			klog.Info("All objects consumed from pit")
			w.done <- true
			break
		}

		if len(esResults.Hits.Hits) != 500 {
			klog.InfoS("wait for next check", "sleep", w.refreshRate.String())
			time.Sleep(w.refreshRate)
		}

		// The continue token represents teh last sort value from the last hit.
		// Which itself gets used in the next es query as search_after
		// If there is no hit there will be no continue token as this means we reached the end of available results
		if len(hit.Sort) > 0 {
			b, err := json.Marshal(hit.Sort)
			if err != nil {
				w.errorAndAbort(err)
				return
			}

			options.Continue = string(b)
		}

		// For the next search request the PIT from the previous search response needs to be taken as it can change over time
		if w.pit.ID != "" {
			w.pit.ID = esResults.PitID
		}
	}
}

func (w *streamer) Stop() {
	_, err := w.f.es.ClosePointInTime()
	if err != nil {
		klog.ErrorS(err, "failed to close pit")
	}

}

func (w *streamer) ResultChan() <-chan watch.Event {
	return w.ch
}

type PITStream struct {
	f       *elasticsearchREST
	options *metainternalversion.ListOptions
	context context.Context
}

var _ rest.ResourceStreamer = &PITStream{}

func (obj *PITStream) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}
func (obj *PITStream) DeepCopyObject() runtime.Object {
	panic("rest.PITStream does not implement DeepCopyObject")
}

// InputStream returns a stream with the contents of the URL location. If no location is provided,
// a null stream is returned.
func (s *PITStream) InputStream(ctx context.Context, apiVersion, acceptHeader string) (stream io.ReadCloser, flush bool, contentType string, err error) {
	jw := &streamer{
		usePIT:      true,
		refreshRate: 0,
		f:           s.f,
		ch:          make(chan watch.Event, 500),
		done:        make(chan bool),
	}

	pipe := &Pipe{
		streamer: jw,
	}

	go func() {
		s.options.Limit = 500
		jw.Start(s.context, s.options)
	}()

	return pipe, false, "application/json", nil
}

type Pipe struct {
	streamer *streamer
}

func (p *Pipe) Read(dst []byte) (n int, err error) {
	read := func(doc watch.Event) (int, error) {
		b, err := json.Marshal(metav1.WatchEvent{
			Type:   string(doc.Type),
			Object: runtime.RawExtension{Object: doc.Object},
		})

		s := copy(dst, b)
		return s, err
	}

	select {
	case doc, ok := <-p.streamer.ResultChan():
		if ok {
			return read(doc)
		}

		return 0, io.EOF
	case <-p.streamer.done:
		close(p.streamer.ch)
	}

	return 0, nil
}

func (p *Pipe) Close() error {
	p.streamer.Stop()
	return nil
}
