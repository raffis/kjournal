package storage

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/validation"
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
	backendRegistry utils.Registry[*configv1alpha1.Backend]
	apiRegistry     utils.Registry[*configv1alpha1.API]
}

func NewProvider(conf configv1alpha1.APIServerConfig) (Provider, error) {
	p := &provider{
		backendRegistry: utils.NewRegistry[*configv1alpha1.Backend](),
		apiRegistry:     utils.NewRegistry[*configv1alpha1.API](),
	}

	for _, v := range conf.Backends {
		backend := v
		if errs := validation.NameIsDNSSubdomain(backend.Name, false); len(errs) > 0 {
			return nil, fmt.Errorf("%w: can not add key %s (%s)", ErrNameInvalid, backend.Name, strings.Join(errs, ";"))
		}

		if err := p.backendRegistry.Add(backend.Name, &backend); err != nil {
			return nil, err
		}
	}

	for _, v := range conf.Apis {
		apiBinding := v
		if err := p.apiRegistry.Add(apiBinding.Name, &apiBinding); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (p *provider) Provide(obj resource.Object, scheme *runtime.Scheme, getter generic.RESTOptionsGetter) (rest.Storage, error) {
	var (
		key        = obj.GetGroupVersionResource().Resource
		apiBinding *configv1alpha1.API
		backend    *configv1alpha1.Backend
		err        error
	)

	if apiBinding, err = p.apiRegistry.Get(key); err != nil {
		return nil, fmt.Errorf("%w: no api binding found for %s", err, key)
	}

	if backend, err = p.backendRegistry.Get(apiBinding.Backend.Name); err != nil {
		return nil, fmt.Errorf("%w: no backend found named %s", err, apiBinding.Backend.Name)
	}

	fmt.Printf("\n\n#############################################333 init %#v -------------- %#v --  %#v", obj, apiBinding, key)
	provider, err := Providers.Get(backend.Type)
	if err != nil {
		return nil, fmt.Errorf("%w: unsupported provder", err)
	}

	if v, ok := obj.(MappableObject); ok {
		v.WithFieldMap(apiBinding.FieldMap)
	}

	return provider(obj, scheme, getter, backend, apiBinding)
}
