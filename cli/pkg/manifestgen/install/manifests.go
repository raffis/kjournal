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
	"fmt"
	"os"
	"path"

	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/raffis/kjournal/cli/pkg/manifestgen/kustomization"
)

/*
func fetch(ctx context.Context, url, version, dir string) error {
	ghURL := fmt.Sprintf("%s/latest/download/manifests.tar.gz", url)
	if strings.HasPrefix(version, "v") {
		ghURL = fmt.Sprintf("%s/download/%s/manifests.tar.gz", url, version)
	}

	req, err := http.NewRequest("GET", ghURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request for %s, error: %w", ghURL, err)
	}

	// download
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to download manifests.tar.gz from %s, error: %w", ghURL, err)
	}
	defer resp.Body.Close()

	// check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download manifests.tar.gz from %s, status: %s", ghURL, resp.Status)
	}

	// extract
	if _, err = untar.Untar(resp.Body, dir); err != nil {
		return fmt.Errorf("failed to untar manifests.tar.gz from %s, error: %w", ghURL, err)
	}

	return nil
}*/

func generate(base string, options Options) error {
	if err := execTemplate(options, namespaceTmpl, path.Join(base, "namespace.yaml")); err != nil {
		return fmt.Errorf("generate namespace failed: %w", err)
	}

	if err := execTemplate(options, labelsTmpl, path.Join(base, "labels.yaml")); err != nil {
		return fmt.Errorf("generate labels failed: %w", err)
	}

	if err := execTemplate(options, kustomizationTmpl, path.Join(base, "kustomization.yaml")); err != nil {
		return fmt.Errorf("generate kustomization failed: %w", err)
	}

	if err := os.MkdirAll(path.Join(base, "roles"), os.ModePerm); err != nil {
		return fmt.Errorf("generate roles failed: %w", err)
	}

	return nil
}

func build(base, output string) error {
	resources, err := kustomization.Build(base)
	if err != nil {
		return err
	}

	//outputBase := filepath.Dir(strings.TrimSuffix(output, string(filepath.Separator)))
	fs := filesys.MakeFsOnDisk()
	if err != nil {
		return err
	}
	if err = fs.WriteFile(output, resources); err != nil {
		return err
	}

	return nil
}
