package k8s

import (
	"context"
	"dalang/config"
	rke2 "dalang/rke2"
	_ "dalang/setup"
	testUtil "dalang/test/test-util"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

//go:embed deployment.sample.yaml
var sampleDeploymentYaml string

//go:embed service.sample.yaml
var sampleServiceYaml string

//go:embed ingress.sample.yaml
var sampleIngressYaml string

func TestK8sDeployment(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var stackName = "testK8sDeployment"
	var opts = testUtil.GetPulumiStackArgs(stackName)

	numServers := 1
	numAgents := 1
	numHosts := numServers + numAgents

	deployFunc := func(ctx *pulumi.Context) error {
		t.Log("Deploying high availability RKE2")

		firstHostRes, firstHostConnArgs, serverIp, _, err := rke2.DeployHighAvailability(ctx, numServers, numAgents)
		if err != nil {
			t.Fatal("Error setting up high availability RKE2: ", err)
		}

		t.Log("Getting kubeconfig")
		kubeconfigRaw, err := rke2.GetKubeConfig(ctx, firstHostConnArgs, firstHostRes)
		if err != nil {
			t.Log("Error getting kubeconfig: ", err)
			return err
		}

		assert.NotEmpty(t, *kubeconfigRaw)

		kubeconfig := rke2.UpdateKubeConfigServerIP(*kubeconfigRaw, *serverIp)

		cwd, _ := os.Getwd()
		if os.Getenv("GO_ENV") == "development" {
			_ = os.WriteFile(filepath.Join(cwd, "kubeconfig.dev.yaml"), []byte(kubeconfig), 0644)
		}

		clientset, err := rke2.GetClientSet(kubeconfig)
		if err != nil {
			t.Log("Error getting clientset: ", err)
			return err
		}

		depType := appsv1.Deployment{}

		deploymentObj, err := rke2.ConvertYamlToObj(sampleDeploymentYaml, &depType)
		if err != nil {
			t.Log("Error converting yaml to obj: ", err)
			t.Fatal(err)
		}

		depObj := deploymentObj.(*appsv1.Deployment)

		t.Log("Applying sample deployment")
		err = ApplyDeployment(clientset, depObj)
		if err != nil {
			t.Log("Error applying deployment: ", err)
			return err
		}

		assert.Nil(t, err)

		svcType := apiv1.Service{}

		serviceObj, err := rke2.ConvertYamlToObj(sampleServiceYaml, &svcType)
		if err != nil {
			t.Log("Error converting yaml to obj: ", err)
			t.Fatal(err)
		}

		svcObj := serviceObj.(*apiv1.Service)

		t.Log("Applying sample service")
		err = ApplyService(clientset, svcObj)
		if err != nil {
			t.Log("Error applying service: ", err)
			return err
		}

		ingType := networkingv1.Ingress{}

		ingressObj, err := rke2.ConvertYamlToObj(sampleIngressYaml, &ingType)
		if err != nil {
			t.Log("Error converting yaml to obj: ", err)
			t.Fatal(err)
		}

		ingObj := ingressObj.(*networkingv1.Ingress)

		t.Log("Applying sample ingress")
		err = ApplyIngress(clientset, ingObj)
		if err != nil {
			t.Log("Error applying service: ", err)
			return err
		}

		assert.Nil(t, err)

		// The docs don't say where it is, but I found the tgz archive here: https://artifacthub.io/packages/helm/ingress-nginx/ingress-nginx
		// It turns out, we don't have to install an ingress controller because it's installed by default with RKE2

		// ingressControllerFilepath := filepath.Join(cwd, "nginx-ingress-controller", "ingress-nginx-4.7.0.tgz")

		// InstallChart(ingressControllerFilepath, "nginx-ingress-controller", "ingress-nginx", &kubeconfig)

		return nil
	}

	stack, err := auto.UpsertStackInlineSource(ctx, stackName, config.Config.ProjectName, deployFunc, opts...)
	if err != nil {
		t.Fatal("Error with UpsertStackInlineSource: ", err)
	}

	// -- remove pulumi stack --
	// defer testUtil.RemoveStack(t, ctx, stack)

	// -- pulumi up --
	res, _ := testUtil.UpStack(t, ctx, stack)

	assert.Equal(t, "update", res.Summary.Kind)
	assert.Equal(t, "succeeded", res.Summary.Result)

	for i := 1; i < numHosts+1; i++ {
		serverConnectInfoKey := fmt.Sprintf("%s-server-%d-connect-info", config.Config.ProjectName, i)
		serverConnectInfo := res.Outputs[serverConnectInfoKey].Value.(map[string]interface{})

		assert.NotEmpty(t, serverConnectInfo["host"])
		assert.NotEmpty(t, serverConnectInfo["port"])
		assert.NotEmpty(t, serverConnectInfo["privateKey"])
		assert.Equal(t, serverConnectInfo["user"], "root")
	}

	// -- pulumi destroy --

	// dRes, _ := testUtil.DestroyStack(t, ctx, stack)

	// assert.Equal(t, "destroy", dRes.Summary.Kind)
	// assert.Equal(t, "succeeded", dRes.Summary.Result)
}
