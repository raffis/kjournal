package config

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/validation"
)

var (
	ErrBackendNotFound      = errors.New("backend not found.")
	ErrBackendAlreadyExists = errors.New("backend with the same name already exists.")
	ErrBackendNameInvalid   = errors.New("invalid backend name provided.")
)

type BackendRegistry interface {
	AddBackend(Backend) error
	GetBackends() []Backend
}

type backendRegistry struct {
	store []Backend
}

func NewBackendRegistry() BackendRegistry {
	return &backendRegistry{}
}

func (r *backendRegistry) AddBackend(backend Backend) error {
	for _, v := range r.store {
		if v.Name == backend.Name {
			return fmt.Errorf("%w: can not add backend %s", ErrBackendAlreadyExists, backend.Name)
		}
	}

	if errs := validation.NameIsDNSSubdomain(backend.Name, false); len(errs) > 0 {
		return fmt.Errorf("%w: can not add backend %s (%s)", ErrBackendNameInvalid, strings.Join(errs, ";"))
	}

	r.store = append(r.store, backend)
	return nil
}

func (r *backendRegistry) GetBackend(name string) (Backend, error) {
	for _, v := range r.store {
		if v.Name == backend.Name {
			return v, nil
		}
	}

	return nil, fmt.Errorf("%w: name %s not found", ErrBackendNotFound, backend.Name)
}
