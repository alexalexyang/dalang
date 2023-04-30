# Intro

I want to set up a k8s cluster with:

- High availability and fault tolerance
  - 3 nodes

- Security
  - RKE2
  - HashiCorp Vault
  - Static analysis
  - SPIFFE

- Monitoring and logging
  - Fluent bit and/or FluentD
  - Prometheus

- CI/CD
  - FluxCD

- A few apps in it
  - Keycloak?
  - My other apps

- Infrastructure orchestrator
  - [Crossplane](https://www.crossplane.io/)


## Method

Domain-driven design, but loosely for now.


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


## Steps

- Install Pulumi

- Create 1 server in Hetzner

- Create multiple servers in Hetzner

- Create a network in Hetzner
  - Done. But not sure if needed. This was just an experiment.

- Create a load balancer

- Install RKE2

- Install Rancher

- Delete Pulumi?


## Notes

`pulumi.Run` requires doing `pulumi up` on CLI. For programmatic stuff, use the [automation API](https://www.pulumi.com/docs/guides/automation-api/).

[Set state checkpoints path](https://www.pulumi.com/docs/reference/cli/pulumi_login/): if this is not done, we get error `pulumi_access_token must be set for login during non-interactive cli sessions`

[Use `file://<fs-path>` to specify storage for local back end](https://www.pulumi.com/docs/intro/concepts/state/#using-a-self-managed-backend)

## References

[Inline program example](https://github.com/pulumi/automation-api-examples/blob/main/go/inline_program/main.go)

[Pulumi releases](https://www.pulumi.com/docs/get-started/install/versions/)