package multisig

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/bls-go-binary/bls"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/hashing"
)

const hasherOutputSize = 16

// BlsMultiSigner provides an implements of the crypto.LowLevelSignerBLS interface
type BlsMultiSigner struct {
	Hasher hashing.Hasher // 16bytes output hasher!
}

// SignShare produces a BLS signature share (single BLS signature) over a given message
func (bms *BlsMultiSigner) SignShare(privKey crypto.PrivateKey, message []byte) ([]byte, error) {
	blsSingleSigner := singlesig.NewBlsSigner()

	return blsSingleSigner.Sign(privKey, message)
}

// VerifySigShare verifies a BLS signature share (single BLS signature) over a given message
func (bms *BlsMultiSigner) VerifySigShare(pubKey crypto.PublicKey, message []byte, sig []byte) error {
	blsSingleSigner := singlesig.NewBlsSigner()

	return blsSingleSigner.Verify(pubKey, message, sig)
}

// VerifySigBytes provides an "cheap" integrity check of a signature given as a byte array
// It does not validate the signature over a message, only verifies that it is a signature
func (bms *BlsMultiSigner) VerifySigBytes(suite crypto.Suite, sig []byte) error {
	if check.IfNil(suite) {
		return crypto.ErrNilSuite
	}
	if sig == nil || len(sig) == 0 {
		return crypto.ErrNilSignature
	}
	_, ok := suite.GetUnderlyingSuite().(*mcl.SuiteBLS12)
	if !ok {
		return crypto.ErrInvalidSuite
	}

	_, err := bms.sigBytesToPoint(sig)

	return err
}

// AggregateSignatures produces an aggregation of single BLS signatures over the same message
func (bms *BlsMultiSigner) AggregateSignatures(
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
	_, ok := suite.GetUnderlyingSuite().(*mcl.SuiteBLS12)
	if !ok {
		return nil, crypto.ErrInvalidSuite
	}

	sigsBLS, err := bms.prepareSignatures(suite, signatures, pubKeysSigners)
	if err != nil {
		return nil, err
	}

	aggSigBLS := &bls.Sign{}
	bls.BlsAggregateSignatures(aggSigBLS, sigsBLS)
	aggSigBytes := aggSigBLS.Serialize()

	return aggSigBytes, nil
}

// VerifyAggregatedSig verifies if a BLS aggregated signature is valid over a given message
func (bms *BlsMultiSigner) VerifyAggregatedSig(
	suite crypto.Suite,
	pubKeys []crypto.PublicKey,
	aggSigBytes []byte,
	msg []byte,
) error {
	if check.IfNil(suite) {
		return crypto.ErrNilSuite
	}
	if pubKeys == nil || len(pubKeys) == 0 {
		return crypto.ErrNilPublicKeys
	}
	if aggSigBytes == nil || len(aggSigBytes) == 0 {
		return crypto.ErrNilSignature
	}
	if msg == nil || len(msg) == 0 {
		return crypto.ErrNilMessage
	}

	_, ok := suite.GetUnderlyingSuite().(*mcl.SuiteBLS12)
	if !ok {
		return crypto.ErrInvalidSuite
	}

	pubKeyPoint := make([]crypto.Point, len(pubKeys))
	for i, pubKey := range pubKeys {
		if check.IfNil(pubKey) {
			return crypto.ErrNilPublicKey
		}

		pubKeyPoint[i] = pubKey.Point()
	}

	preparedPubKeys, err := preparePublicKeys(pubKeyPoint, bms.Hasher, suite)
	if err != nil {
		return err
	}

	aggSig := &bls.Sign{}
	err = aggSig.Deserialize(aggSigBytes)
	if err != nil {
		return err
	}

	res := bls.BlsFastAggregateVerify(aggSig, preparedPubKeys, string(msg))
	if !res {
		return crypto.ErrAggSigNotValid
	}

	return nil
}

