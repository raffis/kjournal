package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// APIServerConfig
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type APIServerConfig struct {
	metav1.TypeMeta `json:",inline"`
	Backend         Backend `json:"backend"`
	Apis            []API   `json:"apis"`
}

type Backend struct {
	Elasticsearch *BackendElasticsearch `json:"elasticsearch"`
	GCloud        *BackendGCloud        `json:"gcloud"`
}

type BackendElasticsearch struct {
	URL              []string `json:"url"`
	AllowInsecureTLS bool     `json:"allowInsecureTLS"`
	CACert           string   `json:"cacert"`
}

type BackendGCloud struct {
	APIKey string `json:"apiKey"`
}

type API struct {
	Resource         string              `json:"resource"`
	FieldMap         map[string][]string `json:"fieldMap"`
	DropFields       []string            `json:"dropFields"`
	Filter           map[string]string   `json:"filter"`
	Backend          ApiBackend          `json:"backend"`
	DefaultTimeRange string              `json:"defaultTimeRange"`
}

type ApiBackend struct {
	Elasticsearch ApiBackendElasticsearch `json:"elasticsearch"`
}

type ApiBackendElasticsearch struct {
	Index           string          `json:"index"`
	RefreshRate     metav1.Duration `json:"refreshRate"`
	TimestampFields []string        `json:"timestampFields"`
	BulkSize        int64           `json:"bulkSize"`
}
