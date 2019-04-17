package singlesig_test

import (
"testing"

"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
mock2 "github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/mock"
"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/singlesig"
"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/mock"
"github.com/stretchr/testify/assert"
)

func TestBLSSigner_SignNilPrivateKeyShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	signature, err := signer.Sign(nil, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrNilPrivateKey, err)
}

func TestBLSSigner_SignPrivateKeyNilSuiteShouldErr(t *testing.T) {
	suite := kyber.NewSuitePairingBn256()
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
	signer := &singlesig.BlsSingleSigner{}
	signature, err := signer.Sign(privKeyNilSuite, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrNilSuite, err)
}

func TestBLSSigner_SignPrivateKeyNilScalarShouldErr(t *testing.T) {
	suite := kyber.NewSuitePairingBn256()
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
	signer := &singlesig.BlsSingleSigner{}
	signature, err := signer.Sign(privKeyNilSuite, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrNilPrivateKeyScalar, err)
}

func TestBLSSigner_SignInvalidScalarShouldErr(t *testing.T) {
	suite := kyber.NewSuitePairingBn256()
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
	signer := &singlesig.BlsSingleSigner{}
	signature, err := signer.Sign(privKeyNilSuite, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrInvalidPrivateKey, err)
}

func TestBLSSigner_SignInvalidSuiteShouldErr(t *testing.T) {
	suite := kyber.NewSuitePairingBn256()
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
	signer := &singlesig.BlsSingleSigner{}

	signature, err := signer.Sign(privKeyNilSuite, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrInvalidSuite, err)
}

func signBLS(msg []byte, signer crypto.SingleSigner, t *testing.T) (
	pubKey crypto.PublicKey,
	privKey crypto.PrivateKey,
	signature []byte,
	err error) {

	suite := kyber.NewSuitePairingBn256()
	kg := signing.NewKeyGenerator(suite)
	privKey, pubKey = kg.GeneratePair()

	signature, err = signer.Sign(privKey, msg)

	assert.NotNil(t, signature)
	assert.Nil(t, err)

	return pubKey, privKey, signature, err
}

func TestBLSSigner_SignOK(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)

	err = signer.Verify(pubKey, msg, signature)

	assert.Nil(t, err)
}

func TestBLSSigner_VerifyNilSuiteShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)

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

func TestBLSSigner_VerifyNilPublicKeyShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	_, _, signature, err := signBLS(msg, signer, t)

	err = signer.Verify(nil, msg, signature)

	assert.Equal(t, crypto.ErrNilPublicKey, err)
}

func TestBLSSigner_VerifyNilMessageShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)

	err = signer.Verify(pubKey, nil, signature)

	assert.Equal(t, crypto.ErrNilMessage, err)
}

func TestBLSSigner_VerifyNilSignatureShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, _, err := signBLS(msg, signer, t)

	err = signer.Verify(pubKey, msg, nil)

	assert.Equal(t, crypto.ErrNilSignature, err)
}

func TestBLSSigner_VerifyInvalidSuiteShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)

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

func TestBLSSigner_VerifyPublicKeyInvalidPointShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)

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

func TestBLSSigner_VerifyInvalidPublicKeyShouldErr(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)

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

func TestBLSSigner_VerifyOK(t *testing.T) {
	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)

	err = signer.Verify(pubKey, msg, signature)

	assert.Nil(t, err)
}
