package storage

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"

	configv1alpha1 "github.com/raffis/kjournal/pkg/apis/config/v1alpha1"
	"github.com/raffis/kjournal/pkg/utils"
)

var (
	ErrUnsupportedBackend = errors.New("unsupported backend")
	ErrNameInvalid        = errors.New("invalid key provided")
	Providers             utils.Registry[RestProvider]
)

func init() {
	Providers = utils.NewRegistry[RestProvider]()
}

type RestProvider func(obj resource.Object, scheme *runtime.Scheme, getter generic.RESTOptionsGetter, backend *configv1alpha1.Backend, apiBinding *configv1alpha1.API) (rest.Storage, error)

type Provider interface {
	Provide(obj resource.Object, scheme *runtime.Scheme, getter generic.RESTOptionsGetter) (rest.Storage, error)
}

type MappableObject interface {
	WithFieldMap(map[string]string)
}

type provider struct {
	backend      *configv1alpha1.Backend
	restProvider RestProvider
	apiRegistry  utils.Registry[*configv1alpha1.API]
}

func NewProvider(conf configv1alpha1.APIServerConfig) (Provider, error) {
	p := &provider{
		backend:     &conf.Backend,
		apiRegistry: utils.NewRegistry[*configv1alpha1.API](),
	}

	t, err := getType(conf.Backend)
	if err != nil {
		return nil, err
	}

	provider, err := Providers.Get(t)
	if err != nil {
		return nil, fmt.Errorf("%w: unsupported provder", err)
	}

	p.restProvider = provider

	for _, v := range conf.Apis {
		apiBinding := v
		if err := p.apiRegistry.Add(apiBinding.Resource, &apiBinding); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (p *provider) Provide(obj resource.Object, scheme *runtime.Scheme, getter generic.RESTOptionsGetter) (rest.Storage, error) {
	var (
		key        = obj.GetGroupVersionResource().Resource
		apiBinding *configv1alpha1.API
		err        error
	)

	if apiBinding, err = p.apiRegistry.Get(key); err != nil {
		return nil, fmt.Errorf("%w: no api binding found for %s", err, key)
	}

	if v, ok := obj.(MappableObject); ok {
		v.WithFieldMap(apiBinding.FieldMap)
	}

	return p.restProvider(obj, scheme, getter, p.backend, apiBinding)
}

func getType(conf configv1alpha1.Backend) (string, error) {
	if conf.Elasticsearch != nil {
		return "elasticsearch", nil
	}

	return "", ErrUnsupportedBackend
}
