package mock

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

type SingleSignerMock struct {
	SignStub   func(private crypto.PrivateKey, msg []byte) ([]byte, error)
	VerifyStub func(public crypto.PublicKey, msg []byte, sig []byte) error
}

func (s *SingleSignerMock) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	return s.SignStub(private, msg)
}

func (s *SingleSignerMock) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	return s.VerifyStub(public, msg, sig)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SingleSignerMock) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}
