package elasticsearch

import (
	"context"
	"encoding/json"
	"io"
	"time"

	statuserr "k8s.io/apimachinery/pkg/api/errors"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/klog/v2"
)

type pit struct {
	ID string `json:"id"`
}

type stream struct {
	usePIT      bool
	rest        *elasticsearchREST
	refreshRate time.Duration
	ch          chan watch.Event
	done        chan bool
	pit         pit
}

func (s *stream) errorAndAbort(err error) {
	status := statuserr.NewBadRequest(err.Error()).Status()
	s.ch <- watch.Event{
		Type:   watch.Error,
		Object: &status,
	}
}

func (s *stream) Start(ctx context.Context, options *metainternalversion.ListOptions) {
	if s.usePIT {
		res, err := s.rest.es.OpenPointInTime([]string{s.rest.opts.Backend.Index}, "5m")
		if err != nil {
			s.errorAndAbort(err)
			return
		}

		if err := json.NewDecoder(res.Body).Decode(&s.pit); err != nil {
			s.errorAndAbort(err)
			return
		}
	}

	for {
		query, err := queryFromListOptions(ctx, options, s.rest)
		if err != nil {
			return
		}

		if s.pit.ID != "" {
			query["pit"] = map[string]interface{}{
				"id":         s.pit.ID,
				"keep_alive": "5m",
			}
		}

		esResults, err := s.rest.fetch(ctx, query, options)
		if err != nil {
			s.errorAndAbort(err)
			return
		}

		var hit esHit
		for _, hit = range esResults.Hits.Hits {
			decodedObj, err := s.rest.decodeFrom(hit)
			if err != nil {
				break
			}

			s.ch <- watch.Event{
				Type:   watch.Added,
				Object: decodedObj,
			}
		}

		if len(esResults.Hits.Hits) != int(s.rest.opts.Backend.BulkSize) && s.refreshRate == 0 {
			klog.Info("All objects consumed from stream")
			s.done <- true
			break
		}

		if len(esResults.Hits.Hits) != int(s.rest.opts.Backend.BulkSize) {
			klog.InfoS("wait for next check", "sleep", s.refreshRate.String())
			time.Sleep(s.refreshRate)
		}

		// The continue token represents the last sort value from the last hit.
		// Which itself gets used in the next es query as search_after
		// If there is no hit there will be no continue token as this means we reached the end of available results
		if len(hit.Sort) > 0 {
			b, err := json.Marshal(hit.Sort)
			if err != nil {
				s.errorAndAbort(err)
				return
			}

			options.Continue = string(b)
		}

		// For the next search request the PIT from the previous search response needs to be taken as it can change over time
		if s.pit.ID != "" {
			s.pit.ID = esResults.PitID
		}
	}
}

func (s *stream) Stop() {
	_, err := s.rest.es.ClosePointInTime()
	if err != nil {
		klog.ErrorS(err, "failed to close pit")
	}

}

func (s *stream) ResultChan() <-chan watch.Event {
	return s.ch
}

type pitStream struct {
	rest    *elasticsearchREST
	options *metainternalversion.ListOptions
	context context.Context
}

var _ rest.ResourceStreamer = &pitStream{}

func (obj *pitStream) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

func (obj *pitStream) DeepCopyObject() runtime.Object {
	panic("rest.PITStream does not implement DeepCopyObject")
}

func (s *pitStream) InputStream(ctx context.Context, apiVersion, acceptHeader string) (io.ReadCloser, bool, string, error) {
	stream := &stream{
		usePIT:      true,
		refreshRate: 0,
		rest:        s.rest,
		ch:          make(chan watch.Event, int(s.rest.opts.Backend.BulkSize)),
		done:        make(chan bool),
	}

	r := &streamReader{
		stream: stream,
	}

	go func() {
		s.options.Limit = s.rest.opts.Backend.BulkSize
		stream.Start(s.context, s.options)
	}()

	return r, false, "application/json", nil
}

type streamReader struct {
	stream *stream
}

func (r *streamReader) Read(dst []byte) (n int, err error) {
	read := func(doc watch.Event) (int, error) {
		b, err := json.Marshal(metav1.WatchEvent{
			Type:   string(doc.Type),
			Object: runtime.RawExtension{Object: doc.Object},
		})

		s := copy(dst, b)
		return s, err
	}

	select {
	case doc, ok := <-r.stream.ResultChan():
		if ok {
			return read(doc)
		}

		return 0, io.EOF
	case <-r.stream.done:
		close(r.stream.ch)
	}

	return 0, nil
}

func (r *streamReader) Close() error {
	r.stream.Stop()
	return nil
}
