package kustomization

import "sigs.k8s.io/kustomize/kyaml/filesys"

type Options struct {
	FileSystem filesys.FileSystem
	BaseDir    string
	TargetPath string
}

func MakeDefaultOptions() Options {
	return Options{
		FileSystem: filesys.MakeFsOnDisk(),
		BaseDir:    "",
		TargetPath: "",
	}
}
