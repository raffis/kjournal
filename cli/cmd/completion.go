package main

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates completion scripts for various shells",
	Long:  "The completion sub-command generates completion scripts for various shells",
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func contextsCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	rawConfig, err := kubeconfigArgs.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return completionError(err)
	}

	var comps []string

	for name := range rawConfig.Contexts {
		if strings.HasPrefix(name, toComplete) {
			comps = append(comps, name)
		}
	}

	return comps, cobra.ShellCompDirectiveNoFileComp
}

func resourceNamesCompletionFunc(gvk schema.GroupVersionKind) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ctx, cancel := context.WithTimeout(context.Background(), rootArgs.timeout)
		defer cancel()

		cfg, err := KubeConfig(kubeconfigArgs, kubeclientOptions)
		if err != nil {
			return completionError(err)
		}

		mapper, err := kubeconfigArgs.ToRESTMapper()
		if err != nil {
			return completionError(err)
		}

		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return completionError(err)
		}

		client, err := dynamic.NewForConfig(cfg)
		if err != nil {
			return completionError(err)
		}

		var dr dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			dr = client.Resource(mapping.Resource).Namespace(*kubeconfigArgs.Namespace)
		} else {
			dr = client.Resource(mapping.Resource)
		}

		list, err := dr.List(ctx, metav1.ListOptions{})
		if err != nil {
			return completionError(err)
		}

		var comps []string

		for _, item := range list.Items {
			name := item.GetName()

			if strings.HasPrefix(name, toComplete) {
				comps = append(comps, name)
			}
		}

		return comps, cobra.ShellCompDirectiveNoFileComp
	}
}

func completionError(err error) ([]string, cobra.ShellCompDirective) {
	cobra.CompError(err.Error())
	return nil, cobra.ShellCompDirectiveError
}
