package p2p_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

var isP2PMessengerException = false

func init() {
	absPath, _ := filepath.Abs("../p2pMessengerTestException")

	_, err := ioutil.ReadFile(absPath)
	if err == nil {
		isP2PMessengerException = true

		fmt.Println("### Running with P2PMessenger exception ###")
	}
}
