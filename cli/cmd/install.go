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
	"fmt"
	"os"

	"github.com/raffis/kjournal/cli/pkg/manifestgen"
	"github.com/raffis/kjournal/cli/pkg/manifestgen/install"
	"github.com/spf13/cobra"
)

type installFlags struct {
	componentsOnly     bool
	withCertmanager    bool
	withConfigTemplate string
	export             bool
	version            string
	manifestsPath      string
}

var installArgs installFlags

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
	installCmd.PersistentFlags().StringVarP(&installArgs.withConfigTemplate, "with-config-template", "", "", "skip the header when printing the results")
	installCmd.PersistentFlags().BoolVarP(&installArgs.withCertmanager, "with-certmanager", "", false, "skip the header when printing the results")
	installCmd.PersistentFlags().BoolVar(&installArgs.export, "export", false,
		"write the install manifests to stdout and exit")
	installCmd.PersistentFlags().StringVarP(&installArgs.version, "version", "v", "",
		"toolkit version, when specified the manifests are downloaded from https://github.com/fluxcd/flux2/releases")

	rootCmd.AddCommand(installCmd)
}

func installCmdRun(cmd *cobra.Command, args []string) error {
	/*ctx, cancel := context.WithTimeout(context.Background(), rootArgs.timeout)
	defer cancel()
	*/
	/*components := append(installArgs.defaultComponents, installArgs.extraComponents...)
	err := utils.ValidateComponents(components)
	if err != nil {
		return err
	}*/

	/*	if ver, err := getVersion(installArgs.version); err != nil {
			return err
		} else {
			installArgs.version = ver
		}
	*/
	if !installArgs.export {
		logger.Generatef("generating manifests")
	}

	tmpDir, err := manifestgen.MkdirTempAbs("", *kubeconfigArgs.Namespace)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	manifestsBase := ""
	/*if isEmbeddedVersion(installArgs.version) {
		if err := writeEmbeddedManifests(tmpDir); err != nil {
			return err
		}
		manifestsBase = tmpDir
	}*/

	opts := install.Options{
		//		BaseURL:                installArgs.manifestsPath,
		Version:   installArgs.version,
		Namespace: *kubeconfigArgs.Namespace,
		//		Components:             components,
		//		Registry:               installArgs.registry,
		//		ImagePullSecret:        installArgs.imagePullSecret,
		//		WatchAllNamespaces:     installArgs.watchAllNamespaces,
		//		NetworkPolicy:          installArgs.networkPolicy,
		//		LogLevel:               installArgs.logLevel.String(),
		//		NotificationController: rootArgs.defaults.NotificationController,
		ManifestFile: fmt.Sprintf("%s.yaml", *kubeconfigArgs.Namespace),
		Timeout:      rootArgs.timeout,
		//		ClusterDomain:          installArgs.clusterDomain,
		//s		TolerationKeys:         installArgs.tolerationKeys,
	}

	if installArgs.manifestsPath == "" {
		opts.BaseURL = install.MakeDefaultOptions().BaseURL
	}

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

	/*
		applyOutput, err := utils.Apply(ctx, kubeconfigArgs, kubeclientOptions, tmpDir, filepath.Join(tmpDir, manifest.Path))
		if err != nil {
			return fmt.Errorf("install failed: %w", err)
		}

		fmt.Fprintln(os.Stderr, applyOutput)

		kubeConfig, err := utils.KubeConfig(kubeconfigArgs, kubeclientOptions)
		if err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
		statusChecker, err := status.NewStatusChecker(kubeConfig, 5*time.Second, rootArgs.timeout, logger)
		if err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
		componentRefs, err := buildComponentObjectRefs(components...)
		if err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
		logger.Waitingf("verifying installation")
		if err := statusChecker.Assess(componentRefs...); err != nil {
			return fmt.Errorf("install failed")
		}*/

	logger.Successf("install finished")
	return nil
}
