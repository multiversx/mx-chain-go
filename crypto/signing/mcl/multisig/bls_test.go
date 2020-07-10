package multisig_test

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/mock"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/multisig"
	"github.com/herumi/bls-go-binary/bls"
	"github.com/stretchr/testify/require"
)

const testMessage = "message"

func createMockSuite(innerSuite interface{}) crypto.Suite {
	validSuite := mcl.NewSuiteBLS12()

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
	suite := mcl.NewSuiteBLS12()
	kg = signing.NewKeyGenerator(suite)
	hasher := &mock.HasherSpongeMock{}
	llSigner = &multisig.BlsMultiSigner{Hasher: hasher}

	privKey, pubKey = kg.GeneratePair()

	return privKey, pubKey, kg, llSigner
}

func createSigSharesBLS(
	nbSigs uint16,
	message []byte,
) (pubKeys []crypto.PublicKey, sigShares [][]byte) {
	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)
	hasher := &mock.HasherSpongeMock{}

	pubKeys = make([]crypto.PublicKey, nbSigs)
	sigShares = make([][]byte, nbSigs)
	llSigner := &multisig.BlsMultiSigner{Hasher: hasher}

	for i := uint16(0); i < nbSigs; i++ {
		sk, pk := kg.GeneratePair()
		pubKeys[i] = pk
		sigShares[i], _ = llSigner.SignShare(sk, message)

	}

	return pubKeys, sigShares
}

func sigBytesToPointG1(sig []byte) (*mcl.PointG1, error) {
	sigBLS := &bls.Sign{}
	err := sigBLS.Deserialize(sig)
	if err != nil {
		return nil, err
	}

	sigPointG1 := mcl.NewPointG1()
	sigPointG1.G1 = bls.CastFromSign(sigBLS)

	return sigPointG1, nil
}

func TestMultiSignerBLS_SignShareNilPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	_, _, _, lls := genSigParamsBLS()

	sig, err := lls.SignShare(nil, msg)

	require.Equal(t, crypto.ErrNilPrivateKey, err)
	require.Nil(t, sig)
}

func TestMultiSignerBLS_SignShareInvalidPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
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

func TestMultiSignerBLS_SignShareNilMsgShouldErr(t *testing.T) {
	t.Parallel()
	privKey, _, _, lls := genSigParamsBLS()
	sig, err := lls.SignShare(privKey, nil)

	require.Equal(t, crypto.ErrNilMessage, err)
	require.Nil(t, sig)
}

func TestMultiSignerBLS_SignShareOK(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	privKey, _, _, lls := genSigParamsBLS()
	sig, err := lls.SignShare(privKey, msg)

	require.NotNil(t, sig)
	require.Nil(t, err)
}

func TestMultiSignerBLS_VerifySigShareNilPubKeyShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	privKey, _, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	err := lls.VerifySigShare(nil, msg, sig)

	require.Equal(t, crypto.ErrNilPublicKey, err)
}

func TestMultiSignerBLS_VerifySigShareInvalidPubKeyShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
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
			return mcl.NewSuiteBLS12()
		},
	}

	err := lls.VerifySigShare(pubKey, msg, sig)

	require.Equal(t, crypto.ErrInvalidPublicKey, err)
}

func TestMultiSignerBLS_VerifySigShareNilMsgShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	privKey, pubKey, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	err := lls.VerifySigShare(pubKey, nil, sig)

	require.Equal(t, crypto.ErrNilMessage, err)
}

func TestMultiSignerBLS_VerifySigShareNilSigShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	_, pubKey, _, lls := genSigParamsBLS()
	err := lls.VerifySigShare(pubKey, msg, nil)

	require.Equal(t, crypto.ErrNilSignature, err)
}

func TestMultiSignerBLS_VerifySigShareInvalidSigShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	privKey, pubKey, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	// change the message so signature becomes invalid
	msg2 := []byte("message2")
	err := lls.VerifySigShare(pubKey, msg2, sig)

	require.Equal(t, crypto.ErrSigNotValid, err)
}

func TestMultiSignerBLS_VerifySigShareOK(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	privKey, pubKey, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	err := lls.VerifySigShare(pubKey, msg, sig)

	require.Nil(t, err)
}

