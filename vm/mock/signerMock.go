package mock

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

type SignerMock struct {
	SignCalled   func(private crypto.PrivateKey, msg []byte) ([]byte, error)
	VerifyCalled func(public crypto.PublicKey, msg []byte, sig []byte) error
}

func (s *SignerMock) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	if s.SignCalled == nil {
		return nil, nil
	}

	return s.SignCalled(private, msg)
}

func (s *SignerMock) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	if s.VerifyCalled == nil {
		return nil
	}

	return s.VerifyCalled(public, msg, sig)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SignerMock) IsInterfaceNil() bool {
	return s == nil
}
