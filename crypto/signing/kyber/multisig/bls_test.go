package multisig_test

import (
	"crypto/cipher"
	"reflect"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/multisig"
	smock "github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/mock"
	"github.com/stretchr/testify/assert"
	"go.dedis.ch/kyber/v3/pairing"
)

const initScalar = 2
const initPointX = 3
const initPointY = 4

var invalidStr = []byte("invalid key")

func unmarshalPrivate(val []byte) (int, error) {
	if reflect.DeepEqual(invalidStr, val) {
		return 0, crypto.ErrInvalidPrivateKey
	}

	return initScalar, nil
}

func marshalPrivate(x int) ([]byte, error) {
	res := []byte(strconv.Itoa(x))
	return res, nil
}

func unmarshalPublic(val []byte) (x, y int, err error) {
	if reflect.DeepEqual(invalidStr, val) {
		return 0, 0, crypto.ErrInvalidPublicKey
	}
	return initPointX, initPointY, nil
}

func marshalPublic(x, y int) ([]byte, error) {
	resStr := strconv.Itoa(x)
	resStr += strconv.Itoa(y)
	res := []byte(resStr)

	return res, nil
}

func createScalar() crypto.Scalar {
	return &smock.ScalarMock{
		X:                   initScalar,
		UnmarshalBinaryStub: unmarshalPrivate,
		MarshalBinaryStub:   marshalPrivate,
	}
}

func createPoint() crypto.Point {
	return &smock.PointMock{
		X:                   initPointX,
		Y:                   initPointY,
		UnmarshalBinaryStub: unmarshalPublic,
		MarshalBinaryStub:   marshalPublic,
	}
}

func createKeyPair(_ cipher.Stream) (crypto.Scalar, crypto.Point) {
	scalar := createScalar()
	point, _ := createPoint().Mul(scalar)
	return scalar, point
}

func createMockSuite(innerSuite interface{}) crypto.Suite {
	suite := &smock.SuiteMock{
		CreateKeyPairStub: createKeyPair,
		CreateScalarStub:  createScalar,
		CreatePointStub:   createPoint,
		GetUnderlyingSuiteStub: func() interface{} {
			return innerSuite
		},
	}

	return suite
}

func genSigParamsBLS() (
	privKey crypto.PrivateKey,
	pubKey crypto.PublicKey,
	kg crypto.KeyGenerator,
	llSigner crypto.LowLevelSignerBLS,
) {
	suite := kyber.NewSuitePairingBn256()
	kg = signing.NewKeyGenerator(suite)
	llSigner = &multisig.KyberMultiSignerBLS{}

	privKey, pubKey = kg.GeneratePair()

	return privKey, pubKey, kg, llSigner
}

func createSigSharesBLS(
	nbSigs uint16,
	message []byte,
) (pubKeys []crypto.PublicKey, sigShares [][]byte) {
	suite := kyber.NewSuitePairingBn256()
	kg := signing.NewKeyGenerator(suite)

	pubKeys = make([]crypto.PublicKey, nbSigs)
	sigShares = make([][]byte, nbSigs)
	llSigner := &multisig.KyberMultiSignerBLS{}

	for i := uint16(0); i < nbSigs; i++ {
		sk, pk := kg.GeneratePair()
		pubKeys[i] = pk
		sigShares[i], _ = llSigner.SignShare(sk, message)

	}

	return pubKeys, sigShares
}

func pointsFromPubKeys(pubKeys []crypto.PublicKey) (pubKeysPoints []crypto.Point) {
	pubKeysPoints = make([]crypto.Point, len(pubKeys))

	for i := range pubKeys {
		pubKeysPoints[i] = pubKeys[i].Point()
	}

	return pubKeysPoints
}

func TestKyberMultiSignerBLS_SignShareNilPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	_, _, _, lls := genSigParamsBLS()

	sig, err := lls.SignShare(nil, msg)

	assert.Nil(t, sig)
	assert.Equal(t, crypto.ErrNilPrivateKey, err)
}

