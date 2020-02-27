package multisig_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/mock"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/multisig"
	"github.com/stretchr/testify/require"
)

const initPointX = 3
const initPointY = 4

var invalidStr = []byte("invalid key")

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

func createPoint() crypto.Point {
	return &mock.PointMock{
		X:                   initPointX,
		Y:                   initPointY,
		UnmarshalBinaryStub: unmarshalPublic,
		MarshalBinaryStub:   marshalPublic,
	}
}

func createMockSuite(innerSuite interface{}) crypto.Suite {
	validSuite := kyber.NewSuitePairingBn256()

	suite := &mock.SuiteMock{
		CreateKeyPairStub: validSuite.CreateKeyPair,
		CreateScalarStub:  validSuite.CreateScalar,
		CreatePointStub:   validSuite.CreatePoint,
		GetUnderlyingSuiteStub: func() interface{} {
			if innerSuite == "invalid suite" {
				return innerSuite
			}
			return validSuite
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
	hasher := &mock.HasherSpongeMock{}
	llSigner = &multisig.KyberMultiSignerBLS{Hasher: hasher}

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

	require.Equal(t, crypto.ErrNilPrivateKey, err)
	require.Nil(t, sig)
}

func TestKyberMultiSignerBLS_SignShareInvalidPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	_, _, _, lls := genSigParamsBLS()
	pk := &mock.PrivateKeyStub{
		ScalarStub: func() crypto.Scalar {
			return &mock.ScalarMock{}
		},
	}

	sig, err := lls.SignShare(pk, msg)

	require.Equal(t, crypto.ErrInvalidPrivateKey, err)
	require.Nil(t, sig)
}

func TestKyberMultiSignerBLS_SignShareNilMsgShouldErr(t *testing.T) {
	t.Parallel()

	privKey, _, _, lls := genSigParamsBLS()
	sig, err := lls.SignShare(privKey, nil)

	require.Equal(t, crypto.ErrNilMessage, err)
	require.Nil(t, sig)
}

func TestKyberMultiSignerBLS_SignShareOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, _, _, lls := genSigParamsBLS()
	sig, err := lls.SignShare(privKey, msg)

	require.Nil(t, err)
	require.NotNil(t, sig)

}

func TestKyberMultiSignerBLS_VerifySigShareNilPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, _, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	err := lls.VerifySigShare(nil, msg, sig)

	require.Equal(t, crypto.ErrNilPublicKey, err)
}

func TestKyberMultiSignerBLS_VerifySigShareInvalidPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, _, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)

	pubKey := &mock.PublicKeyStub{
		ToByteArrayStub: func() (bytes []byte, err error) {
			return []byte("invalid key"), nil
		},
		PointStub: func() crypto.Point {
			return &mock.PointMock{}
		},
		SuiteStub: func() crypto.Suite {
			return kyber.NewSuitePairingBn256()
		},
	}

	err := lls.VerifySigShare(pubKey, msg, sig)

	require.Equal(t, crypto.ErrInvalidPublicKey, err)
}

func TestKyberMultiSignerBLS_VerifySigShareNilMsgShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	err := lls.VerifySigShare(pubKey, nil, sig)

	require.Equal(t, crypto.ErrNilMessage, err)
}

func TestKyberMultiSignerBLS_VerifySigShareNilSigShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	_, pubKey, _, lls := genSigParamsBLS()
	err := lls.VerifySigShare(pubKey, msg, nil)

	require.Equal(t, crypto.ErrNilSignature, err)
}

func TestKyberMultiSignerBLS_VerifySigShareInvalidSigShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	// change the message so signature becomes invalid
	msg2 := []byte("message2")
	err := lls.VerifySigShare(pubKey, msg2, sig)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "invalid signature")
}

func TestKyberMultiSignerBLS_VerifySigShareOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	err := lls.VerifySigShare(pubKey, msg, sig)

	require.Nil(t, err)
}

