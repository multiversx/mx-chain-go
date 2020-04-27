package singlesig_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/mock"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/singlesig"
	"github.com/stretchr/testify/require"
)

func TestBLSSigner_SignNilPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := singlesig.NewBlsSigner()
	signature, err := signer.Sign(nil, msg)

	require.Nil(t, signature)
	require.Equal(t, crypto.ErrNilPrivateKey, err)
}

func TestBLSSigner_SignPrivateKeyNilScalarShouldErr(t *testing.T) {
	t.Parallel()

	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)
	privKey, _ := kg.GeneratePair()

	privKeyNilSuite := &mock.PrivateKeyStub{
		SuiteStub: func() crypto.Suite {
			return suite
		},
		ToByteArrayStub: privKey.ToByteArray,
		ScalarStub: func() crypto.Scalar {
			return nil
		},
		GeneratePublicStub: privKey.GeneratePublic,
	}

	msg := []byte("message to be signed")
	signer := singlesig.NewBlsSigner()
	signature, err := signer.Sign(privKeyNilSuite, msg)

	require.Nil(t, signature)
	require.Equal(t, crypto.ErrNilPrivateKeyScalar, err)
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
	signer := singlesig.NewBlsSigner()
	signature, err := signer.Sign(privKeyNilSuite, msg)

	require.Nil(t, signature)
	require.Equal(t, crypto.ErrInvalidPrivateKey, err)
}

func signBLS(msg []byte, signer crypto.SingleSigner, t *testing.T) (
	pubKey crypto.PublicKey,
	privKey crypto.PrivateKey,
	signature []byte,
	err error,
) {

	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)
	privKey, pubKey = kg.GeneratePair()

	signature, err = signer.Sign(privKey, msg)

	require.NotNil(t, signature)
	require.Nil(t, err)

	return pubKey, privKey, signature, err
}

func TestBLSSigner_SignOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := singlesig.NewBlsSigner()
	pubKey, _, signature, err := signBLS(msg, signer, t)
	require.Nil(t, err)

	err = signer.Verify(pubKey, msg, signature)

	require.Nil(t, err)
}

func TestBLSSigner_VerifyNilPublicKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := singlesig.NewBlsSigner()
	_, _, signature, err := signBLS(msg, signer, t)
	require.Nil(t, err)

	err = signer.Verify(nil, msg, signature)

	require.Equal(t, crypto.ErrNilPublicKey, err)
}

func TestBLSSigner_VerifyNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := singlesig.NewBlsSigner()
	pubKey, _, signature, err := signBLS(msg, signer, t)
	require.Nil(t, err)
	err = signer.Verify(pubKey, nil, signature)

	require.Equal(t, crypto.ErrNilMessage, err)
}

func TestBLSSigner_VerifyNilSignatureShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := singlesig.NewBlsSigner()
	pubKey, _, _, err := signBLS(msg, signer, t)
	require.Nil(t, err)
	err = signer.Verify(pubKey, msg, nil)

	require.Equal(t, crypto.ErrNilSignature, err)
}

func TestBLSSigner_VerifyPublicKeyInvalidPointShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := singlesig.NewBlsSigner()
	pubKey, _, signature, err := signBLS(msg, signer, t)
	require.Nil(t, err)

	pubKeyInvalidSuite := &mock.PublicKeyStub{
		SuiteStub:       pubKey.Suite,
		ToByteArrayStub: pubKey.ToByteArray,
		PointStub: func() crypto.Point {
			return nil
		},
	}

	err = signer.Verify(pubKeyInvalidSuite, msg, signature)

	require.Equal(t, crypto.ErrNilPublicKeyPoint, err)
}

func TestBLSSigner_VerifyInvalidPublicKeyShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := singlesig.NewBlsSigner()
	pubKey, _, signature, err := signBLS(msg, signer, t)
	require.Nil(t, err)
	pubKeyInvalidSuite := &mock.PublicKeyStub{
		SuiteStub:       pubKey.Suite,
		ToByteArrayStub: pubKey.ToByteArray,
		PointStub: func() crypto.Point {
			return &mock.PointMock{}
		},
	}

	err = signer.Verify(pubKeyInvalidSuite, msg, signature)

	require.Equal(t, crypto.ErrInvalidPublicKey, err)
}

func TestBLSSigner_VerifyOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := singlesig.NewBlsSigner()
	pubKey, _, signature, err := signBLS(msg, signer, t)
	require.Nil(t, err)

	err = signer.Verify(pubKey, msg, signature)

	require.Nil(t, err)
}

func TestBLSSigner_SignVerifyWithReconstructedPubKeyOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := singlesig.NewBlsSigner()
	pubKey, _, signature, err := signBLS(msg, signer, t)
	require.Nil(t, err)

	pubKeyBytes, err := pubKey.Point().MarshalBinary()
	require.Nil(t, err)

	// reconstruct publicKey
	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)
	pubKey2, err := kg.PublicKeyFromByteArray(pubKeyBytes)
	require.Nil(t, err)

	// reconstructed public key needs to match original
	// and be able to verify
	err = signer.Verify(pubKey2, msg, signature)

	require.Nil(t, err)
}

func TestBLSSigner_VerifyInvalidSignatureShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message to be signed")
	signer := singlesig.NewBlsSigner()
	pubKey, _, signature, err := signBLS(msg, signer, t)
	require.Nil(t, err)

	// invalidate the signature by changing the message
	msg[0] ^= msg[0]

	err = signer.Verify(pubKey, msg, signature)
	require.Equal(t, crypto.ErrSigNotValid, err)
}

func TestBLSSigner_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var llSig *singlesig.BlsSingleSigner
	require.True(t, check.IfNil(llSig))
	llSig = &singlesig.BlsSingleSigner{}

	require.False(t, check.IfNil(llSig))
}
