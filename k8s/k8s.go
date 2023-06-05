package k8s

import (
	"log"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	// "helm.sh/helm/v3/pkg/kube"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Installs a helm chart from a tgz archive. E.g.: nginx ingress controller.
func InstallChart(dirPath string, release string, namespace string, kubeconfig *string) error {
	log.Println("Loading helm charts in dir: ", dirPath)
	chart, err := loader.Load(dirPath)
	if err != nil {
		log.Println("Error loading helm charts: ", err)
		return err
	}

	actionConfig := new(action.Configuration)

	cwd, _ := os.Getwd()

	kubeconfigPath := filepath.Join(cwd, "kubeconfig.dev.yaml")

	log.Println("Initialising action configuration")
	if err := actionConfig.Init(
		&genericclioptions.ConfigFlags{
			Namespace: &namespace,

			// The docs don't say it but this is the path to the kubeconfig file, not the file as a string.
			KubeConfig: &kubeconfigPath,
		},

		// This might work too.
		// kube.GetConfig(kubeconfigPath, "", namespace),
		namespace,

		// No idea what helm driver does. Docs don't say.
		"configmap",
		log.Printf,
	); err != nil {
		log.Println("Error initialising the action configuration: ", err)
		return err
	}

	log.Println("Creating helm install client")
	client := action.NewInstall(actionConfig)
	client.ReleaseName = release
	client.Namespace = namespace
	client.CreateNamespace = true

	log.Println("Installing helm chart")
	if _, err := client.Run(chart, nil); err != nil {
		log.Println("Error installing helm chart: ", err)
		return err
	}

	return nil
}
