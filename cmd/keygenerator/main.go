package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/schnorr"
)

func main() {
	generator := schnorr.NewKeyGenerator()
	sk, pk := generator.GeneratePair()
	path, err := filepath.Abs("")
	if err != nil {
		fmt.Println("Invalid file path")
		return
	}
	skBytes, err := sk.ToByteArray()
	if err != nil {
		fmt.Println("Cound not convert sk to byte array")
	}
	pkBytes, err := pk.ToByteArray()
	if err != nil {
		fmt.Println("Cound not convert pk to byte array")
	}

	base64sk := make([]byte, base64.StdEncoding.EncodedLen(len(skBytes)))
	base64.StdEncoding.Encode(base64sk, []byte(skBytes))

	base64pk := make([]byte, base64.StdEncoding.EncodedLen(len(pkBytes)))
	base64.StdEncoding.Encode(base64pk, []byte(pkBytes))

	err = ioutil.WriteFile(path + "/sk", base64sk, 0644)
	err = ioutil.WriteFile(path + "/pk", base64pk, 0644)
}
