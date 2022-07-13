package storage

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
		res, err := w.f.es.OpenPointInTime([]string{w.f.opts.Index}, "5m")
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
