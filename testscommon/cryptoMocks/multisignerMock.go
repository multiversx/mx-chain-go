package cryptoMocks

import (
	"github.com/ElrondNetwork/elrond-go-crypto"
)

const signatureSize = 48

// MultisignerMock is used to mock the multisignature scheme
type MultisignerMock struct {
	aggSig  []byte
	sigs    [][]byte
	pubkeys []string
	selfId  uint16

	VerifyCalled                           func(msg []byte, bitmap []byte) error
	CommitmentHashCalled                   func(index uint16) ([]byte, error)
	CreateSignatureShareCalled             func(msg []byte, bitmap []byte) ([]byte, error)
	VerifySignatureShareCalled             func(index uint16, sig []byte, msg []byte, bitmap []byte) error
	AggregateSigsCalled                    func(bitmap []byte) ([]byte, error)
	SignatureShareCalled                   func(index uint16) ([]byte, error)
	CreateCalled                           func(pubKeys []string, index uint16) (crypto.MultiSigner, error)
	ResetCalled                            func(pubKeys []string, index uint16) error
	CreateAndAddSignatureShareForKeyCalled func(message []byte, privateKey crypto.PrivateKey, pubKeyBytes []byte) ([]byte, error)
	StoreSignatureShareCalled              func(index uint16, sig []byte) error
}

// NewMultiSigner -
func NewMultiSigner(consensusSize uint32) *MultisignerMock {
	multisigner := &MultisignerMock{}
	multisigner.sigs = make([][]byte, consensusSize)
	multisigner.pubkeys = make([]string, consensusSize)

	multisigner.aggSig = make([]byte, signatureSize)
	copy(multisigner.aggSig, "aggregated signature")

	return multisigner
}

// Create -
func (mm *MultisignerMock) Create(pubKeys []string, index uint16) (crypto.MultiSigner, error) {
	if mm.CreateCalled != nil {
		return mm.CreateCalled(pubKeys, index)
	}

	multiSig := NewMultiSigner(uint32(len(pubKeys)))

	multiSig.selfId = index
	multiSig.pubkeys = pubKeys

	return multiSig, nil
}

// Reset -
func (mm *MultisignerMock) Reset(pubKeys []string, index uint16) error {
	if mm.ResetCalled != nil {
		return mm.ResetCalled(pubKeys, index)
	}

	mm.sigs = make([][]byte, len(pubKeys))
	mm.pubkeys = make([]string, len(pubKeys))
	mm.selfId = index
	mm.pubkeys = pubKeys

	for i := 0; i < len(pubKeys); i++ {
		mm.sigs[i] = mm.aggSig
	}

	mm.selfId = index
	mm.pubkeys = pubKeys

	return nil
}

// SetAggregatedSig -
func (mm *MultisignerMock) SetAggregatedSig(aggSig []byte) error {
	mm.aggSig = aggSig

	return nil
}

// Verify -
func (mm *MultisignerMock) Verify(msg []byte, bitmap []byte) error {
	if mm.VerifyCalled != nil {
		return mm.VerifyCalled(msg, bitmap)
	}

	return nil
}

// CreateSignatureShare creates a partial signature
func (mm *MultisignerMock) CreateSignatureShare(msg []byte, bitmap []byte) ([]byte, error) {
	if mm.CreateSignatureShareCalled != nil {
		return mm.CreateSignatureShareCalled(msg, bitmap)
	}

	return mm.aggSig, nil
}

// StoreSignatureShare -
func (mm *MultisignerMock) StoreSignatureShare(index uint16, sig []byte) error {
	if mm.StoreSignatureShareCalled != nil {
		return mm.StoreSignatureShareCalled(index, sig)
	}

	if index >= uint16(len(mm.pubkeys)) {
		return crypto.ErrIndexOutOfBounds
	}

	mm.sigs[index] = sig
	return nil
}

// VerifySignatureShare -
func (mm *MultisignerMock) VerifySignatureShare(index uint16, sig []byte, msg []byte, bitmap []byte) error {
	if mm.VerifySignatureShareCalled != nil {
		return mm.VerifySignatureShareCalled(index, sig, msg, bitmap)
	}

	return nil
}

// AggregateSigs -
func (mm *MultisignerMock) AggregateSigs(bitmap []byte) ([]byte, error) {
	if mm.AggregateSigsCalled != nil {
		return mm.AggregateSigsCalled(bitmap)
	}

	return mm.aggSig, nil
}

// SignatureShare -
func (mm *MultisignerMock) SignatureShare(index uint16) ([]byte, error) {
	if mm.SignatureShareCalled != nil {
		return mm.SignatureShareCalled(index)
	}

	if index >= uint16(len(mm.sigs)) {
		return nil, crypto.ErrIndexOutOfBounds
	}

	return mm.sigs[index], nil
}

// CreateAndAddSignatureShareForKey -
func (mm *MultisignerMock) CreateAndAddSignatureShareForKey(message []byte, privateKey crypto.PrivateKey, pubKeyBytes []byte) ([]byte, error) {
	if mm.CreateAndAddSignatureShareForKeyCalled != nil {
		return mm.CreateAndAddSignatureShareForKeyCalled(message, privateKey, pubKeyBytes)
	}

	return nil, nil
}

// IsInterfaceNil -
func (mm *MultisignerMock) IsInterfaceNil() bool {
	return mm == nil
}
