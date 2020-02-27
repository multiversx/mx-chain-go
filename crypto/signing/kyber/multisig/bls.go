package multisig

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign/bls"
)

const hasherOutputSize = 16

// KyberMultiSignerBLS provides an implements of the crypto.LowLevelSignerBLS interface
type KyberMultiSignerBLS struct {
	Hasher hashing.Hasher // 16bytes output hasher!
}

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
	if check.IfNil(suite) {
		return crypto.ErrNilSuite
	}
	if sig == nil {
		return crypto.ErrNilSignature
	}

	_, err := kms.sigBytesToKyberPoint(suite, sig)

	return err
}

// AggregateSignatures produces an aggregation of single BLS signatures over the same message
func (kms *KyberMultiSignerBLS) AggregateSignatures(
	suite crypto.Suite,
	signatures [][]byte,
	pubKeysSigners []crypto.PublicKey,
) ([]byte, error) {
	if check.IfNil(suite) {
		return nil, crypto.ErrNilSuite
	}
	if signatures == nil {
		return nil, crypto.ErrNilSignaturesList
	}
	if pubKeysSigners == nil {
		return nil, crypto.ErrNilPublicKeys
	}

	prepSigs, err := kms.prepareSignatures(suite, signatures, pubKeysSigners)
	if err != nil {
		return nil, err
	}

	aggSigs, err := kms.aggregatePreparedSignatures(suite, prepSigs...)
	if err != nil {
		return nil, err
	}

	return aggSigs, nil
}

// VerifyAggregatedSig verifies if a BLS aggregated signature is valid over a given message
func (kms *KyberMultiSignerBLS) VerifyAggregatedSig(
	suite crypto.Suite,
	pubKeys []crypto.PublicKey,
	aggSigBytes []byte,
	msg []byte,
) error {
	if check.IfNil(suite) {
		return crypto.ErrNilSuite
	}
	kSuite, ok := suite.GetUnderlyingSuite().(pairing.Suite)
	if !ok {
		return crypto.ErrInvalidSuite
	}

	pubKeysPoints := make([]crypto.Point, len(pubKeys))
	for i, pubKey := range pubKeys {
		if check.IfNil(pubKey) {
			return crypto.ErrNilPublicKey
		}
		pubKeysPoints[i] = pubKey.Point()
	}

	aggPointsBytes, err := kms.aggregatePublicKeys(suite, pubKeysPoints)
	if err != nil {
		return err
	}

	aggKPoint := kSuite.G2().Point()
	err = aggKPoint.UnmarshalBinary(aggPointsBytes)
	if err != nil {
		return err
	}

	return bls.Verify(kSuite, aggKPoint, msg, aggSigBytes)
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

func (kms *KyberMultiSignerBLS) prepareSignatures(
	suite crypto.Suite,
	signatures [][]byte,
	pubKeysSigners []crypto.PublicKey,
) ([][]byte, error) {
	prepSigs := make([][]byte, 0)
	for i := range signatures {
		hPk, err := hashPublicKeyPoint(kms.Hasher, pubKeysSigners[i].Point())
		if err != nil {
			return nil, err
		}
		// H1(pubKey_i)*sig_i
		s, err := kms.scalarMulSig(suite, hPk, signatures[i])
		if err != nil {
			return nil, err
		}

		prepSigs = append(prepSigs, s)
	}

	if len(prepSigs) == 0 {
		return nil, crypto.ErrNilSignaturesList
	}

	return prepSigs, nil
}

// aggregatePreparedSignatures produces an aggregation of single BLS signatures over the same message
func (kms *KyberMultiSignerBLS) aggregatePreparedSignatures(suite crypto.Suite, sigs ...[]byte) ([]byte, error) {
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

func (kms *KyberMultiSignerBLS) aggregatePublicKeys(
	suite crypto.Suite,
	pubKeys []crypto.Point,
) ([]byte, error) {

	prepPubKeysPoints, err := preparePublicKeys(pubKeys, kms.Hasher, suite)
	if err != nil {
		return nil, err
	}

	aggPointsBytes, err := kms.aggregatePreparedPublicKeys(suite, prepPubKeysPoints...)

	return aggPointsBytes, err
}

func preparePublicKeys(
	pubKeys []crypto.Point,
	hasher hashing.Hasher,
	suite crypto.Suite,
) ([]crypto.Point, error) {
	prepPubKeysPoints := make([]crypto.Point, 0)
	for i := range pubKeys {
		// t_i = H(pubKey_i)
		hPk, err := hashPublicKeyPoint(hasher, pubKeys[i])
		if err != nil {
			return nil, err
		}

		// t_i*pubKey_i
		prepPoint, err := scalarMulPk(suite, hPk, pubKeys[i])
		if err != nil {
			return nil, err
		}

		prepPubKeysPoints = append(prepPubKeysPoints, prepPoint)
	}

	return prepPubKeysPoints, nil
}

// aggregatePreparedPublicKeys produces an aggregation of BLS public keys (points)
func (kms *KyberMultiSignerBLS) aggregatePreparedPublicKeys(suite crypto.Suite, pubKeys ...crypto.Point) ([]byte, error) {
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

// scalarMulPk returns the result of multiplying a scalar given as a bytes array, with a BLS public key (point)
func scalarMulPk(suite crypto.Suite, scalarBytes []byte, pk crypto.Point) (crypto.Point, error) {
	if pk == nil || pk.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	kScalar, err := createScalar(suite, scalarBytes)
	if err != nil {
		return nil, err
	}

	pkPoint, err := pk.Mul(kScalar)

	return pkPoint, nil
}

// scalarMulSig returns the result of multiplying a scalar given as a bytes array, with a BLS single signature
func (kms *KyberMultiSignerBLS) scalarMulSig(suite crypto.Suite, scalarBytes []byte, sig []byte) ([]byte, error) {
	if check.IfNil(suite) {
		return nil, crypto.ErrNilSuite
	}
	if scalarBytes == nil {
		return nil, crypto.ErrNilParam
	}

	scalar, err := createScalar(suite, scalarBytes)
	if err != nil {
		return nil, err
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

// hashPublicKeyPoint hashes a BLS public key (point) into a byte array (32 bytes length)
func hashPublicKeyPoint(hasher hashing.Hasher, pubKeyPoint crypto.Point) ([]byte, error) {
	if check.IfNil(hasher) {
		return nil, crypto.ErrNilHasher
	}
	if hasher.Size() != hasherOutputSize {
		return nil, crypto.ErrWrongSizeHasher
	}
	if check.IfNil(pubKeyPoint) {
		return nil, crypto.ErrNilPublicKeyPoint
	}

	pointBytes, err := pubKeyPoint.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// H1(pubkey_i)
	h := hasher.Compute(string(pointBytes))
	// accepted length 32, copy the hasherOutputSize bytes and have rest 0
	h32 := make([]byte, 32)
	copy(h32[hasherOutputSize:], h)

	return h32, nil
}

// createScalar creates crypto.Scalar from a byte array
func createScalar(suite crypto.Suite, scalarBytes []byte) (crypto.Scalar, error) {
	if check.IfNil(suite) {
		return nil, crypto.ErrNilSuite
	}

	scalar := suite.CreateScalar()
	err := scalar.UnmarshalBinary(scalarBytes)
	if err != nil {
		return nil, err
	}

	return scalar, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (kms *KyberMultiSignerBLS) IsInterfaceNil() bool {
	return kms == nil
}
