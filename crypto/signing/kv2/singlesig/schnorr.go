package singlesig

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"gopkg.in/dedis/kyber.v2"
	"gopkg.in/dedis/kyber.v2/group/edwards25519"
	"gopkg.in/dedis/kyber.v2/sign/schnorr"
)

type SchnorrSigner struct {
}

// Sign Signs a message with using a single signature schnorr scheme
func (s *SchnorrSigner) Sign(suite crypto.Suite, private crypto.Scalar, msg []byte) ([]byte, error) {

	if suite == nil {
		return nil, crypto.ErrNilSuite
	}

	if private == nil {
		return nil, crypto.ErrNilPrivateKey
	}

	kScalar, ok := private.GetUnderlyingObj().(kyber.Scalar)

	if !ok {
		return nil, crypto.ErrInvalidPrivateKey
	}

	kSuite, ok := suite.GetUnderlyingSuite().(*edwards25519.SuiteEd25519)

	if !ok {
		return nil, crypto.ErrInvalidSuite
	}

	return schnorr.Sign(kSuite, kScalar, msg)
}

// Verify verifies a signature using a single signature schnorr scheme
func (s *SchnorrSigner) Verify(suite crypto.Suite, public crypto.Point, msg []byte, sig []byte) error {
	if suite == nil {
		return crypto.ErrNilSuite
	}

	if public == nil {
		return crypto.ErrNilPublicKey
	}

	if msg == nil {
		return crypto.ErrNilMessage
	}

	if sig == nil {
		return crypto.ErrNilSignature
	}

	kSuite, ok := suite.GetUnderlyingSuite().(*edwards25519.SuiteEd25519)

	if !ok {
		return crypto.ErrInvalidSuite
	}

	kPoint, ok := public.GetUnderlyingObj().(kyber.Point)

	if !ok {
		return crypto.ErrInvalidPublicKey
	}

	return schnorr.Verify(kSuite, kPoint, msg, sig)
}
