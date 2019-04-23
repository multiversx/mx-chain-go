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

	createKeyPair := func(stream cipher.Stream) (crypto.Scalar, crypto.Point) {
		scalar := createScalar()
		point, _ := createPoint().Mul(scalar)
		return scalar, point
	}

	suite := &mock.GeneratorSuite{
		CreateKeyStub: createKeyPair,
		SuiteMock: mock.SuiteMock{
			CreateScalarStub: createScalar,
		},
	}

	kg := signing.NewKeyGenerator(suite)
	privKey, pubKey := kg.GeneratePair()

	sc, _ := privKey.Scalar().(*mock.ScalarMock)
	po, _ := pubKey.Point().(*mock.PointMock)

	assert.Equal(t, 10, sc.X)
	assert.Equal(t, 20, po.X)
	assert.Equal(t, 30, po.Y)
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

	assert.Equal(t, 20, sc.X)
	assert.Equal(t, 40, po.X)
	assert.Equal(t, 60, po.Y)
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

	assert.Nil(t, err)

	sc, _ := privKey.Scalar().(*mock.ScalarMock)

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

	assert.Nil(t, err)

	sc, _ := pubKey.Point().(*mock.PointMock)

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
