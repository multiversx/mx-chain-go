package cryptoMocks

import (
	"bytes"
)

const signatureSize = 48

// MultisignerMock is used to mock the multisignature scheme
type MultisignerMock struct {
	CreateSignatureShareCalled func(privateKeyBytes []byte, message []byte) ([]byte, error)
	VerifySignatureShareCalled func(publicKey []byte, message []byte, sig []byte) error
	AggregateSigsCalled        func(pubKeysSigners [][]byte, signatures [][]byte) ([]byte, error)
	VerifyAggregatedSigCalled  func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error
}

// NewMultiSigner -
func NewMultiSigner() *MultisignerMock {
	return &MultisignerMock{}
}

// CreateSignatureShare -
func (mm *MultisignerMock) CreateSignatureShare(privateKeyBytes []byte, message []byte) ([]byte, error) {
	if mm.CreateSignatureShareCalled != nil {
		return mm.CreateSignatureShareCalled(privateKeyBytes, message)
	}

	return bytes.Repeat([]byte{0xAA}, signatureSize), nil
}

// VerifySignatureShare -
func (mm *MultisignerMock) VerifySignatureShare(publicKey []byte, message []byte, sig []byte) error {
	if mm.VerifySignatureShareCalled != nil {
		return mm.VerifySignatureShareCalled(publicKey, message, sig)
	}
	return nil
}

// AggregateSigs -
func (mm *MultisignerMock) AggregateSigs(pubKeysSigners [][]byte, signatures [][]byte) ([]byte, error) {
	if mm.AggregateSigsCalled != nil {
		return mm.AggregateSigsCalled(pubKeysSigners, signatures)
	}

	return bytes.Repeat([]byte{0xAA}, signatureSize), nil
}

// VerifyAggregatedSig -
func (mm *MultisignerMock) VerifyAggregatedSig(pubKeysSigners [][]byte, message []byte, aggSig []byte) error {
	if mm.VerifyAggregatedSigCalled != nil {
		return mm.VerifyAggregatedSigCalled(pubKeysSigners, message, aggSig)
	}
	return nil
}

// IsInterfaceNil -
func (mm *MultisignerMock) IsInterfaceNil() bool {
	return mm == nil
}
