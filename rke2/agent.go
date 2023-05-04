package rke2

import (
	_ "embed"
	"fmt"
	"log"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gopkg.in/yaml.v3"
)

//go:embed agent-install-script.sh
var agentInstallScript string

// Installs RKE2 agent on remote host
func InstallAgent(ctx *pulumi.Context, connection *remote.ConnectionArgs, dependsOn []pulumi.Resource) (pulumi.Resource, error) {

	log.Println("Copying RKE2 agent install script to remote host")
	copyScriptRes, err := remote.NewCommand(ctx, "copy-rke2-agent-install-script", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String("cat > /root/rke2-agent-install-script.sh && chmod u+x /root/rke2-agent-install-script.sh"),
		Stdin:      pulumi.String(agentInstallScript),
		Delete:     pulumi.String("rm /root/rke2-agent-install-script.sh"),
	}, pulumi.DeleteBeforeReplace(true), pulumi.DependsOn(dependsOn))
	if err != nil {
		return nil, err
	}

	ctx.Export("copy-rke2-agent-install-script-stdout", copyScriptRes.Stdout)
	ctx.Export("copy-rke2-agent-install-script-stderr", copyScriptRes.Stderr)

	log.Println("Running RKE2 agent install script on remote host")
	runScriptRes, err := remote.NewCommand(ctx, "run-rke2-agent-install-script", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String(". /root/rke2-agent-install-script.sh"),
		Triggers:   pulumi.All(copyScriptRes.Create, copyScriptRes.Stdin, copyScriptRes.Delete),
	}, pulumi.DependsOn([]pulumi.Resource{copyScriptRes}), pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return nil, err
	}

	ctx.Export("run-rke2-agent-install-script-stdout", runScriptRes.Stdout)
	ctx.Export("run-rke2-agent-install-script-stderr", runScriptRes.Stderr)

	return runScriptRes, nil
}

func StartAgent(ctx *pulumi.Context, connection *remote.ConnectionArgs, dependsOn []pulumi.Resource, serverIp string, serverToken string) (pulumi.Resource, *string, error) {

	mkconfigDir, err := remote.NewCommand(ctx, "make-agent-config-dir", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String("mkdir -p /etc/rancher/rke2/"),
	}, pulumi.DependsOn(dependsOn), pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return nil, nil, err
	}

	// Create config file

	configStruct := struct {
		Server string `yaml:"server"`
		Token  string `yaml:"token"`
	}{
		Server: fmt.Sprintf("https://%s:9345", serverIp),
		Token:  serverToken,
	}

	log.Println("Config struct: ", configStruct)

	config, err := yaml.Marshal(configStruct)
	if err != nil {
		log.Fatal(err)
	}

	writeFileCmd := fmt.Sprintf("cat > /etc/rancher/rke2/config.yaml <<EOF\n%s\nEOF", string(config))

	writeConfigFile, err := remote.NewCommand(ctx, "write-agent-config-file", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String(writeFileCmd),
		Triggers:   pulumi.All(mkconfigDir.Create, mkconfigDir.Stdin),
	}, pulumi.DependsOn([]pulumi.Resource{mkconfigDir}), pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return nil, nil, err
	}

	log.Println("Starting RKE2 agent on remote host")

	startAgent, err := remote.NewCommand(ctx, "start-rke2-agent", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String("systemctl start rke2-agent.service"),
		Triggers:   pulumi.All(writeConfigFile.Create, writeConfigFile.Stdin),
	}, pulumi.DependsOn([]pulumi.Resource{writeConfigFile}), pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return nil, nil, err
	}

	log.Println("Checking RKE2 agent status on remote host")

	statusRes, err := remote.NewCommand(ctx, "is-rke2-agent-active", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String("systemctl is-active rke2-agent.service"),
		Triggers:   pulumi.All(startAgent.Create, startAgent.Stdin),
	}, pulumi.DependsOn([]pulumi.Resource{startAgent}), pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return nil, nil, err
	}

	ctx.Export("is-rke2-agent-active", statusRes.Stdout)
	statusChan := make(chan string)

	statusRes.Stdout.ApplyT(func(status string) string {
		statusChan <- strings.TrimRight(status, "\n")
		return status
	})

	status := <-statusChan
	close(statusChan)

	if status != "active" {
		log.Println("RKE2 agent is not active: ", status)
		return nil, nil, fmt.Errorf("RKE2 agent is not active: %s", status)
	}

	log.Println("RKE2 agent is active: ", status)

	return statusRes, &status, nil
}
