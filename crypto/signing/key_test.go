package signing_test

import (
	"crypto/cipher"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/stretchr/testify/assert"
)

var invalidStr = []byte("invalid key")

const initScalar = 10
const initPointX = 2
const initPointY = 3

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
	return &mock.ScalarMock{
		X:                   initScalar,
		UnmarshalBinaryStub: unmarshalPrivate,
		MarshalBinaryStub:   marshalPrivate,
	}
}

func createPoint() crypto.Point {
	return &mock.PointMock{
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

func createMockSuite() crypto.Suite {
	suite := &mock.SuiteMock{
		CreateKeyPairStub: createKeyPair,
		CreateScalarStub:  createScalar,
		CreatePointStub:   createPoint,
	}

	return suite
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

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	privKey, pubKey := kg.GeneratePair()

	sc, _ := privKey.Scalar().(*mock.ScalarMock)
	po, _ := pubKey.Point().(*mock.PointMock)

	assert.Equal(t, initScalar, sc.X)
	assert.Equal(t, initScalar*initPointX, po.X)
	assert.Equal(t, initScalar*initPointY, po.Y)
}

func TestKeyGenerator_GeneratePairNonGeneratorOK(t *testing.T) {
	t.Parallel()

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	privKey, pubKey := kg.GeneratePair()

	sc, _ := privKey.Scalar().(*mock.ScalarMock)
	po, _ := pubKey.Point().(*mock.PointMock)

	assert.Equal(t, initScalar, sc.X)
	assert.Equal(t, initScalar*initPointX, po.X)
	assert.Equal(t, initScalar*initPointY, po.Y)
}

func TestKeyGenerator_PrivateKeyFromByteArrayNilArrayShouldErr(t *testing.T) {
	t.Parallel()

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	privKey, err := kg.PrivateKeyFromByteArray(nil)

	assert.Nil(t, privKey)
	assert.Equal(t, crypto.ErrInvalidParam, err)
}

func TestKeyGenerator_PrivateKeyFromByteArrayInvalidArrayShouldErr(t *testing.T) {
	t.Parallel()

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	privKeyBytes := invalidStr
	privKey, err := kg.PrivateKeyFromByteArray(privKeyBytes)

	assert.Nil(t, privKey)
	assert.Equal(t, crypto.ErrInvalidPrivateKey, err)
}

func TestKeyGenerator_PrivateKeyFromByteArrayOK(t *testing.T) {
	t.Parallel()

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	privKeyBytes := []byte("valid key")
	privKey, err := kg.PrivateKeyFromByteArray(privKeyBytes)

	assert.Nil(t, err)

	sc, _ := privKey.Scalar().(*mock.ScalarMock)

	assert.Equal(t, sc.X, initScalar)
}

func TestKeyGenerator_PublicKeyFromByteArrayNilArrayShouldErr(t *testing.T) {
	t.Parallel()

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	pubKey, err := kg.PublicKeyFromByteArray(nil)

	assert.Nil(t, pubKey)
	assert.Equal(t, crypto.ErrInvalidParam, err)
}

func TestKeyGenerator_PublicKeyFromByteArrayInvalidArrayShouldErr(t *testing.T) {
	t.Parallel()

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	pubKeyBytes := invalidStr
	pubKey, err := kg.PublicKeyFromByteArray(pubKeyBytes)

	assert.Nil(t, pubKey)
	assert.Equal(t, crypto.ErrInvalidPublicKey, err)
}

func TestKeyGenerator_PublicKeyFromByteArrayOK(t *testing.T) {
	t.Parallel()

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	pubKeyBytes := []byte("valid key")
	pubKey, err := kg.PublicKeyFromByteArray(pubKeyBytes)

	assert.Nil(t, err)

	po, _ := pubKey.Point().(*mock.PointMock)

	assert.Equal(t, initPointX, po.X)
	assert.Equal(t, initPointY, po.Y)
}

func TestKeyGenerator_SuiteOK(t *testing.T) {
	t.Parallel()

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	s1 := kg.Suite()

	assert.Equal(t, suite, s1)
}

func TestPrivateKey_ToByteArrayOK(t *testing.T) {
	t.Parallel()

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()
	privKeyBytes, err := privKey.ToByteArray()

	assert.Nil(t, err)
	assert.Equal(t, []byte(strconv.Itoa(initScalar)), privKeyBytes)
}

func TestPrivateKey_GeneratePublicOK(t *testing.T) {
	t.Parallel()

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()    // privkey = scalar*basePoint.X; basePoint.X = 1, basePoint.Y = 1
	pubkey := privKey.GeneratePublic() // pubKey = privKey * BasePoint.Y
	pubKeyBytes, _ := pubkey.Point().MarshalBinary()

	assert.Equal(t, []byte(fmt.Sprintf("%d%d", initScalar, initScalar)), pubKeyBytes)
}

func TestPrivateKey_SuiteOK(t *testing.T) {
	t.Parallel()

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()

	s2 := privKey.Suite()

	assert.Equal(t, suite, s2)
}

func TestPrivateKey_Scalar(t *testing.T) {
	t.Parallel()

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()
	sc := privKey.Scalar()
	x := sc.(*mock.ScalarMock).X

	assert.Equal(t, initScalar, x)
}

func TestPublicKey_ToByteArrayOK(t *testing.T) {
	t.Parallel()

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	_, pubKey := kg.GeneratePair()
	pubKeyBytes, err := pubKey.ToByteArray()

	assert.Nil(t, err)
	assert.Equal(t, []byte(fmt.Sprintf("%d%d", initScalar*initPointX, initScalar*initPointY)), pubKeyBytes)
}

func TestPublicKey_SuiteOK(t *testing.T) {
	t.Parallel()

	suite := createMockSuite()
	kg := signing.NewKeyGenerator(suite)
	_, pubKey := kg.GeneratePair()
	s2 := pubKey.Suite()

	assert.Equal(t, suite, s2)
}
