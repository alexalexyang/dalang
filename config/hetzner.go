package config

import (
	"dalang/util"

	_ "embed"
	"log"

	"gopkg.in/yaml.v3"
)

//go:embed hetzner.yaml
var hetznerYaml []byte

type hetznerStruct struct {
	ServerName     string `yaml:"serverName"`
	ServerType     string `yaml:"serverType"`
	ServerLocation string `yaml:"serverLocation"`
	Zone           string `yaml:"zone"`
	OsImage        string `yaml:"osImage"`
}

var HetznerConfig = hetznerStruct{}

func SetHetznerConfig() {
	// -- Get config from config.yaml --
	err := yaml.Unmarshal(hetznerYaml, &HetznerConfig)
	util.IfFatalErr(err)

	if HetznerConfig.ServerName == "" {
		log.Fatal("Not found: Hetzner server name")
	}

	log.Println("Found: Hetzner server name")

	if HetznerConfig.ServerType == "" {
		log.Fatal("Not found: Hetzner server type")
	}

	log.Println("Found: Hetzner server type")

	if HetznerConfig.ServerLocation == "" {
		log.Fatal("Not found: Hetzner server location")
	}

	log.Println("Found: Hetzner server location")

	if HetznerConfig.Zone == "" {
		log.Fatal("Not found: Hetzner server zone")
	}

	log.Println("Found: Hetzner server zone")

	if HetznerConfig.OsImage == "" {
		log.Fatal("Not found: Hetzner server OS image")
	}

	log.Println("Found: Hetzner server OS image")
}