func TestMultiSignerBLS_AggregateSignaturesNilSuiteShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(nil, sigShares, pubKeys)

	require.Equal(t, crypto.ErrNilSuite, err)
}

func TestMultiSignerBLS_AggregateSignaturesInvalidSuiteShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	suite := createMockSuite("invalid suite")
	_, err := llSig.AggregateSignatures(suite, sigShares, pubKeys)

	require.Equal(t, crypto.ErrInvalidSuite, err)
}

func TestMultiSignerBLS_AggregateSignaturesNilSigsShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), nil, pubKeys)

	require.Equal(t, crypto.ErrNilSignaturesList, err)
}

func TestMultiSignerBLS_AggregateSignaturesEmptySigsShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), [][]byte{[]byte("")}, pubKeys)

	require.Equal(t, crypto.ErrNilSignature, err)
}

func TestMultiSignerBLS_AggregateSignaturesInvalidSigsShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	// make first sig share invalid
	sigShares[0] = []byte("invalid signature")
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "err blsSignatureDeserialize")
}

func TestMultiSignerBLS_AggregateSignaturesEmptyPubKeysShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, nil)

	require.Equal(t, crypto.ErrNilPublicKeys, err)
}

func TestMultiSignerBLS_AggregateSignaturesOK(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)

	require.Nil(t, err)
}

func TestMultiSignerBLS_VerifyAggregatedSigNilSuiteShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	err := llSig.VerifyAggregatedSig(nil, pubKeys, aggSig, msg)

	require.Equal(t, crypto.ErrNilSuite, err)
}

func TestMultiSignerBLS_VerifyAggregatedSigInvalidSuiteShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	suite := createMockSuite("invalid suite")
	err := llSig.VerifyAggregatedSig(suite, pubKeys, aggSig, msg)

	require.Equal(t, crypto.ErrInvalidSuite, err)
}

func TestMultiSignerBLS_VerifyAggregatedSigNilPubKeysShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), nil, aggSig, msg)

	require.Equal(t, crypto.ErrNilPublicKeys, err)
}

func TestMultiSignerBLS_VerifyAggregatedSigNilAggSigBytesShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), pubKeys, nil, msg)

	require.Equal(t, crypto.ErrNilSignature, err)
}

func TestMultiSignerBLS_VerifyAggregatedSigInvalidAggSigBytesShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)

	//make aggregated sig invalid
	aggSig := []byte("invalid aggregated signature")
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), pubKeys, aggSig, msg)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "err blsSignatureDeserialize")
}

func TestMultiSignerBLS_VerifyAggregatedSigNilMsgShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), pubKeys, aggSig, nil)

	require.Equal(t, crypto.ErrNilMessage, err)
}

func TestMultiSignerBLS_VerifyAggregatedSigInvalidPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	pubKeys[1] = nil
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), pubKeys, aggSig, msg)

	require.Equal(t, crypto.ErrNilPublicKey, err)
}

func TestMultiSignerBLS_VerifyAggregatedSigInvalidForMessageShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	msg2 := []byte("message2")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, err := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	require.Nil(t, err)

	err = llSig.VerifyAggregatedSig(pubKeys[0].Suite(), pubKeys, aggSig, msg2)

	require.Equal(t, crypto.ErrAggSigNotValid, err)
}

func TestMultiSignerBLS_VerifyAggregatedSigOK(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, err := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	require.Nil(t, err)

	err = llSig.VerifyAggregatedSig(pubKeys[0].Suite(), pubKeys, aggSig, msg)

	require.Nil(t, err)
}

func TestMultiSignerBLS_ScalarMulSigNilScalarShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	privKey, pubKey, _, llSig := genSigParamsBLS()
	sig, _ := llSig.SignShare(privKey, msg)
	bmsSigner := llSig.(*multisig.BlsMultiSigner)

	sigPointG1, err := sigBytesToPointG1(sig)
	require.Nil(t, err)

	res, err := bmsSigner.ScalarMulSig(pubKey.Suite(), nil, sigPointG1)

	require.Equal(t, crypto.ErrNilParam, err)
	require.Nil(t, res)
}

