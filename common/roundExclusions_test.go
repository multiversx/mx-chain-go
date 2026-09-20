package common_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
)

func TestNewRoundExclusionHandler(t *testing.T) {
	t.Parallel()

	t.Run("rejects reversed interval", func(t *testing.T) {
		_, err := common.NewRoundExclusionHandler([]config.HardforkRoundExclusionConfig{{StartRound: 8, EndRound: 7}})
		require.ErrorIs(t, err, common.ErrInvalidHardforkRoundExclusion)
	})

	t.Run("rejects overlap", func(t *testing.T) {
		_, err := common.NewRoundExclusionHandler([]config.HardforkRoundExclusionConfig{
			{StartRound: 10, EndRound: 20},
			{StartRound: 5, EndRound: 10},
		})
		require.ErrorIs(t, err, common.ErrOverlappingHardforkRoundExclusions)
	})

	t.Run("uses sorted inclusive intervals", func(t *testing.T) {
		intervals := []config.HardforkRoundExclusionConfig{
			{StartRound: 20, EndRound: 25},
			{StartRound: 5, EndRound: 10},
		}
		handler, err := common.NewRoundExclusionHandler(intervals)
		require.NoError(t, err)
		intervals[0] = config.HardforkRoundExclusionConfig{}
		require.False(t, handler.IsRoundExcluded(4))
		require.True(t, handler.IsRoundExcluded(5))
		require.True(t, handler.IsRoundExcluded(10))
		require.False(t, handler.IsRoundExcluded(11))
		require.True(t, handler.IsRoundExcluded(20))
		require.True(t, handler.IsRoundExcluded(25))
		require.False(t, handler.IsRoundExcluded(26))
	})
}

func TestResolveRoundExclusionHandler(t *testing.T) {
	t.Parallel()

	defaultHandler, err := common.ResolveRoundExclusionHandler()
	require.NoError(t, err)
	require.False(t, defaultHandler.IsRoundExcluded(1))

	configuredHandler, err := common.NewRoundExclusionHandler([]config.HardforkRoundExclusionConfig{
		{StartRound: 1, EndRound: 1},
	})
	require.NoError(t, err)
	resolvedHandler, err := common.ResolveRoundExclusionHandler(configuredHandler)
	require.NoError(t, err)
	require.Same(t, configuredHandler, resolvedHandler)

	_, err = common.ResolveRoundExclusionHandler(nil)
	require.ErrorIs(t, err, common.ErrNilRoundExclusionHandler)
	_, err = common.ResolveRoundExclusionHandler(configuredHandler, configuredHandler)
	require.ErrorIs(t, err, common.ErrNilRoundExclusionHandler)
}
