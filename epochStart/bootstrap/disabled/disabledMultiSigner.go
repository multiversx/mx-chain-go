package disabled

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

type multiSigner struct {
}

// NewMultiSigner returns a new instance of multiSigner
func NewMultiSigner() *multiSigner {
	return &multiSigner{}
}

// Create -
func (m *multiSigner) Create(_ []string, _ uint16) (crypto.MultiSigner, error) {
	return nil, nil
}

// SetAggregatedSig -
func (m *multiSigner) SetAggregatedSig([]byte) error {
	return nil
}

// Verify -
func (m *multiSigner) Verify(_ []byte, _ []byte) error {
	return nil
}

// Reset -
func (m *multiSigner) Reset(_ []string, _ uint16) error {
	return nil
}

// CreateSignatureShare -
func (m *multiSigner) CreateSignatureShare(_ []byte, _ []byte) ([]byte, error) {
	return nil, nil
}

// StoreSignatureShare -
func (m *multiSigner) StoreSignatureShare(_ uint16, _ []byte) error {
	return nil
}

// SignatureShare -
func (m *multiSigner) SignatureShare(_ uint16) ([]byte, error) {
	return nil, nil
}

// VerifySignatureShare -
func (m *multiSigner) VerifySignatureShare(_ uint16, _ []byte, _ []byte, _ []byte) error {
	return nil
}

// AggregateSigs -
func (m *multiSigner) AggregateSigs(_ []byte) ([]byte, error) {
	return nil, nil
}

// IsInterfaceNil -
func (m *multiSigner) IsInterfaceNil() bool {
	return m == nil
}
