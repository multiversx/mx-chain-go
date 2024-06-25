package cryptoMocks

import (
	"github.com/multiversx/mx-chain-crypto-go"
)

// SignerStub -
type SignerStub struct {
	SignCalled   func(private crypto.PrivateKey, msg []byte) ([]byte, error)
	VerifyCalled func(public crypto.PublicKey, msg []byte, sig []byte) error
}

// Sign -
func (s *SignerStub) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	return s.SignCalled(private, msg)
}

// Verify -
func (s *SignerStub) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	if s.VerifyCalled != nil {
		return s.VerifyCalled(public, msg, sig)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SignerStub) IsInterfaceNil() bool {
	return s == nil
}
