package deployHetznerServer

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SSHKey struct {
	Public  pulumi.StringOutput
	Private pulumi.StringOutput
	Name    pulumi.StringOutput
}

// Generate random SSH key
func GetSSHKey(ctx *pulumi.Context) (*SSHKey, error) {
	log.Println("Generating SSH key")
	newkeys, err := tls.NewPrivateKey(ctx, "private-key", &tls.PrivateKeyArgs{
		Algorithm:  pulumi.String("ECDSA"),
		EcdsaCurve: pulumi.String("P521"),
	})
	if err != nil {
		log.Printf("Error creating Pulumi SSH key: %s\n", err)
		return nil, err
	}

	key := SSHKey{
		Public:  newkeys.PublicKeyOpenssh,
		Private: newkeys.PrivateKeyOpenssh,
	}

	log.Println("SSH key generated: ", key.Public, " ", key.Private)

	if os.Getenv("GO_ENV") == "development" {
		_ = key.Private.ApplyT(func(key string) string {
			wd, _ := os.Getwd()
			_ = os.WriteFile(filepath.Join(wd, "hetzner-private-key"), []byte(key), 0644)

			return key
		})
	}

	// get identifier from the middle of the public key
	key.Name = key.Public.ApplyT(func(pkey string) string {
		middle := int(len(pkey) / 2)

		return "pulumi-key-" + pkey[middle-4:middle+4]
	}).(pulumi.StringOutput)

	log.Println("SSH key name: ", key.Name)

	return &key, nil
}
