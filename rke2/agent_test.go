package rke2

import (
	"context"
	"dalang/config"
	hetzner "dalang/hetzner/deploy-server"
	_ "dalang/setup"
	testUtil "dalang/test/test-util"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

func TestInstallRke2Agent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var stackName = "testInstallRke2Agent"
	var opts = testUtil.GetPulumiStackArgs(stackName)

	deployFunc := func(ctx *pulumi.Context) error {
		// -- Set up RKE2 server server --
		t.Log("Set up RKE2 server server")

		serverKeyPair, err := hetzner.CreateSSHKey(ctx)
		if err != nil {
			t.Log("Error with DeployNetworkFunc: ", err)
			return err
		}

		serverSSHKey, err := hetzner.UploadSSHKey(ctx, serverKeyPair)

		if err != nil {
			t.Log("Error with UploadSSHKey: ", err)
			return err
		}

		serverHost, err := hetzner.DeployServer(ctx, serverSSHKey, 1)
		if err != nil {
			t.Log("Error with DeployNetworkFunc: ", err)
			return err
		}

		serverConnectInfo := remote.ConnectionArgs{
			Host:       serverHost.Ipv4Address,
			Port:       pulumi.Float64(22),
			User:       pulumi.String("root"),
			PrivateKey: serverKeyPair.Private,
		}

		ctx.Export(fmt.Sprintf("%s-server-%d-connect-info", config.Config.ProjectName, 1), serverConnectInfo)

		installServerRes, err := InstallServer(ctx, &serverConnectInfo, []pulumi.Resource{serverHost})
		if err != nil {
			t.Log("Error installing RKE2 server: ", err)
			return err
		}

		serverToken, err := GetRke2ServerToken(ctx, &serverConnectInfo, installServerRes)
		if err != nil {
			t.Log("Error getting RKE2 server token: ", err)
			return err
		}

		assert.NotEmpty(t, *serverToken)

		serverChan := make(chan string)
		serverHost.Ipv4Address.ApplyT(func(ip string) string {
			serverChan <- ip
			return ip
		})

		serverIp := <-serverChan
		close(serverChan)

		t.Log("server token: ", *serverToken)
		t.Log("server ip: ", serverIp)

		// -- Set up RKE2 agent server --
		t.Log("Set up RKE2 agent server")

		agentHost, err := hetzner.DeployServer(ctx, serverSSHKey, 1)
		if err != nil {
			t.Log("Error with DeployNetworkFunc: ", err)
			return err
		}

		t.Log("Agent server deployed")

		agentConnectInfo := remote.ConnectionArgs{
			Host:       agentHost.Ipv4Address,
			Port:       pulumi.Float64(22),
			User:       pulumi.String("root"),
			PrivateKey: serverKeyPair.Private,
		}

		agentChan := make(chan string)
		agentHost.Ipv4Address.ApplyT(func(ip string) string {
			t.Log("agent ip in ApplyT: ", ip)
			agentChan <- ip
			return ip
		})

		agentIp := <-agentChan
		close(agentChan)

		t.Log("agent ip: ", agentIp)

		if os.Getenv("GO_ENV") == "development" {
			cmdRes := exec.Command("sh", "-c", "chmod 777 hetzner-private-key")
			t.Log("cmdRes: ", cmdRes)

			t.Logf("SSH command: ssh-keygen -R %s && ssh -o \"StrictHostKeyChecking no\" -i rke2/hetzner-private-key root@%s", agentIp, agentIp)
		}

		runScriptRes, err := InstallAgent(ctx, agentIp, &agentConnectInfo, []pulumi.Resource{agentHost})
		if err != nil {
			t.Log("Error installing RKE2 agent: ", err)
			return err
		}

		_, agentStatus, err := StartAgent(ctx, agentIp, &agentConnectInfo, []pulumi.Resource{runScriptRes}, serverIp, *serverToken)
		if err != nil {
			t.Log("Error starting RKE2 agent: ", err)
			t.Fail()
		}

		t.Log("agentStatus: ", *agentStatus)

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

	// -- pulumi destroy --

	dRes, _ := testUtil.DestroyStack(t, ctx, stack)

	assert.Equal(t, "destroy", dRes.Summary.Kind)
	assert.Equal(t, "succeeded", dRes.Summary.Result)
}
