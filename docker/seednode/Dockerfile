FROM golang:1.13.6 as builder

WORKDIR /go/elrond-go
COPY . .
RUN GO111MODULE=on go mod vendor
# Seed node
WORKDIR /go/elrond-go/cmd/seednode
RUN go build

# ===== SECOND STAGE ======
FROM ubuntu
COPY --from=builder /go/elrond-go/cmd/seednode /go/elrond-go/cmd/seednode

WORKDIR /go/elrond-go/cmd/seednode/
EXPOSE 10000
ENTRYPOINT ["./seednode"]
