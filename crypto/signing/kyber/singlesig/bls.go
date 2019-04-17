package singlesig

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign/bls"
)

// BlsSingleSigner is a SingleSigner implementation that uses a BLS signature scheme
type BlsSingleSigner struct {
}

// Sign Signs a message using a single signature BLS scheme
func (s *BlsSingleSigner) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	if private == nil {
		return nil, crypto.ErrNilPrivateKey
	}

	scalar := private.Scalar()
	if scalar == nil {
		return nil, crypto.ErrNilPrivateKeyScalar
	}

	kScalar, ok := scalar.GetUnderlyingObj().(kyber.Scalar)

	if !ok {
		return nil, crypto.ErrInvalidPrivateKey
	}

	suite := private.Suite()
	if suite == nil {
		return nil, crypto.ErrNilSuite
	}

	kSuite, ok := suite.GetUnderlyingSuite().(pairing.Suite)

	if !ok {
		return nil, crypto.ErrInvalidSuite
	}

	return bls.Sign(kSuite, kScalar, msg)
}

// Verify verifies a signature using a single signature BLS scheme
func (s *BlsSingleSigner) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
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

	kSuite, ok := suite.GetUnderlyingSuite().(pairing.Suite)

	if !ok {
		return crypto.ErrInvalidSuite
	}

	point := public.Point()
	if point == nil {
		return crypto.ErrNilPublicKeyPoint
	}

	kPoint, ok := point.GetUnderlyingObj().(kyber.Point)

	if !ok {
		return crypto.ErrInvalidPublicKey
	}

	return bls.Verify(kSuite, kPoint, msg, sig)
}
