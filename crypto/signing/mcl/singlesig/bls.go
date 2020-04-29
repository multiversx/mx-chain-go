package singlesig

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/herumi/bls-go-binary/bls"
)

var _ crypto.SingleSigner = (*BlsSingleSigner)(nil)

// BlsSingleSigner is a SingleSigner implementation that uses a BLS signature scheme
type BlsSingleSigner struct {
}

// NewBlsSigner creates a BLS single signer instance
func NewBlsSigner() *BlsSingleSigner {
	return &BlsSingleSigner{}
}

// Sign Signs a message using a single signature BLS scheme
func (s *BlsSingleSigner) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	if check.IfNil(private) {
		return nil, crypto.ErrNilPrivateKey
	}
	if len(msg) == 0 {
		return nil, crypto.ErrNilMessage
	}

	scalar := private.Scalar()
	if check.IfNil(scalar) {
		return nil, crypto.ErrNilPrivateKeyScalar
	}

	mclScalar, ok := scalar.(*mcl.Scalar)
	if !ok || !IsSecretKeyValid(mclScalar) {
		return nil, crypto.ErrInvalidPrivateKey
	}

	sk := bls.CastToSecretKey(mclScalar.Scalar)
	sig := sk.Sign(string(msg))

	return sig.Serialize(), nil
}

// Verify verifies a signature using a single signature BLS scheme
func (s *BlsSingleSigner) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	if check.IfNil(public) {
		return crypto.ErrNilPublicKey
	}
	if len(msg) == 0 {
		return crypto.ErrNilMessage
	}
	if len(sig) == 0 {
		return crypto.ErrNilSignature
	}

	point := public.Point()
	if check.IfNil(point) {
		return crypto.ErrNilPublicKeyPoint
	}

	pubKeyPoint, isPoint := point.(*mcl.PointG2)
	if !isPoint || !IsPubKeyPointValid(pubKeyPoint) {
		return crypto.ErrInvalidPublicKey
	}

	mclPubKey := bls.CastToPublicKey(pubKeyPoint.G2)
	signature := &bls.Sign{}

	err := signature.Deserialize(sig)
	if err != nil {
		return err
	}

	if !IsSigValidPoint(signature) {
		return crypto.ErrBLSInvalidSignature
	}

	if signature.Verify(mclPubKey, string(msg)) {
		return nil
	}

	return crypto.ErrSigNotValid
}

// IsPubKeyPointValid validates the public key is a valid point on G2
func IsPubKeyPointValid(pubKeyPoint *mcl.PointG2) bool {
	return !pubKeyPoint.IsZero() && pubKeyPoint.IsValidOrder() && pubKeyPoint.IsValid()
}

// IsSigValidPoint validates that the signature isi a valid point on G1
func IsSigValidPoint(sig *bls.Sign) bool {
	g1Sig := bls.CastFromSign(sig)
	return !g1Sig.IsZero() && g1Sig.IsValidOrder() && g1Sig.IsValid()
}

// IsSecretKeyValid validates  that the scalar is a valid secret key
func IsSecretKeyValid(scalar *mcl.Scalar) bool {
	return !scalar.Scalar.IsZero() && scalar.Scalar.IsValid()
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *BlsSingleSigner) IsInterfaceNil() bool {
	return s == nil
}
