package networksharding

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding/sorting"
	"github.com/libp2p/go-libp2p-core/peer"
)

type listNoKadSharder struct {
	selfPeerId      peer.ID
	maxPeerCount    int
	computeDistance func(src peer.ID, dest peer.ID) *big.Int
}

// NewListNoKadSharder creates a new sharder instance that is shard agnostic
func NewListNoKadSharder(
	selfPeerId peer.ID,
	maxPeerCount int,
) (*listNoKadSharder, error) {
	if maxPeerCount < minAllowedConnectedPeers {
		return nil, fmt.Errorf("%w, maxPeerCount should be at least %d", p2p.ErrInvalidValue, minAllowedConnectedPeers)
	}

	return &listNoKadSharder{
		selfPeerId:      selfPeerId,
		maxPeerCount:    maxPeerCount,
		computeDistance: computeDistanceByCountingBits,
	}, nil
}

// ComputeEvictionList returns the eviction list
func (lnks *listNoKadSharder) ComputeEvictionList(pidList []peer.ID) []peer.ID {
	list := lnks.convertList(pidList)
	_, evictionProposed := evict(list, lnks.maxPeerCount)

	return evictionProposed
}

func (lnks *listNoKadSharder) convertList(peers []peer.ID) sorting.PeerDistances {
	list := sorting.PeerDistances{}

	for _, p := range peers {
		pd := &sorting.PeerDistance{
			ID:       p,
			Distance: lnks.computeDistance(p, lnks.selfPeerId),
		}
		list = append(list, pd)
	}

	return list
}

// Has returns true if provided pid is among the provided list
func (lnks *listNoKadSharder) Has(pid peer.ID, list []peer.ID) bool {
	return has(pid, list)
}

// SetPeerShardResolver sets the peer shard resolver for this sharder. Doesn't do anything in this implementation
func (lnks *listNoKadSharder) SetPeerShardResolver(_ p2p.PeerShardResolver) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (lnks *listNoKadSharder) IsInterfaceNil() bool {
	return lnks == nil
}
