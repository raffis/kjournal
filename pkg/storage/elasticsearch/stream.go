package elasticsearch

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	statuserr "k8s.io/apimachinery/pkg/api/errors"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"k8s.io/apimachinery/pkg/watch"
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

	var wg sync.WaitGroup
	wg.Add(1)
	batchResults := make(chan esResults, 2)

	// read ahead buffer
	go func() {
		for {
			klog.Info("start list query", "options", options)
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

			batchResults <- esResults

			if len(esResults.Hits.Hits) != int(s.rest.opts.Backend.BulkSize) && s.refreshRate == 0 {
				klog.Info("All objects consumed from stream")
				s.done <- true
				wg.Done()
				break
			}

			if len(esResults.Hits.Hits) != int(s.rest.opts.Backend.BulkSize) {
				klog.InfoS("wait for next check", "sleep", s.refreshRate.String())
				time.Sleep(s.refreshRate)
			}

			// The continue token represents the last sort value from the last hit.
			// Which itself gets used in the next es query as search_after
			// If there is no hit there will be no continue token as this means we reached the end of available results
			if len(esResults.Hits.Hits) > 0 {
				hit := esResults.Hits.Hits[len(esResults.Hits.Hits)-1]
				if len(hit.Sort) > 0 {
					b, err := json.Marshal(hit.Sort)
					if err != nil {
						s.errorAndAbort(err)
						return
					}

					options.Continue = string(b)
				}
			}

			// For the next search request the PIT from the previous search response needs to be taken as it can change over time
			if s.pit.ID != "" {
				s.pit.ID = esResults.PitID
			}
		}
	}()

	// loop over batched results
	for esResults := range batchResults {
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
	}

	wg.Wait()
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
