FROM golang:1.20.7 AS builder
RUN apt-get update && apt-get install -y
WORKDIR /go/mx-chain-go
COPY . .
WORKDIR /go/mx-chain-go/cmd/termui
RUN go build -v
RUN mkdir -p /lib_amd64 /lib_arm64
RUN cp /go/pkg/mod/github.com/multiversx/$(cat /go/mx-chain-go/go.mod | grep mx-chain-vm-v | sort -n | tail -n -1| awk -F '/' '{print$3}'| sed 's/ /@/g')/wasmer/libwasmer_linux_amd64.so /lib_amd64/
RUN cp /go/pkg/mod/github.com/multiversx/$(cat /go/mx-chain-go/go.mod | grep mx-chain-vm-v | sort -n | tail -n -1 | awk -F '/' '{print$3}' | sed 's/ /@/g')/wasmer/libwasmer_linux_arm64_shim.so /lib_arm64/


# ===== SECOND STAGE ======
FROM ubuntu:22.04
ARG TARGETARCH
COPY --from=builder /go/mx-chain-go/cmd/termui /go/mx-chain-go/cmd/termui

# Copy architecture-specific files
COPY --from=builder "/lib_${TARGETARCH}/*" "/lib/"

WORKDIR /go/mx-chain-go/cmd/termui/
ENTRYPOINT ["./termui"]
