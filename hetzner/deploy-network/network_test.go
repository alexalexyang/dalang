package deployHetznerNetwork

import (
	"context"
	_ "dalang/setup"
	testUtil "dalang/test/test-util"

	"log"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

func TestCreateHetznerNetwork(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	projectName := "test-project"
	stackName := "test"
	opts := testUtil.GetPulumiStackArgs(stackName)

	deployFunc := func(ctx *pulumi.Context) error {
		_, err := DeployNetworkFunc(ctx, 1)
		if err != nil {
			t.Log("Error with DeployNetworkFunc ,", err)
			return err
		}

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

	networkName := res.Outputs["network-1-name"].Value.(string)
	networkId := res.Outputs["network-1-id"].Value.(string)

	t.Log("Network 1 name: ", networkName)
	t.Log("Network 1 id: ", networkId)
	assert.Equal(t, "demo-network-1", networkName)

	// -- pulumi destroy --

	dRes, _ := testUtil.DestroyStack(t, ctx, stack)

	assert.Equal(t, "destroy", dRes.Summary.Kind)
	assert.Equal(t, "succeeded", dRes.Summary.Result)

}
