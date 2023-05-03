package rke2

import (
	_ "embed"
	"fmt"
	"log"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

//go:embed rke2-server-install-script.sh
var rke2ServerInstallScript string

//go:embed rke2-agent-install-script.sh
var rke2AgentInstallScript string

func InstallRke2(ctx *pulumi.Context, connection *remote.ConnectionArgs, dependsOn []pulumi.Resource) (*remote.Command, error) {

	log.Println("Copy RKE2 server install script to remote host")
	copyScriptRes, err := remote.NewCommand(ctx, "copy-rke2-server-install-script", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String("cat > /root/rke2-server-install-script.sh && chmod u+x /root/rke2-server-install-script.sh"),
		Stdin:      pulumi.String(rke2ServerInstallScript),
		Delete:     pulumi.String("rm /root/rke2-server-install-script.sh"),
	}, pulumi.DeleteBeforeReplace(true), pulumi.DependsOn(dependsOn))
	if err != nil {
		return nil, err
	}

	ctx.Export("copy-rke2-server-install-script-stdout", copyScriptRes.Stdout)
	ctx.Export("copy-rke2-server-install-script-stderr", copyScriptRes.Stderr)

	log.Println("Run RKE2 server install script on remote host")

	runScriptRes, err := remote.NewCommand(ctx, "run-rke2-server-install-script", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String(". /root/rke2-server-install-script.sh"),
		Triggers:   pulumi.All(copyScriptRes.Create, copyScriptRes.Stdin, copyScriptRes.Delete),
	}, pulumi.DependsOn([]pulumi.Resource{copyScriptRes}), pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return nil, err
	}

	ctx.Export("run-rke2-server-install-script-stdout", runScriptRes.Stdout)
	ctx.Export("run-rke2-server-install-script-stderr", runScriptRes.Stderr)

	return runScriptRes, nil
}

// `dependsOn` has to be the result of running the RKE2 server install script.
// That is, `runScriptRes` above.
func GetRke2ServerToken(ctx *pulumi.Context, connection *remote.ConnectionArgs, dependsOn *remote.Command) (*string, error) {
	log.Println("Check if RKE2 server is active")

	rke2StatusRes, err := remote.NewCommand(ctx, "is-rke2-server-active", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String("systemctl is-active rke2-server.service"),
		Triggers:   pulumi.All(dependsOn.Create, dependsOn.Stdin),
	}, pulumi.DependsOn([]pulumi.Resource{dependsOn}))
	if err != nil {
		return nil, err
	}

	ctx.Export("is-rke2-server-active", rke2StatusRes.Stdout)

	rke2StatusChan := make(chan string)

	rke2StatusRes.Stdout.ApplyT(func(status string) string {
		rke2StatusChan <- strings.TrimRight(status, "\n")
		return status
	})

	rke2Status := <-rke2StatusChan

	if rke2Status != "active" {
		log.Println("RKE2 server is not active: ", rke2Status)
		return nil, fmt.Errorf("RKE2 server is not active: %s", rke2Status)
	}

	log.Println("RKE2 server is active: ", rke2Status)
	close(rke2StatusChan)

	log.Println("Get RKE2 registration token")

	serverTokenRes, err := remote.NewCommand(ctx, "get-registration-token", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String("cat /var/lib/rancher/rke2/server/node-token"),
		Triggers:   pulumi.All(rke2StatusRes.Create, rke2StatusRes.Stdin),
	}, pulumi.DependsOn([]pulumi.Resource{rke2StatusRes}))
	if err != nil {
		return nil, err
	}

	ctx.Export("serverTokenRes", serverTokenRes.Stdout)

	rke2TokenChan := make(chan string)

	serverTokenRes.Stdout.ApplyT(func(status string) string {
		rke2TokenChan <- strings.TrimRight(status, "\n")
		return status
	})

	rke2Token := <-rke2TokenChan
	close(rke2TokenChan)

	return &rke2Token, nil
}

func InstallAgent(ctx *pulumi.Context, connection *remote.ConnectionArgs, dependsOn []pulumi.Resource) (pulumi.Resource, error) {

	log.Println("Copying RKE2 agent install script to remote host")
	res1, err := remote.NewCommand(ctx, "copy-rke2-agent-install-script", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String("cat > /root/rke2-agent-install-script.sh && chmod u+x /root/rke2-agent-install-script.sh"),
		Stdin:      pulumi.String(rke2AgentInstallScript),
		Delete:     pulumi.String("rm /root/rke2-agent-install-script.sh"),
	}, pulumi.DeleteBeforeReplace(true), pulumi.DependsOn(dependsOn))
	if err != nil {
		return nil, err
	}

	log.Println("Running RKE2 agent install script on remote host")

	ctx.Export("copy-rke2-agent-install-script-stdout", res1.Stdout)
	ctx.Export("copy-rke2-agent-install-script-stderr", res1.Stderr)

	res2, err := remote.NewCommand(ctx, "run-rke2-agent-install-script", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String(". /root/rke2-agent-install-script.sh"),
		Triggers:   pulumi.All(res1.Create, res1.Stdin, res1.Delete),
	}, pulumi.DependsOn([]pulumi.Resource{res1}), pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return nil, err
	}

	ctx.Export("run-rke2-agent-install-script-stdout", res2.Stdout)
	ctx.Export("run-rke2-agent-install-script-stderr", res2.Stderr)

	res3, err := remote.NewCommand(ctx, "is-rke2-agent-active", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String("systemctl is-active rke2-agent.service"),
		Triggers:   pulumi.All(res2.Create, res2.Stdin),
	}, pulumi.DependsOn([]pulumi.Resource{res2}), pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return nil, err
	}

	ctx.Export("is-rke2-agent-active", res3.Stdout)

	return res3, nil
}
