package singlesig_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2"
	mock2 "github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2/singlesig"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/mock"
	"github.com/stretchr/testify/assert"
)

func TestSchnorrSigner_SignNilPrivateKeyShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	signature, err := signer.Sign(nil, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrNilPrivateKey, err)
}

func TestSchnorrSigner_SignPrivateKeyNilSuiteShouldErr(t *testing.T) {
	suite := kv2.NewBlakeSHA256Ed25519()
	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()

	privKeyNilSuite := &mock2.PrivateKeyStub{
		SuiteStub: func() crypto.Suite {
			return nil
		},
		ToByteArrayStub:    privKey.ToByteArray,
		ScalarStub:         privKey.Scalar,
		GeneratePublicStub: privKey.GeneratePublic,
	}

	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	signature, err := signer.Sign(privKeyNilSuite, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrNilSuite, err)
}

func TestSchnorrSigner_SignPrivateKeyNilScalarShouldErr(t *testing.T) {
	suite := kv2.NewBlakeSHA256Ed25519()
	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()

	privKeyNilSuite := &mock2.PrivateKeyStub{
		SuiteStub:       privKey.Suite,
		ToByteArrayStub: privKey.ToByteArray,
		ScalarStub: func() crypto.Scalar {
			return nil
		},
		GeneratePublicStub: privKey.GeneratePublic,
	}

	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	signature, err := signer.Sign(privKeyNilSuite, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrNilPrivateKeyScalar, err)
}

func TestSchnorrSigner_SignInvalidScalarShouldErr(t *testing.T) {
	suite := kv2.NewBlakeSHA256Ed25519()
	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()

	privKeyNilSuite := &mock2.PrivateKeyStub{
		SuiteStub:       privKey.Suite,
		ToByteArrayStub: privKey.ToByteArray,
		ScalarStub: func() crypto.Scalar {
			return &mock.ScalarMock{}
		},
		GeneratePublicStub: privKey.GeneratePublic,
	}

	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	signature, err := signer.Sign(privKeyNilSuite, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrInvalidPrivateKey, err)
}

func TestSchnorrSigner_SignInvalidSuiteShouldErr(t *testing.T) {
	suite := kv2.NewBlakeSHA256Ed25519()
	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()

	invalidSuite := &mock.SuiteMock{
		GetUnderlyingSuiteStub: func() interface{} {
			return 0
		},
	}

	privKeyNilSuite := &mock2.PrivateKeyStub{
		SuiteStub: func() crypto.Suite {
			return invalidSuite
		},
		ToByteArrayStub:    privKey.ToByteArray,
		ScalarStub:         privKey.Scalar,
		GeneratePublicStub: privKey.GeneratePublic,
	}

	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}

	signature, err := signer.Sign(privKeyNilSuite, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrInvalidSuite, err)
}

func sign(msg []byte, signer crypto.SingleSigner, t *testing.T) (
	pubKey crypto.PublicKey,
	privKey crypto.PrivateKey,
	signature []byte,
	err error) {

	suite := kv2.NewBlakeSHA256Ed25519()
	kg := signing.NewKeyGenerator(suite)
	privKey, pubKey = kg.GeneratePair()

	signature, err = signer.Sign(privKey, msg)

	assert.NotNil(t, signature)
	assert.Nil(t, err)

	return pubKey, privKey, signature, err
}

func TestSchnorrSigner_SignOK(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	pubKey, _, signature, err := sign(msg, signer, t)

	err = signer.Verify(pubKey, msg, signature)

	assert.Nil(t, err)
}

func TestSchnorrSigner_VerifyNilSuiteShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	pubKey, _, signature, err := sign(msg, signer, t)

	pubKeyNilSuite := &mock2.PublicKeyStub{
		SuiteStub: func() crypto.Suite {
			return nil
		},
		ToByteArrayStub: pubKey.ToByteArray,
		PointStub:       pubKey.Point,
	}

	err = signer.Verify(pubKeyNilSuite, msg, signature)

	assert.Equal(t, crypto.ErrNilSuite, err)
}

func TestSchnorrSigner_VerifyNilPublicKeyShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	_, _, signature, err := sign(msg, signer, t)

	err = signer.Verify(nil, msg, signature)

	assert.Equal(t, crypto.ErrNilPublicKey, err)
}

func TestSchnorrSigner_VerifyNilMessageShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	pubKey, _, signature, err := sign(msg, signer, t)

	err = signer.Verify(pubKey, nil, signature)

	assert.Equal(t, crypto.ErrNilMessage, err)
}

func TestSchnorrSigner_VerifyNilSignatureShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	pubKey, _, _, err := sign(msg, signer, t)

	err = signer.Verify(pubKey, msg, nil)

	assert.Equal(t, crypto.ErrNilSignature, err)
}

func TestSchnorrSigner_VerifyInvalidSuiteShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	pubKey, _, signature, err := sign(msg, signer, t)

	invalidSuite := &mock.SuiteMock{
		GetUnderlyingSuiteStub: func() interface{} {
			return 0
		},
	}

	pubKeyInvalidSuite := &mock2.PublicKeyStub{
		SuiteStub: func() crypto.Suite {
			return invalidSuite
		},
		ToByteArrayStub: pubKey.ToByteArray,
		PointStub:       pubKey.Point,
	}

	err = signer.Verify(pubKeyInvalidSuite, msg, signature)

	assert.Equal(t, crypto.ErrInvalidSuite, err)
}

func TestSchnorrSigner_VerifyPublicKeyInvalidPointShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	pubKey, _, signature, err := sign(msg, signer, t)

	pubKeyInvalidSuite := &mock2.PublicKeyStub{
		SuiteStub:       pubKey.Suite,
		ToByteArrayStub: pubKey.ToByteArray,
		PointStub: func() crypto.Point {
			return nil
		},
	}

	err = signer.Verify(pubKeyInvalidSuite, msg, signature)

	assert.Equal(t, crypto.ErrNilPublicKeyPoint, err)
}

func TestSchnorrSigner_VerifyInvalidPublicKeyShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	pubKey, _, signature, err := sign(msg, signer, t)

	pubKeyInvalidSuite := &mock2.PublicKeyStub{
		SuiteStub:       pubKey.Suite,
		ToByteArrayStub: pubKey.ToByteArray,
		PointStub: func() crypto.Point {
			return &mock.PointMock{}
		},
	}

	err = signer.Verify(pubKeyInvalidSuite, msg, signature)

	assert.Equal(t, crypto.ErrInvalidPublicKey, err)
}

func TestSchnorrSigner_VerifyOK(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.SchnorrSigner{}
	pubKey, _, signature, err := sign(msg, signer, t)

	err = signer.Verify(pubKey, msg, signature)

	assert.Nil(t, err)
}
