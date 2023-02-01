FROM golang:1.17.6 as builder
RUN apt-get update && apt-get install -y
WORKDIR /go/mx-chain-go
COPY . .
WORKDIR /go/mx-chain-go/cmd/termui
RUN go build -i -v
RUN cp /go/pkg/mod/github.com/!elrond!network/arwen-wasm-vm@$(cat /go/mx-chain-go/go.mod | grep arwen-wasm-vm | sed 's/.* //' | tail -n 1)/wasmer/libwasmer_linux_amd64.so /lib/libwasmer_linux_amd64.so || :

# ===== SECOND STAGE ======
FROM ubuntu:22.04
COPY --from=builder /go/mx-chain-go/cmd/termui /go/mx-chain-go/cmd/termui
COPY --from=builder "/lib/libwasmer_linux_amd64.so" "/lib/libwasmer_linux_amd64.so"
WORKDIR /go/mx-chain-go/cmd/termui/
ENTRYPOINT ["./termui"]