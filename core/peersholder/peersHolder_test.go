package peersholder

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/require"
)

func TestNewPeersHolder(t *testing.T) {
	ph := NewPeersHolder(nil)
	require.False(t, check.IfNil(ph))
}

func TestPeersHolder_PutNotPreferredPeerShouldNotAdd(t *testing.T) {
	t.Parallel()

	prefPeers := [][]byte{[]byte("peer0"), []byte("peer1")}
	ph := NewPeersHolder(prefPeers)

	ph.Put([]byte("different pub key"), "pid", 1)

	peers := ph.Get()
	require.Zero(t, len(peers))
}

func TestPeersHolder_PutNewPeerInfoShouldAdd(t *testing.T) {
	t.Parallel()

	prefPeers := [][]byte{[]byte("peer0"), []byte("peer1")}
	ph := NewPeersHolder(prefPeers)

	ph.Put([]byte("peer0"), "pid", 1)

	peers := ph.Get()
	require.Equal(t, peers[1][0], core.PeerID("pid"))
}

func TestPeersHolder_PutOldPeerInfoSameShardShouldNotChange(t *testing.T) {
	t.Parallel()

	prefPeers := [][]byte{[]byte("peer0"), []byte("peer1")}
	ph := NewPeersHolder(prefPeers)

	ph.Put([]byte("peer0"), "pid", 1)
	ph.Put([]byte("peer0"), "pid", 1)

	peers := ph.Get()
	require.Equal(t, 1, len(peers))
	require.Equal(t, 1, len(peers[1]))
	require.Equal(t, peers[1][0], core.PeerID("pid"))
}

func TestPeersHolder_PutOldPeerInfoNewShardShouldUpdate(t *testing.T) {
	t.Parallel()

	prefPeers := [][]byte{[]byte("peer0"), []byte("peer1")}
	ph := NewPeersHolder(prefPeers)

	ph.Put([]byte("peer0"), "pid", 0)
	ph.Put([]byte("peer0"), "pid", 1)

	peers := ph.Get()
	require.Equal(t, 1, len(peers))
	require.Equal(t, 1, len(peers[1]))
	require.Equal(t, peers[1][0], core.PeerID("pid"))
}

func TestPeersHolder_RemovePeerIDNotFound(t *testing.T) {
	t.Parallel()

	prefPeers := [][]byte{[]byte("peer0"), []byte("peer1")}
	ph := NewPeersHolder(prefPeers)

	ph.Put([]byte("peer0"), "pid", 3)
	ph.Remove("new peer id")

	peers := ph.Get()
	require.Equal(t, 1, len(peers))
}

func TestPeersHolder_RemovePeerShouldWork(t *testing.T) {
	t.Parallel()

	prefPeers := [][]byte{[]byte("peer0"), []byte("peer1")}
	ph := NewPeersHolder(prefPeers)

	ph.Put([]byte("peer0"), "pid", 3)
	ph.Remove("pid")

	peers := ph.Get()
	require.Equal(t, 0, len(peers))
}

func TestPeersHolder_RemovePeerShouldNotAffectOrder(t *testing.T) {
	t.Parallel()

	prefPeers := [][]byte{[]byte("peer0"), []byte("peer1"), []byte("peer2")}
	ph := NewPeersHolder(prefPeers)

	ph.Put([]byte("peer0"), "pid0", 3)
	ph.Put([]byte("peer1"), "pid1", 3)
	ph.Put([]byte("peer2"), "pid2", 3)

	ph.Remove("pid1")

	peers := ph.Get()
	require.Equal(t, 1, len(peers))
	require.Equal(t, []core.PeerID{"pid0", "pid2"}, peers[3])
}

func TestPeersHolder_Clear(t *testing.T) {
	t.Parallel()

	prefPeers := [][]byte{[]byte("peer0"), []byte("peer1"), []byte("peer1")}
	ph := NewPeersHolder(prefPeers)

	ph.Put([]byte("peer0"), "pid0", 0)
	ph.Put([]byte("peer1"), "pid1", 1)
	ph.Put([]byte("peer2"), "pid2", 2)
	ph.Put([]byte("peer2"), "pid2", 1)

	ph.Clear()

	peers := ph.Get()
	require.Zero(t, len(peers))
}

func TestPeersHolder_RemovePeerShouldNotRemovePubKeyAsPreferred(t *testing.T) {
	t.Parallel()

	prefPeers := [][]byte{[]byte("peer0"), []byte("peer1"), []byte("peer2")}
	ph := NewPeersHolder(prefPeers)

	ph.Put([]byte("peer0"), "pid0", 3)
	ph.Put([]byte("peer1"), "pid1", 3)
	ph.Put([]byte("peer2"), "pid2", 3)

	ph.Remove("pid1")

	peers := ph.Get()
	require.Equal(t, 1, len(peers))
	require.Equal(t, []core.PeerID{"pid0", "pid2"}, peers[3])

	ph.Lock()
	val, found := ph.pubKeysToPeerIDs["peer1"]
	require.True(t, found)
	require.Nil(t, val)
}
