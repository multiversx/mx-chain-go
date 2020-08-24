package singlesig_test

import (
	goEd25519 "crypto/ed25519"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/mock"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519/singlesig"
	"github.com/stretchr/testify/assert"
)

func TestEd25519SignerSign_NilPrivateKeyShoudErr(t *testing.T) {
	message := []byte("message to sign")
	signer := &singlesig.Ed25519Signer{}

	_, err := signer.Sign(nil, message)
	assert.Equal(t, crypto.ErrNilPrivateKey, err)
}

func TestEd25519SignerSign_InvalidPrivateKeyTypeShoudErr(t *testing.T) {
	message := []byte("message to sign")
	signer := &singlesig.Ed25519Signer{}

	scalar := &mock.ScalarMock{
		GetUnderlyingObjStub: func() interface{} {
			return "this is not a byte array"
		},
	}

	privateKey := &mock.PrivateKeyStub{
		ScalarStub: func() crypto.Scalar {
			return scalar
		},
	}

	_, err := signer.Sign(privateKey, message)
	assert.Equal(t, crypto.ErrInvalidPrivateKey, err)
}

func TestEd25519SignerSign_InvalidPrivateKeyLengthShoudErr(t *testing.T) {
	message := []byte("message to sign")
	signer := &singlesig.Ed25519Signer{}

	scalar := &mock.ScalarMock{
		GetUnderlyingObjStub: func() interface{} {
			return goEd25519.PrivateKey("incorrect length")
		},
	}

	privateKey := &mock.PrivateKeyStub{
		ScalarStub: func() crypto.Scalar {
			return scalar
		},
	}

	_, err := signer.Sign(privateKey, message)
	assert.Equal(t, crypto.ErrInvalidPrivateKey, err)
}

func TestEd25519SignerSign_CorrectParamsShouldNotError(t *testing.T) {
	suite := ed25519.NewEd25519()
	keyGenerator := signing.NewKeyGenerator(suite)
	privateKey, _ := keyGenerator.GeneratePair()
	message := []byte("message to sign")
	signer := &singlesig.Ed25519Signer{}
	_, err := signer.Sign(privateKey, message)
	assert.Nil(t, err)
}

func TestEd25519SignerVerify_NilPublicKeyShouldErr(t *testing.T) {
	signer := &singlesig.Ed25519Signer{}

	err := signer.Verify(nil, []byte(""), []byte(""))
	assert.Equal(t, crypto.ErrNilPublicKey, err)
}

func TestEd25519SignerVerify_InvalidPublicKeyTypeShouldErr(t *testing.T) {
	signer := &singlesig.Ed25519Signer{}

	publicKey := &mock.PublicKeyStub{
		PointStub: func() crypto.Point {
			return &mock.PointMock{
				GetUnderlyingObjStub: func() interface{} {
					return "this is not a byte array"
				},
			}
		},
	}

	err := signer.Verify(publicKey, []byte(""), []byte(""))
	assert.Equal(t, crypto.ErrInvalidPublicKey, err)
}

func TestEd25519SignerVerify_InvalidPublicKeyLengthShouldErr(t *testing.T) {
	signer := &singlesig.Ed25519Signer{}

	publicKey := &mock.PublicKeyStub{
		PointStub: func() crypto.Point {
			return &mock.PointMock{
				GetUnderlyingObjStub: func() interface{} {
					return goEd25519.PublicKey("incorrect length")
				},
			}
		},
	}

	err := signer.Verify(publicKey, []byte(""), []byte(""))
	assert.Equal(t, crypto.ErrInvalidPublicKey, err)
}

func TestEd25519SignerVerify_InvalidSigError(t *testing.T) {
	suite := ed25519.NewEd25519()
	keyGenerator := signing.NewKeyGenerator(suite)
	privateKey, publicKey := keyGenerator.GeneratePair()
	message := []byte("message to sign")
	alteredMessage := []byte("message to sign altered")
	signer := &singlesig.Ed25519Signer{}
	sig, _ := signer.Sign(privateKey, message)
	err := signer.Verify(publicKey, alteredMessage, sig)
	assert.Equal(t, crypto.ErrEd25519InvalidSignature, err)
}

func TestEd25519SignerVerify_InvalidSigErrorForDifferentPubKey(t *testing.T) {
	suite := ed25519.NewEd25519()
	keyGenerator := signing.NewKeyGenerator(suite)
	privateKey, _ := keyGenerator.GeneratePair()
	_, publicKey2 := keyGenerator.GeneratePair()
	message := []byte("message to sign")
	alteredMessage := []byte("message to sign altered")
	signer := &singlesig.Ed25519Signer{}
	sig, _ := signer.Sign(privateKey, message)
	err := signer.Verify(publicKey2, alteredMessage, sig)
	assert.Equal(t, crypto.ErrEd25519InvalidSignature, err)
}

func TestEd25519SignerVerify_CorrectSignature(t *testing.T) {
	suite := ed25519.NewEd25519()
	keyGenerator := signing.NewKeyGenerator(suite)
	privateKey, publicKey := keyGenerator.GeneratePair()
	message := []byte("message to sign")
	signer := &singlesig.Ed25519Signer{}
	sig, _ := signer.Sign(privateKey, message)
	err := signer.Verify(publicKey, message, sig)
	assert.Nil(t, err)
}
