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
	fmt.Printf("\n STORE KEY %#v -- %#v -- %#v\n", key, r.store, value)

	r.store[key] = value
	return nil
}

func (r *registry[T]) Get(key string) (T, error) {
	fmt.Printf("\nGET KEY %#v -- %#v -- %#v\n", key, r.store, r.store[key])

	for k, v := range r.store {
		fmt.Printf("DEBUG key: %#v -- value: %#v\n\n", k, v)
	}

	if _, ok := r.store[key]; ok {
		return r.store[key], nil
	}

	var result T
	return result, fmt.Errorf("%w: name %s not found", ErrNotFound, key)
}
