package main

import (
	"bytes"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/urfave/cli"
)

type cfg struct {
	numKeys    int
	keyType    string
	consoleOut bool
	noSplit    bool
}

const validatorType = "validator"
const walletType = "wallet"
const bothType = "both"

type key struct {
	skBytes []byte
	pkBytes []byte
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
		Name: "key-type",
		Usage: fmt.Sprintf(
			"What king of keys should generate. Available options: %s, %s, %s",
			validatorType,
			walletType,
			bothType),
		Value:       "validator",
		Destination: &argsConfig.keyType,
	}
	// consoleOut is the flag that, if active, will print everything on the console, not on a physical file
	consoleOut = cli.BoolFlag{
		Name:        "console-out",
		Usage:       "Boolean option that will enable printing the generated keys directly on the console",
		Destination: &argsConfig.consoleOut,
	}
	// noSplit is the flag that, if active, will generate the keys in the same file
	noSplit = cli.BoolFlag{
		Name:        "no-split",
		Usage:       "Boolean option that will make each generated key added in the same file",
		Destination: &argsConfig.noSplit,
	}

	argsConfig = &cfg{}

	walletKeyFilenameTemplate    = "walletKey%s.pem"
	validatorKeyFilenameTemplate = "validatorKey%s.pem"

	log = logger.GetOrCreate("keygenerator")

	validatorPubKeyConverter, _ = pubkeyConverter.NewHexPubkeyConverter(blsPubkeyLen)
	walletPubKeyConverter, _    = pubkeyConverter.NewBech32PubkeyConverter(txSignPubkeyLen)
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
		consoleOut,
		noSplit,
	}

	app.Action = func(_ *cli.Context) error {
		return process()
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error("error generating files", "error", err)

		os.Exit(1)
	}
}

func process() error {
	validatorKeys, walletKeys, err := generateKeys(argsConfig.keyType, argsConfig.numKeys)
	if err != nil {
		return err
	}

	return outputKeys(validatorKeys, walletKeys, argsConfig.consoleOut, argsConfig.noSplit)
}

func generateKeys(typeKey string, numKeys int) ([]key, []key, error) {
	if numKeys < 1 {
		return nil, nil, fmt.Errorf("number of keys should be a number greater or equal to 1")
	}

	validatorKeys := make([]key, 0)
	walletKeys := make([]key, 0)
	var err error

	blockSigningGenerator := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	txSigningGenerator := signing.NewKeyGenerator(ed25519.NewEd25519())

	for i := 0; i < numKeys; i++ {
		switch typeKey {
		case validatorType:
			validatorKeys, err = generateKey(blockSigningGenerator, validatorKeys)
			if err != nil {
				return nil, nil, err
			}
		case walletType:
			walletKeys, err = generateKey(txSigningGenerator, walletKeys)
			if err != nil {
				return nil, nil, err
			}
		case bothType:
			validatorKeys, err = generateKey(blockSigningGenerator, validatorKeys)
			if err != nil {
				return nil, nil, err
			}

			walletKeys, err = generateKey(txSigningGenerator, walletKeys)
			if err != nil {
				return nil, nil, err
			}
		default:
			return nil, nil, fmt.Errorf("unknown key type %s", argsConfig.keyType)
		}
	}

	return validatorKeys, walletKeys, nil
}

func generateKey(keyGen crypto.KeyGenerator, list []key) ([]key, error) {
	sk, pk := keyGen.GeneratePair()
	skBytes, err := sk.ToByteArray()
	if err != nil {
		return nil, err
	}

	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return nil, err
	}

	list = append(
		list,
		key{
			skBytes: skBytes,
			pkBytes: pkBytes,
		},
	)

	return list, nil
}

func outputKeys(
	validatorKeys []key,
	walletKeys []key,
	consoleOut bool,
	noSplit bool,
) error {
	if consoleOut {
		return printKeys(validatorKeys, walletKeys)
	}

	return saveKeys(validatorKeys, walletKeys, noSplit)
}

