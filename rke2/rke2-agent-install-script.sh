#!/bin/bash
# https://docs.rke2.io/install/quickstart

curl -sfL https://get.rke2.io | INSTALL_RKE2_TYPE="agent" sh -

systemctl enable rke2-agent.service

# Configure rke2 agent service

# mkdir -p /etc/rancher/rke2/
# vim /etc/rancher/rke2/config.yaml

# Content for config.yaml
# server: https://<server>:9345
# token: <token from server node>

systemctl start rke2-agent.service

