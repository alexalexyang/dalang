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



## Optional

`GO_ENV` environment variable: if set to `development`, will write the private key to file so we can manually log into the deployed server.


## Notes

`pulumi.Run` requires doing `pulumi up` on CLI. For programmatic stuff, use the [automation API](https://www.pulumi.com/docs/guides/automation-api/).

[Set state checkpoints path](https://www.pulumi.com/docs/reference/cli/pulumi_login/): if this is not done, we get error `pulumi_access_token must be set for login during non-interactive cli sessions`

[Use `file://<fs-path>` to specify storage for local back end](https://www.pulumi.com/docs/intro/concepts/state/#using-a-self-managed-backend)

## References

[Inline program example](https://github.com/pulumi/automation-api-examples/blob/main/go/inline_program/main.go)

[Pulumi releases](https://www.pulumi.com/docs/get-started/install/versions/)
