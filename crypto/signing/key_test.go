package signing_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/mock"
	"github.com/stretchr/testify/assert"
	"reflect"
)

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
