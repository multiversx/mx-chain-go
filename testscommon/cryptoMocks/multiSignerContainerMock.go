package cryptoMocks

import crypto "github.com/multiversx/mx-chain-crypto-go"

// MultiSignerContainerMock -
type MultiSignerContainerMock struct {
	MultiSigner crypto.MultiSigner
}

// NewMultiSignerContainerMock -
func NewMultiSignerContainerMock(multiSigner crypto.MultiSigner) *MultiSignerContainerMock {
	return &MultiSignerContainerMock{MultiSigner: multiSigner}
}

// GetMultiSigner -
func (mscm *MultiSignerContainerMock) GetMultiSigner(_ uint32) (crypto.MultiSigner, error) {
	return mscm.MultiSigner, nil
}

// IsInterfaceNil -
func (mscm *MultiSignerContainerMock) IsInterfaceNil() bool {
	return mscm == nil
}
