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

	// -- Get Pulumi --
	getPulumi.GetPulumi()
	getPulumi.GetPulumiWorkspaceBackendDir()

	log.Println("Environment set up")
}
