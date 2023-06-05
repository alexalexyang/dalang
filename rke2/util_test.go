package rke2

import (
	"context"
	"dalang/config"
	hetzner "dalang/hetzner/deploy-server"
	_ "dalang/setup"
	testUtil "dalang/test/test-util"
	_ "embed"
	"fmt"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	"testing"
)

//go:embed deployment.sample.yaml
var sampleDeploymentYaml string

func TestConvertYamlToObj(t *testing.T) {
	t.Parallel()

	k8sObjType := appsv1.Deployment{}

	deploymentObj, err := ConvertYamlToObj(sampleDeploymentYaml, &k8sObjType)
	if err != nil {
		t.Log("Error converting yaml to obj: ", err)
		t.Fatal(err)
	}

	depObj := deploymentObj.(*appsv1.Deployment)

	assert.Equal(t, "nginx:1.14.2", depObj.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "512Mi", depObj.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String())
}

func TestGetKubeConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var stackName = "testGetKubeConfig"
	var opts = testUtil.GetPulumiStackArgs(stackName)

	deployFunc := func(ctx *pulumi.Context) error {
		sshKeyPair, err := hetzner.CreateSSHKey(ctx)
		if err != nil {
			t.Log("Error with DeployNetworkFunc: ", err)
			return err
		}

		sshKey, err := hetzner.UploadSSHKey(ctx, sshKeyPair)

		if err != nil {
			t.Log("Error with UploadSSHKey: ", err)
			return err
		}

		server, err := hetzner.DeployServer(ctx, sshKey, 1)
		if err != nil {
			t.Log("Error with DeployNetworkFunc: ", err)
			return err
		}

		connectInfo := remote.ConnectionArgs{
			Host:       server.Ipv4Address,
			Port:       pulumi.Float64(22),
			User:       pulumi.String("root"),
			PrivateKey: sshKeyPair.Private,
		}

		ctx.Export(fmt.Sprintf("%s-server-%d-connect-info", config.Config.ProjectName, 1), connectInfo)

		installServerRes, err := InstallServer(ctx, &connectInfo, []pulumi.Resource{server})
		if err != nil {
			t.Log("Error installing RKE2 server: ", err)
			return err
		}

		kubeconfig, err := GetKubeConfig(ctx, &connectInfo, installServerRes)
		if err != nil {
			t.Log("Error getting kubeconfig: ", err)
			return err
		}
		t.Log("kubeconfig: ", *kubeconfig)

		assert.NotEmpty(t, *kubeconfig)

		serverChan := make(chan string)
		server.Ipv4Address.ApplyT(func(ip string) string {
			serverChan <- ip
			return ip
		})

		serverIp := <-serverChan
		close(serverChan)

		t.Log("server ip: ", serverIp)
		assert.NotEmpty(t, serverIp)

		updatedKubeConfig := UpdateKubeConfigServerIP(*kubeconfig, serverIp)
		t.Log("updated kubeconfig: ", updatedKubeConfig)
		assert.NotEmpty(t, updatedKubeConfig)

		return nil
	}

	stack, err := auto.UpsertStackInlineSource(ctx, stackName, config.Config.ProjectName, deployFunc, opts...)
	if err != nil {
		t.Fatal("Error with UpsertStackInlineSource: ", err)
	}

	// -- remove pulumi stack --
	defer testUtil.RemoveStack(t, ctx, stack)

	// -- pulumi up --
	res, _ := testUtil.UpStack(t, ctx, stack)

	assert.Equal(t, "update", res.Summary.Kind)
	assert.Equal(t, "succeeded", res.Summary.Result)

	// -- pulumi destroy --

	dRes, _ := testUtil.DestroyStack(t, ctx, stack)

	assert.Equal(t, "destroy", dRes.Summary.Kind)
	assert.Equal(t, "succeeded", dRes.Summary.Result)
}
