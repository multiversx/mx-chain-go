package multisig_test

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/mock"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/bls-go-binary/bls"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/multisig"
	"github.com/stretchr/testify/require"
)

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
	bls.BlsSignatureToG1(sigBLS, sigPointG1.G1)

	return sigPointG1, nil
}

func TestMultiSignerBLS_SignShareNilPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	_, _, _, lls := genSigParamsBLS()

	sig, err := lls.SignShare(nil, msg)

	require.Equal(t, crypto.ErrNilPrivateKey, err)
	require.Nil(t, sig)
}

func TestMultiSignerBLS_SignShareInvalidPrivateKeyShouldErr(t *testing.T) {
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

func TestMultiSignerBLS_SignShareNilMsgShouldErr(t *testing.T) {
	t.Parallel()

	privKey, _, _, lls := genSigParamsBLS()
	sig, err := lls.SignShare(privKey, nil)

	require.Equal(t, crypto.ErrNilMessage, err)
	require.Nil(t, sig)
}

func TestMultiSignerBLS_SignShareOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, _, _, lls := genSigParamsBLS()
	sig, err := lls.SignShare(privKey, msg)

	require.NotNil(t, sig)
	require.Nil(t, err)
}

func TestMultiSignerBLS_VerifySigShareNilPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, _, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	err := lls.VerifySigShare(nil, msg, sig)

	require.Equal(t, crypto.ErrNilPublicKey, err)
}

func TestMultiSignerBLS_VerifySigShareInvalidPubKeyShouldErr(t *testing.T) {
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
			return mcl.NewSuiteBLS12()
		},
	}

	err := lls.VerifySigShare(pubKey, msg, sig)

	require.Equal(t, crypto.ErrInvalidPublicKey, err)
}

func TestMultiSignerBLS_VerifySigShareNilMsgShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	err := lls.VerifySigShare(pubKey, nil, sig)

	require.Equal(t, crypto.ErrNilMessage, err)
}

func TestMultiSignerBLS_VerifySigShareNilSigShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	_, pubKey, _, lls := genSigParamsBLS()
	err := lls.VerifySigShare(pubKey, msg, nil)

	require.Equal(t, crypto.ErrNilSignature, err)
}

func TestMultiSignerBLS_VerifySigShareInvalidSigShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	// change the message so signature becomes invalid
	msg2 := []byte("message2")
	err := lls.VerifySigShare(pubKey, msg2, sig)

	require.Equal(t, crypto.ErrSigNotValid, err)
	require.NotNil(t, err)
}

func TestMultiSignerBLS_VerifySigShareOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)
	err := lls.VerifySigShare(pubKey, msg, sig)

	require.Nil(t, err)
}

func TestMultiSignerBLS_VerifySigBytesNilSigShouldErr(t *testing.T) {
	t.Parallel()

	_, pubKey, _, lls := genSigParamsBLS()
	err := lls.VerifySigBytes(pubKey.Suite(), nil)

	require.Equal(t, crypto.ErrNilSignature, err)
}

func TestMultiSignerBLS_VerifySigBytesInvalidSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, _, _, lls := genSigParamsBLS()
	sig, _ := lls.SignShare(privKey, msg)

	suite := createMockSuite("invalid suite")
	err := lls.VerifySigBytes(suite, sig)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "suite is invalid")
}

func TestMultiSignerBLS_AggregateSignaturesNilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(nil, sigShares, pubKeys)

	require.Equal(t, crypto.ErrNilSuite, err)
}

func TestMultiSignerBLS_AggregateSignaturesInvalidSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	suite := createMockSuite("invalid suite")
	_, err := llSig.AggregateSignatures(suite, sigShares, pubKeys)

	require.Equal(t, crypto.ErrInvalidSuite, err)
}

func TestMultiSignerBLS_AggregateSignaturesNilSigsShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), nil, pubKeys)

	require.Equal(t, crypto.ErrNilSignaturesList, err)
}

func TestMultiSignerBLS_AggregateSignaturesEmptySigsShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), [][]byte{[]byte("")}, pubKeys)

	require.Equal(t, crypto.ErrNilSignature, err)
}

func TestMultiSignerBLS_AggregateSignaturesInvalidSigsShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	// make first sig share invalid
	sigShares[0] = []byte("invalid signature")
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "err blsSignatureDeserialize")
}

func TestMultiSignerBLS_AggregateSignaturesOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	_, err := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)

	require.Nil(t, err)
}

func TestMultiSignerBLS_VerifyAggregatedSigNilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	err := llSig.VerifyAggregatedSig(nil, pubKeys, aggSig, msg)

	require.Equal(t, crypto.ErrNilSuite, err)
}

func TestMultiSignerBLS_VerifyAggregatedSigInvalidSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
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

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), nil, aggSig, msg)

	require.Equal(t, crypto.ErrNilPublicKeys, err)
}

func TestMultiSignerBLS_VerifyAggregatedSigNilAggSigBytesShouldErr(t *testing.T) {
	t.Parallel()
	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, _ := createSigSharesBLS(20, msg)
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), pubKeys, nil, msg)

	require.Equal(t, crypto.ErrNilSignature, err)
}

func TestMultiSignerBLS_VerifyAggregatedSigInvalidAggSigBytesShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
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

	msg := []byte("message")
	hasher := &mock.HasherSpongeMock{}
	llSig := &multisig.BlsMultiSigner{Hasher: hasher}
	pubKeys, sigShares := createSigSharesBLS(20, msg)
	aggSig, _ := llSig.AggregateSignatures(pubKeys[0].Suite(), sigShares, pubKeys)
	err := llSig.VerifyAggregatedSig(pubKeys[0].Suite(), pubKeys, aggSig, nil)

	require.Equal(t, crypto.ErrNilMessage, err)
}

func TestMultiSignerBLS_VerifyAggregatedSigOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
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

	msg := []byte("message")
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
	rs := pubKey.Suite().RandomStream()
	scalar, _ := pubKey.Suite().CreateScalar().Pick(rs)
	bmsSigner := llSig.(*multisig.BlsMultiSigner)
	scalarBytes, _ := scalar.MarshalBinary()
	res, err := bmsSigner.ScalarMulSig(pubKey.Suite(), scalarBytes, nil)

	require.Equal(t, crypto.ErrNilSignature, err)
	require.Nil(t, res)
}

func TestMultiSignerBLS_ScalarMulSigOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	privKey, pubKey, _, llSig := genSigParamsBLS()
	sig, _ := llSig.SignShare(privKey, msg)
	rs := pubKey.Suite().RandomStream()
	scalar, _ := pubKey.Suite().CreateScalar().Pick(rs)
	mclScalar, _ := scalar.(*mcl.MclScalar)
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
