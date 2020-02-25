package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding/sorting"
	"github.com/libp2p/go-libp2p-core/peer"
)

// PrioSharderStub -
type PrioSharderStub struct {
	GetShardCalled             func(id peer.ID) uint32
	GetDistanceCalled          func(a, b interface{}) *big.Int
	SortListCalled             func(peers []peer.ID, ref peer.ID) ([]peer.ID, bool)
	SetPeerShardResolverCalled func(psp p2p.PeerShardResolver) error
}

// GetShard -
func (pss *PrioSharderStub) GetShard(id peer.ID) uint32 {
	if pss.GetShardCalled != nil {
		return pss.GetShardCalled(id)
	}

	return 0
}

// GetDistance -
func (pss *PrioSharderStub) GetDistance(a, b sorting.SortingID) *big.Int {
	if pss.GetDistanceCalled != nil {
		return pss.GetDistanceCalled(a, b)
	}

	return big.NewInt(0)
}

// SortList -
func (pss *PrioSharderStub) SortList(peers []peer.ID, ref peer.ID) ([]peer.ID, bool) {
	if pss.SortListCalled != nil {
		return pss.SortListCalled(peers, ref)
	}

	return make([]peer.ID, 0), true
}

// SetPeerShardResolver -
func (pss *PrioSharderStub) SetPeerShardResolver(psp p2p.PeerShardResolver) error {
	if pss.SetPeerShardResolverCalled != nil {
		return pss.SetPeerShardResolverCalled(psp)
	}

	return nil
}

// IsInterfaceNil -
func (pss *PrioSharderStub) IsInterfaceNil() bool {
	return pss == nil
}
