FROM golang:1.15.6-alpine3.12 AS build-gcsfuse

ARG gcsfuse_version

RUN apk add --update --no-cache fuse fuse-dev git

WORKDIR ${GOPATH}

# Create Tmp Bin Dir
RUN mkdir /tmp/bin

# Install gcsfuse using the specified version or commit hash
RUN go get -u github.com/googlecloudplatform/gcsfuse
RUN go install github.com/googlecloudplatform/gcsfuse/tools/build_gcsfuse
RUN mkdir /tmp/gcsfuse
RUN build_gcsfuse ${GOPATH}/src/github.com/googlecloudplatform/gcsfuse /tmp/gcsfuse ${gcsfuse_version} -ldflags "-X main.gcsfuseVersion=${gcsfuse_version}"

FROM golang:1.15.6-alpine3.12

RUN apk add --update --no-cache fuse fuse-dev git python3 python3-dev py3-pip bash build-base docker

COPY --from=build-gcsfuse /tmp/gcsfuse/bin/* /usr/local/bin/
COPY --from=build-gcsfuse /tmp/gcsfuse/sbin/* /sbin/

RUN python3 -m pip install --upgrade pip setuptools

RUN mkdir /driver
WORKDIR /driver

COPY requirements.txt /tmp/requirements.txt

RUN python3 -m pip install --upgrade --ignore-installed -r /tmp/requirements.txt

RUN rm /tmp/requirements.txt