func TestKyberMultiSignerBLS_SignShareInvalidPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	_, _, _, lls := genSigParamsBLS()
	pk := &mock.PrivateKeyStub{
		ScalarHandler: func() crypto.Scalar {
			return &smock.ScalarMock{}
		},
	}

	sig, err := lls.SignShare(pk, msg)

	assert.Nil(t, sig)
	assert.Equal(t, crypto.ErrInvalidPrivateKey, err)
}

func TestKyberMultiSignerBLS_SignShareNilMsgShouldErr(t *testing.T) {
	t.Parallel()

	privKey, _, _, lls := genSigParamsBLS()
	sig, err := lls.SignShare(privKey, nil)

	assert.Nil(t, sig)
	assert.Equal(t, crypto.ErrNilMessage, err)
}

func TestKyberMultiSignerBLS_SignShareOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, _, _, lls := genSigParamsBLS()
	sig, err := lls.SignShare(privKey, msg)

	assert.NotNil(t, sig)
	assert.Nil(t, err)
}

func TestKyberMultiSignerBLS_VerifySigShareNilPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, _, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	err := lls.VerifySigShare(nil, msg, sig)

	assert.Equal(t, crypto.ErrNilPublicKey, err)
}

func TestKyberMultiSignerBLS_VerifySigShareInvalidPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, _, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	suite := createMockSuite(pairing.NewSuiteBn256())
	kg := signing.NewKeyGenerator(suite)
	pubKeyBytes := []byte("key")
	pubKey, err := kg.PublicKeyFromByteArray(pubKeyBytes)
	err = lls.VerifySigShare(pubKey, msg, sig)

	assert.Equal(t, crypto.ErrInvalidPublicKey, err)
}

func TestKyberMultiSignerBLS_VerifySigShareNilMsgShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	err := lls.VerifySigShare(pubKey, nil, sig)

	assert.Equal(t, crypto.ErrNilMessage, err)
}

func TestKyberMultiSignerBLS_VerifySigShareNilSigShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	_, pubKey, _, lls := genSigParamsBLS()
	err := lls.VerifySigShare(pubKey, msg, nil)

	assert.Equal(t, crypto.ErrNilSignature, err)
}

func TestKyberMultiSignerBLS_VerifySigShareInvalidSigShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	// change the message so signature becomes invalid
	msg2 := []byte("message2")
	err := lls.VerifySigShare(pubKey, msg2, sig)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid signature")
}

func TestKyberMultiSignerBLS_VerifySigShareOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	err := lls.VerifySigShare(pubKey, msg, sig)

	assert.Nil(t, err)
}

func TestKyberMultiSignerBLS_VerifySigBytesNilSigShouldErr(t *testing.T) {
	t.Parallel()

	_, pubKey, _, lls := genSigParamsBLS()
	err := lls.VerifySigBytes(pubKey.Suite(), nil)

	assert.Equal(t, crypto.ErrNilSignature, err)
}

func TestKyberMultiSignerBLS_VerifySigBytesInvalidSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, _, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)

	suite := createMockSuite("invalid suite")
	err := lls.VerifySigBytes(suite, sig)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "suite is invalid")
}

func TestKyberMultiSignerBLS_VerifySigBytesInvalidSigShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	// change the signature so that it becomes invalid
	sig[0] = sig[0] ^ 0xFF
	err := lls.VerifySigBytes(pubKey.Suite(), sig)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "malformed point")
}

func TestKyberMultiSignerBLS_AggregateSignaturesNilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	_, sigShares := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(nil, sigShares...)

	assert.Equal(t, crypto.ErrNilSuite, err)
}

func TestKyberMultiSignerBLS_AggregateSignaturesInvalidSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	_, sigShares := createSigSharesBLS(20, msg)
	suite := createMockSuite("invalid suite")
	_, err := llSig.AggregateSignatures(suite, sigShares...)

	assert.Equal(t, crypto.ErrInvalidSuite, err)
}

