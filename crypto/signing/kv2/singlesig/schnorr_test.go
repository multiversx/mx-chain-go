package singlesig_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2/singlesig"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/mock"
	"github.com/stretchr/testify/assert"
)

func TestSchnorrSigner_SignNilSuiteShouldErr(t *testing.T) {
	suite := kv2.NewBlakeSHA256Ed25519()

	randStream := suite.RandomStream()
	scalar, _ := suite.CreateScalar().Pick(randStream)
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	signature, err := signer.Sign(nil, scalar, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrNilSuite, err)
}

func TestSchnorrSigner_SignNilScalarShouldErr(t *testing.T) {
	suite := kv2.NewBlakeSHA256Ed25519()

	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	signature, err := signer.Sign(suite, nil, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrNilPrivateKey, err)
}

func TestSchnorrSigner_SignInvalidScalarShouldErr(t *testing.T) {
	suite := kv2.NewBlakeSHA256Ed25519()

	msg := []byte("message to be signed")
	scalar := &mock.ScalarMock{}
	signer := &singlesig.SchnorrSigner{}
	signature, err := signer.Sign(suite, scalar, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrInvalidPrivateKey, err)
}

func TestSchnorrSigner_SignInvalidSuiteShouldErr(t *testing.T) {
	suite := kv2.NewBlakeSHA256Ed25519()

	msg := []byte("message to be signed")
	scalar := suite.CreateScalar()
	signer := &singlesig.SchnorrSigner{}
	invalidSuite := &mock.SuiteMock{
		GetUnderlyingSuiteStub: func() interface{} {
			return 0
		},
	}
	signature, err := signer.Sign(invalidSuite, scalar, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrInvalidSuite, err)
}

func sign(msg []byte, signer crypto.SingleSigner, t *testing.T) (
	pubKey crypto.Point,
	privKey crypto.Scalar,
	suite crypto.Suite,
	signature []byte,
	err error) {

	suite = kv2.NewBlakeSHA256Ed25519()
	randStream := suite.RandomStream()
	scalar, _ := suite.CreateScalar().Pick(randStream)
	signature, err = signer.Sign(suite, scalar, msg)

	assert.NotNil(t, signature)
	assert.Nil(t, err)

	pubKey = suite.CreatePoint().Base()
	pubKey, _ = pubKey.Mul(scalar)

	return pubKey, scalar, suite, signature, err
}

func TestSchnorrSigner_SignOK(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	pubKey, _, suite, signature, err := sign(msg, signer, t)

	err = signer.Verify(suite, pubKey, msg, signature)

	assert.Nil(t, err)
}

func TestSchnorrSigner_VerifyNilSuiteShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	pubKey, _, _, signature, err := sign(msg, signer, t)

	err = signer.Verify(nil, pubKey, msg, signature)

	assert.Equal(t, crypto.ErrNilSuite, err)
}

func TestSchnorrSigner_VerifyNilPointShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	_, _, suite, signature, err := sign(msg, signer, t)

	err = signer.Verify(suite, nil, msg, signature)

	assert.Equal(t, crypto.ErrNilPublicKey, err)
}

func TestSchnorrSigner_VerifyNilMessageShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	pubKey, _, suite, signature, err := sign(msg, signer, t)

	err = signer.Verify(suite, pubKey, nil, signature)

	assert.Equal(t, crypto.ErrNilMessage, err)
}

func TestSchnorrSigner_VerifyNilSignatureShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	pubKey, _, suite, _, err := sign(msg, signer, t)

	err = signer.Verify(suite, pubKey, msg, nil)

	assert.Equal(t, crypto.ErrNilSignature, err)
}

func TestSchnorrSigner_VerifyInvalidSuiteShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	pubKey, _, _, signature, err := sign(msg, signer, t)

	invalidSuite := &mock.SuiteMock{
		GetUnderlyingSuiteStub: func() interface{} {
			return 0
		},
	}

	err = signer.Verify(invalidSuite, pubKey, msg, signature)

	assert.Equal(t, crypto.ErrInvalidSuite, err)
}

func TestSchnorrSigner_VerifyInvalidPointShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	_, _, suite, signature, err := sign(msg, signer, t)

	invalidPubKey := &mock.PointMock{}

	err = signer.Verify(suite, invalidPubKey, msg, signature)

	assert.Equal(t, crypto.ErrInvalidPublicKey, err)
}

func TestSchnorrSigner_VerifyOK(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	pubKey, _, suite, signature, err := sign(msg, signer, t)

	err = signer.Verify(suite, pubKey, msg, signature)

	assert.Nil(t, err)
}
