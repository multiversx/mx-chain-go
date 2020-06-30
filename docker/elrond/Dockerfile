FROM golang:1.13.6 as builder

WORKDIR /go/elrond-go
COPY . .
RUN GO111MODULE=on go mod vendor
#Elrond node
WORKDIR /go/elrond-go/cmd/node
RUN go build -i -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"

# ===== SECOND STAGE ======
FROM ubuntu
COPY --from=builder "/go/elrond-go/cmd/node" "/go/elrond-go/cmd/node/"
#TODO: make copy independent from specific arwen-wasm-wm version
COPY --from=builder "/go/pkg/mod/github.com/!elrond!network/arwen-wasm-vm@v0.3.19/wasmer/libwasmer_linux_amd64.so" "/go/pkg/mod/github.com/!elrond!network/arwen-wasm-vm@v0.3.19/wasmer/libwasmer_linux_amd64.so"
WORKDIR /go/elrond-go/cmd/node/
EXPOSE 8080
ENTRYPOINT ["/go/elrond-go/cmd/node/node"]
