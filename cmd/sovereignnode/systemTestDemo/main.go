package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/urfave/cli"
)

var (
	log = logger.GetOrCreate("checker")
)

const (
	proxyPem          = "scripts/testnet/testnet-local/sandbox/proxy/config/walletKey.pem"
	subscribedAddress = "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
	valueToSend       = "5000000000000000"
	proxyUrl          = "http://127.0.0.1:7960"
	txGasLimit        = 55141500
)

func main() {
	app := cli.NewApp()
	app.Name = "Tool to send a move balance transaction to one of the subscribed sovereign addresses"
	app.Usage = "This tool only works if a local testnet, a sovereign shard and a sovereign notifier are started . See README.md"
	app.Action = func(c *cli.Context) error {
		return startProcess()
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
		return
	}
}

func startProcess() error {
	address, privateKey, err := getAddressAndSK(proxyPem)
	if err != nil {
		return err
	}

	txHash, err := sendAmount(address, privateKey, valueToSend, subscribedAddress)
	if err != nil {
		return err
	}

	resp, err := http.Get(fmt.Sprintf("%s/transaction/%s?withResults=true", proxyUrl, txHash))
	if err != nil {
		return err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Warn("could not close response body", "error", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Info("finished successfully", "tx result from api", string(body))
	return nil
}
