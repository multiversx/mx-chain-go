package mock

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

type Signer struct {
	SignStub   func(suite crypto.Suite, private crypto.Scalar, msg []byte) ([]byte, error)
	VerifyStub func(suite crypto.Suite, public crypto.Point, msg []byte, sig []byte) error
}

func (s *Signer) Sign(suite crypto.Suite, private crypto.Scalar, msg []byte) ([]byte, error) {
	return s.SignStub(suite, private, msg)
}

func (s *Signer) Verify(suite crypto.Suite, public crypto.Point, msg []byte, sig []byte) error {
	return s.VerifyStub(suite, public, msg, sig)
}
