package disabled

import crypto "github.com/multiversx/mx-chain-crypto-go"

type disabledMultiSignerContainer struct {
	multiSigner crypto.MultiSigner
}

// NewMultiSignerContainer creates a disabled multi signer container
func NewMultiSignerContainer() *disabledMultiSignerContainer {
	return &disabledMultiSignerContainer{
		multiSigner: NewMultiSigner(),
	}
}

// GetMultiSigner returns a disabled multi signer as this is a disabled component
func (dmsc *disabledMultiSignerContainer) GetMultiSigner(_ uint32) (crypto.MultiSigner, error) {
	return dmsc.multiSigner, nil
}

// IsInterfaceNil returns true if the underlying object is nil
func (dmsc *disabledMultiSignerContainer) IsInterfaceNil() bool {
	return dmsc == nil
}
