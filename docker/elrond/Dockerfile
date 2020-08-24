FROM golang:1.14-stretch as builder
MAINTAINER ElrondNetwork

RUN apt-get update && apt-get install -y
WORKDIR /go/elrond-go
COPY . .
RUN GO111MODULE=on go mod vendor
#Elrond node
WORKDIR /go/elrond-go/cmd/node
RUN go build -i -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"
RUN cp /go/pkg/mod/github.com/!elrond!network/arwen-wasm-vm@$(cat /go/elrond-go/go.mod | grep arwen-wasm-vm | sed 's/.* //')/wasmer/libwasmer_linux_amd64.so /lib/libwasmer_linux_amd64.so
WORKDIR /go/elrond-go
RUN go get github.com/ElrondNetwork/arwen-wasm-vm/cmd/arwen@$(cat /go/elrond-go/go.mod | grep arwen-wasm-vm | sed 's/.* //')
RUN go build -o ./arwen github.com/ElrondNetwork/arwen-wasm-vm/cmd/arwen
RUN cp /go/elrond-go/arwen /go/elrond-go/cmd/node/
WORKDIR /go/elrond-go/cmd/node
# ===== SECOND STAGE ======
FROM ubuntu:18.04
COPY --from=builder "/go/elrond-go/cmd/node" "/go/elrond-go/cmd/node/"
COPY --from=builder "/lib/libwasmer_linux_amd64.so" "/lib/libwasmer_linux_amd64.so"
WORKDIR /go/elrond-go/cmd/node/
EXPOSE 8080
ENTRYPOINT ["/go/elrond-go/cmd/node/node"]
