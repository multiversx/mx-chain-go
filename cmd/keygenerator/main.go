package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/urfave/cli"
)

type cfg struct {
	numKeys int
	keyType string
}

const keysFolderPattern = "node-%d"
const blsPubkeyLen = 96
const txSignPubkeyLen = 32

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

	// numKeys defines a flag for setting how many keys should generate
	numKeys = cli.IntFlag{
		Name:        "num-keys",
		Usage:       "How many keys should generate. Example: 1",
		Value:       1,
		Destination: &argsConfig.numKeys,
	}

	// keyType defines a flag for setting what keys should generate
	keyType = cli.StringFlag{
		Name:        "key-type",
		Usage:       "What king of keys should generate. Available options: validator, wallet, both",
		Value:       "validator",
		Destination: &argsConfig.keyType,
	}

	argsConfig = &cfg{}

	walletKeyFileName    = "walletKey.pem"
	validatorKeyFileName = "validatorKey.pem"

	log = logger.GetOrCreate("keygenerator")
)

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = fileGenHelpTemplate
	app.Name = "Key generation Tool"
	app.Version = "v1.0.0"
	app.Usage = "This binary will generate a validatorKey.pem and walletKey.pem, each containing private key(s)"
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}
	app.Flags = []cli.Flag{
		numKeys,
		keyType,
	}

	app.Action = func(_ *cli.Context) error {
		return generateAllFiles()
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error("error generating files", "error", err)

		os.Exit(1)
	}
}

func generateFolder(index int, numKeys int) (string, error) {
	absPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if numKeys > 1 {
		absPath = filepath.Join(absPath, fmt.Sprintf(keysFolderPattern, index))
	}

	log.Info("generating files in", "folder", absPath)

	err = os.MkdirAll(absPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	return absPath, nil
}

func generateKeys(keyGen crypto.KeyGenerator) ([]byte, []byte, error) {
	sk, pk := keyGen.GeneratePair()
	skBytes, err := sk.ToByteArray()
	if err != nil {
		return nil, nil, err
	}

	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return nil, nil, err
	}

	return skBytes, pkBytes, nil
}

func backupFileIfExists(filename string) {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return
		}
	}
	//if we reached here the file probably exists, make a timestamped backup
	_ = os.Rename(filename, filename+"."+fmt.Sprintf("%d", time.Now().Unix()))
}

func generateAllFiles() error {
	for i := 0; i < argsConfig.numKeys; i++ {
		err := generateOneSetOfFiles(i, argsConfig.numKeys)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateOneSetOfFiles(index int, numKeys int) error {
	switch argsConfig.keyType {
	case "validator":
		return generateBlockKey(index, numKeys)
	case "wallet":
		return generateTxKey(index, numKeys)
	case "both":
		err := generateBlockKey(index, numKeys)
		if err != nil {
			return err
		}

		return generateTxKey(index, numKeys)
	default:
		return fmt.Errorf("unknown key type %s", argsConfig.keyType)
	}
}

func generateBlockKey(index int, numKeys int) error {
	pubkeyConverter, err := factory.NewPubkeyConverter(
		config.PubkeyConfig{
			Length: blsPubkeyLen,
			Type:   factory.HexFormat,
		},
	)
	if err != nil {
		return err
	}

	genForBlockSigningSk := signing.NewKeyGenerator(mcl.NewSuiteBLS12())

	return generateAndSave(index, numKeys, validatorKeyFileName, genForBlockSigningSk, pubkeyConverter)
}

func generateTxKey(index int, numKeys int) error {
	pubkeyConverter, err := factory.NewPubkeyConverter(
		config.PubkeyConfig{
			Length: txSignPubkeyLen,
			Type:   factory.Bech32Format,
		},
	)
	if err != nil {
		return err
	}

	genForBlockSigningSk := signing.NewKeyGenerator(ed25519.NewEd25519())

	return generateAndSave(index, numKeys, walletKeyFileName, genForBlockSigningSk, pubkeyConverter)
}

func generateAndSave(index int, numKeys int, baseFilename string, genForBlockSigningSk crypto.KeyGenerator, pubkeyConverter core.PubkeyConverter) error {
	folder, err := generateFolder(index, numKeys)
	if err != nil {
		return err
	}

	filename := filepath.Join(folder, baseFilename)
	backupFileIfExists(filename)

	err = os.Remove(filename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, core.FileModeUserReadWrite)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	sk, pk, err := generateKeys(genForBlockSigningSk)
	if err != nil {
		return err
	}

	pkString := pubkeyConverter.Encode(pk)

	return core.SaveSkToPemFile(file, pkString, []byte(hex.EncodeToString(sk)))
}
