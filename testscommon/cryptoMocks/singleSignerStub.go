package cryptoMocks

import "github.com/multiversx/mx-chain-crypto-go"

// SingleSignerStub -
type SingleSignerStub struct {
	SignCalled   func(private crypto.PrivateKey, msg []byte) ([]byte, error)
	VerifyCalled func(public crypto.PublicKey, msg []byte, sig []byte) error
}

// Sign -
func (s *SingleSignerStub) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	if s.SignCalled != nil {
		return s.SignCalled(private, msg)
	}

	return nil, nil
}

// Verify -
func (s *SingleSignerStub) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	if s.VerifyCalled != nil {
		return s.VerifyCalled(public, msg, sig)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SingleSignerStub) IsInterfaceNil() bool {
	return s == nil
}
