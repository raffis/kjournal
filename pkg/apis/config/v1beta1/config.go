package v1beta1

type Config struct {
	Backend Backend  `json:"backend"`
	Buckets []Bucket `json:"buckets"`
}

type Backend struct {
	Type          string               `json:"type"`
	Elasticsearch BackendElasticsearch `json:"elasticsearch"`
}

type BackendElasticsearch struct {
}

type Bucket struct {
	Type       string        `json:"type"`
	Name       string        `json:"name"`
	Namespaced bool          `json:"namespaced"`
	Backend    BucketBackend `json:"backend"`
}

type BucketBackend struct {
	Elasticsearch BucketBackendElasticsearch `json:"elasticsearch"`
}

type BucketBackendElasticsearch struct {
	Index string `json:"index"`
}
