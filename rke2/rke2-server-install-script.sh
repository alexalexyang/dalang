#!/bin/bash
# https://docs.rke2.io/install/quickstart

curl -sfL https://get.rke2.io | sh -

systemctl enable rke2-server.service

systemctl start rke2-server.service

