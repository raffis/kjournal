package v1beta1

import (
	"encoding/json"
	"time"

	"github.com/raffis/kjournal/pkg/apis/core/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ContainerLog struct {
	v1alpha1.ContainerLog
	Payload map[string]interface{} `json:"payload"`
	//Env          map[string]interface{} `json:"env"`
	fieldMap map[string]string
}

func (in *ContainerLog) UnmarshalJSON(bs []byte) (err error) {
	if err = json.Unmarshal(bs, &in.Payload); err != nil {
		return err
	}

	/*if v, ok := in.Payload["kubernetes"]; ok {
		in.Env = v.(map[string]interface{})
		delete(in.Payload, "kubernetes")
	}*/

	if v, ok := in.Payload["@es_ts"]; ok {
		ts, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return err
		}

		in.CreationTimestamp = v1.Time{Time: ts}
		delete(in.Payload, "@es_ts")
	}

	if v, ok := in.Payload["pod_name"]; ok {
		in.Pod = v.(string)
		delete(in.Payload, "pod_name")
	}

	if v, ok := in.Payload["container_name"]; ok {
		in.Container = v.(string)
		delete(in.Payload, "container_name")
	}

	if v, ok := in.Payload["namespace_name"]; ok {
		in.Namespace = v.(string)
		delete(in.Payload, "namespace_name")
	}

	in.TypeMeta.Kind = "Log"
	in.TypeMeta.APIVersion = "container.kjournal/v1beta1"

	return nil
}

func (in *ContainerLog) WithFieldMap(fieldMap map[string]string) *ContainerLog {
	in.fieldMap = fieldMap
	return in
}

func (in *ContainerLog) New() runtime.Object {
	return &ContainerLog{}
}

func (in *ContainerLog) NewList() runtime.Object {
	return &ContainerLogList{}
}

type ContainerLogList struct {
	v1alpha1.ContainerLogList
	Items []ContainerLog `json:"items"`
}
