FROM golang:1.17.6 as builder
MAINTAINER MultiversX

RUN apt-get update && apt-get install -y
WORKDIR /go/mx-chain-go
COPY . .
RUN GO111MODULE=on go mod tidy
# Multiversx node
WORKDIR /go/mx-chain-go/cmd/node
RUN go build -i -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"
RUN cp /go/pkg/mod/github.com/multiversx/wasm-vm@$(cat /go/mx-chain-go-go/go.mod | grep wasm-vm | sed 's/.* //' | tail -n 1)/wasmer/libwasmer_linux_amd64.so /lib/libwasmer_linux_amd64.so

WORKDIR /go/mx-chain-go/cmd/node
# ===== SECOND STAGE ======
FROM ubuntu:18.04
COPY --from=builder "/go/mx-chain-go/cmd/node" "/go/mx-chain-go/cmd/node/"
COPY --from=builder "/lib/libwasmer_linux_amd64.so" "/lib/libwasmer_linux_amd64.so"
WORKDIR /go/mx-chain-go/cmd/node/
EXPOSE 8080
ENTRYPOINT ["/go/mx-chain-go/cmd/node/node"]
