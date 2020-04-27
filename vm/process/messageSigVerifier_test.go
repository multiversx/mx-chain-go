package process

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	"github.com/stretchr/testify/assert"
)

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
