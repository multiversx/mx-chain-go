package peer

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

// bootstrapPeerTypeProvider implements PeerTypeProviderHandler interface
type bootstrapPeerTypeProvider struct {
}

// NewBootstrapPeerTypeProvider returns a new instance of bootstrapPeerTypeProvider
// It should be used for bootstrap only!
func NewBootstrapPeerTypeProvider() *bootstrapPeerTypeProvider {
	return &bootstrapPeerTypeProvider{}
}

// ComputeForPubKey returns common.ObserverList and shard 0 always
func (bptp *bootstrapPeerTypeProvider) ComputeForPubKey(_ []byte) (common.PeerType, uint32, error) {
	return common.ObserverList, 0, nil
}

// GetAllPeerTypeInfos returns an empty slice
func (bptp *bootstrapPeerTypeProvider) GetAllPeerTypeInfos() []*state.PeerTypeInfo {
	return make([]*state.PeerTypeInfo, 0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bptp *bootstrapPeerTypeProvider) IsInterfaceNil() bool {
	return bptp == nil
}
