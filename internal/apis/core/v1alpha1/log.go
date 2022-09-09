package v1alpha1

import (
	"encoding/json"
	"time"

	"github.com/raffis/kjournal/pkg/apis/core/v1alpha1"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/runtime"
)

type Log struct {
	v1alpha1.Log
	Payload  json.RawMessage `json:"payload"`
	fieldMap map[string]string
}

func (in *Log) UnmarshalJSON(bs []byte) (err error) {
	for k, v := range in.fieldMap {
		switch k {
		case "metadata.creationTimestamp":
			ts, err := time.Parse(time.RFC3339, gjson.Get(string(bs), v).Str)
			if err != nil {
				return err
			}

			in.CreationTimestamp.Time = ts
		case "payload":
			if v == "." || v == "" {
				in.Payload = bs
			} else {
				in.Payload = json.RawMessage(gjson.Get(string(bs), v).Raw)
			}
		}
	}

	in.TypeMeta.Kind = "Log"
	in.TypeMeta.APIVersion = "core.kjournal/v1alpha1"

	return nil
}

func (in *Log) WithFieldMap(fieldMap map[string]string) {
	in.fieldMap = fieldMap
}

func (in *Log) New() runtime.Object {
	return &Log{
		fieldMap: in.fieldMap,
	}
}

func (in *Log) NewList() runtime.Object {
	return &LogList{}
}

type LogList struct {
	v1alpha1.LogList
	Items []Log `json:"items"`
}