func TestKyberMultiSignerBLS_VerifySigBytesNilSigShouldErr(t *testing.T) {
	t.Parallel()

	_, pubKey, _, lls := genSigParamsBLS()
	err := lls.VerifySigBytes(pubKey.Suite(), nil)

	require.Equal(t, crypto.ErrNilSignature, err)
}

func TestKyberMultiSignerBLS_VerifySigBytesInvalidSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, _, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)

	suite := createMockSuite("invalid suite")
	err := lls.VerifySigBytes(suite, sig)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "suite is invalid")
}

func TestKyberMultiSignerBLS_VerifySigBytesInvalidSigShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	// change the signature so that it becomes invalid
	sig[0] = sig[0] ^ 0xFF
	err := lls.VerifySigBytes(pubKey.Suite(), sig)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "malformed point")
}

func TestKyberMultiSignerBLS_AggregateSignaturesNilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(nil, sigShares, pubKeys)

	require.Equal(t, crypto.ErrNilSuite, err)
}

func TestKyberMultiSignerBLS_AggregateSignaturesInvalidSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	suite := createMockSuite("invalid suite")
	_, err := llSig.AggregateSignatures(suite, sigShares, pubKeys)

	require.Equal(t, crypto.ErrInvalidSuite, err)
}

func TestKyberMultiSignerBLS_AggregateSignaturesNilSigsShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), nil, pubKeys)

	require.Equal(t, crypto.ErrNilSignaturesList, err)
}

func TestKyberMultiSignerBLS_AggregateSignaturesEmptySigsShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), [][]byte{[]byte("")}, pubKeys)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "not enough data")
}

func TestKyberMultiSignerBLS_AggregateSignaturesInvalidSigsShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	// make first sig share invalid
	sigShares[0][0] = sigShares[0][0] ^ 0xFF
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "malformed point")
}

func TestKyberMultiSignerBLS_AggregateSignaturesOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)

	require.Nil(t, err)
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigNilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	pubKeysPoints := make([]crypto.Point, len(pubKeys))

	for i := range pubKeys {
		pubKeysPoints = append(pubKeysPoints, pubKeys[i].Point())
	}

	err := llSig.VerifyAggregatedSig(nil, pubKeys, aggSig, msg)

	require.Equal(t, crypto.ErrNilSuite, err)
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigInvalidSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	suite := createMockSuite("invalid suite")
	err := llSig.VerifyAggregatedSig(suite, pubKeys, aggSig, msg)

	require.Equal(t, crypto.ErrInvalidSuite, err)
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigNilPubKeysShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), nil, aggSig, msg)

	require.NotNil(t, err)
	require.Equal(t, crypto.ErrNilPublicKeys, err)
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigNilAggSigBytesShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), pubKeys, nil, msg)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "not enough data")
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigInvalidAggSigBytesShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)

	//make aggregated sig invalid
	aggSig[0] = aggSig[0] ^ 0xFF
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), pubKeys, aggSig, msg)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "malformed point")
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigNilMsgShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), pubKeys, aggSig, nil)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "invalid signature")
}

func TestKyberMultiSignerBLS_VerifyAggregatedSigOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), pubKeys, aggSig, msg)

	require.Nil(t, err)
}

func TestKyberMultiSignerBLS_AggregatePreparedPublicKeysNilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	pubKeysPoints := pointsFromPubKeys(pubKeys)
	aggPks, err := llSig.AggregatePreparedPublicKeys(nil, pubKeysPoints...)

	require.Equal(t, crypto.ErrNilSuite, err)
	require.Nil(t, aggPks)
}

func TestKyberMultiSignerBLS_AggregatePreparedPublicKeysInvalidSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	pubKeysPoints := pointsFromPubKeys(pubKeys)
	suite := createMockSuite("invalid suite")
	aggPks, err := llSig.AggregatePreparedPublicKeys(suite, pubKeysPoints...)

	require.Equal(t, crypto.ErrInvalidSuite, err)
	require.Nil(t, aggPks)
}

