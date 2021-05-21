package peersholder

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/require"
)

func TestPeersHolderOperations(t *testing.T) {
	t.Parallel()

	ph := NewPeersHolder()
	require.False(t, check.IfNil(ph))

	pk, pid := "pk", "pid"
	pk1, pid1 := "pk1", "pid1"

	ph.Add(pk, pid)
	res, ok := ph.GetPeerIDForPublicKey(pk)
	require.True(t, ok)
	require.Equal(t, res, pid)

	res, ok = ph.GetPublicKeyForPeerID(pid)
	require.True(t, ok)
	require.Equal(t, res, pk)

	_, ok = ph.GetPublicKeyForPeerID(pid1)
	require.False(t, ok)
	_, ok = ph.GetPeerIDForPublicKey(pk1)
	require.False(t, ok)

	ph.Add(pk1, pid1)

	ok = ph.DeletePublicKey("invalid pk")
	require.False(t, ok)

	ok = ph.DeletePeerID("invalid peer id")
	require.False(t, ok)

	ok = ph.DeletePeerID(pid)
	_, ok = ph.GetPublicKeyForPeerID(pid)
	require.False(t, ok)
	_, ok = ph.GetPeerIDForPublicKey(pk)
	require.False(t, ok)

	ok = ph.DeletePublicKey(pk1)
	_, ok = ph.GetPublicKeyForPeerID(pid1)
	require.False(t, ok)
	_, ok = ph.GetPeerIDForPublicKey(pk1)
	require.False(t, ok)

	ph.Clear()

	require.Zero(t, ph.Len())
}

func TestPeersHolder_ConcurrentSafe(t *testing.T) {
	t.Parallel()

	ph := NewPeersHolder()

	for i := 0; i < 1000; i++ {
		go func(i int) {
			modRes := i % 7
			switch modRes {
			case 0:
				ph.Add("pk", "pid")
			case 1:
				ph.DeletePeerID("pid")
			case 2:
				ph.DeletePublicKey("pk")
			case 3:
				ph.GetPeerIDForPublicKey("pk")
			case 4:
				ph.GetPublicKeyForPeerID("pid")
			case 5:
				ph.Clear()
			case 6:
				ph.Len()
			}
		}(i)
	}
}
