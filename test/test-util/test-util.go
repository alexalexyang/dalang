package testUtil

import (
	"context"
	"dalang/util"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/stretchr/testify/assert"
)

func GetPulumiStackArgs(stackName string) []auto.LocalWorkspaceOption {

	cwd, err := util.GetCwd()
	if err != nil {
		log.Fatal("Cannot get CWD: ", err)
	}

	packageDir := path.Dir(*cwd)
	projDir := path.Dir(packageDir)
	workspaceBackendPath := filepath.Join(projDir, "pulumi-backend")
	log.Println("Workspace backend path: ", workspaceBackendPath)

	// Specify a local backend instead of using the service.
	project := auto.Project(
		workspace.Project{
			Name:    "test-project",
			Runtime: workspace.NewProjectRuntimeInfo("go", nil),
			Backend: &workspace.ProjectBackend{
				URL: fmt.Sprintf("file://%s", workspaceBackendPath),
			},
		},
	)

	// Point to correct workdir
	workDir := auto.WorkDir(workspaceBackendPath)

	secretsProvider := auto.SecretsProvider("passphrase")

	stackSettings := auto.Stacks(map[string]workspace.ProjectStack{
		stackName: {SecretsProvider: "passphrase"},
	})

	envVars := auto.EnvVars(map[string]string{
		"PULUMI_CONFIG_PASSPHRASE": "password",
	})

	return []auto.LocalWorkspaceOption{project, workDir, secretsProvider, stackSettings, envVars}
}

func UpStack(t *testing.T, ctx context.Context, stack auto.Stack) (auto.UpResult, error) {
	res, err := stack.Up(ctx)
	if err != nil {
		t.Errorf("up failed, err: %v", err)
		t.FailNow()
	}

	return res, err
}

func DestroyStack(t *testing.T, ctx context.Context, stack auto.Stack) (auto.DestroyResult, error) {
	dRes, err := stack.Destroy(ctx)
	if err != nil {
		t.Errorf("destroy failed, err: %v", err)
		t.FailNow()
	}

	return dRes, err
}

func RemoveStack(t *testing.T, ctx context.Context, stack auto.Stack) {
	err := os.Unsetenv("PULUMI_CONFIG_PASSPHRASE")
	assert.Nil(t, err, "failed to unset EnvVar.")

	// -- pulumi stack rm --
	err = stack.Workspace().RemoveStack(ctx, stack.Name())
	assert.Nil(t, err, "failed to remove stack. Resources have leaked.")
}
