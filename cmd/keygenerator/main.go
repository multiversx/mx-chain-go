package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2"
)

// PkSkPairsToGenerate holds the number of pairs sk/pk that will be generated
const PkSkPairsToGenerate = 21

func main() {

	path, err := filepath.Abs("")
	if err != nil {
		fmt.Println("Invalid file path")
		return
	}

	fsk, err := os.OpenFile(path+"/sk", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	fpk, err := os.OpenFile(path+"/pk", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	suite := kv2.NewBlakeSHA256Ed25519()

	for i := 0; i < PkSkPairsToGenerate; i++ {
		generator := signing.NewKeyGenerator(suite)
		sk, pk := generator.GeneratePair()
		skBytes, err2 := sk.ToByteArray()
		if err2 != nil {
			fmt.Println("Cound not convert sk to byte array")
		}
		pkBytes, err2 := pk.ToByteArray()
		if err2 != nil {
			fmt.Println("Cound not convert pk to byte array")
		}

		skHex := []byte(hex.EncodeToString(skBytes))
		pkHex := []byte(hex.EncodeToString(pkBytes))

		if _, err3 := fsk.Write(append(skHex, '\n')); err3 != nil {
			fmt.Println(err3)
		}

		if _, err3 := fpk.Write(append(pkHex, '\n')); err3 != nil {
			fmt.Println(err3)
		}
	}

	if err4 := fsk.Close(); err4 != nil {
		fmt.Println(err4)
	}

	if err4 := fpk.Close(); err4 != nil {
		fmt.Println(err4)
	}
}
