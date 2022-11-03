package utils

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound      = errors.New("key not found")
	ErrAlreadyExists = errors.New("entry with the same name already exists")
)

type Registry[T any] interface {
	Add(key string, value T) error
	Get(key string) (T, error)
	MustRegister(key string, value T)
}

type registry[T any] struct {
	store map[string]T
}

func NewRegistry[T any]() Registry[T] {
	return &registry[T]{
		store: make(map[string]T),
	}
}

func (r *registry[T]) Add(key string, value T) error {
	if _, ok := r.store[key]; ok {
		return fmt.Errorf("%w: can not add key %s", ErrAlreadyExists, key)
	}

	r.store[key] = value
	return nil
}

func (r *registry[T]) MustRegister(key string, value T) {
	if err := r.Add(key, value); err != nil {
		panic(err)
	}
}

func (r *registry[T]) Get(key string) (T, error) {
	if _, ok := r.store[key]; ok {
		return r.store[key], nil
	}

	var result T
	return result, fmt.Errorf("%w: name %s not found", ErrNotFound, key)
}
