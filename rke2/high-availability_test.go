package rke2

import (
	"context"
	"dalang/config"
	_ "dalang/setup"
	testUtil "dalang/test/test-util"
	"fmt"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

func TestHighAvailability(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var stackName = "testHighAvailability"
	var opts = testUtil.GetPulumiStackArgs(stackName)

	numServers := 1
	numAgents := 2
	numHosts := numServers + numAgents

	deployFunc := func(ctx *pulumi.Context) error {
		_, _, serverIp, serverToken, err := DeployHighAvailability(ctx, numServers, numAgents)
		if err != nil {
			t.Fatal("Error setting up high availability RKE2: ", err)
		}

		assert.NotEmpty(t, serverIp)
		assert.NotEmpty(t, serverToken)

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

	for i := 1; i < numHosts+1; i++ {
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
