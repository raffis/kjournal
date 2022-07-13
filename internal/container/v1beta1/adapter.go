package v1beta1

import (
	"encoding/json"
	"time"

	"github.com/raffis/kjournal/pkg/apis/container/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Log struct {
	v1beta1.Log
	Unstructured map[string]interface{} `json:"unstructured"`
	Env          map[string]interface{} `json:"env"`
}

func (in *Log) UnmarshalJSON(bs []byte) (err error) {
	if err = json.Unmarshal(bs, &in.Unstructured); err != nil {
		return err
	}

	if v, ok := in.Unstructured["kubernetes"]; ok {
		in.Env = v.(map[string]interface{})
		delete(in.Unstructured, "kubernetes")
	}

	if v, ok := in.Unstructured["@es_ts"]; ok {
		ts, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return err
		}

		in.Metadata.CreationTimestamp = v1.Time{Time: ts}
		delete(in.Unstructured, "@es_ts")
	}

	if v, ok := in.Env["pod_name"]; ok {
		in.Pod = v.(string)
		delete(in.Env, "pod_name")
	}

	if v, ok := in.Env["container_name"]; ok {
		in.Container = v.(string)
		delete(in.Env, "container_name")
	}

	if v, ok := in.Env["namespace_name"]; ok {
		in.Metadata.Namespace = v.(string)
		delete(in.Env, "namespace_name")
	}

	in.TypeMeta.Kind = "Log"
	in.TypeMeta.APIVersion = "container.kjournal/v1beta1"

	return nil
}

func (in *Log) New() runtime.Object {
	return &Log{}
}

func (in *Log) NewList() runtime.Object {
	return &LogList{}
}

type LogList struct {
	v1beta1.LogList
	Items []Log `json:"items"`
}
