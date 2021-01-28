FROM golang:1.15.7 as builder
MAINTAINER ElrondNetwork

RUN apt-get update && apt-get install -y
WORKDIR /go/elrond-go
COPY . .
RUN GO111MODULE=on go mod vendor
# Seed node
WORKDIR /go/elrond-go/cmd/seednode
RUN go build

# ===== SECOND STAGE ======
FROM ubuntu:18.04
COPY --from=builder /go/elrond-go/cmd/seednode /go/elrond-go/cmd/seednode

WORKDIR /go/elrond-go/cmd/seednode/
EXPOSE 10000
ENTRYPOINT ["./seednode"]
