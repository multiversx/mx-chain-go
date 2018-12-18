package p2p_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

//TODO remove this skip when things get clearer why the tests sometimes fail
//Task: EN-575
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
