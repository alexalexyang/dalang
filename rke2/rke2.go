package rke2

import (
	_ "embed"
	"log"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

//go:embed rke2-server-install-script.sh
var rke2ServerInstallScript string

func InstallServer(ctx *pulumi.Context, connection *remote.ConnectionArgs, dependsOn []pulumi.Resource) (pulumi.Resource, error) {

	log.Println("Copying RKE2 Server Install Script to remote host")
	res1, err := remote.NewCommand(ctx, "copy-rke2-server-install-script", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String("cat > /root/rke2-server-install-script.sh && chmod u+x /root/rke2-server-install-script.sh"),
		Stdin:      pulumi.String(rke2ServerInstallScript),
		Delete:     pulumi.String("rm /root/rke2-server-install-script.sh"),
	}, pulumi.DeleteBeforeReplace(true), pulumi.DependsOn(dependsOn))
	if err != nil {
		return nil, err
	}

	log.Println("Running RKE2 Server Install Script on remote host")

	ctx.Export("copy-rke2-server-install-script-stdout", res1.Stdout)
	ctx.Export("copy-rke2-server-install-script-stderr", res1.Stderr)

	res2, err := remote.NewCommand(ctx, "run-rke2-server-install-script", &remote.CommandArgs{
		Connection: connection,
		Create:     pulumi.String(". /root/rke2-server-install-script.sh"),
		Triggers:   pulumi.All(res1.Create, res1.Stdin, res1.Delete),
	}, pulumi.DependsOn([]pulumi.Resource{res1}), pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return nil, err
	}

	ctx.Export("run-rke2-server-install-script-stdout", res2.Stdout)
	ctx.Export("run-rke2-server-install-script-stderr", res2.Stderr)

	return res2, nil
}
