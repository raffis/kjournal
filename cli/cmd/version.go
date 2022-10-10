package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type versionFlags struct {
	client bool
	output string
}

var versionArgs versionFlags

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Long:  "The version command prints the cli version",
	//ValidArgsFunction: resourceNamesCompletionFunc(logsv1beta1.GroupVersion.WithKind(logsv1beta1.LogKind)),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf(`{"version":"%s","sha":"%s","date":"%s"}`+"\n", version, commit, date)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

/*
func versionCmdRun(cmd *cobra.Command, args []string) error {
	if versionArgs.output != "yaml" && versionArgs.output != "json" {
		return fmt.Errorf("--output must be json or yaml, not %s", versionArgs.output)
	}

	ctx, cancel := context.WithTimeout(context.Background(), rootArgs.timeout)
	defer cancel()

	info := map[string]string{}
	info["flux"] = rootArgs.defaults.Version

	if !versionArgs.client {
		kubeClient, err := utils.KubeClient(kubeconfigArgs, kubeclientOptions)
		if err != nil {
			return err
		}

		selector := client.MatchingLabels{manifestgen.PartOfLabelKey: manifestgen.PartOfLabelValue}
		var list v1.DeploymentList
		if err := kubeClient.List(ctx, &list, client.InNamespace(*kubeconfigArgs.Namespace), selector); err != nil {
			return err
		}

		if len(list.Items) == 0 {
			return fmt.Errorf("no deployments found in %s namespace", *kubeconfigArgs.Namespace)
		}

		for _, d := range list.Items {
			for _, c := range d.Spec.Template.Spec.Containers {
				name, tag, err := splitImageStr(c.Image)
				if err != nil {
					return err
				}
				info[name] = tag
			}
		}
	}

	var marshalled []byte
	var err error

	if versionArgs.output == "json" {
		marshalled, err = json.MarshalIndent(&info, "", "  ")
		marshalled = append(marshalled, "\n"...)
	} else {
		marshalled, err = yaml.Marshal(&info)
	}

	if err != nil {
		return err
	}

	rootCmd.Print(string(marshalled))
	return nil
}

func splitImageStr(image string) (string, string, error) {
	imageArr := strings.Split(image, ":")
	if len(imageArr) < 2 {
		return "", "", fmt.Errorf("missing image tag in image %s", image)
	}

	name, tag := imageArr[0], imageArr[1]
	nameArr := strings.Split(name, "/")
	return nameArr[len(nameArr)-1], tag, nil
}
*/
