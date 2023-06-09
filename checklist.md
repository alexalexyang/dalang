# Checklist

## Provisioning

### [Pulumi](https://www.pulumi.com/)

- [x] Install it
- [x] Create local back end

## Platform

### [Hetzner](https://www.hetzner.com/)

- Server
  - [x] Integration tests, positive cases only
  - [x] Deploy single server
  - [x] Deploy multiple servers of same type with one SSH key
  - [x] SSH key
  - [ ] Disable root login
  - [ ] Close all unused ports
  - [ ] Further server hardening

- Load balancer

- Network
  - Not sure if needed. Was an experiment.
  - [x] Integration tests, positive cases only
  - [x] Deploy network

## Kubernetes distribution

### [RKE2](https://docs.rke2.io/)

- RKE2 server
  - [x] Integration tests, positive cases only
  - [x] Install in remote host

- RKE2 agent
  - [x] Integration tests, positive cases only
  - [x] Install in remote host

- High availability
  - [x] Integration tests, positive cases only
  - [ ] Server
  - [x] Agent

- App test
  - [x] Integration tests, positive cases only
  - [x] Sample deployment
    - [x] Manually try port-forwarding
  - [x] Sample service
    - [x] Manually try port-forwarding
  - [x] Ingress
  - [ ] Ingress controller
  - [ ] Ingress class
  - [ ] Load balancer

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