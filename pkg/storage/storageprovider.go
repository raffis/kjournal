package storage

import (
	"sync"

	"github.com/raffis/kjournal/pkg/apiserver"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	registryrest "k8s.io/apiserver/pkg/registry/rest"
)

// SingletonProvider ensures different versions of the same resource share storage
type SingletonProvider struct {
	sync.Once
	Provider apiserver.StorageProvider
	storage  registryrest.Storage
	err      error
}

func (s *SingletonProvider) Get(
	scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (registryrest.Storage, error) {
	s.Once.Do(func() {
		s.storage, s.err = s.Provider(scheme, optsGetter)
	})
	return s.storage, s.err
}
