# Intro

Let's build a full cloud platform using the Kubernetes ecosystem!

[Checklist](checklist.md) tracks the stack and its statuses.

This is just for learning. If you really want to create your own cloud, you should probably try something like [Crossplane](https://www.crossplane.io/).

## Method

Domain-driven, but forgivingly.


## Required

config/config.yaml:
```yaml
pulumiReleaseUrl: https://get.pulumi.com/releases/sdk/pulumi-v3.64.0-darwin-x64.tar.gz
```

config/secrets.yaml:
```yaml
# Allow binary to access Hetzner.
hcloudToken: abcde12345
```

config/hetzner.yaml:
- ```
  serverName: "default-server-name"
  serverType: "cx11"
  serverLocation: "hel1"
  zone: "europe-north1"
  osImage: "ubuntu-22.04"
  ```
- [Hetzner servers and pricing](https://docs.hetzner.com/cloud/servers/overview/#shared-vcpu)
- [Minimum recommendation for RKE2](https://docs.rke2.io/install/requirements#hardware): CX21

rke2/sample-deployment.yaml:
- Used by rke2 `util_test.go` and `deployment_test.go`
- The sample `nginx` one from the [Kubernetes page on deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) works

## Optional

`GO_ENV` environment variable: if set to `development`, will write the private key to file so we can manually log into the deployed server.


## Notes on Pulumi

`pulumi.Run` requires doing `pulumi up` on CLI. For programmatic stuff, use the [automation API](https://www.pulumi.com/docs/guides/automation-api/).

[Set state checkpoints path](https://www.pulumi.com/docs/reference/cli/pulumi_login/): if this is not done, we get error `pulumi_access_token must be set for login during non-interactive cli sessions`

[Use `file://<fs-path>` to specify storage for local back end](https://www.pulumi.com/docs/intro/concepts/state/#using-a-self-managed-backend)

## Note on VPS providers

I've used Hetzner in this project. But, long story short, whenever I ask them questions their replies are really rude and dismissive, and they also ignore my questions. Apparently, they're famous for this. Their policies are also misleading. For instance, their site says we can ask for limits to be raised after a month with them and paying our first invoice, but that’s not true.

I got concerned because my company has used them for client projects. I'm not comfortable recommending Hetzner to clients if the platform is a blocker we can't resolve. So, I finally asked my colleagues to recommend alternatives and they sent these two links:

[European cloud platforms](https://european-alternatives.eu/category/cloud-computing-platforms)

[European VPS hosts](https://european-alternatives.eu/category/vps-virtual-private-server-hosters)

A few colleagues add that Scaleway is not recommended. Seems like people haven't been happy with the way they bring up servers. I don't have details in this area, but I've used them a little in the last year and their documentation is really not great. You should research them more before deciding.

A few colleagues say that OVHcloud is pretty good. I haven't tried them. But when I finally get time to refactor Hetzner out of the picture, I'll try them.

## References

[Inline program example](https://github.com/pulumi/automation-api-examples/blob/main/go/inline_program/main.go)

[Pulumi releases](https://www.pulumi.com/docs/get-started/install/versions/)
