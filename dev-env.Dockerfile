FROM golang:1.13.6-alpine3.11 AS build-gcsfuse

ARG gcsfuse_version

RUN apk add --update --no-cache fuse fuse-dev git upx

WORKDIR ${GOPATH}

# Create Tmp Bin Dir
RUN mkdir /tmp/bin

# Install gcsfuse using the specified version or commit hash
RUN go get -u github.com/googlecloudplatform/gcsfuse
RUN go install github.com/googlecloudplatform/gcsfuse/tools/build_gcsfuse
RUN mkdir /tmp/gcsfuse
RUN build_gcsfuse ${GOPATH}/src/github.com/googlecloudplatform/gcsfuse /tmp/gcsfuse ${gcsfuse_version} -ldflags "-X main.gcsfuseVersion=${gcsfuse_version}"

FROM golang:1.13.6-alpine3.11

RUN apk add --update --no-cache fuse fuse-dev git upx python3 python3-dev py3-pip bash build-base docker

COPY --from=build-gcsfuse /tmp/gcsfuse/bin/gcsfuse /usr/local/bin/gcsfuse

RUN python3 -m pip install --upgrade pip setuptools

RUN mkdir /driver
WORKDIR /driver

COPY requirements.txt /tmp/requirements.txt

RUN python3 -m pip install --upgrade -r /tmp/requirements.txt

RUN rm /tmp/requirements.txt
