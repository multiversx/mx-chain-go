package sorting

import (
	"fmt"
	"math/big"
	"sort"
	"testing"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

func createPeerDistance(distance int) *PeerDistance {
	return &PeerDistance{
		ID:       peer.ID(fmt.Sprintf("pid_%d", distance)),
		Distance: big.NewInt(int64(distance)),
	}
}

func TestPeerDistances_Sort(t *testing.T) {
	t.Parallel()

	pid4 := createPeerDistance(4)
	pid0 := createPeerDistance(0)
	pid100 := createPeerDistance(100)
	pid1 := createPeerDistance(1)
	pid2 := createPeerDistance(2)

	pids := PeerDistances{pid4, pid0, pid100, pid1, pid2}
	sort.Sort(pids)

	assert.Equal(t, pid0, pids[0])
	assert.Equal(t, pid1, pids[1])
	assert.Equal(t, pid2, pids[2])
	assert.Equal(t, pid4, pids[3])
	assert.Equal(t, pid100, pids[4])
	assert.Equal(t, 5, len(pids))
}
