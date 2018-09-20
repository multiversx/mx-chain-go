package p2p_test

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestPanicOnGarbageIn(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	p2p.NewClusterParameter("ERR", 1, 2)

}

func TestGeneratePeersAndAddresses(t *testing.T) {
	cp := p2p.NewClusterParameter("0.0.0.0", 4000, 4005)

	assert.Equal(t, 6, len(cp.Peers()))
	assert.Equal(t, 6, len(cp.Addrs()))

	for idx, peer := range cp.Peers() {
		fmt.Printf("Peer %s has address %s\n", peer.Pretty(), cp.Addrs()[idx])
	}

}
