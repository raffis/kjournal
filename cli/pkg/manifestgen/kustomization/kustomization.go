package kustomization

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/raffis/kjournal/cli/pkg/manifestgen"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/provider"
	kustypes "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/yaml"
)

// Generate scans the given directory for Kubernetes manifests and creates a
// konfig.DefaultKustomizationFileName file, including all discovered manifests
// as resources.
func Generate(options Options) (*manifestgen.Manifest, error) {
	kfile := filepath.Join(options.TargetPath, konfig.DefaultKustomizationFileName())
	abskfile := filepath.Join(options.BaseDir, kfile)

	scan := func(base string) ([]string, error) {
		var paths []string
		pvd := provider.NewDefaultDepProvider()
		rf := pvd.GetResourceFactory()
		err := options.FileSystem.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if path == base {
				return nil
			}
			if info.IsDir() {
				// If a sub-directory contains an existing Kustomization file, add the
				// directory as a resource and do not decent into it.
				for _, kfilename := range konfig.RecognizedKustomizationFileNames() {
					if options.FileSystem.Exists(filepath.Join(path, kfilename)) {
						paths = append(paths, path)
						return filepath.SkipDir
					}
				}
				return nil
			}
			fContents, err := options.FileSystem.ReadFile(path)
			if err != nil {
				return err
			}
			if _, err := rf.SliceFromBytes(fContents); err != nil {
				return nil
			}
			paths = append(paths, path)
			return nil
		})
		return paths, err
	}

	if _, err := os.Stat(abskfile); err != nil {
		abs, err := filepath.Abs(filepath.Dir(abskfile))
		if err != nil {
			return nil, err
		}

		files, err := scan(abs)
		if err != nil {
			return nil, err
		}

		f, err := options.FileSystem.Create(abskfile)
		if err != nil {
			return nil, err
		}
		if err = f.Close(); err != nil {
			return nil, err
		}

		kus := kustypes.Kustomization{
			TypeMeta: kustypes.TypeMeta{
				APIVersion: kustypes.KustomizationVersion,
				Kind:       kustypes.KustomizationKind,
			},
		}

		var resources []string
		for _, file := range files {
			relP, err := filepath.Rel(abs, file)
			if err != nil {
				return nil, err
			}
			resources = append(resources, relP)
		}

		kus.Resources = resources
		kd, err := yaml.Marshal(kus)
		if err != nil {
			return nil, err
		}

		return &manifestgen.Manifest{
			Path:    kfile,
			Content: string(kd),
		}, nil
	}

	kd, err := os.ReadFile(abskfile)
	if err != nil {
		return nil, err
	}
	return &manifestgen.Manifest{
		Path:    kfile,
		Content: string(kd),
	}, nil
}

// kustomizeBuildMutex is a workaround for a concurrent map read and map write bug.
// TODO(stefan): https://github.com/kubernetes-sigs/kustomize/issues/3659
var kustomizeBuildMutex sync.Mutex

// Build takes the path to a directory with a konfig.RecognizedKustomizationFileNames,
// builds it, and returns the resulting manifests as multi-doc YAML. It restricts the
// Kustomize file system to the parent directory of the base.
func Build(base string) ([]byte, error) {
	// TODO(hidde): drop this when consumers have moved away to BuildWithRoot.
	parent := filepath.Dir(strings.TrimSuffix(base, string(filepath.Separator)))
	return BuildWithRoot(parent, base)
}

// BuildWithRoot takes the path to a directory with a konfig.RecognizedKustomizationFileNames,
// builds it, and returns the resulting manifests as multi-doc YAML.
// The Kustomize file system is restricted to root.
func BuildWithRoot(root, base string) ([]byte, error) {
	kustomizeBuildMutex.Lock()
	defer kustomizeBuildMutex.Unlock()

	fs := filesys.MakeFsOnDisk()

	var kfile string
	for _, f := range konfig.RecognizedKustomizationFileNames() {
		if kf := filepath.Join(base, f); fs.Exists(kf) {
			kfile = kf
			break
		}
	}
	if kfile == "" {
		return nil, fmt.Errorf("%s not found", konfig.DefaultKustomizationFileName())
	}

	// TODO(hidde): work around for a bug in kustomize causing it to
	//  not properly handle absolute paths on Windows.
	//  Convert the path to a relative path to the working directory
	//  as a temporary fix:
	//  https://github.com/kubernetes-sigs/kustomize/issues/2789
	if filepath.IsAbs(base) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		base, err = filepath.Rel(wd, base)
		if err != nil {
			return nil, err
		}
	}

	buildOptions := &krusty.Options{
		DoLegacyResourceSort: true,
		LoadRestrictions:     kustypes.LoadRestrictionsNone,
		AddManagedbyLabel:    false,
		DoPrune:              false,
		PluginConfig:         kustypes.DisabledPluginConfig(),
	}

	k := krusty.MakeKustomizer(buildOptions)
	m, err := k.Run(fs, base)
	if err != nil {
		return nil, err
	}

	resources, err := m.AsYaml()
	if err != nil {
		return nil, err
	}

	return resources, nil
}
