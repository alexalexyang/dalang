#!/bin/bash
# https://docs.rke2.io/install/quickstart

curl -sfL https://get.rke2.io | INSTALL_RKE2_TYPE="agent" sh -

systemctl enable rke2-agent.service
