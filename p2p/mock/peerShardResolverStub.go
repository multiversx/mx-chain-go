package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

type PeerShardResolverStub struct {
	ByIDCalled func(pid p2p.PeerID) uint32
}

func (psrs *PeerShardResolverStub) ByID(pid p2p.PeerID) uint32 {
	return psrs.ByIDCalled(pid)
}

func (psrs *PeerShardResolverStub) IsInterfaceNil() bool {
	return psrs == nil
}
