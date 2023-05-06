package rke2

import (
	"context"
	"dalang/config"
	_ "dalang/setup"
	testUtil "dalang/test/test-util"
	_ "embed"
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	"testing"
)

func TestK8sDeployment(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var stackName = "testK8sDeployment"
	var opts = testUtil.GetPulumiStackArgs(stackName)

	numServers := 1
	numAgents := 1
	numHosts := numServers + numAgents

	deployFunc := func(ctx *pulumi.Context) error {
		firstHostRes, firstHostConnArgs, serverIp, _, err := DeployHighAvailability(ctx, numServers, numAgents)
		if err != nil {
			t.Fatal("Error setting up high availability RKE2: ", err)
		}

		kubeconfigRaw, err := GetKubeConfig(ctx, firstHostConnArgs, firstHostRes)
		if err != nil {
			t.Log("Error getting kubeconfig: ", err)
			return err
		}

		t.Log("kubeconfig: ", *kubeconfigRaw)
		assert.NotEmpty(t, *kubeconfigRaw)

		kubeconfig := UpdateKubeConfigServerIP(*kubeconfigRaw, *serverIp)

		clientset, err := GetClientSet(kubeconfig)
		if err != nil {
			t.Log("Error getting clientset: ", err)
			return err
		}

		k8sObjType := appsv1.Deployment{}

		deploymentObj, err := ConvertYamlToObj(sampleDeploymentYaml, &k8sObjType)
		if err != nil {
			t.Log("Error converting yaml to obj: ", err)
			t.Fatal(err)
		}

		err = ApplyDeployment(clientset, deploymentObj)
		if err != nil {
			t.Log("Error applying deployment: ", err)
			return err
		}
		assert.Nil(t, err)

		return nil
	}

	stack, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, deployFunc, opts...)
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
