package mock

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

// SignerMock -
type SignerMock struct {
	SignStub   func(private crypto.PrivateKey, msg []byte) ([]byte, error)
	VerifyStub func(public crypto.PublicKey, msg []byte, sig []byte) error
}

// Sign -
func (s *SignerMock) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	if s.SignStub != nil {
		return s.SignStub(private, msg)
	}

	return []byte("signature"), nil
}

// Verify -
func (s *SignerMock) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	if s.VerifyStub != nil {
		return s.VerifyStub(public, msg, sig)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SignerMock) IsInterfaceNil() bool {
	return s == nil
}