func TestKyberMultiSignerBLS_AggregateSignaturesNilSigsShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, _ := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite())

	assert.Equal(t, crypto.ErrNilSignaturesList, err)
}

func TestKyberMultiSignerBLS_AggregateSignaturesEmptySigsShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, _ := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), []byte(""))

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not enough data")
}

func TestKyberMultiSignerBLS_AggregateSignaturesInvalidSigsShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	// make first sig share invalid
	sigShares[0][0] = sigShares[0][0] ^ 0xFF
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares...)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "malformed point")
}

func TestKyberMultiSignerBLS_AggregateSignaturesOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares...)

	assert.Nil(t, err)
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigNilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares...)
	pubKeysPoints := make([]crypto.Point, len(pubKeys))

	for i := range pubKeys {
		pubKeysPoints = append(pubKeysPoints, pubKeys[i].Point())
	}

	aggPks, _ := llSig.AggregatePublicKeys(pubKeys[0].Suite(), pubKeysPoints...)
	err := llSig.VerifyAggregatedSig(nil, aggPks, aggSig, msg)

	assert.Equal(t, crypto.ErrNilSuite, err)
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigInvalidSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares...)
	pubKeysPoints := pointsFromPubKeys(pubKeys)
	aggPks, _ := llSig.AggregatePublicKeys(pubKeys[0].Suite(), pubKeysPoints...)
	err := llSig.VerifyAggregatedSig(nil, aggPks, aggSig, msg)

	assert.Equal(t, crypto.ErrNilSuite, err)
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigNilAggPKBytesShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares...)
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), nil, aggSig, msg)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not enough data")
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigInvalidAggPKBytesShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares...)
	pubKeysPoints := pointsFromPubKeys(pubKeys)
	aggPks, _ := llSig.AggregatePublicKeys(pubKeys[0].Suite(), pubKeysPoints...)
	//make aggregated pubKeys invalid
	aggPks[0] = aggPks[0] ^ 0xFF
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), aggPks, aggSig, msg)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "malformed point")
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigNilAggSigBytesShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, _ := createSigSharesBLS(20, msg)
	pubKeysPoints := pointsFromPubKeys(pubKeys)
	aggPks, _ := llSig.AggregatePublicKeys(pubKeys[0].Suite(), pubKeysPoints...)
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), aggPks, nil, msg)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not enough data")
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigInvalidAggSigBytesShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares...)
	pubKeysPoints := pointsFromPubKeys(pubKeys)
	aggPks, _ := llSig.AggregatePublicKeys(pubKeys[0].Suite(), pubKeysPoints...)
	//make aggregated sig invalid
	aggSig[0] = aggSig[0] ^ 0xFF
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), aggPks, aggSig, msg)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "malformed point")
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigNilMsgShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares...)
	pubKeysPoints := pointsFromPubKeys(pubKeys)
	aggPks, _ := llSig.AggregatePublicKeys(pubKeys[0].Suite(), pubKeysPoints...)
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), aggPks, aggSig, nil)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid signature")
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares...)
	pubKeysPoints := pointsFromPubKeys(pubKeys)
	aggPks, _ := llSig.AggregatePublicKeys(pubKeys[0].Suite(), pubKeysPoints...)
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), aggPks, aggSig, msg)

	assert.Nil(t, err)
}

func TestKyberMultiSignerBLS_AggregatePublicKeysNilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, _ := createSigSharesBLS(20, msg)
	pubKeysPoints := pointsFromPubKeys(pubKeys)
	aggPks, err := llSig.AggregatePublicKeys(nil, pubKeysPoints...)

	assert.Nil(t, aggPks)
	assert.Equal(t, crypto.ErrNilSuite, err)
}

