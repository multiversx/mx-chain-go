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
	CreateCommitmentMock     func() ([]byte, []byte)
	AggregateCommitmentsMock func(bitmap []byte) ([]byte, error)
	CreateSignatureShareMock func(bitmap []byte) ([]byte, error)
	VerifySignatureShareMock func(index uint16, sig []byte, bitmap []byte) error
	AggregateSigsMock        func(bitmap []byte) ([]byte, error)
	AddCommitmentMock        func(index uint16, value []byte) error
	SetCommitmentSecretMock  func([]byte) error
	AddCommitmentHashMock    func(uint16, []byte) error
	CommitmentMock           func(uint16) ([]byte, error)
}

func NewMultiSigner() *BelNevMock {
	multisigner := &BelNevMock{}
	multisigner.commitments = make([][]byte, 21)
	multisigner.sigs = make([][]byte, 21)
	multisigner.pubkeys = make([]string, 21)

	return multisigner
}

// Reset resets the multiSigner and initializes corresponding fields with the given params
func (bnm *BelNevMock) Create(pubKeys []string, index uint16) (crypto.MultiSigner, error) {
	multiSig := NewMultiSigner()

	multiSig.selfId = index
	multiSig.pubkeys = pubKeys

	return multiSig, nil
}

// SetMessage sets the message to be signed
func (bnm *BelNevMock) SetMessage(msg []byte) error {
	bnm.msg = msg

	return nil
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
func (bnm *BelNevMock) CreateCommitment() (commSecret []byte, commitment []byte) {

	return bnm.CreateCommitmentMock()
}

// SetCommitmentSecret sets the committment secret
func (bnm *BelNevMock) SetCommitmentSecret(commSecret []byte) error {
	if bnm.SetCommitmentSecretMock == nil {
		bnm.commSecret = commSecret

		return nil
	} else {
		return bnm.SetCommitmentSecretMock(commSecret)
	}
}

// AddCommitmentHash adds a commitment hash to the list on the specified position
func (bnm *BelNevMock) AddCommitmentHash(index uint16, commHash []byte) error {
	if bnm.AddCommitmentHashMock == nil {
		bnm.commHash = commHash

		return nil
	} else {
		return bnm.AddCommitmentHashMock(index, commHash)
	}
}

// CommitmentHash returns the commitment hash from the list on the specified position
func (bnm *BelNevMock) CommitmentHash(index uint16) ([]byte, error) {
	if bnm.CommitmentHashMock == nil {
		return bnm.commHash, nil
	} else {
		return bnm.CommitmentHashMock(index)
	}
}

// AddCommitment adds a commitment to the list on the specified position
func (bnm *BelNevMock) AddCommitment(index uint16, value []byte) error {
	if bnm.AddCommitmentMock == nil {
		if index >= uint16(len(bnm.commitments)) {
			return crypto.ErrIndexOutOfBounds
		}

		bnm.commitments[index] = value

		return nil
	} else {
		return bnm.AddCommitmentMock(index, value)
	}
}

// Commitment returns the commitment from the list with the specified position
func (bnm *BelNevMock) Commitment(index uint16) ([]byte, error) {
	if bnm.CommitmentMock == nil {
		if index >= uint16(len(bnm.commitments)) {
			return nil, crypto.ErrIndexOutOfBounds
		}

		return bnm.commitments[index], nil
	} else {
		return bnm.CommitmentMock(index)
	}
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

// CreateSignatureShare creates a partial signature
func (bnm *BelNevMock) CreateSignatureShare(bitmap []byte) ([]byte, error) {
	return bnm.CreateSignatureShareMock(bitmap)
}

// AddSignatureShare adds the partial signature of the signer with specified position
func (bnm *BelNevMock) AddSignatureShare(index uint16, sig []byte) error {
	if index >= uint16(len(bnm.pubkeys)) {
		return crypto.ErrIndexOutOfBounds
	}

	bnm.sigs[index] = sig
	return nil
}

// VerifySignatureShare verifies the partial signature of the signer with specified position
func (bnm *BelNevMock) VerifySignatureShare(index uint16, sig []byte, bitmap []byte) error {
	return bnm.VerifySignatureShareMock(index, sig, bitmap)
}

// AggregateSigs aggregates all collected partial signatures
func (bnm *BelNevMock) AggregateSigs(bitmap []byte) ([]byte, error) {
	return bnm.AggregateSigsMock(bitmap)
}

// SignatureShare
func (bnm *BelNevMock) SignatureShare(index uint16) ([]byte, error) {
	if index >= uint16(len(bnm.sigs)) {
		return nil, crypto.ErrIndexOutOfBounds
	}

	return bnm.sigs[index], nil
}
