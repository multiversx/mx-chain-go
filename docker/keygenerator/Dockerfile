FROM golang:1.14-stretch as builder
MAINTAINER ElrondNetwork

RUN apt-get update && apt-get install -y
WORKDIR /go/elrond-go
COPY . .
RUN GO111MODULE=on go mod vendor
# Keygenerator node
WORKDIR /go/elrond-go/cmd/keygenerator
RUN go build

# ===== SECOND STAGE ======
FROM ubuntu:18.04
COPY --from=builder /go/elrond-go/cmd/keygenerator /go/elrond-go/cmd/keygenerator

WORKDIR /go/elrond-go/cmd/keygenerator/
ENTRYPOINT ["./keygenerator"]
