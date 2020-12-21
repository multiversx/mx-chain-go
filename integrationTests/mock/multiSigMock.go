package mock

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

// SignatureSize -
const SignatureSize = 48

// BelNevMock is used to mock belare neven multisignature scheme
type BelNevMock struct {
	aggSig      []byte
	aggCom      []byte
	commSecret  []byte
	commHash    []byte
	commitments [][]byte
	sigs        [][]byte
	pubkeys     []string
	selfId      uint16

	VerifyMock               func(msg []byte, bitmap []byte) error
	CommitmentHashMock       func(index uint16) ([]byte, error)
	CreateCommitmentMock     func() ([]byte, []byte)
	AggregateCommitmentsMock func(bitmap []byte) error
	CreateSignatureShareMock func(msg []byte, bitmap []byte) ([]byte, error)
	VerifySignatureShareMock func(index uint16, sig []byte, msg []byte, bitmap []byte) error
	AggregateSigsMock        func(bitmap []byte) ([]byte, error)
	SignatureShareMock       func(index uint16) ([]byte, error)
	StoreCommitmentMock      func(index uint16, value []byte) error
	StoreCommitmentHashMock  func(uint16, []byte) error
	CommitmentMock           func(uint16) ([]byte, error)
	CreateCalled             func(pubKeys []string, index uint16) (crypto.MultiSigner, error)
	ResetCalled              func(pubKeys []string, index uint16) error
}

// NewMultiSigner -
func NewMultiSigner(nrConsens uint32) *BelNevMock {
	multisigner := &BelNevMock{}
	multisigner.commitments = make([][]byte, nrConsens)
	multisigner.sigs = make([][]byte, nrConsens)
	multisigner.pubkeys = make([]string, nrConsens)

	multisigner.aggCom = []byte("agg commitment")
	multisigner.commHash = []byte("commitment hash")
	multisigner.commSecret = []byte("commitment secret")
	multisigner.aggSig = make([]byte, SignatureSize)
	copy(multisigner.aggSig, []byte("aggregated signature"))

	return multisigner
}

// Create creates a multiSigner using receiver as template and initializes corresponding fields with the given params
func (bnm *BelNevMock) Create(pubKeys []string, index uint16) (crypto.MultiSigner, error) {
	if bnm.CreateCalled != nil {
		return bnm.CreateCalled(pubKeys, index)
	}

	multiSig := NewMultiSigner(uint32(len(pubKeys)))

	multiSig.selfId = index
	multiSig.pubkeys = pubKeys

	return multiSig, nil
}

// Reset -
func (bnm *BelNevMock) Reset(pubKeys []string, index uint16) error {
	if bnm.ResetCalled != nil {
		return bnm.ResetCalled(pubKeys, index)
	}

	bnm.commitments = make([][]byte, len(pubKeys))
	bnm.sigs = make([][]byte, len(pubKeys))
	bnm.pubkeys = make([]string, len(pubKeys))

	for i := 0; i < len(pubKeys); i++ {
		bnm.commitments[i] = bnm.commSecret
		bnm.sigs[i] = bnm.aggSig
	}

	bnm.selfId = index
	bnm.pubkeys = pubKeys

	return nil
}

// SetAggregatedSig sets the aggregated signature according to the given byte array
func (bnm *BelNevMock) SetAggregatedSig(aggSig []byte) error {
	bnm.aggSig = aggSig

	return nil
}

// Verify returns nil if the aggregateed signature is verified for the given public keys
func (bnm *BelNevMock) Verify(msg []byte, bitmap []byte) error {
	if bnm.VerifyMock != nil {
		return bnm.VerifyMock(msg, bitmap)
	}

	return nil
}

// CreateCommitment creates a secret commitment and the corresponding public commitment point
func (bnm *BelNevMock) CreateCommitment() (commSecret []byte, commitment []byte) {
	if bnm.CreateCommitmentMock != nil {
		cs, comm := bnm.CreateCommitmentMock()
		bnm.commSecret = cs
		bnm.commitments[bnm.selfId] = comm

		return commSecret, comm
	}

	return bnm.commSecret, bnm.commitments[bnm.selfId]
}

// StoreCommitmentHash adds a commitment hash to the list on the specified position
func (bnm *BelNevMock) StoreCommitmentHash(index uint16, commHash []byte) error {
	if bnm.StoreCommitmentHashMock == nil {
		bnm.commHash = commHash

		return nil
	}

	return bnm.StoreCommitmentHashMock(index, commHash)
}

// CommitmentHash returns the commitment hash from the list on the specified position
func (bnm *BelNevMock) CommitmentHash(index uint16) ([]byte, error) {
	if bnm.CommitmentHashMock == nil {
		return bnm.commHash, nil
	}

	return bnm.CommitmentHashMock(index)
}

// StoreCommitment adds a commitment to the list on the specified position
func (bnm *BelNevMock) StoreCommitment(index uint16, value []byte) error {
	if bnm.StoreCommitmentMock == nil {
		if index >= uint16(len(bnm.commitments)) {
			return crypto.ErrIndexOutOfBounds
		}

		bnm.commitments[index] = value

		return nil
	}

	return bnm.StoreCommitmentMock(index, value)
}

// Commitment returns the commitment from the list with the specified position
func (bnm *BelNevMock) Commitment(index uint16) ([]byte, error) {
	if bnm.CommitmentMock == nil {
		if index >= uint16(len(bnm.commitments)) {
			return nil, crypto.ErrIndexOutOfBounds
		}

		return bnm.commitments[index], nil
	}

	return bnm.CommitmentMock(index)
}

// AggregateCommitments aggregates the list of commitments
func (bnm *BelNevMock) AggregateCommitments(bitmap []byte) error {
	if bnm.AggregateCommitmentsMock != nil {
		return bnm.AggregateCommitmentsMock(bitmap)
	}
	return nil
}

// CreateSignatureShare creates a partial signature
func (bnm *BelNevMock) CreateSignatureShare(msg []byte, bitmap []byte) ([]byte, error) {
	if bnm.CreateSignatureShareMock != nil {
		return bnm.CreateSignatureShareMock(msg, bitmap)
	}

	return bnm.aggSig, nil
}

// StoreSignatureShare adds the partial signature of the signer with specified position
func (bnm *BelNevMock) StoreSignatureShare(index uint16, sig []byte) error {
	if index >= uint16(len(bnm.pubkeys)) {
		return crypto.ErrIndexOutOfBounds
	}

	bnm.sigs[index] = sig
	return nil
}

// VerifySignatureShare verifies the partial signature of the signer with specified position
func (bnm *BelNevMock) VerifySignatureShare(index uint16, sig []byte, msg []byte, bitmap []byte) error {
	if bnm.VerifySignatureShareMock != nil {
		return bnm.VerifySignatureShareMock(index, sig, msg, bitmap)
	}

	return nil
}

// AggregateSigs aggregates all collected partial signatures
func (bnm *BelNevMock) AggregateSigs(bitmap []byte) ([]byte, error) {
	if bnm.AggregateSigsMock != nil {
		return bnm.AggregateSigsMock(bitmap)
	}
	return bnm.aggSig, nil
}

// SignatureShare -
func (bnm *BelNevMock) SignatureShare(index uint16) ([]byte, error) {
	if bnm.SignatureShareMock != nil {
		return bnm.SignatureShareMock(index)

	}

	if index >= uint16(len(bnm.sigs)) {
		return nil, crypto.ErrIndexOutOfBounds
	}

	return bnm.sigs[index], nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bnm *BelNevMock) IsInterfaceNil() bool {
	return bnm == nil
}
