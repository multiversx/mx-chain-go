package disabled

import (
	"github.com/ElrondNetwork/elrond-go-crypto"
)

type multiSigner struct {
}

// NewMultiSigner returns a new instance of disabled multiSigner
func NewMultiSigner() *multiSigner {
	return &multiSigner{}
}

// Create returns a nil instance and a nil error
func (m *multiSigner) Create(_ []string, _ uint16) (crypto.MultiSigner, error) {
	return nil, nil
}

// SetAggregatedSig returns nil
func (m *multiSigner) SetAggregatedSig([]byte) error {
	return nil
}

// Verify returns nil
func (m *multiSigner) Verify(_ []byte, _ []byte) error {
	return nil
}

// Reset returns nil and does nothing
func (m *multiSigner) Reset(_ []string, _ uint16) error {
	return nil
}

// CreateSignatureShare returns nil byte slice and nil error
func (m *multiSigner) CreateSignatureShare(_ []byte, _ []byte) ([]byte, error) {
	return nil, nil
}

// StoreSignatureShare returns nil
func (m *multiSigner) StoreSignatureShare(_ uint16, _ []byte) error {
	return nil
}

// SignatureShare returns nil byte slice and a nil error
func (m *multiSigner) SignatureShare(_ uint16) ([]byte, error) {
	return nil, nil
}

// VerifySignatureShare returns nil
func (m *multiSigner) VerifySignatureShare(_ uint16, _ []byte, _ []byte, _ []byte) error {
	return nil
}

// AggregateSigs returns nil byte slice and nil error
func (m *multiSigner) AggregateSigs(_ []byte) ([]byte, error) {
	return nil, nil
}

// CreateAndAddSignatureShareForKey will return an empty slice and a nil error
func (m *multiSigner) CreateAndAddSignatureShareForKey(_ []byte, _ crypto.PrivateKey, _ []byte) ([]byte, error) {
	return make([]byte, 0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *multiSigner) IsInterfaceNil() bool {
	return m == nil
}
