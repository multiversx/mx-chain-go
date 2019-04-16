package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
)

// PkSkPairsToGenerate holds the number of pairs sk/pk that will be generated
const PkSkPairsToGenerate = 21

func main() {

	path, err := filepath.Abs("")
	if err != nil {
		fmt.Println("Invalid file path")
		return
	}

	fskPlain, err := os.OpenFile(path+"/skPlainText", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		_ = fskPlain.Close()
	}()

	fskPem, err := os.OpenFile(path+"/skPem", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		_ = fskPem.Close()
	}()

	fpk, err := os.OpenFile(path+"/pkPlainText", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		_ = fpk.Close()
	}()

	suite := kyber.NewBlakeSHA256Ed25519()

	for i := 0; i < PkSkPairsToGenerate; i++ {
		generator := signing.NewKeyGenerator(suite)
		sk, pk := generator.GeneratePair()
		skBytes, err2 := sk.ToByteArray()
		if err2 != nil {
			fmt.Println("Could not convert sk to byte array")
		}
		pkBytes, err2 := pk.ToByteArray()
		if err2 != nil {
			fmt.Println("Could not convert pk to byte array")
		}

		skHex := []byte(hex.EncodeToString(skBytes))
		pkHex := []byte(hex.EncodeToString(pkBytes))

		_, err := fskPlain.Write(append(skHex, '\n'))
		if err != nil {
			fmt.Println(err)
		}

		_, err = fpk.Write(append(pkHex, '\n'))
		if err != nil {
			fmt.Println(err)
		}

		err = core.SaveSkToPemFile(fskPem, string(pkHex), skHex)
		if err != nil {
			fmt.Println(err)
		}
	}
}
