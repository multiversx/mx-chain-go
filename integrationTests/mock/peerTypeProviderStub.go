package mock

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

// PeerTypeProviderStub -
type PeerTypeProviderStub struct {
	ComputeForPubKeyCalled func(pubKey []byte) (common.PeerType, uint32, error)
}

// ComputeForPubKey -
func (p *PeerTypeProviderStub) ComputeForPubKey(pubKey []byte) (common.PeerType, uint32, error) {
	if p.ComputeForPubKeyCalled != nil {
		return p.ComputeForPubKeyCalled(pubKey)
	}

	return "", 0, nil
}

// GetAllPeerTypeInfos -
func (p *PeerTypeProviderStub) GetAllPeerTypeInfos() []*state.PeerTypeInfo {
	return nil
}

// IsInterfaceNil -
func (p *PeerTypeProviderStub) IsInterfaceNil() bool {
	return p == nil
}

// Close  -
func (p *PeerTypeProviderStub) Close() error {
	return nil
}
