package singlesig_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/mock"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/singlesig"
	"github.com/stretchr/testify/assert"
)

func TestBLSSigner_SignNilPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	signature, err := signer.Sign(nil, msg)

	assert.Nil(t, signature)
	assert.Equal(t, crypto.ErrNilPrivateKey, err)
}

func TestBLSSigner_SignPrivateKeyNilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()

	privKeyNilSuite := &mock.PrivateKeyStub{
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
	t.Parallel()

	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()

	privKeyNilSuite := &mock.PrivateKeyStub{
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
	t.Parallel()

	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()

	privKeyNilSuite := &mock.PrivateKeyStub{
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
	t.Parallel()

	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()

	invalidSuite := &mock.SuiteMock{
		GetUnderlyingSuiteStub: func() interface{} {
			return 0
		},
	}

	privKeyNilSuite := &mock.PrivateKeyStub{
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

	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)
	privKey, pubKey = kg.GeneratePair()

	signature, err = signer.Sign(privKey, msg)

	assert.NotNil(t, signature)
	assert.Nil(t, err)

	return pubKey, privKey, signature, err
}

func TestBLSSigner_SignOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)

	err = signer.Verify(pubKey, msg, signature)

	assert.Nil(t, err)
}

func TestBLSSigner_VerifyNilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)

	pubKeyNilSuite := &mock.PublicKeyStub{
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
	t.Parallel()

	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	_, _, signature, err := signBLS(msg, signer, t)

	err = signer.Verify(nil, msg, signature)

	assert.Equal(t, crypto.ErrNilPublicKey, err)
}

func TestBLSSigner_VerifyNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)

	err = signer.Verify(pubKey, nil, signature)

	assert.Equal(t, crypto.ErrNilMessage, err)
}

func TestBLSSigner_VerifyNilSignatureShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, _, err := signBLS(msg, signer, t)

	err = signer.Verify(pubKey, msg, nil)

	assert.Equal(t, crypto.ErrNilSignature, err)
}

func TestBLSSigner_VerifyInvalidSuiteShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)

	invalidSuite := &mock.SuiteMock{
		GetUnderlyingSuiteStub: func() interface{} {
			return 0
		},
	}

	pubKeyInvalidSuite := &mock.PublicKeyStub{
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
	t.Parallel()

	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)

	pubKeyInvalidSuite := &mock.PublicKeyStub{
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
	t.Parallel()

	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)

	pubKeyInvalidSuite := &mock.PublicKeyStub{
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
	t.Parallel()

	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)

	err = signer.Verify(pubKey, msg, signature)

	assert.Nil(t, err)
}

func TestBLSSigner_SignVerifyWithReconstructedPubKeyOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := &singlesig.BlsSingleSigner{}
	pubKey, _, signature, err := signBLS(msg, signer, t)
	assert.Nil(t, err)

	pubKeyBytes, err := pubKey.Point().MarshalBinary()
	assert.Nil(t, err)

	// reconstruct publicKey
	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)
	pubKey2, err := kg.PublicKeyFromByteArray(pubKeyBytes)
	assert.Nil(t, err)

	// reconstructed public key needs to match original
	// and be able to verify
	err = signer.Verify(pubKey2, msg, signature)

	assert.Nil(t, err)
}
