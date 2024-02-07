package bls

import (
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/subRounds"
	"github.com/stretchr/testify/require"
)

func TestSubRoundStartExtraSignersHolder_Reset(t *testing.T) {
	t.Parallel()

	wasResetCalled1 := false
	wasResetCalled2 := false
	expectedPubKeys := []string{"pk1", "pk2"}

	extraSigner1 := &subRounds.SubRoundStartExtraSignatureHandlerMock{
		ResetCalled: func(pubKeys []string) error {
			require.Equal(t, expectedPubKeys, pubKeys)

			wasResetCalled1 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id1"
		},
	}

	extraSigner2 := &subRounds.SubRoundStartExtraSignatureHandlerMock{
		ResetCalled: func(pubKeys []string) error {
			require.Equal(t, expectedPubKeys, pubKeys)

			wasResetCalled2 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id2"
		},
	}

	holder := NewSubRoundStartExtraSignersHolder()
	require.False(t, holder.IsInterfaceNil())

	err := holder.RegisterExtraSigningHandler(extraSigner1)
	require.Nil(t, err)
	err = holder.RegisterExtraSigningHandler(extraSigner2)
	require.Nil(t, err)
	err = holder.RegisterExtraSigningHandler(extraSigner2)
	require.Equal(t, errors.ErrExtraSignerIdAlreadyExists, err)

	err = holder.Reset(expectedPubKeys)
	require.Nil(t, err)
	require.True(t, wasResetCalled1)
	require.True(t, wasResetCalled2)
}