func preparePublicKeys(
	pubKeysPoints []crypto.Point,
	hasher hashing.Hasher,
	suite crypto.Suite,
) ([]bls.PublicKey, error) {
	prepPubKeysPoints := make([]bls.PublicKey, len(pubKeysPoints))
	for i, pubKeyPoint := range pubKeysPoints {
		// t_i = H(pubKey_i)
		hPk, err := hashPublicKeyPoint(hasher, pubKeyPoint)
		if err != nil {
			return nil, err
		}

		// t_i*pubKey_i
		prepPublicKeyPoint, err := scalarMulPk(suite, hPk, pubKeyPoint)
		if err != nil {
			return nil, err
		}

		prepPubKeyG2, ok := prepPublicKeyPoint.GetUnderlyingObj().(*bls.G2)
		if !ok {
			return nil, crypto.ErrInvalidPoint
		}
		bls.BlsG2ToPublicKey(prepPubKeyG2, &prepPubKeysPoints[i])
	}

	return prepPubKeysPoints, nil
}

func (bms *BlsMultiSigner) prepareSignatures(
	suite crypto.Suite,
	signatures [][]byte,
	pubKeysSigners []crypto.PublicKey,
) ([]bls.Sign, error) {
	prepSigs := make([]bls.Sign, 0)
	for i, sig := range signatures {
		sigBLS := &bls.Sign{}
		if sig == nil || len(sig) == 0 {
			return nil, crypto.ErrNilSignature
		}

		err := sigBLS.Deserialize(sig)
		if err != nil {
			return nil, err
		}

		sigPoint := mcl.NewPointG1()
		bls.BlsSignatureToG1(sigBLS, sigPoint.G1)

		pubKeyPoint := pubKeysSigners[i].Point()
		hPk, err := hashPublicKeyPoint(bms.Hasher, pubKeyPoint)
		if err != nil {
			return nil, err
		}
		// H1(pubKey_i)*sig_i
		sPointG1, err := bms.scalarMulSig(suite, hPk, sigPoint)
		if err != nil {
			return nil, err
		}

		bls.BlsG1ToSignature(sPointG1.G1, sigBLS)
		prepSigs = append(prepSigs, *sigBLS)
	}

	if len(prepSigs) == 0 {
		return nil, crypto.ErrNilSignaturesList
	}

	return prepSigs, nil
}

// scalarMulPk returns the result of multiplying a scalar given as a bytes array, with a BLS public key (point)
func scalarMulPk(suite crypto.Suite, scalarBytes []byte, pk crypto.Point) (crypto.Point, error) {
	if pk == nil {
		return nil, crypto.ErrNilParam
	}

	scalar, err := createScalar(suite, scalarBytes)
	if err != nil {
		return nil, err
	}

	pkPoint, err := pk.Mul(scalar)

	return pkPoint, nil
}

// ScalarMulSig returns the result of multiplication of a scalar with a BLS signature
func (bms *BlsMultiSigner) scalarMulSig(suite crypto.Suite, scalarBytes []byte, sigPoint *mcl.PointG1) (*mcl.PointG1, error) {
	if scalarBytes == nil || len(scalarBytes) == 0 {
		return nil, crypto.ErrNilParam
	}
	if sigPoint == nil {
		return nil, crypto.ErrNilSignature
	}

	scalar := suite.CreateScalar()
	sc, _ := scalar.(*mcl.MclScalar)
	err := sc.Scalar.SetString(core.ToHex(scalarBytes), 16)
	if err != nil {
		return nil, crypto.ErrInvalidScalar
	}

	resPoint, err := sigPoint.Mul(scalar)
	if err != nil {
		return nil, err
	}

	resPointG1, ok := resPoint.(*mcl.PointG1)
	if !ok {
		return nil, crypto.ErrInvalidPoint
	}

	return resPointG1, nil
}

// sigBytesToPoint returns the point corresponding to the BLS signature byte array
func (bms *BlsMultiSigner) sigBytesToPoint(sig []byte) (crypto.Point, error) {
	sigBLS := &bls.Sign{}
	err := sigBLS.Deserialize(sig)
	if err != nil {
		return nil, err
	}

	pG1 := mcl.NewPointG1()
	bls.BlsSignatureToG1(sigBLS, pG1.G1)

	return pG1, nil
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
	sc, _ := scalar.(*mcl.MclScalar)

	err := sc.Scalar.SetString(core.ToHex(scalarBytes), 16)
	if err != nil {
		return nil, err
	}

	return scalar, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bms *BlsMultiSigner) IsInterfaceNil() bool {
	return bms == nil
}
