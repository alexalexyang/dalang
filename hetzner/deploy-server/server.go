package deployHetznerServer

import (
	"dalang/config"
	"fmt"
	"log"
	"os"
	"path/filepath"

	gonanoid "github.com/matoous/go-nanoid"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SSHKeyPair struct {
	Public  pulumi.StringOutput
	Private pulumi.StringOutput
	Name    pulumi.StringOutput
}

// Generate random SSH key
func CreateSSHKey(ctx *pulumi.Context) (*SSHKeyPair, error) {
	log.Println("Generating SSH key")
	newkeys, err := tls.NewPrivateKey(ctx, "private-key", &tls.PrivateKeyArgs{
		Algorithm:  pulumi.String("ECDSA"),
		EcdsaCurve: pulumi.String("P521"),
	})
	if err != nil {
		log.Printf("Error creating Pulumi SSH key: %s\n", err)
		return nil, err
	}

	keyPair := SSHKeyPair{
		Public:  newkeys.PublicKeyOpenssh,
		Private: newkeys.PrivateKeyOpenssh,
	}

	log.Println("SSH key generated: ", keyPair.Public, " ", keyPair.Private)

	if os.Getenv("GO_ENV") == "development" {
		_ = keyPair.Private.ApplyT(func(key string) string {
			wd, _ := os.Getwd()
			_ = os.WriteFile(filepath.Join(wd, "hetzner-private-key"), []byte(key), 0644)

			return key
		})
	}

	// get identifier from the middle of the public key
	keyPair.Name = keyPair.Public.ApplyT(func(pkey string) string {
		middle := int(len(pkey) / 2)

		return "pulumi-key-" + pkey[middle-4:middle+4]
	}).(pulumi.StringOutput)

	ctx.Export("keyName", keyPair.Name)
	ctx.Export("publicKey", keyPair.Public)

	log.Println("SSH key name: ", keyPair.Name)

	return &keyPair, nil
}

func UploadSSHKey(ctx *pulumi.Context, key *SSHKeyPair) (*hcloud.SshKey, error) {
	log.Println("Uploading SSH key to Hetzner")

	sshKey, err := hcloud.NewSshKey(ctx, "sshkey", &hcloud.SshKeyArgs{
		PublicKey: key.Public,
		Name:      key.Name,
	})
	if err != nil {
		log.Printf("Error uploading Pulumi SSH key: %s\n", err)
		return nil, err
	}

	ctx.Export("keyNameRes", key.Name)
	ctx.Export("publicKeyRes", key.Public)

	return sshKey, nil
}

func DeployServer(ctx *pulumi.Context, sshKey *hcloud.SshKey, serverNum int) (*hcloud.Server, error) {

	id, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-", 12)
	if err != nil {
		log.Println("Error generating random alphanumeric characters: ", err)
		return nil, err
	}

	serverName := fmt.Sprintf("%s-server-%d-%s", config.Config.ProjectName, serverNum, id)

	log.Println("Deploying server", serverName)

	server, err := hcloud.NewServer(ctx, serverName, &hcloud.ServerArgs{
		Name:       pulumi.String(serverName),
		ServerType: pulumi.String(config.HetznerConfig.ServerType),
		Image:      pulumi.String(config.HetznerConfig.OsImage),
		Location:   pulumi.String(config.HetznerConfig.ServerLocation),
		SshKeys:    pulumi.StringArray{sshKey.ID()},
	})
	if err != nil {
		log.Println("Error creating server: ", err)
		return nil, err
	}

	log.Println("Server", serverName, "deployed")

	return server, nil
}
