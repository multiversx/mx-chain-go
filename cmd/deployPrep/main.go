package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/urfave/cli"
)

var (
	deployPrepHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
VERSION:
   {{.Version}}
   {{end}}
`
	numPairs = cli.IntFlag{
		Name:  "num-pairs",
		Usage: "Number of private/public keys to generate",
		Value: 21,
	}
	mintValue = cli.Uint64Flag{
		Name:  "mint-value",
		Usage: "Initial minting for all public keys generated",
		Value: 1000000000,
	}

	privKeysFilename = "./privkeys.pem"
	genesisFilename  = "./genesis.json"

	errInvalidNumPrivPubKeys = errors.New("invalid number of private/public keys to generate")
	errInvalidMintValue      = errors.New("invalid mint value for generated public keys")
)

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = deployPrepHelpTemplate
	app.Name = "Deploy Preparation Tool"
	app.Version = "v0.0.1"
	app.Usage = "This binary will generate a privkeys.pem and genesis.json files to be used in mass deployment"
	app.Flags = []cli.Flag{numPairs, mintValue}
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return generateFiles(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func generateFiles(ctx *cli.Context) error {
	num := ctx.GlobalInt(numPairs.Name)
	if num < 1 {
		return errInvalidNumPrivPubKeys
	}

	initialMint := ctx.GlobalUint64(mintValue.Name)
	if initialMint < 0 {
		return errInvalidMintValue
	}

	privKeysFile, err := os.OpenFile(
		privKeysFilename,
		os.O_CREATE|os.O_WRONLY,
		0666)
	if err != nil {
		return err
	}
	defer func() {
		err = privKeysFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	genesisFile, err := os.OpenFile(
		genesisFilename,
		os.O_CREATE|os.O_WRONLY,
		0666)
	if err != nil {
		return err
	}
	defer func() {
		err = genesisFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	genesis := &sharding.Genesis{
		StartTime:          0,
		RoundDuration:      6000,
		ConsensusGroupSize: uint32(num),
		MinNodesPerShard:   uint32(num),
		ElasticSubrounds:   true,
		InitialNodes:       make([]*sharding.InitialNode, num),
	}

	suite := kyber.NewBlakeSHA256Ed25519()

	for i := 0; i < num; i++ {
		generator := signing.NewKeyGenerator(suite)
		sk, pk := generator.GeneratePair()
		skBytes, err := sk.ToByteArray()
		if err != nil {
			return err
		}

		pkBytes, err := pk.ToByteArray()
		if err != nil {
			return err
		}

		skHex := []byte(hex.EncodeToString(skBytes))
		pkHex := hex.EncodeToString(pkBytes)

		genesis.InitialNodes[i] = &sharding.InitialNode{
			PubKey:  pkHex,
			Balance: fmt.Sprintf("%d", initialMint),
		}

		err = core.SaveSkToPemFile(privKeysFile, pkHex, skHex)
		if err != nil {
			return err
		}
	}

	genesisBuff, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		return err
	}

	_, err = genesisFile.Write(genesisBuff)
	if err != nil {
		return err
	}

	fmt.Println("Files generated successfully!")
	return nil
}
