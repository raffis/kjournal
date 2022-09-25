package install

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"text/template"
)

var kustomizationTmpl = `---
{{- $registry := .Registry }}
{{- $logLevel := .LogLevel }}
{{- $clusterDomain := .ClusterDomain }}
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {{.Namespace}}
namePrefix: kjournal-

labels:
- pairs:
    app.kubernetes.io/instance: {{.Namespace}}
    app.kubernetes.io/version: "{{.Version}}"
    app.kubernetes.io/part-of: kjournal

resources:
- {{.BaseURL}}/namespace
- {{.BaseURL}}/apiserver
- {{.BaseURL}}/rbac
{{- if .ServiceMonitor }}
#- {{.BaseURL}}/components/prometheus
{{- end }}
{{- if .NetworkPolicy }}
#- {{.BaseURL}}/policies
{{- end }}

components: 
{{- if .ConfigTemplate }}
- {{.BaseURL}}/components/config-templates/{{.ConfigTemplate}}
{{- end }}
{{- if .CertManager }}
- {{.BaseURL}}/components/certmanager

patchesStrategicMerge:
- |
  apiVersion: cert-manager.io/v1
  kind: Certificate
  metadata:
    name: apiserver
    annotations:
      kjournal/cluster-domain: {{.ClusterDomain}}
{{- end }}
`

func execTemplate(obj interface{}, tmpl, filename string) error {
	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		return err
	}

	var data bytes.Buffer
	writer := bufio.NewWriter(&data)
	if err := t.Execute(writer, obj); err != nil {
		return err
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, data.String())
	if err != nil {
		return err
	}

	return file.Sync()
}
