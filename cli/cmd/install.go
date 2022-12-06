package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/raffis/kjournal/cli/internal/utils"
	"github.com/raffis/kjournal/cli/pkg/manifestgen"
	"github.com/raffis/kjournal/cli/pkg/manifestgen/install"
	"github.com/raffis/kjournal/cli/pkg/status"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cli-utils/pkg/object"
)

type installFlags struct {
	withCertManager     bool
	withNetworkPolicies bool
	withConfigTemplate  string
	withServiceMonitor  bool
	export              bool
	version             string
	base                string
	asKustomization     bool
	registry            string
	imagePullSecret     string
}

var installArgs installFlags
var defaults install.Options

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Deploy the kournal apiserver to kubernetes",
	Long:  "Deploys the kjournal to kubernetes",
	Example: `  # Deploy the apiserver to kjournal-system
  kjournal install -n kjournal-system
  
  # Output manifests instead deploying
  kjournal install -n kjournal-system --export
  
  # Deploy with certmanager and prometheus operator support
  kjournal install -n kjournal-system --with-certmanager --with-servicemonitor
  
  # Specify a specific version
  kjournal install -n kjournal-system --version=v0.0.1`,
	RunE: installCmdRun,
}

func init() {
	defaults = install.MakeDefaultOptions()

	installCmd.PersistentFlags().StringVarP(&installArgs.withConfigTemplate, "with-config-template", "", defaults.ConfigTemplate,
		"specify a kjournal config template")
	installCmd.PersistentFlags().BoolVarP(&installArgs.withCertManager, "with-certmanager", "", defaults.CertManager,
		"Enable certmanager support (recomended option)")
	installCmd.PersistentFlags().BoolVarP(&installArgs.withServiceMonitor, "with-servicemonitor", "", defaults.ServiceMonitor,
		"Enable prometheus-operator support (Deploys a ServiceMonitor)")
	installCmd.PersistentFlags().BoolVarP(&installArgs.export, "export", "", false,
		"write the install manifests to stdout and exit")
	installCmd.PersistentFlags().StringVarP(&installArgs.version, "version", "", version,
		"specify a specific kjournal version to install (by default the latest version is used)")
	installCmd.PersistentFlags().BoolVarP(&installArgs.asKustomization, "as-kustomization", "k", defaults.AsKustomization,
		"Print kustomization to stdout and exit")
	installCmd.Flags().StringVar(&installArgs.registry, "registry", defaults.Registry,
		"container registry where the toolkit images are published")
	installCmd.Flags().StringVar(&installArgs.imagePullSecret, "image-pull-secret", defaults.ImagePullSecret,
		"Kubernetes secret name used for pulling the toolkit images from a private registry")
	installCmd.Flags().BoolVar(&installArgs.withNetworkPolicies, "network-policy", defaults.NetworkPolicy,
		"deny ingress access to the toolkit controllers from other namespaces using network policies")
	installCmd.Flags().StringVarP(&installArgs.base, "base", "", defaults.Base,
		"Path or URL to kustomize base")

	rootCmd.AddCommand(installCmd)
}

func installCmdRun(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), rootArgs.timeout)
	defer cancel()

	tmpDir, err := manifestgen.MkdirTempAbs("", *kubeconfigArgs.Namespace)
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	installVersion := installArgs.version
	if installVersion == "latest" {
		latest, err := install.GetLatestVersion()
		if err != nil {
			return err
		}

		installVersion = latest
	}

	manifestsBase := ""
	if installVersion == version || installVersion == "" {
		//manifestsBase = "./"
		err := writeEmbeddedManifests(tmpDir)
		if err != nil {
			return fmt.Errorf("install failed : %w", err)
		}

		installArgs.base = "./config/base"
		manifestsBase = tmpDir
	}

	opts := defaults
	opts.Base = installArgs.base
	opts.AsKustomization = installArgs.asKustomization
	opts.Version = installVersion
	opts.Namespace = *kubeconfigArgs.Namespace
	opts.Registry = installArgs.registry
	opts.ImagePullSecret = installArgs.imagePullSecret
	opts.NetworkPolicy = installArgs.withNetworkPolicies
	opts.CertManager = installArgs.withCertManager
	opts.ServiceMonitor = installArgs.withServiceMonitor
	opts.ConfigTemplate = installArgs.withConfigTemplate
	opts.ManifestFile = fmt.Sprintf("%s.yaml", *kubeconfigArgs.Namespace)

	if opts.Namespace == "" {
		opts.Namespace = defaults.Namespace
	}

	/*if installArgs.manifestsPath == "" {
		opts.BaseURL = install.MakeDefaultOptions().BaseURL
	}*/

	manifest, err := install.Generate(opts, manifestsBase)
	if err != nil {
		return fmt.Errorf("install failed : %w", err)
	}

	if _, err := manifest.WriteFile(tmpDir); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	if installArgs.export {
		fmt.Print(manifest.Content)
		return nil
	} else {
		klog.V(2).InfoS("build manifests", "manifests", manifest.Content)
	}

	logger.Successf("manifests build completed")
	logger.Infof("installing components in %s namespace", opts.Namespace)

	kubeConfig, err := kubeconfigArgs.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	_, err = utils.Apply(ctx, kubeConfig, tmpDir, filepath.Join(tmpDir, manifest.Path))
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	statusChecker, err := status.NewStatusChecker(kubeConfig, 5*time.Second, rootArgs.timeout, logger)
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	apiserver := object.ObjMetadata{
		Namespace: opts.Namespace,
		Name:      "kjournal-apiserver",
		GroupKind: schema.GroupKind{Group: "apps", Kind: "Deployment"},
	}

	logger.Waitingf("verifying installation")
	if err := statusChecker.Assess(apiserver); err != nil {
		return fmt.Errorf("install failed")
	}

	logger.Successf("install finished")
	return nil
}
