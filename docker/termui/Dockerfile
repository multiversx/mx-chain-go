FROM golang:1.20.7 as builder
RUN apt-get update && apt-get install -y
WORKDIR /go/mx-chain-go
COPY . .
WORKDIR /go/mx-chain-go/cmd/termui
RUN go build -v
RUN cp /go/pkg/mod/github.com/multiversx/$(cat /go/mx-chain-go/go.mod | grep mx-chain-vm-v | sort -n | tail -n -1| awk -F '/' '{print$3}'| sed 's/ /@/g')/wasmer/libwasmer_linux_amd64.so /lib/libwasmer_linux_amd64.so

# ===== SECOND STAGE ======
FROM ubuntu:22.04
COPY --from=builder /go/mx-chain-go/cmd/termui /go/mx-chain-go/cmd/termui
COPY --from=builder "/lib/libwasmer_linux_amd64.so" "/lib/libwasmer_linux_amd64.so"
WORKDIR /go/mx-chain-go/cmd/termui/
ENTRYPOINT ["./termui"]
