FROM golang:1.18.2-alpine3.15 AS build-gcsfuse

ARG gcsfuse_version
ARG global_ldflags

RUN apk add --update --no-cache fuse fuse-dev git

WORKDIR ${GOPATH}/src/github.com/googlecloudplatform/gcsfuse

# Create Tmp Bin Dir
RUN mkdir /tmp/bin

# Install gcsfuse using the specified version or commit hash
RUN git clone https://github.com/googlecloudplatform/gcsfuse . && git checkout "v${gcsfuse_version}"
RUN go install ./tools/build_gcsfuse
RUN mkdir /tmp/gcsfuse
RUN build_gcsfuse . /tmp/gcsfuse ${gcsfuse_version} -ldflags "all=${global_ldflags}" -ldflags "-X main.gcsfuseVersion=${gcsfuse_version} ${global_ldflags}"

FROM alpine:3.15

# https://github.com/opencontainers/image-spec/blob/master/annotations.md
LABEL "org.opencontainers.image.authors"="Ofek Lev <ofekmeister@gmail.com>"
LABEL "org.opencontainers.image.description"="CSI driver for Google Cloud Storage"
LABEL "org.opencontainers.image.licenses"="Apache-2.0 OR MIT"
LABEL "org.opencontainers.image.source"="https://github.com/ofek/csi-gcs"
LABEL "org.opencontainers.image.title"="csi-gcs"

RUN apk add --update --no-cache ca-certificates fuse tini && rm -rf /tmp/*

# Allow non-root users to specify the allow_other or allow_root mount options
RUN echo "user_allow_other" > /etc/fuse.conf

# Create directories for mounts and temporary key storage
RUN mkdir -p /var/lib/kubelet/pods /tmp/keys

WORKDIR /

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/driver"]

# Copy the binaries
COPY --from=build-gcsfuse /tmp/gcsfuse/bin/* /usr/local/bin/
COPY --from=build-gcsfuse /tmp/gcsfuse/sbin/* /sbin/
COPY bin/driver /usr/local/bin/
