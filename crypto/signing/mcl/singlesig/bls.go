package singlesig

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/herumi/bls-go-binary/bls"
)

// BlsSingleSigner is a SingleSigner implementation that uses a BLS signature scheme
type BlsSingleSigner struct {
}

func NewBlsSigner() *BlsSingleSigner {
	return &BlsSingleSigner{}
}

// Sign Signs a message using a single signature BLS scheme
// TODO: check when mcl-bls-go lib updates that the secret key and public key types are exported and conversion is done
// correctly between secret key and scalar and public key and point
func (s *BlsSingleSigner) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	if private == nil || private.IsInterfaceNil() {
		return nil, crypto.ErrNilPrivateKey
	}

	if msg == nil {
		return nil, crypto.ErrNilMessage
	}

	scalar := private.Scalar()
	if scalar == nil || scalar.IsInterfaceNil() {
		return nil, crypto.ErrNilPrivateKeyScalar
	}

	mclScalar, ok := scalar.(*mcl.MclScalar)
	if !ok {
		return nil, crypto.ErrInvalidPrivateKey
	}

	// TODO: check after lib update if conversion functions are available
	sk := &bls.SecretKey{}
	_ = sk.Deserialize(mclScalar.Scalar.Serialize())
	sig := sk.Sign(string(msg))

	return sig.Serialize(), nil
}

// Verify verifies a signature using a single signature BLS scheme
func (s *BlsSingleSigner) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	if public == nil || public.IsInterfaceNil() {
		return crypto.ErrNilPublicKey
	}

	if msg == nil {
		return crypto.ErrNilMessage
	}

	if sig == nil {
		return crypto.ErrNilSignature
	}

	suite := public.Suite()
	if suite == nil || suite.IsInterfaceNil() {
		return crypto.ErrNilSuite
	}

	point := public.Point()
	if point == nil || point.IsInterfaceNil() {
		return crypto.ErrNilPublicKeyPoint
	}

	pubKeyPoint, ok := point.(*mcl.PointG2)
	if !ok {
		return crypto.ErrInvalidPublicKey
	}

	pubKeyPointBytes := pubKeyPoint.SerializeUncompressed()

	// TODO: check after lib update if conversion functions are available
	pubKey := &bls.PublicKey{}
	_ = pubKey.Deserialize(pubKeyPointBytes)
	signature := &bls.Sign{}

	err := signature.Deserialize(sig)
	if err != nil {
		return err
	}

	if signature.Verify(pubKey, string(msg)) {
		return nil
	}

	return crypto.ErrSigNotValid
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *BlsSingleSigner) IsInterfaceNil() bool {
	return s == nil
}
