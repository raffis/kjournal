package gcloud

import (
	"context"

	"google.golang.org/api/iterator"
	statuserr "k8s.io/apimachinery/pkg/api/errors"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"k8s.io/apimachinery/pkg/watch"
)

type stream struct {
	rest *gcloudREST
	ch   chan watch.Event
	done chan bool
}

func (s *stream) errorAndAbort(err error) {
	status := statuserr.NewBadRequest(err.Error()).Status()
	s.ch <- watch.Event{
		Type:   watch.Error,
		Object: &status,
	}
}

func (s *stream) Start(ctx context.Context, options *metainternalversion.ListOptions) {

	it := s.rest.client.Entries(ctx)
	it.PageInfo().MaxSize = 1000

	for {
		hit, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			s.errorAndAbort(err)
			return
		}

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

func (s *stream) Stop() {

}

func (s *stream) ResultChan() <-chan watch.Event {
	return s.ch
}
