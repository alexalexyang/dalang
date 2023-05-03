package deployHetznerNetwork

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"dalang/util"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type runFunc func(ctx *pulumi.Context, num int) (*hcloud.Network, error)

func CreateNetwork(runFunc runFunc, numTimes int, destroy bool) (*auto.Stack, *auto.UpResult, error) {

	ctx := context.Background()

	// we use a simple stack name here, but recommend using auto.FullyQualifiedStackName for maximum specificity.
	stackName := "dev"
	projectName := "demo-project"

	cwd, err := util.GetCwd()
	if err != nil {
		log.Println("Cannot get CWD: ", err)
		return nil, nil, err
	}

	workspaceBackendPath := filepath.Join(*cwd, "pulumi-backend")
	log.Println("Workspace backend path: ", workspaceBackendPath)

	// Specify a local backend instead of using the service.
	project := auto.Project(
		workspace.Project{
			Name:    "demo-project",
			Runtime: workspace.NewProjectRuntimeInfo("go", nil),
			Backend: &workspace.ProjectBackend{
				URL: fmt.Sprintf("file://%s", workspaceBackendPath),
			},
		},
	)

	// Point to correct workdir
	workdir := auto.WorkDir(workspaceBackendPath)

	secretsProvider := auto.SecretsProvider("passphrase")

	stackSettings := auto.Stacks(map[string]workspace.ProjectStack{
		stackName: {SecretsProvider: "passphrase"},
	})

	envvars := auto.EnvVars(map[string]string{
		// In a real program, you would feed in the password securely or via the actual environment.
		"PULUMI_CONFIG_PASSPHRASE": "password",
	})

	deployFunc := func(ctx *pulumi.Context) error {
		counter := 0
		for counter < numTimes {
			_, err := runFunc(ctx, counter)
			if err != nil {
				log.Println(err)
				return err
			}
			counter = counter + 1
		}
		return nil
	}

	stack, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, deployFunc, project, secretsProvider, stackSettings, envvars, workdir)
	if err != nil {
		log.Println("Error with UpsertStackInlineSource: ", err)
		return nil, nil, err
	}

	if destroy {
		log.Println("Starting stack destroy")

		// wire up our destroy to stream progress to stdout
		stdoutStreamer := optdestroy.ProgressStreams(os.Stdout)

		// destroy our stack and exit early
		_, err := stack.Destroy(ctx, stdoutStreamer)
		if err != nil {
			log.Printf("Failed to destroy stack: %v", err)
		}
		log.Println("Stack successfully destroyed")
		os.Exit(0)
	}

	stdoutStreamer := optup.ProgressStreams(os.Stdout)

	res, err := stack.Up(ctx, stdoutStreamer)
	if err != nil {
		log.Printf("Failed to update stack: %v\n\n", err)
		os.Exit(1)
	}

	log.Println("Update succeeded!")

	outputs := res.Outputs

	for k, v := range outputs {
		log.Printf("%s = %v)\n", k, v.Value)
	}

	return &stack, &res, nil
}

func DeployNetworkFunc(ctx *pulumi.Context, num int) (*hcloud.Network, error) {

	log.Println("Creating new network...")

	network, err := hcloud.NewNetwork(ctx, fmt.Sprintf("demo-network-%d", num), &hcloud.NetworkArgs{
		IpRange: pulumi.String("10.0.10.0/24"),
		Name:    pulumi.String(fmt.Sprintf("demo-network-%d", num)),
		Labels: pulumi.Map{
			"network": pulumi.String("dev-network"),
		},
	})
	if err != nil {
		return nil, err
	}

	ctx.Export(fmt.Sprintf("network-%d-name", num), network.Name)
	ctx.Export(fmt.Sprintf("network-%d-id", num), network.ID())

	return network, nil

}
