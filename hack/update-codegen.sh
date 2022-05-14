#!/usr/bin/env ash

# docker run --rm -v $PWD:/go/src/github.com/ofek/csi-gcs -w /go/src/github.com/ofek/csi-gcs golang:1.18.2-alpine3.15 ./hack/update-codegen.sh
apk add --update --no-cache fuse fuse-dev git bash
go get -u k8s.io/code-generator@v0.24.0

CLIENTSET_NAME_VERSIONED=clientset \
CLIENTSET_PKG_NAME=clientset \
bash "/go/pkg/mod/k8s.io/code-generator@v0.24.0/generate-groups.sh" deepcopy,client \
  github.com/ofek/csi-gcs/pkg/client github.com/ofek/csi-gcs/pkg/apis \
  "published-volume:v1beta1"
