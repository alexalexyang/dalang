package deployHetznerServer

import (
	"context"
	_ "dalang/setup"
	testUtil "dalang/test/test-util"
	"fmt"
	"strings"

	"log"
	"testing"

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
			t.Log("Error with DeployNetworkFunc ,", err)
			return err
		}

		_, err = UploadSSHKey(ctx, sshKey)

		if err != nil {
			t.Log("Error with UploadSSHKey ,", err)
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

	assert.Equal(t, res.Outputs["keyName"].Value, res.Outputs["keyNameRes"].Value)
	assert.Equal(t, res.Outputs["publicKey"].Value, res.Outputs["publicKeyRes"].Value)

	// -- pulumi destroy --

	dRes, _ := testUtil.DestroyStack(t, ctx, stack)

	assert.Equal(t, "destroy", dRes.Summary.Kind)
	assert.Equal(t, "succeeded", dRes.Summary.Result)
}
