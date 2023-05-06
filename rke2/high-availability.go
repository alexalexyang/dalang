package rke2

import (
	"dalang/config"
	hetzner "dalang/hetzner/deploy-server"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Returns serverIp, serverToken, error
func DeployHighAvailability(ctx *pulumi.Context, numServers int, numAgents int) (*string, *string, error) {

	if numServers > 1 {
		return nil, nil, fmt.Errorf("numServers must be 1 for now as I haven't implemented HA for servers yet")
	}

	numHosts := numServers + numAgents
	hostInfoSlice, _, err := hetzner.DeployServers1SSHKey(ctx, numHosts)
	if err != nil {
		log.Println("Error with DeployNetworkFunc: ", err)
		return nil, nil, err
	}

	for idx, hostInfo := range hostInfoSlice {
		ctx.Export(fmt.Sprintf("%s-server-%d-connect-info", config.Config.ProjectName, idx+1), hostInfo.ConnectArgs)
	}

	serverHostsSlice := hostInfoSlice[:numServers]

	serverIp, serverToken, err := setupServerHosts(ctx, serverHostsSlice, numServers)
	if err != nil {
		log.Println("Error setting up hosts: ", err)
		return nil, nil, err
	}

	if os.Getenv("GO_ENV") == "development" {
		cmdRes := exec.Command("sh", "-c", "chmod 777 hetzner-private-key")
		log.Println("cmdRes: ", cmdRes)

		log.Printf("SSH command for server: ssh-keygen -R %s && ssh -o \"StrictHostKeyChecking no\" -i rke2/hetzner-private-key root@%s", *serverIp, *serverIp)
	}

	if numAgents == 0 {
		log.Println("numAgents is 0, no agents to set up")
		return serverIp, serverToken, nil
	}

	agentHostsSlice := hostInfoSlice[numServers:]

	setupAgentHosts(ctx, agentHostsSlice, numServers, *serverIp, serverToken)

	return serverIp, serverToken, nil
}

// Returns serverIp, serverToken, error
func setupServerHosts(ctx *pulumi.Context, hosts []hetzner.Host, numServers int) (*string, *string, error) {

	log.Println("Setting up first host")

	host1 := hosts[0]

	installServerRes, err := InstallServer(ctx, &host1.ConnectArgs, []pulumi.Resource{host1.Server})
	if err != nil {
		log.Println("Error installing RKE2 server: ", err)
		return nil, nil, err
	}

	serverToken, err := GetRke2ServerToken(ctx, &host1.ConnectArgs, installServerRes)
	if err != nil {
		log.Println("Error getting RKE2 server token: ", err)
		return nil, nil, err
	}

	serverChan := make(chan string)
	host1.Server.Ipv4Address.ApplyT(func(ip string) string {
		serverChan <- ip
		return ip
	})

	serverIp := <-serverChan
	close(serverChan)

	log.Println("server token: ", *serverToken)
	log.Println("server ip: ", serverIp)

	if numServers == 1 {
		log.Println("Only 1 server, no remaining hosts to set up")
		return &serverIp, serverToken, nil
	}

	log.Println("Setting up remaining hosts")

	remainingHosts := hosts[1:]

	for i := 1; i < len(remainingHosts)+1; i++ {
		log.Println("Setting up host ", i+1)

		host := remainingHosts[i-1]

		// Todo: set up RKE2 secondary host
		log.Println("Set up RKE2 server in host: ", host)
	}

	return &serverIp, serverToken, nil
}

func setupAgentHosts(ctx *pulumi.Context, hosts []hetzner.Host, numServers int, serverIp string, serverToken *string) error {

	for _, host := range hosts {

		agentChan := make(chan string)
		host.Server.Ipv4Address.ApplyT(func(ip string) string {
			log.Println("agent ip in ApplyT: ", ip)
			agentChan <- ip
			return ip
		})

		agentIp := <-agentChan
		close(agentChan)

		log.Println("Setting up RKE2 agent in host: ", agentIp)

		if os.Getenv("GO_ENV") == "development" {

			cmdRes := exec.Command("sh", "-c", "chmod 777 hetzner-private-key")
			log.Println("cmdRes: ", cmdRes)

			log.Printf("SSH command for agent: ssh-keygen -R %s && ssh -o \"StrictHostKeyChecking no\" -i rke2/hetzner-private-key root@%s", agentIp, agentIp)
		}

		runScriptRes, err := InstallAgent(ctx, agentIp, &host.ConnectArgs, []pulumi.Resource{host.Server})
		if err != nil {
			log.Println("Error installing RKE2 agent: ", err)
			return err
		}

		_, agentStatus, err := StartAgent(ctx, agentIp, &host.ConnectArgs, []pulumi.Resource{runScriptRes}, serverIp, *serverToken)
		if err != nil {
			log.Println("Error starting RKE2 agent: ", err)
			return err
		}

		log.Println("Status for agent ", agentIp, ":", *agentStatus)
	}

	return nil
}
