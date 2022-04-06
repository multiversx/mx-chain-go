package peersHolder

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewPeersHolder(t *testing.T) {
	t.Parallel()

	ph := NewPeersHolder(nil)
	assert.False(t, check.IfNil(ph))
}

func TestPeersHolder_PutConnectionAddress(t *testing.T) {
	t.Parallel()

	t.Run("not preferred should not add", func(t *testing.T) {
		t.Parallel()

		preferredPeers := [][]byte{[]byte("/ip4/10.100.100.100")}
		ph := NewPeersHolder(preferredPeers)
		assert.False(t, check.IfNil(ph))

		unknownConnection := []byte("/ip4/20.200.200.200/tcp/8080/p2p/some-random-pid") // preferredPeers[0]
		providedPid := core.PeerID("provided pid")
		ph.PutConnectionAddress(providedPid, unknownConnection)

		_, found := ph.tempPeerIDsWaitingForShard[providedPid]
		assert.False(t, found)

		peers := ph.Get()
		assert.Equal(t, 0, len(peers))
	})
	t.Run("new connection should add to intermediate maps", func(t *testing.T) {
		t.Parallel()

		preferredPeers := [][]byte{[]byte("/ip4/10.100.100.100"), []byte("10.100.100.101")}
		ph := NewPeersHolder(preferredPeers)
		assert.False(t, check.IfNil(ph))

		newConnection := []byte("/ip4/10.100.100.100/tcp/38191/p2p/some-random-pid") // preferredPeers[0]
		providedPid := core.PeerID("provided pid")
		ph.PutConnectionAddress(providedPid, newConnection)

		knownSlice, found := ph.tempPeerIDsWaitingForShard[providedPid]
		assert.True(t, found)
		assert.True(t, bytes.Equal(preferredPeers[0], []byte(knownSlice)))

		pInfo := ph.connAddrSliceToPeerInfo[knownSlice]
		assert.Equal(t, providedPid, pInfo.pid)
		assert.Equal(t, uint32(0), pInfo.shardID)

		// not in the final map yet
		peers := ph.Get()
		assert.Equal(t, 0, len(peers))
	})
	t.Run("should update", func(t *testing.T) {
		t.Parallel()

		preferredPeers := [][]byte{[]byte("/ip4/10.100.100.100"), []byte("10.100.100.101"), []byte("16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ")}
		ph := NewPeersHolder(preferredPeers)
		assert.False(t, check.IfNil(ph))

		newConnection := []byte("/ip4/10.100.100.102/tcp/38191/p2p/16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ") // preferredPeers[2]
		providedPid := core.PeerID("provided pid")
		ph.PutConnectionAddress(providedPid, newConnection)

		knownSlice, found := ph.tempPeerIDsWaitingForShard[providedPid]
		assert.True(t, found)
		assert.True(t, bytes.Equal(preferredPeers[2], []byte(knownSlice)))

		pInfo := ph.connAddrSliceToPeerInfo[knownSlice]
		assert.Equal(t, providedPid, pInfo.pid)
		assert.Equal(t, uint32(0), pInfo.shardID)

		ph.PutConnectionAddress(providedPid, newConnection) // try to update with same connection for coverage

		newPid := core.PeerID("new pid")
		ph.PutConnectionAddress(newPid, newConnection)
		knownSlice, found = ph.tempPeerIDsWaitingForShard[providedPid]
		assert.True(t, found)
		assert.True(t, bytes.Equal(preferredPeers[2], []byte(knownSlice)))

		pInfo = ph.connAddrSliceToPeerInfo[knownSlice]
		assert.Equal(t, newPid, pInfo.pid)
		assert.Equal(t, uint32(0), pInfo.shardID)

		// not in the final map yet
		peers := ph.Get()
		assert.Equal(t, 0, len(peers))
	})
}

