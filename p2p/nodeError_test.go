package p2p_test

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"testing"
)

func TestNodeError(t *testing.T) {
	err := p2p.NewNodeError("A", "B", "MES")

	if err.Err != err.Error() {
		t.Fatal("Node error should return Err on Error()")
	}
}
