FROM golang:1.15.7 as builder
MAINTAINER MultiversX

RUN apt-get update && apt-get install -y
WORKDIR /go/mx-chain-go
COPY . .
RUN GO111MODULE=on go mod tidy
# Seed node
WORKDIR /go/mx-chain-go/cmd/seednode
RUN go build

# ===== SECOND STAGE ======
FROM ubuntu:18.04
COPY --from=builder /go/mx-chain-go/cmd/seednode /go/mx-chain-go/cmd/seednode

WORKDIR /go/mx-chain-go/cmd/seednode/
EXPOSE 10000
ENTRYPOINT ["./seednode"]
