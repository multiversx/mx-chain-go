package multisig

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign/bls"
)

func VerifySigBytesBLS(suite crypto.Suite, sig []byte) error {
	if suite == nil {
		return crypto.ErrNilSuite
	}

	if sig == nil {
		return crypto.ErrNilSignature
	}

	_, err := kyberSigPointBLS(suite, sig)

	return err
}

func ScalarMulSigBLS(suite crypto.Suite, scalar crypto.Scalar, sig []byte) ([]byte, error) {
	if sig == nil {
		return nil, crypto.ErrNilParam
	}

	kScalar, ok := scalar.GetUnderlyingObj().(kyber.Scalar)
	if !ok {
		return nil, crypto.ErrInvalidScalar
	}

	sigKPoint, err := kyberSigPointBLS(suite, sig)
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

// AggregateSignatures produces an aggregation of single BLS signatures
func AggregateSignatures(suite crypto.Suite, sigs ...[]byte) ([]byte, error) {
	if suite == nil {
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

// VerifyAggregatedSig verifies if a BLS aggregated signature is valid
func VerifyAggregatedSig(suite crypto.Suite, aggPointsBytes []byte, aggSigBytes []byte, msg []byte) error {
	if suite == nil {
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

// AggregateKyberPublicKeys produces an aggregation of BLS public keys (points)
func AggregateKyberPublicKeys(suite crypto.Suite, pubKeys ...crypto.Point) ([]byte, error) {
	if pubKeys == nil {
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

func kyberSigPointBLS(suite crypto.Suite, sig []byte) (kyber.Point, error) {
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
