package config

import (
	"dalang/util"

	_ "embed"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed secrets.yaml
var secretsYaml []byte

//go:embed config.yaml
var configYaml []byte

type configStruct struct {
	ProjectRootDir   string
	PulumiReleaseUrl string `yaml:"pulumiReleaseUrl"`
	ProjectName      string `yaml:"projectName"`
}

var Config = configStruct{}

type secretsStruct struct {
	// -- Hetzner Cloud --
	HcloudToken string `yaml:"hcloudToken"`
}

func SetSecrets() {
	// -- Get secrets from secrets.yaml --
	secrets := secretsStruct{}

	err := yaml.Unmarshal(secretsYaml, &secrets)
	util.IfFatalErr(err)

	if secrets.HcloudToken == "" {
		log.Println("Not found: Hetzner Cloud token")
	} else {
		log.Println("Found: Hetzner Cloud token")

		// -- Set HCLOUD token from secrets as env var --
		err = os.Setenv("HCLOUD_TOKEN", secrets.HcloudToken)
		util.IfFatalErr(err)
	}
}

func SetConfig() {
	// -- Get config from config.yaml --
	err := yaml.Unmarshal(configYaml, &Config)
	util.IfFatalErr(err)

	if Config.PulumiReleaseUrl == "" {
		log.Fatal("Not found: Pulumi release URL")
	}

	log.Println("Found: Pulumi release URL")

	if Config.ProjectName == "" {
		log.Fatal("Not found: Project name")
	}

	log.Println("Found: Pulumi name")

	projRootDir, err := util.GetProjectRootDir()
	util.IfFatalErr(err)

	Config.ProjectRootDir = *projRootDir

	if Config.ProjectRootDir == "" {
		log.Fatal("Not found: project root directory path")
	}
	log.Println("Found: project root directory path")
}
