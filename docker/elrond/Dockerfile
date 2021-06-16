FROM golang:1.15.7 as builder

RUN apt-get update && apt-get install -y
WORKDIR /go/elrond-go
COPY . .
RUN GO111MODULE=on go mod vendor

# Elrond node
WORKDIR /go/elrond-go/cmd/node
RUN go build -i -v -ldflags="-w -s -X main.appVersion=$(git describe --tags --long --dirty)"

# Build elrond vm
WORKDIR /go/elrond-go
RUN cp $(go list -f '{{.Dir}}' -m github.com/ElrondNetwork/arwen-wasm-vm/v1_3)/wasmer/libwasmer_linux_amd64.so /lib/libwasmer_linux_amd64.so
RUN go build -o ./arwen github.com/ElrondNetwork/arwen-wasm-vm/cmd/arwen

FROM ubuntu:18.04

COPY --from=builder "/go/elrond-go/cmd/node/node" "/usr/bin/node"
COPY --from=builder "/go/elrond-go/arwen" "/usr/bin/arwen"
COPY --from=builder "/lib/libwasmer_linux_amd64.so" "/lib/libwasmer_linux_amd64.so"

EXPOSE 8080
ENTRYPOINT "node"
