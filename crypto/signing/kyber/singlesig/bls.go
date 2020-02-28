package singlesig

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign/bls"
)

// BlsSingleSigner is a SingleSigner implementation that uses a BLS signature scheme
type BlsSingleSigner struct {
}

// Sign Signs a message using a single signature BLS scheme
func (s *BlsSingleSigner) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	if check.IfNil(private) {
		return nil, crypto.ErrNilPrivateKey
	}

	if msg == nil {
		return nil, crypto.ErrNilMessage
	}

	scalar := private.Scalar()
	if check.IfNil(scalar) {
		return nil, crypto.ErrNilPrivateKeyScalar
	}

	kScalar, ok := scalar.GetUnderlyingObj().(kyber.Scalar)
	if !ok {
		return nil, crypto.ErrInvalidPrivateKey
	}

	suite := private.Suite()
	if check.IfNil(suite) {
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
	if check.IfNil(public) {
		return crypto.ErrNilPublicKey
	}
	if msg == nil {
		return crypto.ErrNilMessage
	}
	if sig == nil {
		return crypto.ErrNilSignature
	}

	suite := public.Suite()
	if check.IfNil(suite) {
		return crypto.ErrNilSuite
	}

	point := public.Point()
	if check.IfNil(point) {
		return crypto.ErrNilPublicKeyPoint
	}

	kSuite, ok := suite.GetUnderlyingSuite().(pairing.Suite)
	if !ok {
		return crypto.ErrInvalidSuite
	}

	kPoint, ok := point.GetUnderlyingObj().(kyber.Point)
	if !ok {
		return crypto.ErrInvalidPublicKey
	}

	return bls.Verify(kSuite, kPoint, msg, sig)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *BlsSingleSigner) IsInterfaceNil() bool {
	return s == nil
}
