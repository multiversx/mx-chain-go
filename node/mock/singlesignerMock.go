package mock

import (
	"bytes"
	"errors"

	"github.com/ElrondNetwork/elrond-go/crypto"
)

type SinglesignMock struct {
}

// Sign Signs a message with using a single signature schnorr scheme
func (s *SinglesignMock) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	return []byte("signed"), nil
}

// Verify verifies a signature using a single signature schnorr scheme
func (s *SinglesignMock) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	verSig := []byte("signed")

	if !bytes.Equal(sig, verSig) {
		return crypto.ErrSigNotValid
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SinglesignMock) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}

type SinglesignFailMock struct {
}

// Sign Signs a message with using a single signature schnorr scheme
func (s *SinglesignFailMock) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	return nil, errors.New("signing failure")
}

// Verify verifies a signature using a single signature schnorr scheme
func (s *SinglesignFailMock) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	return errors.New("signature verification failure")
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SinglesignFailMock) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}

type SinglesignStub struct {
	SignCalled   func(private crypto.PrivateKey, msg []byte) ([]byte, error)
	VerifyCalled func(public crypto.PublicKey, msg []byte, sig []byte) error
}

// Sign Signs a message with using a single signature schnorr scheme
func (s *SinglesignStub) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	return s.SignCalled(private, msg)
}

// Verify verifies a signature using a single signature schnorr scheme
func (s *SinglesignStub) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	return s.VerifyCalled(public, msg, sig)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SinglesignStub) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}
