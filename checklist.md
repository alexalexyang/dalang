# Checklist

## Provisioning

### [Pulumi](https://www.pulumi.com/)

- [x] Install it
- [x] Create local back end

## Platform

### [Hetzner](https://www.hetzner.com/)

- Server
  - [x] Integration tests
  - [x] Deploy multiple servers of same type
  - [x] SSH key
  - [ ] Disable root login
  - [ ] Close all unused ports
  - [ ] Further server hardening

- Load balancer

- Network
  - Not sure if needed. Was an experiment.
  - [x] Integration tests
  - [x] Deploy network

## Kubernetes distribution

### [RKE2](https://docs.rke2.io/)

- RKE2 server
  - [x] Integration test
  - [x] Install in remote host

- RKE2 agent
  - [x] Tests
  - [x] Install in remote host

## GUI

### [Rancher](https://www.rancher.com/)

- Not sure if needed

## Logging

### Fluent bit

### FluentD

### Prometheus

## Security

### HashiCorp Vault

### Static analysis

### SPIFFE

### [Firecracker](https://firecracker-microvm.github.io/)

## CI/CD

### FluxCD

## Infrastructure orchestrator

### [Crossplane](https://www.crossplane.io/)

### High availability and fault tolerance

- [ ] 3 nodes

## Sample deployments

Maybe keycloak.