package signing_test

import (
	"crypto/cipher"
	"reflect"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/mock"
	"github.com/stretchr/testify/assert"
)

var invalidStr = []byte("invalid key")

func signOK(suite crypto.Suite, private crypto.Scalar, msg []byte) ([]byte, error) {
	return []byte("signed"), nil
}

func signNOK(suite crypto.Suite, private crypto.Scalar, msg []byte) ([]byte, error) {
	return nil, crypto.ErrInvalidParam
}

func Verify(suite crypto.Suite, public crypto.Point, msg []byte, sig []byte) error {
	if !reflect.DeepEqual(sig, []byte("signed")) {
		return crypto.ErrSigNotValid
	}

	return nil
}

func unmarshalPrivate(val []byte) (int, error) {
	if reflect.DeepEqual(invalidStr, val) {
		return 0, crypto.ErrInvalidPrivateKey
	}

	return 4, nil
}

func marshalPrivate(x int) ([]byte, error) {
	res := []byte(strconv.Itoa(x))
	return res, nil
}

func unmarshalPublic(val []byte) (x, y int, err error) {
	if reflect.DeepEqual(invalidStr, val) {
		return 0, 0, crypto.ErrInvalidPublicKey
	}
	return 4, 5, nil
}

func marshalPublic(x, y int) ([]byte, error) {
	resStr := strconv.Itoa(x)
	resStr += strconv.Itoa(y)
	res := []byte(resStr)

	return res, nil
}

func createScalar() crypto.Scalar {
	return &mock.ScalarMock{
		X:                   10,
		UnmarshalBinaryStub: unmarshalPrivate,
		MarshalBinaryStub:   marshalPrivate,
	}
}

func createPoint() crypto.Point {
	return &mock.PointMock{
		X:                   2,
		Y:                   3,
		UnmarshalBinaryStub: unmarshalPublic,
		MarshalBinaryStub:   marshalPublic,
	}
}

func TestNewKeyGenerator(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{}
	kg := signing.NewKeyGenerator(suite)
	assert.NotNil(t, kg)

	s2 := kg.Suite()
	assert.Equal(t, suite, s2)
}

func TestKeyGenerator_GeneratePairNilSuiteShouldPanic(t *testing.T) {
	t.Parallel()

	kg := signing.NewKeyGenerator(nil)

	assert.Panics(t, func() { kg.GeneratePair() }, "the code did not panic")
}

func TestKeyGenerator_GeneratePairGeneratorOK(t *testing.T) {
	t.Parallel()

	createKey := func(stream cipher.Stream) crypto.Scalar {
		return createScalar()
	}

	suite := &mock.GeneratorSuite{
		CreateKeyStub: createKey,
		SuiteMock: mock.SuiteMock{
			CreatePointStub: createPoint,
		},
	}

	kg := signing.NewKeyGenerator(suite)
	privKey, pubKey := kg.GeneratePair()

	sc, _ := privKey.Scalar().(*mock.ScalarMock)
	po, _ := pubKey.Point().(*mock.PointMock)

	assert.Equal(t, sc.X, 10)
	assert.Equal(t, po.X, 20)
	assert.Equal(t, po.Y, 30)
}

func TestKeyGenerator_GeneratePairNonGeneratorOK(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	privKey, pubKey := kg.GeneratePair()

	sc, _ := privKey.Scalar().(*mock.ScalarMock)
	po, _ := pubKey.Point().(*mock.PointMock)

	assert.Equal(t, sc.X, 20)
	assert.Equal(t, po.X, 40)
	assert.Equal(t, po.Y, 60)
}

func TestKeyGenerator_PrivateKeyFromByteArrayNilArrayShouldErr(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	privKey, err := kg.PrivateKeyFromByteArray(nil)

	assert.Nil(t, privKey)
	assert.Equal(t, crypto.ErrInvalidParam, err)
}

func TestKeyGenerator_PrivateKeyFromByteArrayInvalidArrayShouldErr(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	privKeyBytes := invalidStr
	privKey, err := kg.PrivateKeyFromByteArray(privKeyBytes)

	assert.Nil(t, privKey)
	assert.Equal(t, crypto.ErrInvalidPrivateKey, err)
}

