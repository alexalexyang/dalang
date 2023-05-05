package rke2

import (
	_ "embed"
	"fmt"
	"log"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GetKubeConfig(ctx *pulumi.Context, connection *remote.ConnectionArgs, dependsOn *remote.Command) (*string, error) {
	log.Println("Get RKE2 kubeconfig file")

	kcRes, err := remote.NewCommand(ctx, "get-rke2-kubeconfig-file", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String("cat /etc/rancher/rke2/rke2.yaml"),
	}, pulumi.DependsOn([]pulumi.Resource{dependsOn}))
	if err != nil {
		return nil, err
	}

	kcChan := make(chan string)

	kcRes.Stdout.ApplyT(func(kubeconfigFile string) string {
		kcChan <- kubeconfigFile
		return kubeconfigFile
	})

	kubeconfig := <-kcChan
	close(kcChan)

	return &kubeconfig, nil
}

func UpdateKubeConfigServerIP(kubeconfig string, serverIP string) string {
	return strings.Replace(kubeconfig, "server: https://127.0.0.1:", fmt.Sprintf("server: https://%s:", serverIP), 1)
}
