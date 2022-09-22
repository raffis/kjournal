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
