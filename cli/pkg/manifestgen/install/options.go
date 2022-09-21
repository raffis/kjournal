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

import "time"

type Options struct {
	AsKustomization bool
	BaseURL         string
	Version         string
	Namespace       string
	Registry        string
	ImagePullSecret string
	NetworkPolicy   bool
	CertManager     bool
	LogLevel        string
	ManifestFile    string
	Timeout         time.Duration
	TargetPath      string
	ClusterDomain   string
}

func MakeDefaultOptions() Options {
	return Options{
		Version:         "latest",
		Namespace:       "kjournal-system",
		Registry:        "ghcr.io/raffis/kjournal",
		ImagePullSecret: "",
		NetworkPolicy:   true,
		CertManager:     false,
		AsKustomization: false,
		LogLevel:        "info",
		BaseURL:         "github.com/raffis/kjournal",
		ManifestFile:    "kjournal.yaml",
		Timeout:         time.Minute,
		TargetPath:      "",
		ClusterDomain:   "cluster.local",
	}
}
