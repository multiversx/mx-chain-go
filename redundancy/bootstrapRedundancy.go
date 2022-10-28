package redundancy

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
)

type bootstrapNodeRedundancy struct {
	observerPrivateKey crypto.PrivateKey
}

// NewBootstrapNodeRedundancy returns a new instance of bootstrapNodeRedundancy
// It should be used for bootstrap only!
func NewBootstrapNodeRedundancy(nodePrivateKey crypto.PrivateKey) (*bootstrapNodeRedundancy, error) {
	if check.IfNil(nodePrivateKey) {
		return nil, ErrNilObserverPrivateKey
	}

	return &bootstrapNodeRedundancy{
		observerPrivateKey: nodePrivateKey,
	}, nil
}

// IsRedundancyNode returns false always
func (bnr *bootstrapNodeRedundancy) IsRedundancyNode() bool {
	return false
}

// IsMainMachineActive returns true always
func (bnr *bootstrapNodeRedundancy) IsMainMachineActive() bool {
	return true
}

// ObserverPrivateKey returns node's private key
func (bnr *bootstrapNodeRedundancy) ObserverPrivateKey() crypto.PrivateKey {
	return bnr.observerPrivateKey
}

// IsInterfaceNil returns true if there is no value under the interface
func (bnr *bootstrapNodeRedundancy) IsInterfaceNil() bool {
	return bnr == nil
}
