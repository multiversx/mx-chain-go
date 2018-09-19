package p2p

import (
	"fmt"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestPanicOnGarbageIn(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	NewClusterParameter("ERR", 1, 2)

}

func TestGeneratePeersAndAddresses(t *testing.T) {
	cp := NewClusterParameter("0.0.0.0", 4000, 4005)

	assert.Equal(t, 6, len(cp.Peers()))
	assert.Equal(t, 6, len(cp.Addrs()))

	for idx, peer := range cp.peers {
		fmt.Printf("Peer %s has address %s\n", peer.Pretty(), cp.Addrs()[idx])
	}

}
