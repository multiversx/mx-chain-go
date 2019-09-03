package multisig

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign/bls"
)

// KyberMultiSignerBLS provides an implements of the crypto.LowLevelSignerBLS interface
type KyberMultiSignerBLS struct{}

// SignShare produces a BLS signature share (single BLS signature) over a given message
func (kms *KyberMultiSignerBLS) SignShare(privKey crypto.PrivateKey, message []byte) ([]byte, error) {
	blsSingleSigner := &singlesig.BlsSingleSigner{}

	return blsSingleSigner.Sign(privKey, message)
}

// VerifySigShare verifies a BLS signature share (single BLS signature) over a given message
func (kms *KyberMultiSignerBLS) VerifySigShare(pubKey crypto.PublicKey, message []byte, sig []byte) error {
	blsSingleSigner := &singlesig.BlsSingleSigner{}

	return blsSingleSigner.Verify(pubKey, message, sig)
}

// VerifySigBytes provides an "cheap" integrity check of a signature given as a byte array
// It does not validate the signature over a message, only verifies that it is a signature
func (kms *KyberMultiSignerBLS) VerifySigBytes(suite crypto.Suite, sig []byte) error {
	if suite == nil || suite.IsInterfaceNil() {
		return crypto.ErrNilSuite
	}

	if sig == nil {
		return crypto.ErrNilSignature
	}

	_, err := kms.sigBytesToKyberPoint(suite, sig)

	return err
}

// AggregateSignatures produces an aggregation of single BLS signatures over the same message
func (kms *KyberMultiSignerBLS) AggregateSignatures(suite crypto.Suite, sigs ...[]byte) ([]byte, error) {
	if suite == nil || suite.IsInterfaceNil() {
		return nil, crypto.ErrNilSuite
	}

	if sigs == nil {
		return nil, crypto.ErrNilSignaturesList
	}

	kSuite, ok := suite.GetUnderlyingSuite().(pairing.Suite)
	if !ok {
		return nil, crypto.ErrInvalidSuite
	}

	return bls.AggregateSignatures(kSuite, sigs...)
}

// VerifyAggregatedSig verifies if a BLS aggregated signature is valid over a given message
func (kms *KyberMultiSignerBLS) VerifyAggregatedSig(
	suite crypto.Suite,
	aggPointsBytes []byte,
	aggSigBytes []byte,
	msg []byte,
) error {
	if suite == nil || suite.IsInterfaceNil() {
		return crypto.ErrNilSuite
	}

	kSuite, ok := suite.GetUnderlyingSuite().(pairing.Suite)
	if !ok {
		return crypto.ErrInvalidSuite
	}

	aggKPoint := kSuite.G2().Point()
	err := aggKPoint.UnmarshalBinary(aggPointsBytes)
	if err != nil {
		return err
	}

	return bls.Verify(kSuite, aggKPoint, msg, aggSigBytes)
}

// AggregatePublicKeys produces an aggregation of BLS public keys (points)
func (kms *KyberMultiSignerBLS) AggregatePublicKeys(suite crypto.Suite, pubKeys ...crypto.Point) ([]byte, error) {
	if suite == nil || suite.IsInterfaceNil() {
		return nil, crypto.ErrNilSuite
	}

	if len(pubKeys) == 0 {
		return nil, crypto.ErrNilPublicKeys
	}

	kSuite, ok := suite.GetUnderlyingSuite().(pairing.Suite)
	if !ok {
		return nil, crypto.ErrInvalidSuite
	}

	kyberPoints := make([]kyber.Point, len(pubKeys))
	for i, pubKey := range pubKeys {
		if pubKey == nil {
			return nil, crypto.ErrNilPublicKeyPoint
		}

		kyberPoints[i], ok = pubKey.GetUnderlyingObj().(kyber.Point)
		if !ok {
			return nil, crypto.ErrInvalidPublicKey
		}
	}

	kyberAggPubKey := bls.AggregatePublicKeys(kSuite, kyberPoints...)

	return kyberAggPubKey.MarshalBinary()
}

// ScalarMulSig returns the result of multiplication of a scalar with a BLS signature
func (kms *KyberMultiSignerBLS) ScalarMulSig(suite crypto.Suite, scalar crypto.Scalar, sig []byte) ([]byte, error) {
	if suite == nil || suite.IsInterfaceNil() {
		return nil, crypto.ErrNilSuite
	}

	if scalar == nil || scalar.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	if sig == nil {
		return nil, crypto.ErrNilSignature
	}

	kScalar, ok := scalar.GetUnderlyingObj().(kyber.Scalar)
	if !ok {
		return nil, crypto.ErrInvalidScalar
	}

	sigKPoint, err := kms.sigBytesToKyberPoint(suite, sig)
	if err != nil {
		return nil, err
	}

	resPoint := sigKPoint.Mul(kScalar, sigKPoint)
	resBytes, err := resPoint.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return resBytes, nil
}

// sigBytesToKyberPoint returns the kyber point corresponding to the BLS signature byte array
func (kms *KyberMultiSignerBLS) sigBytesToKyberPoint(suite crypto.Suite, sig []byte) (kyber.Point, error) {
	kSuite, ok := suite.GetUnderlyingSuite().(pairing.Suite)
	if !ok {
		return nil, crypto.ErrInvalidSuite
	}

	sigKPoint := kSuite.G1().Point()
	err := sigKPoint.UnmarshalBinary(sig)
	if err != nil {
		return sigKPoint, err
	}

	return sigKPoint, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (kms *KyberMultiSignerBLS) IsInterfaceNil() bool {
	if kms == nil {
		return true
	}
	return false
}
