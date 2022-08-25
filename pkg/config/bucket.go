package config

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/validation"
)

var (
	ErrBucketNotFound      = errors.New("backend not found.")
	ErrBucketAlreadyExists = errors.New("bucket with the same name already exists.")
	ErrBucketNameInvalid   = errors.New("invalid bucket name provided.")
)

type BucketRegistry interface {
	AddBucket(bucket Bucket) error
	GetBucket(name string) (Bucket, error)
}

type bucketRegistry struct {
	store []Bucket
}

func NewBucketRegistry() BucketRegistry {
	return &bucketRegistry{}
}

func (r *bucketRegistry) AddBucket(bucket Bucket) error {
	for _, v := range r.store {
		if v.Name == bucket.Name {
			return fmt.Errorf("%w: can not add bucket %s", ErrBucketAlreadyExists, bucket.Name)
		}
	}

	if errs := validation.NameIsDNSSubdomain(bucket.Name, false); len(errs) > 0 {
		return fmt.Errorf("%w: can not add bucket %s (%s)", ErrBucketNameInvalid, strings.Join(errs, ";"))
	}

	r.store = append(r.store, bucket)
	return nil
}

func (r *bucketRegistry) GetBucket(name string) (Bucket, error) {
	for _, v := range r.store {
		if v.Name == name {
			return v, nil
		}
	}

	return nil, fmt.Errorf("%w: name %s not found", ErrBucketNotFound, name)
}
