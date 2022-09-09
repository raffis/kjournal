package v1alpha1

import (
	"encoding/json"
	"time"

	"github.com/raffis/kjournal/pkg/apis/core/v1alpha1"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/runtime"
)

type ContainerLog struct {
	v1alpha1.ContainerLog
	Payload  json.RawMessage `json:"payload"`
	fieldMap map[string]string
}

func (in *ContainerLog) UnmarshalJSON(bs []byte) (err error) {
	for k, v := range in.fieldMap {
		switch k {
		case "metadata.namespace":
			in.Namespace = gjson.Get(string(bs), v).Str
		case "metadata.creationTimestamp":
			ts, err := time.Parse(time.RFC3339, gjson.Get(string(bs), v).Str)
			if err != nil {
				return err
			}

			in.CreationTimestamp.Time = ts
		case "pod":
			in.Pod = gjson.Get(string(bs), v).Str
		case "container":
			in.Container = gjson.Get(string(bs), v).Str
		case "payload":
			if v == "." || v == "" {
				in.Payload = bs
			} else {
				in.Payload = json.RawMessage(gjson.Get(string(bs), v).Raw)
			}
		}
	}

	in.TypeMeta.Kind = "ContainerLog"
	in.TypeMeta.APIVersion = "core.kjournal/v1alpha1"

	return nil
}

func (in *ContainerLog) WithFieldMap(fieldMap map[string]string) {
	in.fieldMap = fieldMap
}

func (in *ContainerLog) New() runtime.Object {
	return &ContainerLog{
		fieldMap: in.fieldMap,
	}
}

func (in *ContainerLog) NewList() runtime.Object {
	return &ContainerLogList{}
}

type ContainerLogList struct {
	v1alpha1.ContainerLogList
	Items []ContainerLog `json:"items"`
}
