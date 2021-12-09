FROM golang:1.15.7 as builder
MAINTAINER ElrondNetwork

RUN apt-get update && apt-get install -y
WORKDIR /go/elrond-go
COPY . .
RUN GO111MODULE=on go mod vendor
#Elrond node
WORKDIR /go/elrond-go/cmd/node
RUN go build -i -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"
RUN cp /go/pkg/mod/github.com/!elrond!network/arwen-wasm-vm@$(cat /go/elrond-go/go.mod | grep arwen-wasm-vm | sed 's/.* //' | tail -n 1)/wasmer/libwasmer_linux_amd64.so /lib/libwasmer_linux_amd64.so

WORKDIR /go/elrond-go/cmd/node
# ===== SECOND STAGE ======
FROM ubuntu:18.04
COPY --from=builder "/go/elrond-go/cmd/node" "/go/elrond-go/cmd/node/"
COPY --from=builder "/lib/libwasmer_linux_amd64.so" "/lib/libwasmer_linux_amd64.so"
WORKDIR /go/elrond-go/cmd/node/
EXPOSE 8080
ENTRYPOINT ["/go/elrond-go/cmd/node/node"]
