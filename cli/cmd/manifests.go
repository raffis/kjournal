package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"

	"k8s.io/klog/v2"
)

//go:embed config/base
var embeddedManifests embed.FS

func writeEmbeddedManifests(dir string) error {
	return fs.WalkDir(embeddedManifests, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		klog.V(5).Info("walk embedded manifests", "path", p)
		if d.IsDir() {
			return nil
		}

		data, err := fs.ReadFile(embeddedManifests, p)
		if err != nil {
			return fmt.Errorf("reading file failed: %w", err)
		}

		parent := path.Join(dir, path.Dir(p))
		klog.V(5).Info("create folder", "path", parent)

		if _, err := os.Stat(parent); os.IsNotExist(err) {
			if err := os.MkdirAll(parent, 0750); err != nil {
				return err
			}
		}

		err = os.WriteFile(path.Join(dir, p), data, 0666)
		if err != nil {
			return fmt.Errorf("writing file failed: %w", err)
		}

		return nil
	})

}
