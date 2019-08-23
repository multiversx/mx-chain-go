package singlesig

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/schnorr"
)

// SchnorrSigner is a SingleSigner implementation that uses a Schnorr signature scheme
type SchnorrSigner struct {
}

// Sign Signs a message with using a single signature schnorr scheme
func (s *SchnorrSigner) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	if private == nil {
		return nil, crypto.ErrNilPrivateKey
	}

	scalar := private.Scalar()
	if scalar == nil {
		return nil, crypto.ErrNilPrivateKeyScalar
	}

	suite := private.Suite()
	if suite == nil {
		return nil, crypto.ErrNilSuite
	}

	kSuite, ok := suite.GetUnderlyingSuite().(schnorr.Suite)
	if !ok {
		return nil, crypto.ErrInvalidSuite
	}

	kScalar, ok := scalar.GetUnderlyingObj().(kyber.Scalar)
	if !ok {
		return nil, crypto.ErrInvalidPrivateKey
	}

	return schnorr.Sign(kSuite, kScalar, msg)
}

// Verify verifies a signature using a single signature schnorr scheme
func (s *SchnorrSigner) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	if public == nil {
		return crypto.ErrNilPublicKey
	}

	if msg == nil {
		return crypto.ErrNilMessage
	}

	if sig == nil {
		return crypto.ErrNilSignature
	}

	suite := public.Suite()
	if suite == nil {
		return crypto.ErrNilSuite
	}

	point := public.Point()
	if point == nil {
		return crypto.ErrNilPublicKeyPoint
	}

	kSuite, ok := suite.GetUnderlyingSuite().(schnorr.Suite)
	if !ok {
		return crypto.ErrInvalidSuite
	}

	kPoint, ok := point.GetUnderlyingObj().(kyber.Point)
	if !ok {
		return crypto.ErrInvalidPublicKey
	}

	return schnorr.Verify(kSuite, kPoint, msg, sig)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SchnorrSigner) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}