func TestMultiSignerBLS_ScalarMulSigNilSigShouldErr(t *testing.T) {
	t.Parallel()
	_, pubKey, _, llSig := genSigParamsBLS()
	scalar, _ := pubKey.Suite().CreateScalar().Pick()
	bmsSigner := llSig.(*multisig.BlsMultiSigner)
	scalarBytes, _ := scalar.MarshalBinary()
	res, err := bmsSigner.ScalarMulSig(pubKey.Suite(), scalarBytes, nil)

	require.Equal(t, crypto.ErrNilSignature, err)
	require.Nil(t, res)
}

func TestBlsMultiSigner_ScalarMulSigNilSuiteShouldErr(t *testing.T) {
	t.Parallel()
	privKey, pubKey, _, llSig := genSigParamsBLS()
	msg := []byte(testMessage)
	sig, _ := llSig.SignShare(privKey, msg)

	scalar, _ := pubKey.Suite().CreateScalar().Pick()
	bmsSigner := llSig.(*multisig.BlsMultiSigner)
	mclScalar, _ := scalar.(*mcl.Scalar)
	scalarBytesHexStr := mclScalar.Scalar.GetString(16)

	// odd length hex string fails hex decoding, so make it even
	if len(scalarBytesHexStr)%2 != 0 {
		scalarBytesHexStr = "0" + scalarBytesHexStr
	}

	scalarBytes, err := hex.DecodeString(scalarBytesHexStr)
	require.Nil(t, err)
	sigPointG1, err := sigBytesToPointG1(sig)
	require.Nil(t, err)
	res, err := bmsSigner.ScalarMulSig(nil, scalarBytes, sigPointG1)

	require.Equal(t, crypto.ErrNilSuite, err)
	require.Nil(t, res)
}

func TestMultiSignerBLS_ScalarMulSigOK(t *testing.T) {
	t.Parallel()
	msg := []byte(testMessage)
	privKey, pubKey, _, llSig := genSigParamsBLS()
	sig, _ := llSig.SignShare(privKey, msg)
	scalar, _ := pubKey.Suite().CreateScalar().Pick()
	mclScalar, _ := scalar.(*mcl.Scalar)
	scalarBytesHexStr := mclScalar.Scalar.GetString(16)

	// odd length hex string fails hex decoding, so make it even
	if len(scalarBytesHexStr)%2 != 0 {
		scalarBytesHexStr = "0" + scalarBytesHexStr
	}

	scalarBytes, err := hex.DecodeString(scalarBytesHexStr)
	require.Nil(t, err)
	bmsSigner := llSig.(*multisig.BlsMultiSigner)

	sigPointG1, err := sigBytesToPointG1(sig)
	require.Nil(t, err)

	res, err := bmsSigner.ScalarMulSig(pubKey.Suite(), scalarBytes, sigPointG1)

	require.Nil(t, err)
	require.NotNil(t, res)
}

func TestBlsMultiSigner_VerifySigBytesNilSigShouldErr(t *testing.T) {
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}

	err := llSig.VerifySigBytes(nil, nil)
	require.Equal(t, crypto.ErrNilSignature, err)
}

func TestBlsMultiSigner_VerifySigBytesZeroSigShouldErr(t *testing.T) {
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}

	p1 := mcl.NewPointG1()
	point := p1.Null()
	pointBytes, err := point.MarshalBinary()
	require.Nil(t, err)

	err = llSig.VerifySigBytes(nil, pointBytes)
	require.Equal(t, crypto.ErrBLSInvalidSignature, err)
}

func TestBlsMultiSigner_VerifySigBytesInvalidSigShouldErr(t *testing.T) {
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}

	p1 := mcl.NewPointG1()
	point := p1.Null()
	pointBytes, err := point.MarshalBinary()
	require.Nil(t, err)

	pointBytes[0] = 1
	err = llSig.VerifySigBytes(nil, pointBytes)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "blsSignatureDeserialize")
}

func Test_PreparePublicKeysNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte(testMessage)
	pubKeys, _ := createSigSharesBLS(20, msg)
	prepPubKeys, err := multisig.PreparePublicKeys(pubKeys, nil, pubKeys[0].Suite())
	require.Equal(t, crypto.ErrNilHasher, err)
	require.Nil(t, prepPubKeys)
}

func Test_PreparePublicKeysNilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherSpongeMock{}
	msg := []byte(testMessage)
	pubKeys, _ := createSigSharesBLS(20, msg)
	prepPubKeys, err := multisig.PreparePublicKeys(pubKeys, hasher, nil)
	require.Equal(t, crypto.ErrNilSuite, err)
	require.Nil(t, prepPubKeys)
}

