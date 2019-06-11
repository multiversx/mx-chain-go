package memp2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Initializing_MemP2PNetwork_with_4_Peers(t *testing.T) {
	network := NewMemP2PNetwork()

	peer1 := NewMemP2PMessenger(network)
	peer2 := NewMemP2PMessenger(network)
	peer3 := NewMemP2PMessenger(network)
	peer4 := NewMemP2PMessenger(network)

	assert.Equal(t, 4, len(network.Peers))
	assert.Equal(t, "Peer1", string(peer1.ID()))
	assert.Equal(t, "Peer2", string(peer2.ID()))
	assert.Equal(t, "Peer3", string(peer3.ID()))
	assert.Equal(t, "Peer4", string(peer4.ID()))

	assert.Equal(t, "/memp2p/Peer1", peer1.Address)
	assert.Equal(t, "/memp2p/Peer2", peer2.Address)
	assert.Equal(t, "/memp2p/Peer3", peer3.Address)
	assert.Equal(t, "/memp2p/Peer4", peer4.Address)

	peerIDs := peer4.Peers()
	assert.Equal(t, "Peer1", peerIDs[0])
	assert.Equal(t, "Peer2", peerIDs[1])
	assert.Equal(t, "Peer3", peerIDs[2])
	assert.Equal(t, "Peer4", peerIDs[3])

	assert.Equal(t, 1, len(peer2.Addresses()))
	assert.Equal(t, "/memp2p/Peer2", peer2.Addresses()[0])
}
