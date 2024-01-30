package main

import (
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/marshal/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	addressLen = 32
	hashSize   = 32
)

var (
	log                = logger.GetOrCreate("sovereign-mock-notifier")
	hasher             = blake2b.NewBlake2b()
	marshaller, _      = factory.NewMarshalizer("gogo protobuf")
	pubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(addressLen, "erd")

	wsURL             = "localhost:22111"
	grpcAddress       = ":8085"
	subscribedAddress = "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
)
