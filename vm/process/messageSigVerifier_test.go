package process

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewMessageSigVerifier_NilKeyGenShouldErr(t *testing.T) {
	mv, err := NewMessageSigVerifier(nil, &mock.SignerMock{})

	assert.Nil(t, mv)
	assert.Equal(t, vm.ErrNilKeyGenerator, err)
}

func TestNewMessageSigVerifier_NilSignerShouldErr(t *testing.T) {
	mv, err := NewMessageSigVerifier(&mock.KeyGenMock{}, nil)

	assert.Nil(t, mv)
	assert.Equal(t, vm.ErrNilSingleSigner, err)
}

func TestNewMessageSigVerifier(t *testing.T) {
	mv, err := NewMessageSigVerifier(&mock.KeyGenMock{}, &mock.SignerMock{})

	assert.Nil(t, err)
	assert.NotNil(t, mv)
}

func TestMessageSigVerifier_IsInterfaceNil(t *testing.T) {
	mv, err := NewMessageSigVerifier(&mock.KeyGenMock{}, &mock.SignerMock{})

	assert.Nil(t, err)
	assert.False(t, mv.IsInterfaceNil())

	mv = nil
	assert.True(t, mv.IsInterfaceNil())
}

func TestMessageSigVerifier_VerifyNilPubKey(t *testing.T) {
	mv, _ := NewMessageSigVerifier(&mock.KeyGenMock{}, &mock.SignerMock{})

	err := mv.Verify([]byte("a"), []byte("a"), nil)
	assert.Equal(t, vm.ErrNilPublicKey, err)
}

func TestMessageSigVerifier_VerifyInvalidPubKey(t *testing.T) {
	currErr := errors.New("current error")
	mv, _ := NewMessageSigVerifier(&mock.KeyGenMock{
		PublicKeyFromByteArrayCalled: func(b []byte) (key crypto.PublicKey, err error) {
			return nil, currErr
		},
	}, &mock.SignerMock{})

	err := mv.Verify([]byte("a"), []byte("a"), []byte("a"))
	assert.Equal(t, currErr, err)
}

func TestMessageSigVerifier_VerifyInvalidSingature(t *testing.T) {
	currErr := errors.New("current error")
	mv, _ := NewMessageSigVerifier(
		&mock.KeyGenMock{},
		&mock.SignerMock{VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return currErr
		}},
	)

	err := mv.Verify([]byte("a"), []byte("a"), []byte("a"))
	assert.Equal(t, currErr, err)
}

func TestMessageSigVerifier_Verify(t *testing.T) {
	mv, _ := NewMessageSigVerifier(
		&mock.KeyGenMock{},
		&mock.SignerMock{},
	)

	err := mv.Verify([]byte("a"), []byte("a"), []byte("a"))
	assert.Nil(t, err)
}
