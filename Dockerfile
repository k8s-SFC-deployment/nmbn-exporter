ARG ARCH
FROM golang:1.22.1 AS builder

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=$ARCH

WORKDIR /go/src/github.com/k8s-SFC-deployment/nmbn-exporter

COPY . .

RUN go build -o bin/nmbn-exporter cmd/nmbn-exporter/main.go

FROM $ARCH/alpine:3.16

RUN   apk update \
  &&  apk add iptables

WORKDIR /nmbn-exporter
COPY --from=builder /go/src/github.com/k8s-SFC-deployment/nmbn-exporter/bin/nmbn-exporter .

CMD ["./nmbn-exporter"]