func TestPeersHolder_PutShardID(t *testing.T) {
	t.Parallel()

	t.Run("peer not added in the waiting list should be skipped", func(t *testing.T) {
		t.Parallel()

		preferredPeers := [][]byte{[]byte("/ip4/10.100.100.100")}
		ph := NewPeersHolder(preferredPeers)
		assert.False(t, check.IfNil(ph))

		providedPid := core.PeerID("provided pid")
		providedShardID := uint32(123)
		ph.PutShardID(providedPid, providedShardID)

		peers := ph.Get()
		assert.Equal(t, 0, len(peers))
	})
	t.Run("peer not added in map should be skipped", func(t *testing.T) {
		t.Parallel()

		preferredPeers := [][]byte{[]byte("/ip4/10.100.100.100")}
		ph := NewPeersHolder(preferredPeers)
		assert.False(t, check.IfNil(ph))

		providedPid := core.PeerID("provided pid")
		providedShardID := uint32(123)
		ph.tempPeerIDsWaitingForShard[providedPid] = string(preferredPeers[0])
		ph.PutShardID(providedPid, providedShardID)

		peers := ph.Get()
		assert.Equal(t, 0, len(peers))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		preferredPeers := [][]byte{[]byte("/ip4/10.100.100.100"), []byte("10.100.100.101"), []byte("16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ")}
		ph := NewPeersHolder(preferredPeers)
		assert.False(t, check.IfNil(ph))

		newConnection := []byte("/ip4/10.100.100.101/tcp/38191/p2p/some-random-pid") // preferredPeers[1]
		providedPid := core.PeerID("provided pid")
		ph.PutConnectionAddress(providedPid, newConnection)

		providedShardID := uint32(123)
		ph.PutShardID(providedPid, providedShardID)

		peers := ph.Get()
		assert.Equal(t, 1, len(peers))
		peersInShard, found := peers[providedShardID]
		assert.True(t, found)
		assert.Equal(t, providedPid, peersInShard[0])

		pidData := ph.peerIDs[providedPid]
		assert.Equal(t, preferredPeers[1], []byte(pidData.connectionAddressSlice))
		assert.Equal(t, providedShardID, pidData.shardID)
		assert.Equal(t, 0, pidData.index)

		_, found = ph.tempPeerIDsWaitingForShard[providedPid]
		assert.False(t, found)
	})
}

func TestPeersHolder_Contains(t *testing.T) {
	t.Parallel()

	preferredPeers := [][]byte{[]byte("/ip4/10.100.100.100"), []byte("10.100.100.101")}
	ph := NewPeersHolder(preferredPeers)
	assert.False(t, check.IfNil(ph))

	newConnection := []byte("/ip4/10.100.100.101/tcp/38191/p2p/some-random-pid") // preferredPeers[1]
	providedPid := core.PeerID("provided pid")
	ph.PutConnectionAddress(providedPid, newConnection)

	providedShardID := uint32(123)
	ph.PutShardID(providedPid, providedShardID)

	assert.True(t, ph.Contains(providedPid))

	ph.Remove(providedPid)
	assert.False(t, ph.Contains(providedPid))

	unknownPid := core.PeerID("unknown pid")
	ph.Remove(unknownPid) // for code coverage
}

func TestPeersHolder_Clear(t *testing.T) {
	t.Parallel()

	preferredPeers := [][]byte{[]byte("/ip4/10.100.100.100"), []byte("16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ")}
	ph := NewPeersHolder(preferredPeers)
	assert.False(t, check.IfNil(ph))

	newConnection1 := []byte("/ip4/10.100.100.100/tcp/38191/p2p/some-random-pid") // preferredPeers[0]
	providedPid1 := core.PeerID("provided pid 1")
	ph.PutConnectionAddress(providedPid1, newConnection1)
	providedShardID := uint32(123)
	ph.PutShardID(providedPid1, providedShardID)
	assert.True(t, ph.Contains(providedPid1))

	newConnection2 := []byte("/ip4/10.100.100.102/tcp/38191/p2p/16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ") // preferredPeers[1]
	providedPid2 := core.PeerID("provided pid 1")
	ph.PutConnectionAddress(providedPid2, newConnection2)
	ph.PutShardID(providedPid2, providedShardID)
	assert.True(t, ph.Contains(providedPid2))

	peers := ph.Get()
	assert.Equal(t, 1, len(peers))
	assert.Equal(t, 2, len(peers[providedShardID]))

	ph.Clear()
	peers = ph.Get()
	assert.Equal(t, 0, len(peers))
}
