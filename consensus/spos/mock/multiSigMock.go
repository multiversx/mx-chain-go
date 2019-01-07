package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

// BelNevMock is used to mock belare neven multisignature scheme
type BelNevMock struct {
	msg         []byte
	sigBitmap   []byte
	commBitmap  []byte
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

	VerifyMock               func() error
	CommitmentHashMock       func(index uint16) ([]byte, error)
	CommitmentBitmapMock     func() []byte
	CreateCommitmentMock     func() ([]byte, []byte, error)
	AggregateCommitmentsMock func() ([]byte, error)
	SignPartialMock          func() ([]byte, error)
	SigBitmapMock            func() []byte
	VerifyPartialMock        func(index uint16, sig []byte) error
	AggregateSigsMock        func() ([]byte, error)
}

func NewMultiSigner() *BelNevMock {
	multisigner := &BelNevMock{}
	multisigner.commitments = make([][]byte, 21)
	multisigner.sigs = make([][]byte, 21)
	multisigner.pubkeys = make([]string, 21)

	return multisigner
}

// NewMultiSiger instantiates another multiSigner of the same type
func (bnm *BelNevMock) NewMultiSiger(hasher hashing.Hasher, pubKeys []string, key crypto.PrivateKey, index uint16) (crypto.MultiSigner, error) {
	// take all stubs
	bnm2 := bnm

	bnm2.msg = nil
	bnm2.sigBitmap = nil
	bnm2.aggSig = nil
	bnm2.commSecret = nil
	bnm2.commHash = nil
	bnm2.commitments = make([][]byte, len(pubKeys))
	bnm2.sigs = make([][]byte, len(pubKeys))
	bnm2.hasher = hasher
	bnm2.pubkeys = pubKeys
	bnm2.privKey = key
	bnm2.selfId = index

	return bnm2, nil
}

// SetMessage sets the message to be signed
func (bnm *BelNevMock) SetMessage(msg []byte) {
	bnm.msg = msg
}

// SetSigBitmap sets the signers bitmap. Starting with index 0, each signer has 1 bit according to it's position in the
// signers list, set to 1 if signer's signature is used, 0 if not used
func (bnm *BelNevMock) SetSigBitmap(bmp []byte) error {
	bnm.sigBitmap = bmp

	return nil
}

// SetAggregatedSig sets the aggregated signature according to the given byte array
func (bnm *BelNevMock) SetAggregatedSig(aggSig []byte) error {
	bnm.aggSig = aggSig

	return nil
}

// Verify returns nil if the aggregateed signature is verified for the given public keys
func (bnm *BelNevMock) Verify() error {
	return bnm.VerifyMock()
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

// CommitmentBitmap returns the bitmap with the set
func (bnm *BelNevMock) CommitmentBitmap() []byte {
	return bnm.CommitmentBitmapMock()
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
func (bnm *BelNevMock) AggregateCommitments() ([]byte, error) {
	return bnm.AggregateCommitmentsMock()
}

// SetAggCommitment sets the aggregated commitment for the marked signers in bitmap
func (bnm *BelNevMock) SetAggCommitment(aggCommitment []byte, bitmap []byte) error {
	bnm.aggCom = aggCommitment
	bnm.commBitmap = bitmap

	return nil
}

// SignPartial creates a partial signature
func (bnm *BelNevMock) SignPartial() ([]byte, error) {
	return bnm.SignPartialMock()
}

// SigBitmap returns the bitmap for the set partial signatures
func (bnm *BelNevMock) SigBitmap() []byte {
	return bnm.SigBitmapMock()
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
func (bnm *BelNevMock) VerifyPartial(index uint16, sig []byte) error {
	return bnm.VerifyPartialMock(index, sig)
}

// AggregateSigs aggregates all collected partial signatures
func (bnm *BelNevMock) AggregateSigs() ([]byte, error) {
	return bnm.AggregateSigsMock()
}
