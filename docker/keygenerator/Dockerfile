FROM golang:1.23.6 AS builder

RUN apt-get update && apt-get install -y
WORKDIR /go/mx-chain-go
COPY . .
RUN go mod tidy
# Keygenerator node
WORKDIR /go/mx-chain-go/cmd/keygenerator
RUN go build

# ===== SECOND STAGE ======
FROM ubuntu:22.04
COPY --from=builder /go/mx-chain-go/cmd/keygenerator /go/mx-chain-go/cmd/keygenerator

ENTRYPOINT ["/go/mx-chain-go/cmd/keygenerator/keygenerator"]
