FROM golang:1.13.6-alpine3.11 AS build

ARG version
ARG gcsfuse_version
ARG global_ldflags
ARG upx_flags

ENV DRIVER ${GOPATH}/src/github.com/ofek/csi-gcs/
ENV UPX ${upx_flags}

RUN apk add --update --no-cache fuse fuse-dev git upx

WORKDIR ${GOPATH}

# Create Tmp Bin Dir
RUN mkdir /tmp/bin

# Install gcsfuse using the specified version or commit hash
RUN go get -u github.com/googlecloudplatform/gcsfuse
RUN go install github.com/googlecloudplatform/gcsfuse/tools/build_gcsfuse
RUN mkdir /tmp/gcsfuse
RUN build_gcsfuse ${GOPATH}/src/github.com/googlecloudplatform/gcsfuse /tmp/gcsfuse ${gcsfuse_version} -ldflags "all=${global_ldflags}" -ldflags "-X main.gcsfuseVersion=${gcsfuse_version} ${global_ldflags}"

# We don't need mount(8) compatibility as we call the binary directly, so only copy that
RUN mv /tmp/gcsfuse/bin/gcsfuse /tmp/bin/gcsfuse

RUN mkdir -p ${DRIVER}
WORKDIR ${DRIVER}

COPY go.mod go.sum ${DRIVER}
COPY cmd ${DRIVER}/cmd
COPY pkg ${DRIVER}/pkg

# Build the driver
RUN go build -o /tmp/bin/driver -ldflags "all=${global_ldflags}" -ldflags "-X github.com/ofek/csi-gcs/pkg/driver.driverVersion=${version} ${global_ldflags}" ./cmd

# Compress the binaries
RUN if [ "${UPX}" != "" ]; then \
        upx /tmp/bin/driver \
     && upx /tmp/bin/gcsfuse; \
    fi

FROM alpine:3.11

# https://github.com/opencontainers/image-spec/blob/master/annotations.md
LABEL "org.opencontainers.image.authors"="Ofek Lev <ofekmeister@gmail.com>"
LABEL "org.opencontainers.image.description"="CSI driver for Google Cloud Storage"
LABEL "org.opencontainers.image.licenses"="Apache-2.0 OR MIT"
LABEL "org.opencontainers.image.source"="https://github.com/ofek/csi-gcs"
LABEL "org.opencontainers.image.title"="csi-gcs"

RUN apk add --update --no-cache ca-certificates fuse && rm -rf /tmp/*

# Allow non-root users to specify the allow_other or allow_root mount options
RUN echo "user_allow_other" > /etc/fuse.conf

# Create directories for mounts and temporary key storage
RUN mkdir -p /var/lib/kubelet/pods /tmp/keys

WORKDIR /

ENTRYPOINT ["/usr/local/bin/driver"]

# Copy the binaries
COPY --from=build /tmp/bin/* /usr/local/bin/
