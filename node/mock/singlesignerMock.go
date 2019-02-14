package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"bytes"
)

type SinglesignMock struct {
}

// Sign Signs a message with using a single signature schnorr scheme
func (s *SinglesignMock) Sign(suite crypto.Suite, private crypto.Scalar, msg []byte) ([]byte, error) {
	return []byte("signed" + string(msg)), nil
}

// Verify verifies a signature using a single signature schnorr scheme
func (s *SinglesignMock) Verify(suite crypto.Suite, public crypto.Point, msg []byte, sig []byte) error {
	verSig := []byte("signed" + string(msg))

	if !bytes.Equal(sig, verSig) {
		return crypto.ErrSigNotValid
	}
	return nil
}