func TestKeyGenerator_PrivateKeyFromByteArrayOK(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	privKeyBytes := []byte("valid key")
	privKey, err := kg.PrivateKeyFromByteArray(privKeyBytes)
	sc, _ := privKey.Scalar().(*mock.ScalarMock)

	assert.Nil(t, err)
	assert.Equal(t, sc.X, 4)
}

func TestKeyGenerator_PublicKeyFromByteArrayNilArrayShouldErr(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	pubKey, err := kg.PublicKeyFromByteArray(nil)

	assert.Nil(t, pubKey)
	assert.Equal(t, crypto.ErrInvalidParam, err)
}

func TestKeyGenerator_PublicKeyFromByteArrayInvalidArrayShouldErr(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	pubKeyBytes := invalidStr
	pubKey, err := kg.PublicKeyFromByteArray(pubKeyBytes)

	assert.Nil(t, pubKey)
	assert.Equal(t, crypto.ErrInvalidPublicKey, err)
}

func TestKeyGenerator_PublicKeyFromByteArrayOK(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	pubKeyBytes := []byte("valid key")
	pubKey, err := kg.PublicKeyFromByteArray(pubKeyBytes)
	sc, _ := pubKey.Point().(*mock.PointMock)

	assert.Nil(t, err)
	assert.Equal(t, 4, sc.X)
	assert.Equal(t, 5, sc.Y)
}

func TestKeyGenerator_SuiteOK(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	s1 := kg.Suite()

	assert.Equal(t, suite, s1)
}

func TestPrivateKey_SignNilSignerShouldErr(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()
	msg := []byte("message")
	signature, err := privKey.Sign(msg, nil)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrNilSingleSigner, err)
}

func TestPrivateKey_SignInvalidSignerParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	signer := &mock.Signer{
		SignStub: signNOK,
	}

	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()
	msg := []byte("message")
	signature, err := privKey.Sign(msg, signer)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrInvalidParam, err)
}

func TestPrivateKey_SignOK(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	signer := &mock.Signer{
		SignStub: signOK,
	}

	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()
	msg := []byte("message")
	signature, err := privKey.Sign(msg, signer)

	assert.Equal(t, []byte("signed"), signature)
	assert.Nil(t, err)
}

func TestPrivateKey_ToByteArrayOK(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()
	privKeyBytes, err := privKey.ToByteArray()

	assert.Nil(t, err)
	assert.Equal(t, []byte("20"), privKeyBytes)
}

func TestPrivateKey_GeneratePublicOK(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()
	pubkey := privKey.GeneratePublic()
	pubKeyBytes, _ := pubkey.Point().MarshalBinary()

	assert.Equal(t, []byte("2020"), pubKeyBytes)
}

func TestPrivateKey_SuiteOK(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()

	s2 := privKey.Suite()

	assert.Equal(t, suite, s2)
}

func TestPrivateKey_Scalar(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()
	sc := privKey.Scalar()
	x := sc.(*mock.ScalarMock).X

	assert.Equal(t, 20, x)
}

func TestPublicKey_VerifyNilSignerShouldErr(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	_, pubKey := kg.GeneratePair()
	msg := []byte("message")
	signature := []byte("signature")

	err := pubKey.Verify(msg, signature, nil)

	assert.Equal(t, crypto.ErrNilSingleSigner, err)
}

func TestPublicKey_VerifyOK(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	signer := &mock.Signer{
		SignStub:   signNOK,
		VerifyStub: Verify,
	}

	kg := signing.NewKeyGenerator(suite)
	privKey, pubKey := kg.GeneratePair()
	msg := []byte("message")
	signature, _ := privKey.Sign(msg, signer)
	err := pubKey.Verify(msg, signature, signer)

	assert.Equal(t, crypto.ErrSigNotValid, err)
}

func TestPublicKey_ToByteArrayOK(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	_, pubKey := kg.GeneratePair()
	pubKeyBytes, err := pubKey.ToByteArray()

	assert.Nil(t, err)
	assert.Equal(t, []byte("4060"), pubKeyBytes)
}

func TestPublicKey_SuiteOK(t *testing.T) {
	t.Parallel()

	suite := &mock.SuiteMock{
		CreateScalarStub: createScalar,
		CreatePointStub:  createPoint,
	}

	kg := signing.NewKeyGenerator(suite)
	_, pubKey := kg.GeneratePair()
	s2 := pubKey.Suite()

	assert.Equal(t, suite, s2)
}
