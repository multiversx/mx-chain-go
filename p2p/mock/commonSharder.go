package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

var _ p2p.CommonSharder = (*CommonSharder)(nil)

// CommonSharder -
type CommonSharder struct {
	SetPeerShardResolverCalled func(psp p2p.PeerShardResolver) error
}

// SetPeerShardResolver -
func (cs *CommonSharder) SetPeerShardResolver(psp p2p.PeerShardResolver) error {
	if cs.SetPeerShardResolverCalled != nil {
		return cs.SetPeerShardResolverCalled(psp)
	}

	return nil
}

// IsInterfaceNil -
func (cs *CommonSharder) IsInterfaceNil() bool {
	return cs == nil
}
