package v1alpha1

import (
	"time"

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
}

type BackendElasticsearch struct {
	URL              []string `json:"url"`
	AllowInsecureTLS bool     `json:"allowInsecureTLS"`
	CACert           string   `json:"cacert"`
}

type API struct {
	Resource string            `json:"resource"`
	FieldMap map[string]string `json:"fieldMap"`
	Filter   map[string]string `json:"filter"`
	Backend  ApiBackend        `json:"backend"`
	DocRoot  string            `json:"docRoot"`
}

type ApiBackend struct {
	Elasticsearch BucketBackendElasticsearch `json:"elasticsearch"`
}

type BucketBackendElasticsearch struct {
	Index       string        `json:"index"`
	RefreshRate time.Duration `json:"refreshRate"`
}