func Test_PreparePublicKeysOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherSpongeMock{}
	msg := []byte(testMessage)
	pubKeys, _ := createSigSharesBLS(20, msg)
	prepPubKeys, err := multisig.PreparePublicKeys(pubKeys, hasher, pubKeys[0].Suite())
	require.Nil(t, err)
	require.NotNil(t, prepPubKeys)
}

func Test_PrepareSignaturesNilSignaturesShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}

	msg := []byte(testMessage)
	pubKeys, _ := createSigSharesBLS(20, msg)
	prepSignatures, err := llSig.PrepareSignatures(pubKeys[0].Suite(), nil, pubKeys)
	require.Equal(t, crypto.ErrNilSignaturesList, err)
	require.Nil(t, prepSignatures)
}

func TestBlsMultiSigner_PrepareSignaturesNilSignatureInListShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}

	msg := []byte(testMessage)
	pubKeys, signatures := createSigSharesBLS(20, msg)
	signatures[1] = nil
	prepSignatures, err := llSig.PrepareSignatures(pubKeys[0].Suite(), signatures, pubKeys)
	require.Equal(t, crypto.ErrNilSignature, err)
	require.Nil(t, prepSignatures)
}

func TestBlsMultiSigner_PrepareSignaturesInvalidSignatureInSignaturesShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	msg := []byte(testMessage)
	pubKeys, signatures := createSigSharesBLS(20, msg)

	p1 := mcl.NewPointG1()
	pointSig := p1.Null()
	sigPointBytes, err := pointSig.MarshalBinary()
	require.Nil(t, err)

	signatures[1] = sigPointBytes
	prepSignatures, err := llSig.PrepareSignatures(pubKeys[0].Suite(), signatures, pubKeys)
	require.Equal(t, crypto.ErrBLSInvalidSignature, err)
	require.Nil(t, prepSignatures)
}

func TestBlsMultiSigner_PrepareSignaturesNilPubKeysShouldErr(t *testing.T) {
	t.Parallel()
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}

	msg := []byte(testMessage)
	pubKeys, signatures := createSigSharesBLS(20, msg)
	prepSignatures, err := llSig.PrepareSignatures(pubKeys[0].Suite(), signatures, nil)
	require.Equal(t, crypto.ErrNilPublicKeys, err)
	require.Nil(t, prepSignatures)
}

func TestBlsMultiSigner_PrepareSignaturesNilPubKeyInKeysShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}

	msg := []byte(testMessage)
	pubKeys, signatures := createSigSharesBLS(20, msg)
	pubKeys[1] = nil
	prepSignatures, err := llSig.PrepareSignatures(pubKeys[0].Suite(), signatures, pubKeys)
	require.Equal(t, crypto.ErrNilPublicKey, err)
	require.Nil(t, prepSignatures)
}

func TestBlsMultiSigner_PrepareSignaturesInvalidPubKeyInKeysShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	msg := []byte(testMessage)
	pubKeys, signatures := createSigSharesBLS(20, msg)
	pubKeys[1] = &mock.PublicKeyStub{
		ToByteArrayStub: func() (bytes []byte, err error) {
			return []byte("invalid key"), nil
		},
		PointStub: func() crypto.Point {
			return &mock.PointMock{
				MarshalBinaryStub: func(x, y int) (bytes []byte, err error) {
					return []byte("invalid key"), nil
				},
			}
		},
		SuiteStub: func() crypto.Suite {
			return mcl.NewSuiteBLS12()
		},
	}
	prepSignatures, err := llSig.PrepareSignatures(pubKeys[0].Suite(), signatures, pubKeys)
	require.Equal(t, crypto.ErrInvalidPublicKey, err)
	require.Nil(t, prepSignatures)
}

func Test_ScalarMulPkNilPkShouldErr(t *testing.T) {
	t.Parallel()

	suite := mcl.NewSuiteBLS12()
	scalar := suite.CreateScalar()
	scalarBytes, err := scalar.MarshalBinary()
	require.Nil(t, err)

	point, err := multisig.ScalarMulPk(suite, scalarBytes, nil)
	require.Equal(t, crypto.ErrNilParam, err)
	require.Nil(t, point)
}

