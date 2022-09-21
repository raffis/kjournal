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
	"path"

	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/raffis/kjournal/cli/pkg/manifestgen/kustomization"
)

func generate(base string, options Options) error {
	if err := execTemplate(options, kustomizationTmpl, path.Join(base, "kustomization.yaml")); err != nil {
		return fmt.Errorf("generate kustomization failed: %w", err)
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
