#
# build
#
FROM golang:1.18.3 AS builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o kubelet-reloader main.go

#
# dist
#
FROM docker.m.daocloud.io/library/ubuntu:20.04

RUN apt-get update -q -y && apt-get install -q -y curl && apt clean all

ENTRYPOINT ["/usr/local/bin/kubelet-reloader"]

COPY --from=builder /workspace/kubelet-reloader /usr/local/bin/kubelet-reloader
