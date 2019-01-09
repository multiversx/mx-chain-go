package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

// BelNevMock is used to mock belare neven multisignature scheme
type BelNevMock struct {
	msg         []byte
	aggSig      []byte
	aggCom      []byte
	commSecret  []byte
	commHash    []byte
	commitments [][]byte
	sigs        [][]byte
	pubkeys     []string
	privKey     crypto.PrivateKey
	selfId      uint16
	hasher      hashing.Hasher

	VerifyMock               func(bitmap []byte) error
	CommitmentHashMock       func(index uint16) ([]byte, error)
	CreateCommitmentMock     func() ([]byte, []byte, error)
	AggregateCommitmentsMock func(bitmap []byte) ([]byte, error)
	SignPartialMock          func(bitmap []byte) ([]byte, error)
	VerifyPartialMock        func(index uint16, sig []byte, bitmap []byte) error
	AggregateSigsMock        func(bitmap []byte) ([]byte, error)
}

func NewMultiSigner() *BelNevMock {
	multisigner := &BelNevMock{}
	multisigner.commitments = make([][]byte, 21)
	multisigner.sigs = make([][]byte, 21)
	multisigner.pubkeys = make([]string, 21)

	return multisigner
}

// Reset resets the multiSigner and initializes corresponding fields with the given params
func (bnm *BelNevMock) Reset(pubKeys []string, index uint16) error {
	bnm.msg = nil
	bnm.aggSig = nil
	bnm.commSecret = nil
	bnm.commHash = nil
	bnm.commitments = make([][]byte, len(pubKeys))
	bnm.sigs = make([][]byte, len(pubKeys))
	bnm.pubkeys = pubKeys
	bnm.selfId = index

	return nil
}

// SetMessage sets the message to be signed
func (bnm *BelNevMock) SetMessage(msg []byte) {
	bnm.msg = msg
}

// SetAggregatedSig sets the aggregated signature according to the given byte array
func (bnm *BelNevMock) SetAggregatedSig(aggSig []byte) error {
	bnm.aggSig = aggSig

	return nil
}

// Verify returns nil if the aggregateed signature is verified for the given public keys
func (bnm *BelNevMock) Verify(bitmap []byte) error {
	return bnm.VerifyMock(bitmap)
}

// CreateCommitment creates a secret commitment and the corresponding public commitment point
func (bnm *BelNevMock) CreateCommitment() (commSecret []byte, commitment []byte, err error) {

	return bnm.CreateCommitmentMock()
}

// SetCommitmentSecret sets the committment secret
func (bnm *BelNevMock) SetCommitmentSecret(commSecret []byte) error {
	bnm.commSecret = commSecret

	return nil
}

// AddCommitmentHash adds a commitment hash to the list on the specified position
func (bnm *BelNevMock) AddCommitmentHash(index uint16, commHash []byte) error {
	bnm.commHash = commHash

	return nil
}

// CommitmentHash returns the commitment hash from the list on the specified position
func (bnm *BelNevMock) CommitmentHash(index uint16) ([]byte, error) {
	return bnm.commHash, nil
}

// AddCommitment adds a commitment to the list on the specified position
func (bnm *BelNevMock) AddCommitment(index uint16, value []byte) error {
	if index >= uint16(len(bnm.commitments)) {
		return crypto.ErrInvalidIndex
	}

	bnm.commitments[index] = value

	return nil
}

// Commitment returns the commitment from the list with the specified position
func (bnm *BelNevMock) Commitment(index uint16) ([]byte, error) {
	if index >= uint16(len(bnm.commitments)) {
		return nil, crypto.ErrInvalidIndex
	}

	return bnm.commitments[index], nil
}

// AggregateCommitments aggregates the list of commitments
func (bnm *BelNevMock) AggregateCommitments(bitmap []byte) ([]byte, error) {
	return bnm.AggregateCommitmentsMock(bitmap)
}

// SetAggCommitment sets the aggregated commitment
func (bnm *BelNevMock) SetAggCommitment(aggCommitment []byte) error {
	bnm.aggCom = aggCommitment

	return nil
}

// SignPartial creates a partial signature
func (bnm *BelNevMock) SignPartial(bitmap []byte) ([]byte, error) {
	return bnm.SignPartialMock(bitmap)
}

// AddSignPartial adds the partial signature of the signer with specified position
func (bnm *BelNevMock) AddSignPartial(index uint16, sig []byte) error {
	if index >= uint16(len(bnm.pubkeys)) {
		return crypto.ErrInvalidIndex
	}

	bnm.sigs[index] = sig
	return nil
}

// VerifyPartial verifies the partial signature of the signer with specified position
func (bnm *BelNevMock) VerifyPartial(index uint16, sig []byte, bitmap []byte) error {
	return bnm.VerifyPartialMock(index, sig, bitmap)
}

// AggregateSigs aggregates all collected partial signatures
func (bnm *BelNevMock) AggregateSigs(bitmap []byte) ([]byte, error) {
	return bnm.AggregateSigsMock(bitmap)
}
