package mock

import (
	"bytes"
	"errors"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

// BelNevMock is used to mock belare neven multisignature scheme
type BelNevMock struct {
}

var sigValue = []byte("signature")

func NewMultiSigner() *BelNevMock {
	return &BelNevMock{}
}

// Create resets the multiSigner and initializes corresponding fields with the given params
func (bnm *BelNevMock) Create(pubKeys []string, index uint16) (crypto.MultiSigner, error) {
	multiSig := NewMultiSigner()
	return multiSig, nil
}

// Reset
func (bnm *BelNevMock) Reset(pubKeys []string, index uint16) error {
	return nil
}

// SetAggregatedSig sets the aggregated signature according to the given byte array
func (bnm *BelNevMock) SetAggregatedSig(aggSig []byte) error {
	return nil
}

// Verify returns nil if the aggregateed signature is verified for the given public keys
func (bnm *BelNevMock) Verify(_ []byte, bitmap []byte) error {
	if bytes.Equal(bitmap, sigValue) {
		return nil
	}
	return errors.New("sig not valid")
}

// CreateCommitment creates a secret commitment and the corresponding public commitment point
func (bnm *BelNevMock) CreateCommitment() (commSecret []byte, commitment []byte) {
	return make([]byte, 0), make([]byte, 0)
}

// StoreCommitmentHash adds a commitment hash to the list on the specified position
func (bnm *BelNevMock) StoreCommitmentHash(index uint16, commHash []byte) error {
	return nil
}

// CommitmentHash returns the commitment hash from the list on the specified position
func (bnm *BelNevMock) CommitmentHash(index uint16) ([]byte, error) {
	return make([]byte, 0), nil
}

// StoreCommitment adds a commitment to the list on the specified position
func (bnm *BelNevMock) StoreCommitment(index uint16, value []byte) error {
	return nil
}

// Commitment returns the commitment from the list with the specified position
func (bnm *BelNevMock) Commitment(index uint16) ([]byte, error) {
	return make([]byte, 0), nil
}

// AggregateCommitments aggregates the list of commitments
func (bnm *BelNevMock) AggregateCommitments(bitmap []byte) error {
	return nil
}

// CreateSignatureShare creates a partial signature
func (bnm *BelNevMock) CreateSignatureShare(msg []byte, bitmap []byte) ([]byte, error) {
	return make([]byte, 0), nil
}

// StoreSignatureShare adds the partial signature of the signer with specified position
func (bnm *BelNevMock) StoreSignatureShare(index uint16, sig []byte) error {
	return nil
}

// VerifySignatureShare verifies the partial signature of the signer with specified position
func (bnm *BelNevMock) VerifySignatureShare(index uint16, sig []byte, msg []byte, bitmap []byte) error {
	return nil
}

// AggregateSigs aggregates all collected partial signatures
func (bnm *BelNevMock) AggregateSigs(bitmap []byte) ([]byte, error) {
	return sigValue, nil
}

// SignatureShare
func (bnm *BelNevMock) SignatureShare(index uint16) ([]byte, error) {
	return make([]byte, 0), nil
}