func TestKyberMultiSignerBLS_AggregatePublicKeysInvalidSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, _ := createSigSharesBLS(20, msg)
	pubKeysPoints := pointsFromPubKeys(pubKeys)
	suite := createMockSuite("invalid suite")
	aggPks, err := llSig.AggregatePublicKeys(suite, pubKeysPoints...)

	assert.Nil(t, aggPks)
	assert.Equal(t, crypto.ErrInvalidSuite, err)
}

func TestKyberMultiSignerBLS_AggregatePublicKeysNilPubKeysShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, _ := createSigSharesBLS(20, msg)
	aggPks, err := llSig.AggregatePublicKeys(pubKeys[0].Suite())

	assert.Nil(t, aggPks)
	assert.Equal(t, crypto.ErrNilPublicKeys, err)
}

func TestKyberMultiSignerBLS_AggregatePublicKeysWithNilPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, _ := createSigSharesBLS(20, msg)
	emptyPubKeys := make([]crypto.Point, len(pubKeys))
	aggPks, err := llSig.AggregatePublicKeys(pubKeys[0].Suite(), emptyPubKeys...)

	assert.Nil(t, aggPks)
	assert.Equal(t, crypto.ErrNilPublicKeyPoint, err)
}

func TestKyberMultiSignerBLS_AggregatePublicKeysWithInvalidPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, _ := createSigSharesBLS(20, msg)
	pubKeysPoints := pointsFromPubKeys(pubKeys)
	// make a public key point invalid for the given suite
	pubKeysPoints[0] = createPoint() // this uses the mock suite so should be invalid for real suite
	aggPks, err := llSig.AggregatePublicKeys(pubKeys[0].Suite(), pubKeysPoints...)

	assert.Nil(t, aggPks)
	assert.Equal(t, crypto.ErrInvalidPublicKey, err)
}

func TestKyberMultiSignerBLS_AggregatePublicKeysOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	llSig := &multisig.KyberMultiSignerBLS{}
	pubKeys, _ := createSigSharesBLS(20, msg)
	pubKeysPoints := pointsFromPubKeys(pubKeys)
	aggPks, err := llSig.AggregatePublicKeys(pubKeys[0].Suite(), pubKeysPoints...)

	assert.NotNil(t, aggPks)
	assert.Nil(t, err)
}

func TestKyberMultiSignerBLS_ScalarMulSigNilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, llSig := genSigParamsBLS()
	sig, _ := llSig.SignShare(privKey, msg)
	rs := pubKey.Suite().RandomStream()
	scalar, _ := pubKey.Suite().CreateScalar().Pick(rs)
	res, err := llSig.ScalarMulSig(nil, scalar, sig)

	assert.Nil(t, res)
	assert.Equal(t, crypto.ErrNilSuite, err)
}

func TestKyberMultiSignerBLS_ScalarMulSigNilScalarShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, llSig := genSigParamsBLS()
	sig, _ := llSig.SignShare(privKey, msg)
	res, err := llSig.ScalarMulSig(pubKey.Suite(), nil, sig)

	assert.Nil(t, res)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestKyberMultiSignerBLS_ScalarMulSigNilSigShouldErr(t *testing.T) {
	t.Parallel()

	_, pubKey, _, llSig := genSigParamsBLS()
	rs := pubKey.Suite().RandomStream()
	scalar, _ := pubKey.Suite().CreateScalar().Pick(rs)
	res, err := llSig.ScalarMulSig(pubKey.Suite(), scalar, nil)

	assert.Nil(t, res)
	assert.Equal(t, crypto.ErrNilSignature, err)
}

func TestKyberMultiSignerBLS_ScalarMulSigOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, llSig := genSigParamsBLS()
	sig, _ := llSig.SignShare(privKey, msg)
	rs := pubKey.Suite().RandomStream()
	scalar, _ := pubKey.Suite().CreateScalar().Pick(rs)
	res, err := llSig.ScalarMulSig(pubKey.Suite(), scalar, sig)

	assert.NotNil(t, res)
	assert.Nil(t, err)
}
