package setup

import (
	"dalang/config"
	getPulumi "dalang/pulumi-installer"
	"log"
)

func init() {
	log.Println("Setting up environment")

	config.SetSecrets()

	// Must be called before getting Pulumi
	config.SetConfig()

	config.SetHetznerConfig()

	// -- Get Pulumi --
	getPulumi.GetPulumiWorkspaceBackendDir()
	getPulumi.GetPulumi()

	log.Println("Environment set up")
}
