package p2p_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

var skipP2PMessengerTests = false

func init() {
	p2p.Log.SetLevel(logger.LogDebug)

	absPath, _ := filepath.Abs("../skipP2PMessengerTests")

	_, err := ioutil.ReadFile(absPath)
	if err == nil {
		skipP2PMessengerTests = true

		fmt.Println("### Skiping tests from P2PMessenger struct ###")
	}
}
