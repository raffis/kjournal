package main

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
	k8sversion "k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
)

//type versionFlags struct {
//	client bool
//	output string
//}

//var versionArgs versionFlags

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Long:  "The version command prints the cli version as well as attempts to print the kjournal-apiserver version",
	RunE:  versionCmdRun,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func versionCmdRun(cmd *cobra.Command, args []string) error {
	clientSemantic, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	clientVersion := k8sversion.Info{
		Major:        strconv.Itoa(int(clientSemantic.Major())),
		Minor:        strconv.Itoa(int(clientSemantic.Minor())),
		GitVersion:   version,
		GitCommit:    commit,
		GitTreeState: "clean",
		BuildDate:    date,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
	fmt.Printf("Client Version: %#v\n", clientVersion)

	cfg, err := KubeConfig(kubeconfigArgs, kubeclientOptions)
	if err != nil {
		return err
	}

	cfg.APIPath = ""

	client, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}

	srvVersion, err := client.ServerVersion()
	if err != nil {
		return err
	}

	fmt.Printf("Server Version: %#v\n", *srvVersion)
	return nil
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