func Test_ScalarMulPkNilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	suite := mcl.NewSuiteBLS12()
	scalar := suite.CreateScalar()
	kg := signing.NewKeyGenerator(suite)
	_, pk := kg.GeneratePair()
	scalarBytes, err := scalar.MarshalBinary()
	require.Nil(t, err)

	point, err := multisig.ScalarMulPk(nil, scalarBytes, pk.Point())
	require.Equal(t, crypto.ErrNilSuite, err)
	require.Nil(t, point)
}

func Test_ScalarMulPkOK(t *testing.T) {
	t.Parallel()

	suite := mcl.NewSuiteBLS12()
	scalar := suite.CreateScalar()
	kg := signing.NewKeyGenerator(suite)

	_, pk := kg.GeneratePair()
	require.NotNil(t, pk)

	mclScalar, ok := scalar.GetUnderlyingObj().(*bls.Fr)
	require.True(t, ok)
	scalarHexStr := mclScalar.GetString(16)
	scalarBytes := make([]byte, 32)

	// odd length hex string fails hex decoding, so make it even
	if len(scalarHexStr)%2 != 0 {
		scalarHexStr = "0" + scalarHexStr
	}
	scalarHexStr = "1" + scalarHexStr[1:] //make the first byte non 0 (resulting value will be 32 bytes long in all cases)

	nb, err := hex.Decode(scalarBytes, []byte(scalarHexStr))
	require.Nil(t, err)
	require.Equal(t, 32, nb)

	point, err := multisig.ScalarMulPk(suite, scalarBytes, pk.Point())
	require.Nil(t, err)
	require.NotNil(t, point)
}

func Test_HashPublicKeyPointsNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	msg := testMessage
	pubKeys, _ := createSigSharesBLS(20, []byte(msg))
	concatPubKeys, err := multisig.ConcatPubKeys(pubKeys)
	require.Nil(t, err)

	hash, err := multisig.HashPublicKeyPoints(nil, pubKeys[0].Point(), concatPubKeys)
	require.Equal(t, crypto.ErrNilHasher, err)
	require.Nil(t, hash)
}

func Test_HashPublicKeyPointsNilPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherSpongeMock{}
	msg := testMessage
	pubKeys, _ := createSigSharesBLS(20, []byte(msg))
	concatPubKeys, err := multisig.ConcatPubKeys(pubKeys)
	require.Nil(t, err)

	hash, err := multisig.HashPublicKeyPoints(hasher, nil, concatPubKeys)
	require.Equal(t, crypto.ErrNilPublicKeyPoint, err)
	require.Nil(t, hash)
}

func Test_HashPublicKeyPointsWrongSizeHasherShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	msg := testMessage
	pubKeys, _ := createSigSharesBLS(20, []byte(msg))
	concatPubKeys, err := multisig.ConcatPubKeys(pubKeys)
	require.Nil(t, err)

	hash, err := multisig.HashPublicKeyPoints(hasher, pubKeys[0].Point(), concatPubKeys)
	require.Equal(t, crypto.ErrWrongSizeHasher, err)
	require.Nil(t, hash)
}

func Test_HashPublicKeyPointsNilConcatPubKeysShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	msg := testMessage
	pubKeys, _ := createSigSharesBLS(20, []byte(msg))
	hash, err := multisig.HashPublicKeyPoints(hasher, pubKeys[0].Point(), nil)
	require.Equal(t, crypto.ErrNilParam, err)
	require.Nil(t, hash)
}

func Test_HashPublicKeyPointsOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherSpongeMock{}
	msg := testMessage
	pubKeys, _ := createSigSharesBLS(20, []byte(msg))
	concatPubKeys, err := multisig.ConcatPubKeys(pubKeys)
	require.Nil(t, err)

	hash, err := multisig.HashPublicKeyPoints(hasher, pubKeys[0].Point(), concatPubKeys)
	require.Nil(t, err)
	require.NotNil(t, hash)
}

func TestBlsMultiSigner_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var llSig *multisig.BlsMultiSigner

	require.True(t, check.IfNil(llSig))

	hasher := &mock.HasherSpongeMock{}
	llSig = &multisig.BlsMultiSigner{Hasher: hasher}

	require.False(t, check.IfNil(llSig))
}
