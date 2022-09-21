package utils

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/raffis/kjournal/cli/pkg/manifestgen/kustomization"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/kustomize/api/konfig"
)

func Apply(ctx context.Context, cfg *rest.Config, root, manifestPath string) (objects []*unstructured.Unstructured, err error) {
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return objects, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return objects, err
	}

	objects, err = readObjects(root, manifestPath)
	if err != nil {
		return objects, err
	}

	for _, obj := range objects {
		mapping, err := mapper.RESTMapping(obj.GroupVersionKind().GroupKind(), obj.GroupVersionKind().Version)
		if err != nil {
			return objects, err
		}

		var dr dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			// namespaced resources should specify the namespace
			dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
		} else {
			// for cluster-wide resources
			dr = dyn.Resource(mapping.Resource)
		}

		data, err := json.Marshal(obj)
		if err != nil {
			return objects, err
		}

		_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
			FieldManager: "kjournal",
		})

		if err != nil {
			return objects, err
		}
	}

	return objects, nil
}

func readObjects(root, manifestPath string) ([]*unstructured.Unstructured, error) {
	fi, err := os.Lstat(manifestPath)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() || !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("expected %q to be a file", manifestPath)
	}

	if isRecognizedKustomizationFile(manifestPath) {
		resources, err := kustomization.BuildWithRoot(root, filepath.Dir(manifestPath))
		if err != nil {
			return nil, err
		}
		return ReadObjects(bytes.NewReader(resources))
	}

	ms, err := os.Open(manifestPath)
	if err != nil {
		return nil, err
	}
	defer ms.Close()

	return ReadObjects(bufio.NewReader(ms))
}

func isRecognizedKustomizationFile(path string) bool {
	base := filepath.Base(path)
	for _, v := range konfig.RecognizedKustomizationFileNames() {
		if base == v {
			return true
		}
	}
	return false
}

// ReadObjects decodes the YAML or JSON documents from the given reader into unstructured Kubernetes API objects.
// The documents which do not subscribe to the Kubernetes Object interface, are silently dropped from the result.
func ReadObjects(r io.Reader) ([]*unstructured.Unstructured, error) {
	reader := k8syaml.NewYAMLOrJSONDecoder(r, 2048)
	objects := make([]*unstructured.Unstructured, 0)

	for {
		obj := &unstructured.Unstructured{}
		err := reader.Decode(obj)
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return objects, err
		}

		if obj.IsList() {
			err = obj.EachListItem(func(item runtime.Object) error {
				obj := item.(*unstructured.Unstructured)
				objects = append(objects, obj)
				return nil
			})
			if err != nil {
				return objects, err
			}
			continue
		}

		if IsKubernetesObject(obj) && !IsKustomization(obj) {
			objects = append(objects, obj)
		}
	}

	return objects, nil
}

// IsKubernetesObject checks if the given object has the minimum required fields to be a Kubernetes object.
func IsKubernetesObject(object *unstructured.Unstructured) bool {
	if object.GetName() == "" || object.GetKind() == "" || object.GetAPIVersion() == "" {
		return false
	}
	return true
}

// IsKustomization checks if the given object is a Kustomize config.
func IsKustomization(object *unstructured.Unstructured) bool {
	if object.GetKind() == "Kustomization" && object.GroupVersionKind().GroupKind().Group == "kustomize.config.k8s.io" {
		return true
	}
	return false
}
