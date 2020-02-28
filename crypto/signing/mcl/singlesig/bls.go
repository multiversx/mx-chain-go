package singlesig

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/bls-go-binary/bls"
)

// BlsSingleSigner is a SingleSigner implementation that uses a BLS signature scheme
type BlsSingleSigner struct {
}

func NewBlsSigner() *BlsSingleSigner {
	return &BlsSingleSigner{}
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

	mclScalar, ok := scalar.(*mcl.MclScalar)
	if !ok {
		return nil, crypto.ErrInvalidPrivateKey
	}

	sk := &bls.SecretKey{}
	bls.BlsFrToSecretKey(mclScalar.Scalar, sk)
	sig := sk.Sign(string(msg))

	return sig.Serialize(), nil
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

	point := public.Point()
	if check.IfNil(point) {
		return crypto.ErrNilPublicKeyPoint
	}

	pubKeyPoint, ok := point.(*mcl.PointG2)
	if !ok {
		return crypto.ErrInvalidPublicKey
	}

	mclPubKey := &bls.PublicKey{}

	bls.BlsG2ToPublicKey(pubKeyPoint.G2, mclPubKey)
	signature := &bls.Sign{}

	err := signature.Deserialize(sig)
	if err != nil {
		return err
	}

	if signature.Verify(mclPubKey, string(msg)) {
		return nil
	}

	return crypto.ErrSigNotValid
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *BlsSingleSigner) IsInterfaceNil() bool {
	return s == nil
}
