package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	"github.com/urfave/cli"
)

var (
	fileGenHelpTemplate = `NAME:
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
	consensusType = cli.StringFlag{
		Name:  "consensus-type",
		Usage: "Consensus type to be used and for which, private/public keys, to generate",
		Value: "bls",
	}

	initialBalancesSkFileName = "./initialBalancesSk.pem"
	initialNodesSkFileName    = "./initialNodesSk.pem"
)

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = fileGenHelpTemplate
	app.Name = "Key generation Tool"
	app.Version = "v0.0.1"
	app.Usage = "This binary will generate a initialBalancesSk.pem and initialNodesSk.pem, each containing one private key"
	app.Flags = []cli.Flag{consensusType}
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
	var initialBalancesSkFile, initialNodesSkFile *os.File

	defer func() {
		if initialBalancesSkFile != nil {
			err := initialBalancesSkFile.Close()
			if err != nil {
				fmt.Println(err.Error())
			}
		}

		if initialNodesSkFile != nil {
			err := initialNodesSkFile.Close()
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}()

	err := os.Remove(initialBalancesSkFileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	initialBalancesSkFile, err = os.OpenFile(initialBalancesSkFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	err = os.Remove(initialNodesSkFileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	initialNodesSkFile, err = os.OpenFile(initialNodesSkFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	genForBalanceSk := signing.NewKeyGenerator(getSuiteForBalanceSk())
	consensusType := ctx.GlobalString(consensusType.Name)
	genForBlockSigningSk := signing.NewKeyGenerator(getSuiteForBlockSigningSk(consensusType))

	pkHexBalance, skHex, err := getIdentifierAndPrivateKey(genForBalanceSk)
	if err != nil {
		return err
	}

	err = core.SaveSkToPemFile(initialBalancesSkFile, pkHexBalance, skHex)
	if err != nil {
		return err
	}

	pkHexBlockSigning, skHex, err := getIdentifierAndPrivateKey(genForBlockSigningSk)
	if err != nil {
		return err
	}

	err = core.SaveSkToPemFile(initialNodesSkFile, pkHexBlockSigning, skHex)
	if err != nil {
		return err
	}

	fmt.Println("Files generated successfully.")
	fmt.Printf("\tpk for balance:\t%s\n", pkHexBalance)
	ac, _ := addressConverters.NewPlainAddressConverter(32, "")
	adr, _ := ac.CreateAddressFromHex(pkHexBalance)
	bech32, _ := ac.ConvertToBech32(adr)
	fmt.Printf("\tpk in bech32:\t%s\n", bech32)
	fmt.Printf("\tpk for block signing:\t%s\n", pkHexBlockSigning)
	//the block signing PK results in a bech32 string greater than the standard imposed 90 char limit
	//but signing key is longer and can't be mistaken for a txid as it is the case with the balance one
	//ac, _ = addressConverters.NewPlainAddressConverter(128, "")
	//adr, _ = ac.CreateAddressFromHex(pkHexBlockSigning)
	//bech32, _ = ac.ConvertToBech32(adr)
	//fmt.Printf("\tpk for block signing in bech32: %s\n", bech32)

	return nil
}

func getSuiteForBalanceSk() crypto.Suite {
	return kyber.NewBlakeSHA256Ed25519()
}

func getSuiteForBlockSigningSk(consensusType string) crypto.Suite {
	// TODO: A factory which returns the suite according to consensus type should be created in elrond-go project
	// Ex: crypto.NewSuite(consensusType) crypto.Suite
	switch consensusType {
	case "bls":
		return kyber.NewSuitePairingBn256()
	case "bn":
		return kyber.NewBlakeSHA256Ed25519()
	default:
		return nil
	}
}

func getIdentifierAndPrivateKey(keyGen crypto.KeyGenerator) (string, []byte, error) {
	sk, pk := keyGen.GeneratePair()
	skBytes, err := sk.ToByteArray()
	if err != nil {
		return "", nil, err
	}

	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return "", nil, err
	}

	skHex := []byte(hex.EncodeToString(skBytes))
	pkHex := hex.EncodeToString(pkBytes)

	return pkHex, skHex, nil
}
