FROM golang:1.17.6 as builder

RUN apt-get update && apt-get install -y
WORKDIR /go/mx-chain-go
COPY . .
RUN go mod tidy
# Seed node
WORKDIR /go/mx-chain-go/cmd/seednode
RUN go build

# ===== SECOND STAGE ======
FROM ubuntu:22.04
COPY --from=builder /go/mx-chain-go/cmd/seednode /go/mx-chain-go/cmd/seednode

WORKDIR /go/mx-chain-go/cmd/seednode/
EXPOSE 10000
ENTRYPOINT ["./seednode"]
