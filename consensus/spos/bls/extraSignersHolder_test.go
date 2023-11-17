package bls

import (
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/subRounds"
	"github.com/stretchr/testify/require"
)

func TestNewEmptyExtraSignersHolder(t *testing.T) {
	t.Parallel()

	holder := NewEmptyExtraSignersHolder()
	require.False(t, holder.IsInterfaceNil())

	require.False(t, holder.GetSubRoundStartExtraSignersHolder().IsInterfaceNil())
	require.False(t, holder.GetSubRoundSignatureExtraSignersHolder().IsInterfaceNil())
	require.False(t, holder.GetSubRoundEndExtraSignersHolder().IsInterfaceNil())
}

func TestNewExtraSignersHolder(t *testing.T) {
	t.Parallel()

	t.Run("nil start round holder, should return error", func(t *testing.T) {
		holder, err := NewExtraSignersHolder(
			nil,
			&subRounds.SubRoundSignatureExtraSignersHolderMock{},
			&subRounds.SubRoundEndExtraSignersHolderMock{},
		)
		require.Nil(t, holder)
		require.Equal(t, errors.ErrNilStartRoundExtraSignersHolder, err)
	})
	t.Run("nil sign round holder, should return error", func(t *testing.T) {
		holder, err := NewExtraSignersHolder(
			&subRounds.SubRoundStartExtraSignersHolderMock{},
			nil,
			&subRounds.SubRoundEndExtraSignersHolderMock{},
		)
		require.Nil(t, holder)
		require.Equal(t, errors.ErrNilSignatureRoundExtraSignersHolder, err)
	})
	t.Run("nil end round holder, should return error", func(t *testing.T) {
		holder, err := NewExtraSignersHolder(
			&subRounds.SubRoundStartExtraSignersHolderMock{},
			&subRounds.SubRoundSignatureExtraSignersHolderMock{},
			nil,
		)
		require.Nil(t, holder)
		require.Equal(t, errors.ErrNilEndRoundExtraSignersHolder, err)
	})
	t.Run("should work", func(t *testing.T) {
		holder, err := NewExtraSignersHolder(
			&subRounds.SubRoundStartExtraSignersHolderMock{},
			&subRounds.SubRoundSignatureExtraSignersHolderMock{},
			&subRounds.SubRoundEndExtraSignersHolderMock{},
		)
		require.Nil(t, err)
		require.False(t, holder.IsInterfaceNil())
	})
}

func TestExtraSignersHolder_Getters(t *testing.T) {
	holder, _ := NewExtraSignersHolder(
		&subRounds.SubRoundStartExtraSignersHolderMock{},
		&subRounds.SubRoundSignatureExtraSignersHolderMock{},
		&subRounds.SubRoundEndExtraSignersHolderMock{},
	)

	require.Equal(t, &subRounds.SubRoundStartExtraSignersHolderMock{}, holder.GetSubRoundStartExtraSignersHolder())
	require.Equal(t, &subRounds.SubRoundSignatureExtraSignersHolderMock{}, holder.GetSubRoundSignatureExtraSignersHolder())
	require.Equal(t, &subRounds.SubRoundEndExtraSignersHolderMock{}, holder.GetSubRoundEndExtraSignersHolder())
}
