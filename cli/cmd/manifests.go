package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
)

//go:embed config/base
var embeddedManifests embed.FS

func writeEmbeddedManifests(dir string) error {
	return fs.WalkDir(embeddedManifests, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Printf("read: %v\n", p)
		if d.IsDir() {
			return nil
		}

		data, err := fs.ReadFile(embeddedManifests, p)
		if err != nil {
			return fmt.Errorf("reading file failed: %w", err)
		}

		parent := path.Join(dir, path.Dir(p))
		fmt.Printf("create folder %#v\n", parent)
		if _, err := os.Stat(parent); os.IsNotExist(err) {
			os.MkdirAll(parent, 0750)
		}

		err = os.WriteFile(path.Join(dir, p), data, 0666)
		if err != nil {
			return fmt.Errorf("writing file failed: %w", err)
		}

		return nil
	})

}
