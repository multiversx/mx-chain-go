FROM golang:1.20.7 AS builder

RUN apt-get update && apt-get upgrade -y
WORKDIR /go/mx-chain-go
COPY . .
RUN go mod tidy
# Multiversx node
WORKDIR /go/mx-chain-go/cmd/node
RUN go build -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"

RUN mkdir -p /lib_amd64 /lib_arm64

RUN cp /go/pkg/mod/github.com/multiversx/$(cat /go/mx-chain-go/go.mod | grep mx-chain-vm-v | sort -n | tail -n -1 | awk -F '/' '{print$3}' | sed 's/ /@/g')/wasmer/libwasmer_linux_amd64.so /lib_amd64/
RUN cp /go/pkg/mod/github.com/multiversx/$(cat /go/mx-chain-go/go.mod | grep mx-chain-vm-go | sort -n | tail -n -1 | awk -F '/' '{print$3}' | sed 's/ /@/g')/wasmer2/libvmexeccapi.so /lib_amd64/

RUN cp /go/pkg/mod/github.com/multiversx/$(cat /go/mx-chain-go/go.mod | grep mx-chain-vm-v | sort -n | tail -n -1 | awk -F '/' '{print$3}' | sed 's/ /@/g')/wasmer/libwasmer_linux_arm64_shim.so /lib_arm64/
RUN cp /go/pkg/mod/github.com/multiversx/$(cat /go/mx-chain-go/go.mod | grep mx-chain-vm-go | sort -n | tail -n -1 | awk -F '/' '{print$3}' | sed 's/ /@/g')/wasmer2/libvmexeccapi_arm.so /lib_arm64/

# ===== SECOND STAGE ======
FROM ubuntu:22.04
ARG TARGETARCH
RUN apt-get update && apt-get upgrade -y
COPY --from=builder "/go/mx-chain-go/cmd/node/node" "/go/mx-chain-go/cmd/node/"

# Copy architecture-specific files
COPY --from=builder "/lib_${TARGETARCH}/*" "/lib/"

WORKDIR /go/mx-chain-go/cmd/node/
EXPOSE 8080
ENTRYPOINT ["/go/mx-chain-go/cmd/node/node"]
