package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// NetworkShardingCollectorStub -
type NetworkShardingCollectorStub struct {
	UpdatePeerIdPublicKeyCalled func(pid core.PeerID, pk []byte)
}

// UpdatePeerIdPublicKey -
func (nscs *NetworkShardingCollectorStub) UpdatePeerIdPublicKey(pid core.PeerID, pk []byte) {
	nscs.UpdatePeerIdPublicKeyCalled(pid, pk)
}

// IsInterfaceNil -
func (nscs *NetworkShardingCollectorStub) IsInterfaceNil() bool {
	return nscs == nil
}
