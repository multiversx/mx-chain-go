package mock

import "github.com/ElrondNetwork/elrond-go/core"

// PeerTypeProviderStub -
type PeerTypeProviderStub struct {
	ComputeForPubKeyCalled func(pubKey []byte, shardID uint32) (core.ValidatorList, error)
}

// ComputeForPubKey -
func (p *PeerTypeProviderStub) ComputeForPubKey(pubKey []byte, shardID uint32) (core.ValidatorList, error) {
	if p.ComputeForPubKeyCalled != nil {
		return p.ComputeForPubKeyCalled(pubKey, shardID)
	}

	return "", nil
}

// IsInterfaceNil -
func (p *PeerTypeProviderStub) IsInterfaceNil() bool {
	return p == nil
}
