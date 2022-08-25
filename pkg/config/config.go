package config

import "time"

type Config struct {
	Backends []Backend `json:"backends"`
	Buckets  []Bucket  `json:"buckets"`
}

type Backend struct {
	Type          string `json:"type"`
	Name          string
	Elasticsearch BackendElasticsearch `json:"elasticsearch"`
}

type BackendElasticsearch struct {
	URL string
}

type Bucket struct {
	Type     string        `json:"type"`
	Name     string        `json:"name"`
	Backend  BucketBackend `json:"backend"`
	FieldMap map[string]string
}

type BucketBackend struct {
	Elasticsearch BucketBackendElasticsearch `json:"elasticsearch"`
}

type BucketBackendElasticsearch struct {
	Index       string `json:"index"`
	RefreshRate time.Duration
}
