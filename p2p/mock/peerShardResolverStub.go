package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

// PeerShardResolverStub -
type PeerShardResolverStub struct {
	GetShardIDCalled func(pid p2p.PeerID) uint32
}

// GetShardID -
func (psrs *PeerShardResolverStub) GetShardID(pid p2p.PeerID) uint32 {
	return psrs.GetShardIDCalled(pid)
}

// IsInterfaceNil -
func (psrs *PeerShardResolverStub) IsInterfaceNil() bool {
	return psrs == nil
}
