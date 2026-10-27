package cryptoMocks

import crypto "github.com/multiversx/mx-chain-crypto-go"

// MultiSignerContainerMock -
type MultiSignerContainerMock struct {
	MultiSigner crypto.MultiSignerV2
}

// NewMultiSignerContainerMock -
func NewMultiSignerContainerMock(multiSigner crypto.MultiSignerV2) *MultiSignerContainerMock {
	return &MultiSignerContainerMock{MultiSigner: multiSigner}
}

// GetMultiSigner -
func (mscm *MultiSignerContainerMock) GetMultiSigner(_ uint32) (crypto.MultiSignerV2, error) {
	return mscm.MultiSigner, nil
}

// IsInterfaceNil -
func (mscm *MultiSignerContainerMock) IsInterfaceNil() bool {
	return mscm == nil
}
