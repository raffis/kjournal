/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

commonLabels:
  app.kubernetes.io/instance: {{.Namespace}}
  app.kubernetes.io/version: "{{.Version}}"
  app.kubernetes.io/part-of: kjournal

resources:
- ssh://git@github.com/raffis/kjournal//config/namespace
- ssh://git@github.com/raffis/kjournal//config/apiserver
- ssh://git@github.com/raffis/kjournal//config/apiservice
- ssh://git@github.com/raffis/kjournal//config/rbac
{{- if .NetworkPolicy }}
- ssh://git@github.com/raffis/kjournal//config/policies
{{- end }}

components: 
- ssh://git@github.com/raffis/kjournal//config/certmanager

`
var labelsTmpl = `---
apiVersion: builtin
kind: LabelTransformer
metadata:
  name: labels
labels:
fieldSpecs:
  - path: metadata/labels
    create: true
`

var namespaceTmpl = `---
apiVersion: v1
kind: Namespace
metadata:
  name: {{.Namespace}}
  labels:
    pod-security.kubernetes.io/warn: restricted
    pod-security.kubernetes.io/warn-version: latest
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

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
