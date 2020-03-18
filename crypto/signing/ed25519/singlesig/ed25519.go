package singlesig

import (
	"crypto/ed25519"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
)

// Ed25519Signer exposes the signing and verification functionalities from the ed25519 signature scheme
type Ed25519Signer struct{}

// Sign will sign a message using ed25519 signature scheme
func (e *Ed25519Signer) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	if check.IfNil(private) {
		return nil, crypto.ErrNilPrivateKey
	}

	ed25519Scalar, ok := private.Scalar().GetUnderlyingObj().(ed25519.PrivateKey)
	if !ok {
		return nil, crypto.ErrInvalidPrivateKey
	}
	if len(ed25519Scalar) != ed25519.PrivateKeySize {
		return nil, crypto.ErrInvalidPrivateKey
	}

	sig := ed25519.Sign(ed25519Scalar, msg)

	return sig, nil
}

// Verify verifies a signature using a single signature ed25519 scheme
func (e *Ed25519Signer) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	if check.IfNil(public) {
		return crypto.ErrNilPublicKey
	}

	ed25519Point, ok := public.Point().GetUnderlyingObj().(ed25519.PublicKey)
	if !ok {
		return crypto.ErrInvalidPublicKey
	}
	if len(ed25519Point) != ed25519.PublicKeySize {
		return crypto.ErrInvalidPublicKey
	}

	isValidSig := ed25519.Verify(ed25519Point, msg, sig)
	if !isValidSig {
		return crypto.ErrEd25519InvalidSignature
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (e *Ed25519Signer) IsInterfaceNil() bool {
	return e == nil
}
