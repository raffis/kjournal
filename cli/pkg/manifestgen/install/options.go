package install

import "time"

type Options struct {
	AsKustomization bool
	Base            string
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
		ConfigTemplate:  "elasticsearch-kjournal-structured",
		NetworkPolicy:   true,
		CertManager:     false,
		ServiceMonitor:  false,
		AsKustomization: false,
		LogLevel:        "info",
		Base:            "github.com/raffis/kjournal//config",
		ManifestFile:    "kjournal.yaml",
		Timeout:         time.Minute,
		TargetPath:      "",
		ClusterDomain:   "cluster.local",
	}
}
