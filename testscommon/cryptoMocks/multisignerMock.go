package cryptoMocks

import (
	"bytes"

	crypto "github.com/multiversx/mx-chain-crypto-go"
)

const signatureSize = 48

// MultisignerMock is used to mock the multisignature scheme
type MultisignerMock struct {
	CreateSignatureShareCalled   func(privateKeyBytes []byte, message []byte) ([]byte, error)
	CreateSignatureShareV2Called func(privateKeyBytes crypto.PrivateKey, message []byte) ([]byte, error)
	VerifySignatureShareCalled   func(publicKey []byte, message []byte, sig []byte) error
	VerifySignatureShareV2Called func(publicKey crypto.PublicKey, message []byte, sig []byte) error
	AggregateSigsCalled          func(pubKeysSigners [][]byte, signatures [][]byte) ([]byte, error)
	AggregateSigsV2Called        func(pubKeysSigners []crypto.PublicKey, signatures [][]byte) ([]byte, error)
	VerifyAggregatedSigCalled    func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error
	VerifyAggregatedSigV2Called  func(pubKeysSigners []crypto.PublicKey, message []byte, aggSig []byte) error
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

// CreateSignatureShareV2 -
func (mm *MultisignerMock) CreateSignatureShareV2(privateKeyBytes crypto.PrivateKey, message []byte) ([]byte, error) {
	if mm.CreateSignatureShareV2Called != nil {
		return mm.CreateSignatureShareV2Called(privateKeyBytes, message)
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

// VerifySignatureShareV2 -
func (mm *MultisignerMock) VerifySignatureShareV2(publicKey crypto.PublicKey, message []byte, sig []byte) error {
	if mm.VerifySignatureShareV2Called != nil {
		return mm.VerifySignatureShareV2Called(publicKey, message, sig)
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

// AggregateSigsV2 -
func (mm *MultisignerMock) AggregateSigsV2(pubKeysSigners []crypto.PublicKey, signatures [][]byte) ([]byte, error) {
	if mm.AggregateSigsV2Called != nil {
		return mm.AggregateSigsV2Called(pubKeysSigners, signatures)
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

// VerifyAggregatedSigV2 -
func (mm *MultisignerMock) VerifyAggregatedSigV2(pubKeysSigners []crypto.PublicKey, message []byte, aggSig []byte) error {
	if mm.VerifyAggregatedSigV2Called != nil {
		return mm.VerifyAggregatedSigV2Called(pubKeysSigners, message, aggSig)
	}
	return nil
}

// IsInterfaceNil -
func (mm *MultisignerMock) IsInterfaceNil() bool {
	return mm == nil
}