func printKeys(validatorKeys []key, walletKeys []key) error {
	if len(validatorKeys)+len(walletKeys) == 0 {
		return fmt.Errorf("internal error: no keys to print")
	}

	var errFound error
	if len(validatorKeys) > 0 {
		err := printSliceKeys("Validator keys:", validatorKeys, validatorPubKeyConverter)
		if err != nil {
			errFound = err
		}
	}
	if len(walletKeys) > 0 {
		err := printSliceKeys("Wallet keys:", walletKeys, walletPubKeyConverter)
		if err != nil {
			errFound = err
		}
	}

	return errFound
}

func printSliceKeys(message string, sliceKeys []key, converter core.PubkeyConverter) error {
	data := []string{message + "\n"}

	for _, k := range sliceKeys {
		buf := bytes.NewBuffer(make([]byte, 0))
		err := writeKeyToStream(buf, k, converter)
		if err != nil {
			return err
		}

		data = append(data, string(buf.Bytes()))
	}

	log.Info(strings.Join(data, ""))
	return nil
}

func writeKeyToStream(writer io.Writer, key key, pubkeyConverter core.PubkeyConverter) error {
	if check.IfNilReflect(writer) {
		return fmt.Errorf("nil writer")
	}

	pkString := pubkeyConverter.Encode(key.pkBytes)

	blk := pem.Block{
		Type:  "PRIVATE KEY for " + pkString,
		Bytes: []byte(hex.EncodeToString(key.skBytes)),
	}

	return pem.Encode(writer, &blk)
}

func saveKeys(validatorKeys []key, walletKeys []key, noSplit bool) error {
	if len(validatorKeys)+len(walletKeys) == 0 {
		return fmt.Errorf("internal error: no keys to save")
	}

	var errFound error
	if len(validatorKeys) > 0 {
		err := saveSliceKeys(validatorKeyFilenameTemplate, validatorKeys, validatorPubKeyConverter, noSplit)
		if err != nil {
			errFound = err
		}
	}
	if len(walletKeys) > 0 {
		err := saveSliceKeys(walletKeyFilenameTemplate, walletKeys, walletPubKeyConverter, noSplit)
		if err != nil {
			errFound = err
		}
	}

	return errFound
}

func saveSliceKeys(baseFilenameTemplate string, keys []key, pubkeyConverter core.PubkeyConverter, noSplit bool) error {
	var file *os.File
	var err error
	for i, k := range keys {
		shouldCreateFile := !noSplit || i == 0
		if shouldCreateFile {
			file, err = generateFile(i, len(keys), noSplit, baseFilenameTemplate)
			if err != nil {
				return err
			}
		}

		err = writeKeyToStream(file, k, pubkeyConverter)
		if err != nil {
			return err
		}
		if !noSplit {
			err = file.Close()
			if err != nil {
				return err
			}
		}
	}

	if noSplit {
		return file.Close()
	}

	return nil
}

func generateFile(index int, numKeys int, noSplit bool, baseFilenameTemplate string) (*os.File, error) {
	folder, err := generateFolder(index, numKeys, noSplit)
	if err != nil {
		return nil, err
	}

	filename := filepath.Join(folder, baseFilenameTemplate)
	backupFileIfExists(filename)
	//replace the %s with empty string
	filename = fmt.Sprintf(filename, "")
	err = os.Remove(filename)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, core.FileModeUserReadWrite)
}

func generateFolder(index int, numKeys int, noSplit bool) (string, error) {
	absPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	shouldCreateDirectory := numKeys > 1 && !noSplit
	if shouldCreateDirectory {
		absPath = filepath.Join(absPath, fmt.Sprintf(keysFolderPattern, index))
	}

	log.Info("generating files in", "folder", absPath)

	err = os.MkdirAll(absPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	return absPath, nil
}

func backupFileIfExists(filenameTemplate string) {
	existingFilename := fmt.Sprintf(filenameTemplate, "")
	if _, err := os.Stat(existingFilename); err != nil {
		if os.IsNotExist(err) {
			return
		}
	}
	//if we reached here the file probably exists, make a timestamped backup
	_ = os.Rename(existingFilename, fmt.Sprintf(filenameTemplate, fmt.Sprintf("_%d", time.Now().Unix())))
}
