package rke2

import (
	"context"
	"dalang/config"
	hetzner "dalang/hetzner/deploy-server"
	_ "dalang/setup"
	testUtil "dalang/test/test-util"

	"fmt"
	"log"
	"testing"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

var projectName = "test-project"

func TestInstallRke2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var stackName = "testInstallRke2"
	var opts = testUtil.GetPulumiStackArgs(stackName)

	deployFunc := func(ctx *pulumi.Context) error {
		sshKeyPair, err := hetzner.CreateSSHKey(ctx)
		if err != nil {
			t.Log("Error with DeployNetworkFunc ,", err)
			return err
		}

		sshKey, err := hetzner.UploadSSHKey(ctx, sshKeyPair)

		if err != nil {
			t.Log("Error with UploadSSHKey ,", err)
			return err
		}

		server, err := hetzner.DeployServer(ctx, sshKey, 1)
		if err != nil {
			t.Log("Error with DeployNetworkFunc ,", err)
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

	log.Println("Creating or selecting stack: ", stackName)

	stack, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, deployFunc, opts...)
	if err != nil {
		t.Fatal("Error with UpsertStackInlineSource ,", err)
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
