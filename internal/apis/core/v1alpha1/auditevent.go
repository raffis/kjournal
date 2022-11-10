package v1alpha1

import (
	"context"
	"encoding/json"

	"github.com/raffis/kjournal/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AuditEventList struct {
	v1alpha1.AuditEventList
	Items []AuditEvent `json:"items"`
}

type AuditEvent struct {
	v1alpha1.AuditEvent `json:",inline"`
}

func (in *AuditEvent) UnmarshalJSON(bs []byte) error {
	return json.Unmarshal(bs, &in.AuditEvent.Event)
}

func (in *AuditEvent) New() runtime.Object {
	return &AuditEvent{
		AuditEvent: v1alpha1.AuditEvent{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AuditEvent",
				APIVersion: "core.kjournal/v1alpha1",
			},
		},
	}
}

func (in *AuditEvent) NewList() runtime.Object {
	return &AuditEventList{}
}

var auditEventTableColums = []metav1.TableColumnDefinition{
	{Name: "REVEIVED", Type: "", Format: "name", Description: "s"},
	{Name: "VERB", Type: "string", Format: "name", Description: "s"},
	{Name: "STATUS", Type: "string", Description: "The reference to the service that hosts this API endpoint."},
	{Name: "USER", Type: "string", Description: "Whether this service is available."},
}

func (in *AuditEvent) asCells() []interface{} {
	return []interface{}{in.RequestReceivedTimestamp, in.Verb, in.ResponseStatus.Code, in.User.Username}
}

// ConvertToTable implements the TableConvertor interface for REST.
func (in *AuditEvent) ConvertToTable(ctx context.Context, tableOptions runtime.Object) (*metav1.Table, error) {
	table := &metav1.Table{
		ColumnDefinitions: auditEventTableColums,
		TypeMeta:          in.TypeMeta,
	}

	rows := make([]metav1.TableRow, 0, 1)
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: in},
		Cells:  in.asCells(),
	}

	rows = append(rows, row)
	table.Rows = rows
	return table, nil
}

// ConvertToTable implements the TableConvertor interface for REST.
func (in *AuditEventList) ConvertToTable(ctx context.Context, tableOptions runtime.Object) (*metav1.Table, error) {
	table := &metav1.Table{
		ColumnDefinitions: auditEventTableColums,
		TypeMeta:          in.TypeMeta,
	}

	rows := make([]metav1.TableRow, 0, 1)

	for _, v := range in.Items {
		row := metav1.TableRow{
			Object: runtime.RawExtension{Object: &v},
			Cells:  v.asCells(),
		}
		rows = append(rows, row)
	}

	table.Rows = rows
	return table, nil
}
