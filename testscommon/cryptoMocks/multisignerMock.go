package cryptoMocks

import (
	"github.com/ElrondNetwork/elrond-go-crypto"
)

const signatureSize = 48

// MultisignerMock is used to mock the multisignature scheme
type MultisignerMock struct {
	aggSig      []byte
	aggCom      []byte
	commSecret  []byte
	commHash    []byte
	commitments [][]byte
	sigs        [][]byte
	pubkeys     []string
	selfId      uint16

	VerifyCalled                           func(msg []byte, bitmap []byte) error
	CommitmentHashCalled                   func(index uint16) ([]byte, error)
	CreateCommitmentCalled                 func() ([]byte, []byte)
	AggregateCommitmentsCalled             func(bitmap []byte) error
	CreateSignatureShareCalled             func(msg []byte, bitmap []byte) ([]byte, error)
	VerifySignatureShareCalled             func(index uint16, sig []byte, msg []byte, bitmap []byte) error
	AggregateSigsCalled                    func(bitmap []byte) ([]byte, error)
	SignatureShareCalled                   func(index uint16) ([]byte, error)
	StoreCommitmentCalled                  func(index uint16, value []byte) error
	StoreCommitmentHashCalled              func(uint16, []byte) error
	CommitmentCalled                       func(uint16) ([]byte, error)
	CreateCalled                           func(pubKeys []string, index uint16) (crypto.MultiSigner, error)
	ResetCalled                            func(pubKeys []string, index uint16) error
	CreateAndAddSignatureShareForKeyCalled func(message []byte, privateKey crypto.PrivateKey, pubKeyBytes []byte) ([]byte, error)
}

// NewMultiSigner -
func NewMultiSigner(consensusSize uint32) *MultisignerMock {
	multisigner := &MultisignerMock{}
	multisigner.commitments = make([][]byte, consensusSize)
	multisigner.sigs = make([][]byte, consensusSize)
	multisigner.pubkeys = make([]string, consensusSize)

	multisigner.aggCom = []byte("agg commitment")
	multisigner.commHash = []byte("commitment hash")
	multisigner.commSecret = []byte("commitment secret")
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

	mm.commitments = make([][]byte, len(pubKeys))
	mm.sigs = make([][]byte, len(pubKeys))
	mm.pubkeys = make([]string, len(pubKeys))
	mm.selfId = index
	mm.pubkeys = pubKeys

	for i := 0; i < len(pubKeys); i++ {
		mm.commitments[i] = mm.commSecret
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

// CreateCommitment -
func (mm *MultisignerMock) CreateCommitment() (commSecret []byte, commitment []byte) {
	if mm.CreateCommitmentCalled != nil {
		cs, comm := mm.CreateCommitmentCalled()
		mm.commSecret = cs
		mm.commitments[mm.selfId] = comm

		return commSecret, comm
	}

	return mm.commSecret, mm.commitments[mm.selfId]
}

// StoreCommitmentHash adds a commitment hash to the list on the specified position
func (mm *MultisignerMock) StoreCommitmentHash(index uint16, commHash []byte) error {
	if mm.StoreCommitmentHashCalled == nil {
		mm.commHash = commHash

		return nil
	}

	return mm.StoreCommitmentHashCalled(index, commHash)
}

// CommitmentHash returns the commitment hash from the list on the specified position
func (mm *MultisignerMock) CommitmentHash(index uint16) ([]byte, error) {
	if mm.CommitmentHashCalled == nil {
		return mm.commHash, nil
	}

	return mm.CommitmentHashCalled(index)
}

// StoreCommitment adds a commitment to the list on the specified position
func (mm *MultisignerMock) StoreCommitment(index uint16, value []byte) error {
	if mm.StoreCommitmentCalled == nil {
		if index >= uint16(len(mm.commitments)) {
			return crypto.ErrIndexOutOfBounds
		}

		mm.commitments[index] = value

		return nil
	}

	return mm.StoreCommitmentCalled(index, value)
}

// Commitment returns the commitment from the list with the specified position
func (mm *MultisignerMock) Commitment(index uint16) ([]byte, error) {
	if mm.CommitmentCalled == nil {
		if index >= uint16(len(mm.commitments)) {
			return nil, crypto.ErrIndexOutOfBounds
		}

		return mm.commitments[index], nil
	}

	return mm.CommitmentCalled(index)
}

// AggregateCommitments aggregates the list of commitments
func (mm *MultisignerMock) AggregateCommitments(bitmap []byte) error {
	if mm.AggregateCommitmentsCalled != nil {
		return mm.AggregateCommitmentsCalled(bitmap)
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
