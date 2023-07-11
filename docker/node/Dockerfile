FROM golang:1.17.6 as builder

RUN apt-get update && apt-get install -y
WORKDIR /go/mx-chain-go
COPY . .
RUN go mod tidy
# Multiversx node
WORKDIR /go/mx-chain-go/cmd/node
RUN go build -i -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"
RUN cp /go/pkg/mod/github.com/multiversx/$(cat /go/mx-chain-go/go.mod | grep mx-chain-vm-v | sort -n | tail -n -1| awk -F '/' '{print$3}'| sed 's/ /@/g')/wasmer/libwasmer_linux_amd64.so /lib/libwasmer_linux_amd64.so
RUN cp /go/pkg/mod/github.com/multiversx/$(cat /go/mx-chain-go/go.mod | grep mx-chain-vm-go | sort -n | tail -n -1| awk -F '/' '{print$3}'| sed 's/ /@/g')/wasmer2/libvmexeccapi.so /lib/libvmexeccapi.so

WORKDIR /go/mx-chain-go/cmd/node
# ===== SECOND STAGE ======
FROM ubuntu:22.04
COPY --from=builder "/go/mx-chain-go/cmd/node" "/go/mx-chain-go/cmd/node/"
COPY --from=builder "/lib/libwasmer_linux_amd64.so" "/lib/libwasmer_linux_amd64.so"
COPY --from=builder "/lib/libvmexeccapi.so" "/lib/libvmexeccapi.so"
WORKDIR /go/mx-chain-go/cmd/node/
EXPOSE 8080
ENTRYPOINT ["/go/mx-chain-go/cmd/node/node", "-no-key"]
