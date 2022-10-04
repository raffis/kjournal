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
	Short: "Install install install",
	Long:  "The install command fetchs install from namespaced resources",
	Example: `  # Stream all install install from the namespace mynamespace
  kjournal install -n mynamespace
  
  # Stream install from the last 48 hours
  kjournal install -n mynamespace --since 48h
  
  # Stream install for all deployments
  kjournal install -n mynamespace deployments
  
  # Stream install for a pod named abc
  kjournal install -n mynamespace pods/abc`,
	//ValidArgsFunction: resourceNamesCompletionFunc(installv1.GroupVersion.WithKind(corev1alpha1.EventKind)),
	RunE: installCmdRun,
}

func init() {
	defaults = install.MakeDefaultOptions()

	installCmd.PersistentFlags().StringVarP(&installArgs.withConfigTemplate, "with-config-template", "", "",
		"specify a kjournal config template")
	installCmd.PersistentFlags().BoolVarP(&installArgs.withCertManager, "with-certmanager", "", defaults.CertManager,
		"Enable certmanager support (recomended option)")
	installCmd.PersistentFlags().BoolVarP(&installArgs.withServiceMonitor, "with-servicemonitor", "", defaults.ServiceMonitor,
		"Enable prometheus-operator support (Deploys a ServiceMonitor)")
	installCmd.PersistentFlags().BoolVarP(&installArgs.export, "export", "", false,
		"write the install manifests to stdout and exit")
	installCmd.PersistentFlags().StringVarP(&installArgs.version, "version", "", "",
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

	version := installArgs.version
	if version == "" {
		latest, err := install.GetLatestVersion()
		if err != nil {
			return err
		}

		version = latest
	}

	manifestsBase := ""
	if version == VERSION || version == "" {
		//manifestsBase = "./"
		err := writeEmbeddedManifests(tmpDir)
		if err != nil {
			return fmt.Errorf("install failed : %w", err)
		}

		installArgs.base = "./config"
		manifestsBase = tmpDir
	}

	opts := install.Options{
		Base:            installArgs.base,
		AsKustomization: installArgs.asKustomization,
		Version:         version,
		Namespace:       *kubeconfigArgs.Namespace,
		Registry:        installArgs.registry,
		ImagePullSecret: installArgs.imagePullSecret,
		NetworkPolicy:   installArgs.withNetworkPolicies,
		CertManager:     installArgs.withCertManager,
		ServiceMonitor:  installArgs.withServiceMonitor,
		ManifestFile:    fmt.Sprintf("%s.yaml", *kubeconfigArgs.Namespace),
	}

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
	} else if rootArgs.verbose {
		fmt.Print(manifest.Content)
	}

	logger.Successf("manifests build completed")
	logger.Actionf("installing components in %s namespace", *kubeconfigArgs.Namespace)

	kubeConfig, err := kubeconfigArgs.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	applyOutput, err := utils.Apply(ctx, kubeConfig, tmpDir, filepath.Join(tmpDir, manifest.Path))
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	fmt.Fprintln(os.Stderr, applyOutput)

	statusChecker, err := status.NewStatusChecker(kubeConfig, 5*time.Second, rootArgs.timeout, logger)
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	apiserver := object.ObjMetadata{
		Namespace: *kubeconfigArgs.Namespace,
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