func TestKyberMultiSignerBLS_AggregatePreparedPublicKeysNilPubKeysShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	aggPks, err := llSig.AggregatePreparedPublicKeys(pubKeys[0].Suite())

	require.Equal(t, crypto.ErrNilPublicKeys, err)
	require.Nil(t, aggPks)
}

func TestKyberMultiSignerBLS_AggregatePreparedPublicKeysWithNilPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	emptyPubKeys := make([]crypto.Point, len(pubKeys))
	aggPks, err := llSig.AggregatePreparedPublicKeys(pubKeys[0].Suite(), emptyPubKeys...)

	require.Equal(t, crypto.ErrNilPublicKeyPoint, err)
	require.Nil(t, aggPks)
}

func TestKyberMultiSignerBLS_AggregatePreparedPublicKeysWithInvalidPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	pubKeysPoints := pointsFromPubKeys(pubKeys)
	// make a public key point invalid for the given suite
	pubKeysPoints[0] = createPoint() // this uses the mock suite so should be invalid for real suite
	aggPks, err := llSig.AggregatePreparedPublicKeys(pubKeys[0].Suite(), pubKeysPoints...)

	require.Equal(t, crypto.ErrInvalidPublicKey, err)
	require.Nil(t, aggPks)
}

func TestKyberMultiSignerBLS_AggregatePreparedPublicKeysOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.KyberMultiSignerBLS{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	pubKeysPoints := pointsFromPubKeys(pubKeys)
	aggPks, err := llSig.AggregatePreparedPublicKeys(pubKeys[0].Suite(), pubKeysPoints...)

	require.NotNil(t, aggPks)
	require.Nil(t, err)
}

func TestKyberMultiSignerBLS_ScalarMulSigNilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, llSig := genSigParamsBLS()
	sig, _ := llSig.SignShare(privKey, msg)
	rs := pubKey.Suite().RandomStream()
	scalar, _ := pubKey.Suite().CreateScalar().Pick(rs)
	scalarBytes, _ := scalar.MarshalBinary()
	kmsSigner := llSig.(*multisig.KyberMultiSignerBLS)
	res, err := kmsSigner.ScalarMulSig(nil, scalarBytes, sig)

	require.Equal(t, crypto.ErrNilSuite, err)
	require.Nil(t, res)
}

func TestKyberMultiSignerBLS_ScalarMulSigNilScalarShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, llSig := genSigParamsBLS()
	sig, _ := llSig.SignShare(privKey, msg)
	kmsSigner := llSig.(*multisig.KyberMultiSignerBLS)
	res, err := kmsSigner.ScalarMulSig(pubKey.Suite(), nil, sig)

	require.Equal(t, crypto.ErrNilParam, err)
	require.Nil(t, res)
}

func TestKyberMultiSignerBLS_ScalarMulSigNilSigShouldErr(t *testing.T) {
	t.Parallel()

	_, pubKey, _, llSig := genSigParamsBLS()
	rs := pubKey.Suite().RandomStream()
	scalar, _ := pubKey.Suite().CreateScalar().Pick(rs)
	kmsSigner := llSig.(*multisig.KyberMultiSignerBLS)
	scalarBytes, _ := scalar.MarshalBinary()
	res, err := kmsSigner.ScalarMulSig(pubKey.Suite(), scalarBytes, nil)

	require.Equal(t, crypto.ErrNilSignature, err)
	require.Nil(t, res)
}

func TestKyberMultiSignerBLS_ScalarMulSigOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, llSig := genSigParamsBLS()
	sig, _ := llSig.SignShare(privKey, msg)
	rs := pubKey.Suite().RandomStream()
	scalar, _ := pubKey.Suite().CreateScalar().Pick(rs)
	scalarBytes, _ := scalar.MarshalBinary()
	kmsSigner := llSig.(*multisig.KyberMultiSignerBLS)
	res, err := kmsSigner.ScalarMulSig(pubKey.Suite(), scalarBytes, sig)

	require.Nil(t, err)
	require.NotNil(t, res)
}
