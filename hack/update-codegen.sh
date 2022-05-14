#!/usr/bin/env ash

apk add --update --no-cache bash
go get -u k8s.io/code-generator@v0.24.0

CLIENTSET_NAME_VERSIONED=clientset \
CLIENTSET_PKG_NAME=clientset \
bash "/go/pkg/mod/k8s.io/code-generator@v0.24.0/generate-groups.sh" deepcopy,client \
  github.com/ofek/csi-gcs/pkg/client github.com/ofek/csi-gcs/pkg/apis \
  "published-volume:v1beta1"
