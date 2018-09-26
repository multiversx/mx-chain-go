package p2p_test

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"testing"
)

func TestFailNewConnectParams(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	//invalid port
	p2p.NewConnectParams(65536)
}
