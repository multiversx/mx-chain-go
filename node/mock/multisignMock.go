package mock

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

// MultisignMock -
type MultisignMock struct {
}

// Create -
func (mm *MultisignMock) Create(pubKeys []string, index uint16) (crypto.MultiSigner, error) {
	panic("implement me")
}

// Reset -
func (mm *MultisignMock) Reset(pubKeys []string, index uint16) error {
	panic("implement me")
}

// SetAggregatedSig -
func (mm *MultisignMock) SetAggregatedSig([]byte) error {
	panic("implement me")
}

// Verify -
func (mm *MultisignMock) Verify(msg []byte, bitmap []byte) error {
	panic("implement me")
}

// CreateCommitment -
func (mm *MultisignMock) CreateCommitment() (commSecret []byte, commitment []byte) {
	panic("implement me")
}

// StoreCommitmentHash -
func (mm *MultisignMock) StoreCommitmentHash(index uint16, commHash []byte) error {
	panic("implement me")
}

// CommitmentHash -
func (mm *MultisignMock) CommitmentHash(index uint16) ([]byte, error) {
	panic("implement me")
}

// StoreCommitment -
func (mm *MultisignMock) StoreCommitment(index uint16, value []byte) error {
	panic("implement me")
}

// Commitment -
func (mm *MultisignMock) Commitment(index uint16) ([]byte, error) {
	panic("implement me")
}

// AggregateCommitments -
func (mm *MultisignMock) AggregateCommitments(bitmap []byte) error {
	panic("implement me")
}

// CreateSignatureShare -
func (mm *MultisignMock) CreateSignatureShare(msg []byte, bitmap []byte) ([]byte, error) {
	panic("implement me")
}

// StoreSignatureShare -
func (mm *MultisignMock) StoreSignatureShare(index uint16, sig []byte) error {
	panic("implement me")
}

// VerifySignatureShare -
func (mm *MultisignMock) VerifySignatureShare(index uint16, sig []byte, msg []byte, bitmap []byte) error {
	panic("implement me")
}

// SignatureShare -
func (mm *MultisignMock) SignatureShare(index uint16) ([]byte, error) {
	panic("implement me")
}

// AggregateSigs -
func (mm *MultisignMock) AggregateSigs(bitmap []byte) ([]byte, error) {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (mm *MultisignMock) IsInterfaceNil() bool {
	if mm == nil {
		return true
	}
	return false
}
