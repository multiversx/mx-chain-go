package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/schnorr"
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

	for i := 0; i < PkSkPairsToGenerate; i++ {
		generator := schnorr.NewKeyGenerator()
		sk, pk := generator.GeneratePair()
		skBytes, err2 := sk.ToByteArray()
		if err2 != nil {
			fmt.Println("Cound not convert sk to byte array")
		}
		pkBytes, err2 := pk.ToByteArray()
		if err2 != nil {
			fmt.Println("Cound not convert pk to byte array")
		}

		base64sk := make([]byte, base64.StdEncoding.EncodedLen(len(skBytes)))
		base64.StdEncoding.Encode(base64sk, []byte(skBytes))

		base64pk := make([]byte, base64.StdEncoding.EncodedLen(len(pkBytes)))
		base64.StdEncoding.Encode(base64pk, []byte(pkBytes))

		if _, err3 := fsk.Write(append(base64sk, '\n')); err3 != nil {
			fmt.Println(err3)
		}

		if _, err3 := fpk.Write(append(base64pk, '\n')); err3 != nil {
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
