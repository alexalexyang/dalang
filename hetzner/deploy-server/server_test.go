package deployHetznerServer

import (
	"context"
	"dalang/config"
	_ "dalang/setup"
	testUtil "dalang/test/test-util"
	"fmt"
	"strings"

	"testing"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

var projectName = "test-project"

func TestCreatePulumiSSHKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var stackName = "testCreateSSHKeys"
	var opts = testUtil.GetPulumiStackArgs(stackName)

	deployFunc := func(ctx *pulumi.Context) error {
		_, err := CreateSSHKey(ctx)
		if err != nil {
			t.Log("Error with DeployNetworkFunc: ", err)
			return err
		}

		return nil
	}

	stack, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, deployFunc, opts...)
	if err != nil {
		t.Fatal("Error with UpsertStackInlineSource: ", err)
	}

	// -- remove pulumi stack --
	defer testUtil.RemoveStack(t, ctx, stack)

	// -- pulumi up --
	res, _ := testUtil.UpStack(t, ctx, stack)

	assert.Equal(t, "update", res.Summary.Kind)
	assert.Equal(t, "succeeded", res.Summary.Result)

	publicKey := fmt.Sprintf("%v", res.Outputs["publicKey"].Value)

	assert.Equal(t, true, strings.Contains(publicKey, "ecdsa-sha2-nistp521"))

	// -- pulumi destroy --

	dRes, _ := testUtil.DestroyStack(t, ctx, stack)

	assert.Equal(t, "destroy", dRes.Summary.Kind)
	assert.Equal(t, "succeeded", dRes.Summary.Result)
}

func TestUploadSSHKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var stackName = "testUploadSSHKey"
	var opts = testUtil.GetPulumiStackArgs(stackName)

	deployFunc := func(ctx *pulumi.Context) error {
		sshKey, err := CreateSSHKey(ctx)
		if err != nil {
			t.Log("Error with DeployNetworkFunc: ", err)
			return err
		}

		_, err = UploadSSHKey(ctx, sshKey)

		if err != nil {
			t.Log("Error with UploadSSHKey: ", err)
			return err
		}

		return nil
	}

	stack, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, deployFunc, opts...)
	if err != nil {
		t.Fatal("Error with UpsertStackInlineSource: ", err)
	}

	// -- remove pulumi stack --
	defer testUtil.RemoveStack(t, ctx, stack)

	// -- pulumi up --
	res, _ := testUtil.UpStack(t, ctx, stack)

	assert.Equal(t, "update", res.Summary.Kind)
	assert.Equal(t, "succeeded", res.Summary.Result)

	assert.Equal(t, res.Outputs["keyName"].Value, res.Outputs["keyNameRes"].Value)
	assert.Equal(t, res.Outputs["publicKey"].Value, res.Outputs["publicKeyRes"].Value)

	// -- pulumi destroy --

	dRes, _ := testUtil.DestroyStack(t, ctx, stack)

	assert.Equal(t, "destroy", dRes.Summary.Kind)
	assert.Equal(t, "succeeded", dRes.Summary.Result)
}

func TestDeployOneHetznerServer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var stackName = "testDeployOneHetznerServer"
	var opts = testUtil.GetPulumiStackArgs(stackName)

	deployFunc := func(ctx *pulumi.Context) error {
		sshKeyPair, err := CreateSSHKey(ctx)
		if err != nil {
			t.Log("Error with DeployNetworkFunc: ", err)
			return err
		}

		sshKey, err := UploadSSHKey(ctx, sshKeyPair)

		if err != nil {
			t.Log("Error with UploadSSHKey: ", err)
			return err
		}

		server, err := DeployServer(ctx, sshKey, 1)
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

		return nil
	}

	stack, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, deployFunc, opts...)
	if err != nil {
		t.Fatal("Error with UpsertStackInlineSource: ", err)
	}

	// -- remove pulumi stack --
	defer testUtil.RemoveStack(t, ctx, stack)

	// -- pulumi up --
	res, _ := testUtil.UpStack(t, ctx, stack)

	assert.Equal(t, "update", res.Summary.Kind)
	assert.Equal(t, "succeeded", res.Summary.Result)

	serverConnectInfoKey := fmt.Sprintf("%s-server-%d-connect-info", config.Config.ProjectName, 1)
	serverConnectInfo := res.Outputs[serverConnectInfoKey].Value.(map[string]interface{})

	assert.NotEmpty(t, serverConnectInfo["host"])
	assert.NotEmpty(t, serverConnectInfo["port"])
	assert.NotEmpty(t, serverConnectInfo["privateKey"])
	assert.Equal(t, serverConnectInfo["user"], "root")

	// -- pulumi destroy --

	dRes, _ := testUtil.DestroyStack(t, ctx, stack)

	assert.Equal(t, "destroy", dRes.Summary.Kind)
	assert.Equal(t, "succeeded", dRes.Summary.Result)
}

func TestMultipleHetznerServers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var stackName = "testMultipleHetznerServers"
	var opts = testUtil.GetPulumiStackArgs(stackName)

	numServers := 2

	deployFunc := func(ctx *pulumi.Context) error {
		serverInfoSlice, _, err := DeployServers1SSHKey(ctx, numServers)
		if err != nil {
			t.Log("Error with DeployNetworkFunc: ", err)
			return err
		}

		assert.Equal(t, len(serverInfoSlice), numServers)

		for idx, serverInfo := range serverInfoSlice {
			ctx.Export(fmt.Sprintf("%s-server-%d-connect-info", config.Config.ProjectName, idx+1), serverInfo.ConnectArgs)
		}

		return nil
	}

	stack, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, deployFunc, opts...)
	if err != nil {
		t.Fatal("Error with UpsertStackInlineSource: ", err)
	}

	// -- remove pulumi stack --
	defer testUtil.RemoveStack(t, ctx, stack)

	// -- pulumi up --
	res, _ := testUtil.UpStack(t, ctx, stack)

	assert.Equal(t, "update", res.Summary.Kind)
	assert.Equal(t, "succeeded", res.Summary.Result)

	for i := 1; i < numServers+1; i++ {
		serverConnectInfoKey := fmt.Sprintf("%s-server-%d-connect-info", config.Config.ProjectName, i)
		serverConnectInfo := res.Outputs[serverConnectInfoKey].Value.(map[string]interface{})

		assert.NotEmpty(t, serverConnectInfo["host"])
		assert.NotEmpty(t, serverConnectInfo["port"])
		assert.NotEmpty(t, serverConnectInfo["privateKey"])
		assert.Equal(t, serverConnectInfo["user"], "root")
	}

	// -- pulumi destroy --

	dRes, _ := testUtil.DestroyStack(t, ctx, stack)

	assert.Equal(t, "destroy", dRes.Summary.Kind)
	assert.Equal(t, "succeeded", dRes.Summary.Result)
}
