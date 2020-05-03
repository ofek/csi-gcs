#!/usr/bin/env bash

CLIENTSET_NAME_VERSIONED=clientset \
CLIENTSET_PKG_NAME=clientset \
bash "$HOME/go/pkg/mod/k8s.io/code-generator@v0.15.12-beta.0/generate-groups.sh" deepcopy,client \
  github.com/ofek/csi-gcs/pkg/client github.com/ofek/csi-gcs/pkg/apis \
  "published-volume:v1beta1" 