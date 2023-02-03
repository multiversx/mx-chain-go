package mock

import (
	"github.com/multiversx/mx-chain-crypto-go"
)

// SignerMock -
type SignerMock struct {
	SignStub   func(private crypto.PrivateKey, msg []byte) ([]byte, error)
	VerifyStub func(public crypto.PublicKey, msg []byte, sig []byte) error
}

// Sign -
func (s *SignerMock) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	return s.SignStub(private, msg)
}

// Verify -
func (s *SignerMock) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	return s.VerifyStub(public, msg, sig)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SignerMock) IsInterfaceNil() bool {
	return s == nil
}
