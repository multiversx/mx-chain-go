package mock

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

// MultisignMock -
type MultisignMock struct {
}

// Create -
func (mm *MultisignMock) Create(_ []string, _ uint16) (crypto.MultiSigner, error) {
	panic("implement me")
}

// Reset -
func (mm *MultisignMock) Reset(_ []string, _ uint16) error {
	panic("implement me")
}

// SetAggregatedSig -
func (mm *MultisignMock) SetAggregatedSig([]byte) error {
	panic("implement me")
}

// Verify -
func (mm *MultisignMock) Verify(_ []byte, _ []byte) error {
	panic("implement me")
}

// CreateCommitment -
func (mm *MultisignMock) CreateCommitment() (commSecret []byte, commitment []byte) {
	panic("implement me")
}

// StoreCommitmentHash -
func (mm *MultisignMock) StoreCommitmentHash(_ uint16, _ []byte) error {
	panic("implement me")
}

// CommitmentHash -
func (mm *MultisignMock) CommitmentHash(_ uint16) ([]byte, error) {
	panic("implement me")
}

// StoreCommitment -
func (mm *MultisignMock) StoreCommitment(_ uint16, _ []byte) error {
	panic("implement me")
}

// Commitment -
func (mm *MultisignMock) Commitment(_ uint16) ([]byte, error) {
	panic("implement me")
}

// AggregateCommitments -
func (mm *MultisignMock) AggregateCommitments(_ []byte) error {
	panic("implement me")
}

// CreateSignatureShare -
func (mm *MultisignMock) CreateSignatureShare(_ []byte, _ []byte) ([]byte, error) {
	panic("implement me")
}

// StoreSignatureShare -
func (mm *MultisignMock) StoreSignatureShare(_ uint16, _ []byte) error {
	panic("implement me")
}

// VerifySignatureShare -
func (mm *MultisignMock) VerifySignatureShare(_ uint16, _ []byte, _ []byte, _ []byte) error {
	panic("implement me")
}

// SignatureShare -
func (mm *MultisignMock) SignatureShare(_ uint16) ([]byte, error) {
	panic("implement me")
}

// AggregateSigs -
func (mm *MultisignMock) AggregateSigs(_ []byte) ([]byte, error) {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (mm *MultisignMock) IsInterfaceNil() bool {
	return mm == nil
}
