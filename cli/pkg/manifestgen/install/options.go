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
	ServiceMonitor  bool
	ConfigTemplate  string
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
		ConfigTemplate:  "",
		NetworkPolicy:   true,
		CertManager:     false,
		ServiceMonitor:  false,
		AsKustomization: false,
		LogLevel:        "info",
		BaseURL:         "github.com/raffis/kjournal//config",
		ManifestFile:    "kjournal.yaml",
		Timeout:         time.Minute,
		TargetPath:      "",
		ClusterDomain:   "cluster.local",
	}
}
