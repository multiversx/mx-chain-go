package cryptoMocks

import (
<<<<<<< HEAD
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
=======
	"bytes"
>>>>>>> rc/v1.4.0
)

const signatureSize = 48

// MultisignerMock is used to mock the multisignature scheme
type MultisignerMock struct {
<<<<<<< HEAD
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
	SetAggregatedSigCalled                 func(sig []byte) error
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
	if mm.SetAggregatedSigCalled != nil {
		return mm.SetAggregatedSigCalled(aggSig)
	}

	mm.aggSig = aggSig

	return nil
=======
	CreateSignatureShareCalled func(privateKeyBytes []byte, message []byte) ([]byte, error)
	VerifySignatureShareCalled func(publicKey []byte, message []byte, sig []byte) error
	AggregateSigsCalled        func(pubKeysSigners [][]byte, signatures [][]byte) ([]byte, error)
	VerifyAggregatedSigCalled  func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error
}

// NewMultiSigner -
func NewMultiSigner() *MultisignerMock {
	return &MultisignerMock{}
>>>>>>> rc/v1.4.0
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
