package rke2

import (
	"context"
	"dalang/config"
	hetzner "dalang/hetzner/deploy-server"
	_ "dalang/setup"
	testUtil "dalang/test/test-util"
	"os"
	"os/exec"
	"strings"

	"fmt"
	"log"
	"testing"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

// func TestMain(m *testing.M) {
// 	err := os.Setenv("GO_ENV", "development")
// 	if err != nil {
// 		log.Println("Error setting GO_ENV: ", err)
// 	}

// 	code := m.Run()
// 	os.Exit(code)
// }

var projectName = "test-project"

func TestInstallRke2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var stackName = "testInstallRke2"
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

		installServerRes, err := InstallRke2(ctx, &connectInfo, []pulumi.Resource{server})
		if err != nil {
			t.Log("Error installing RKE2 server: ", err)
			return err
		}

		serverToken, err := GetRke2ServerToken(ctx, &connectInfo, installServerRes)
		if err != nil {
			t.Log("Error getting RKE2 server token: ", err)
			return err
		}

		t.Log("serverToken: ", *serverToken)

		return nil
	}

	log.Println("Creating or selecting stack: ", stackName)

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

	isRke2ServerActive := fmt.Sprintf("%s", res.Outputs["is-rke2-server-active"].Value)

	assert.Equal(t, "active", strings.TrimRight(isRke2ServerActive, "\n"))

	rke2RegistrationToken := fmt.Sprintf("%s", res.Outputs["rke2-registration-token"].Value)
	t.Log("rke2RegistrationToken: ", rke2RegistrationToken)

	assert.NotEmpty(t, rke2RegistrationToken)

	if os.Getenv("GO_ENV") == "development" {
		cmdRes := exec.Command("sh", "-c", "chmod 777 hetzner-private-key")
		t.Log("cmdRes: ", cmdRes)

		t.Logf("SSH command: ssh-keygen -R %s && ssh -o \"StrictHostKeyChecking no\" -i rke2/hetzner-private-key root@%s", serverConnectInfo["host"], serverConnectInfo["host"])
	}

	// -- pulumi destroy --

	dRes, _ := testUtil.DestroyStack(t, ctx, stack)

	assert.Equal(t, "destroy", dRes.Summary.Kind)
	assert.Equal(t, "succeeded", dRes.Summary.Result)
}
