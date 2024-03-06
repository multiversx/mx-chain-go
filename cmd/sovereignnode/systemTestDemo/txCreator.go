package main

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/blockchain/cryptoProvider"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/interactors"
)

var (
	suite  = ed25519.NewEd25519()
	keyGen = signing.NewKeyGenerator(suite)
)

func sendAmount(
	address core.AddressHandler,
	privateKey []byte,
	value string,
	receiverAddress string,
) (string, error) {
	args := blockchain.ArgsProxy{
		ProxyURL:            proxyUrl,
		CacheExpirationTime: time.Minute,
		EntityType:          core.Proxy,
	}

	proxy, err := blockchain.NewProxy(args)
	if err != nil {
		return "", err
	}

	txBuilder, err := builders.NewTxBuilder(cryptoProvider.NewSigner())
	if err != nil {
		return "", err
	}

	ti, err := interactors.NewTransactionInteractor(proxy, txBuilder)
	if err != nil {
		return "", err
	}

	networkConfig, err := proxy.GetNetworkConfig(context.Background())
	if err != nil {
		return "", err
	}

	transactionArguments, _, err := proxy.GetDefaultTransactionArguments(context.Background(), address, networkConfig)
	if err != nil {
		return "", err
	}

	transactionArguments.Value = value
	transactionArguments.GasLimit = txGasLimit
	transactionArguments.Receiver = receiverAddress

	holder, err := cryptoProvider.NewCryptoComponentsHolder(keyGen, privateKey)
	if err != nil {
		return "", err
	}

	err = ti.ApplySignature(holder, nil)
	if err != nil {
		return "", err
	}

	hash, err := proxy.SendTransaction(context.Background(), nil)
	if err != nil {
		return "", err
	}
	log.Info("sent transaction", "tx hash", hash)

	log.Info("waiting 5 rounds...")
	time.Sleep(time.Millisecond * time.Duration(networkConfig.RoundDuration*5))

	return hash, nil
}
