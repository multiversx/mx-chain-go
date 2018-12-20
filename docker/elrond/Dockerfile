FROM centos:7
MAINTAINER ElrondNetwork
ARG CACHE_CANCELER=unknown

# prepare
RUN yum install -y wget git curl gcc golang && \
    rm -rf /var/lib/apt/lists/*

# install go
RUN wget https://storage.googleapis.com/golang/go1.11.1.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.11.1.linux-amd64.tar.gz && \
    rm go1.11.1.linux-amd64.tar.gz


# configure go
RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
ENV GOPATH /go
ENV PATH /usr/local/go/bin:/go/bin:$PATH
WORKDIR /go

# install godep
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# fetch repository
WORKDIR src/github.com/ElrondNetwork
RUN ls

RUN git clone -b master https://elrondtester:a73e39f0fe868e584d3c27b3b4d8aacb2e925532@github.com/ElrondNetwork/elrond-go-sandbox
WORKDIR elrond-go-sandbox
RUN >> skipP2PMessengerTests
RUN ls
RUN dep ensure