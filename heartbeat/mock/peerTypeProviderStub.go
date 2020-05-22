package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// PeerTypeProviderStub -
type PeerTypeProviderStub struct {
	ComputeForPubKeyCalled func(pubKey []byte) (core.PeerType, uint32, error)
}

// ComputeForPubKey -
func (p *PeerTypeProviderStub) ComputeForPubKey(pubKey []byte) (core.PeerType, uint32, error) {
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
