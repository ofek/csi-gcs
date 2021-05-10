FROM golang:1.15.6-alpine3.12 as builder

ARG global_ldflags

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY cmd/webhook cmd/webhook
COPY pkg/ pkg/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build \
  -ldflags "all=${global_ldflags}" -a -o wbkserver ./cmd/webhook/...

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/wbkserver .
USER 65532:65532

ENTRYPOINT ["/wbkserver"]