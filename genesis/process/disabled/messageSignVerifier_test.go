package disabled

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessageSignVerifier(t *testing.T) {
	t.Parallel()

	t.Run("nil key gen should error", func(t *testing.T) {
		t.Parallel()

		sv, err := NewMessageSignVerifier(nil)
		require.Equal(t, genesis.ErrNilKeyGenerator, err)
		require.Nil(t, sv)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		sv, err := NewMessageSignVerifier(&mock.KeyGeneratorStub{})
		require.NoError(t, err)
		require.NotNil(t, sv)
	})
}

func TestMessageSignVerifier_Verify(t *testing.T) {
	t.Parallel()

	t.Run("empty key should error", func(t *testing.T) {
		t.Parallel()

		sv, _ := NewMessageSignVerifier(&mock.KeyGeneratorStub{})
		require.NotNil(t, sv)

		err := sv.Verify(nil, nil, nil)
		require.Equal(t, genesis.ErrEmptyPubKey, err)
	})
	t.Run("invalid pub key should error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		sv, _ := NewMessageSignVerifier(&mock.KeyGeneratorStub{
			CheckPublicKeyValidCalled: func(b []byte) error {
				return expectedErr
			},
		})
		err := sv.Verify(nil, nil, []byte("pk"))
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		sv, _ := NewMessageSignVerifier(&mock.KeyGeneratorStub{
			CheckPublicKeyValidCalled: func(b []byte) error {
				return nil
			},
		})
		err := sv.Verify(nil, nil, []byte("pk"))
		assert.NoError(t, err)
	})
}

func TestMessageSignVerifier_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var sv *messageSignVerifier
	require.True(t, sv.IsInterfaceNil())

	sv, _ = NewMessageSignVerifier(&mock.KeyGeneratorStub{})
	require.False(t, sv.IsInterfaceNil())
}
