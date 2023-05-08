package rke2

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

// Has no tests
func GetClientSet(kubeconfig string) (*kubernetes.Clientset, error) {

	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		log.Println("Error creating kubeconfig from string: ", err)
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("Error creating clientset: ", err)
		return nil, err
	}

	return clientset, nil
}

func ConvertYamlToObj(yamlString string, k8sObjType interface{}) (interface{}, error) {
	decoder := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(yamlString)), 1000)

	if err := decoder.Decode(&k8sObjType); err != nil {
		log.Println("Error decoding yaml file: ", err)
		return nil, err
	}

	return k8sObjType, nil
}
