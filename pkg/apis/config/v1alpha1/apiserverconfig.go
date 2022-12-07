package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// APIServerConfig
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type APIServerConfig struct {
	metav1.TypeMeta `json:",inline,omitempty"`
	Backend         Backend `json:"backend,omitempty"`
	Apis            []API   `json:"apis,omitempty"`
}

type Backend struct {
	Elasticsearch *BackendElasticsearch `json:"elasticsearch,omitempty"`
}

type TLS struct {
	AllowInsecure bool   `json:"allowInsecure,omitempty"`
	CACert        string `json:"caCert,omitempty"`
	ServerName    string `json:"serverName,omitempty"`
}

type BackendElasticsearch struct {
	URL []string `json:"url,omitempty"`
	TLS TLS      `json:"tls,omitempty"`
}

type API struct {
	Resource         string              `json:"resource,omitempty"`
	FieldMap         map[string][]string `json:"fieldMap,omitempty"`
	DropFields       []string            `json:"dropFields,omitempty"`
	Filter           string              `json:"filter,omitempty"`
	Backend          ApiBackend          `json:"backend,omitempty"`
	DefaultTimeRange string              `json:"defaultTimeRange,omitempty"`
}

type ApiBackend struct {
	Elasticsearch ApiBackendElasticsearch `json:"elasticsearch,omitempty"`
}

type ApiBackendElasticsearch struct {
	Index           string          `json:"index,omitempty"`
	RefreshRate     metav1.Duration `json:"refreshRate,omitempty"`
	TimestampFields []string        `json:"timestampFields,omitempty"`
	BulkSize        int64           `json:"bulkSize,omitempty"`
}
