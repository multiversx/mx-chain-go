package main

import (
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

	err = ioutil.WriteFile(path + "/sk", skBytes, 0644)
	err = ioutil.WriteFile(path + "/pk", pkBytes, 0644)
}
