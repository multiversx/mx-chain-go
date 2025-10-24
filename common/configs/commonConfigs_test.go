package configs_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common/configs"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
)

func TestNewCommonConfigsHandler(t *testing.T) {
	t.Parallel()

	t.Run("should return error for empty config by epoch", func(t *testing.T) {
		t.Parallel()

		pce, err := configs.NewCommonConfigsHandler(nil, nil, nil)
		require.Nil(t, pce)
		require.Equal(t, configs.ErrEmptyCommonConfigsByEpoch, err)
	})

	t.Run("should return error for empty config by round", func(t *testing.T) {
		t.Parallel()

		pce, err := configs.NewCommonConfigsHandler([]config.EpochStartConfigByEpoch{}, nil, nil)
		require.Nil(t, pce)
		require.Equal(t, configs.ErrEmptyCommonConfigsByRound, err)
	})

	t.Run("should return error for duplicated epoch configs", func(t *testing.T) {
		t.Parallel()

		conf := []config.EpochStartConfigByEpoch{
			{EnableEpoch: 0, GracePeriodRounds: 1},
			{EnableEpoch: 0, GracePeriodRounds: 2},
		}
		pce, err := configs.NewCommonConfigsHandler(conf, []config.EpochStartConfigByRound{}, []config.ConsensusConfigByEpoch{})
		require.Nil(t, pce)
		require.Equal(t, configs.ErrDuplicatedEpochConfig, err)
	})

	t.Run("should return error for missing epoch 0 config", func(t *testing.T) {
		t.Parallel()

		conf := []config.EpochStartConfigByEpoch{
			{EnableEpoch: 1, GracePeriodRounds: 1},
			{EnableEpoch: 2, GracePeriodRounds: 2},
		}
		pce, err := configs.NewCommonConfigsHandler(conf, []config.EpochStartConfigByRound{}, []config.ConsensusConfigByEpoch{})
		require.Nil(t, pce)
		require.Equal(t, configs.ErrMissingEpochZeroConfig, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		conf := []config.EpochStartConfigByEpoch{
			{EnableEpoch: 0, GracePeriodRounds: 0},
			{EnableEpoch: 2, GracePeriodRounds: 2},
			{EnableEpoch: 1, GracePeriodRounds: 1},
		}
		confByRound := []config.EpochStartConfigByRound{
			{EnableRound: 0, MaxRoundsWithoutCommittedStartInEpochBlock: 10},
			{EnableRound: 1, MaxRoundsWithoutCommittedStartInEpochBlock: 11},
		}
		consensusConf := []config.ConsensusConfigByEpoch{
			{EnableEpoch: 0, NumRoundsToWaitBeforeSignalingChronologyStuck: 10},
			{EnableEpoch: 1, NumRoundsToWaitBeforeSignalingChronologyStuck: 11},
		}

		pce, err := configs.NewCommonConfigsHandler(conf, confByRound, consensusConf)
		require.NotNil(t, pce)
		require.NoError(t, err)
		require.False(t, pce.IsInterfaceNil())

		require.Equal(t, uint32(0), pce.GetOrderedEpochStartConfigByEpoch(0).GracePeriodRounds)
		require.Equal(t, uint32(1), pce.GetOrderedEpochStartConfigByEpoch(1).GracePeriodRounds)
		require.Equal(t, uint32(2), pce.GetOrderedEpochStartConfigByEpoch(2).GracePeriodRounds)
	})
}

func TestCommonConfigsByEpoch_Getters(t *testing.T) {
	t.Parallel()

	conf := []config.EpochStartConfigByEpoch{
		{EnableEpoch: 0, GracePeriodRounds: 10, ExtraDelayForRequestBlockInfoInMilliseconds: 20},
		{EnableEpoch: 1, GracePeriodRounds: 11, ExtraDelayForRequestBlockInfoInMilliseconds: 21},
		{EnableEpoch: 2, GracePeriodRounds: 12, ExtraDelayForRequestBlockInfoInMilliseconds: 22},
	}

	confByRound := []config.EpochStartConfigByRound{
		{EnableRound: 0, MaxRoundsWithoutCommittedStartInEpochBlock: 30},
		{EnableRound: 1, MaxRoundsWithoutCommittedStartInEpochBlock: 31},
	}

	consensusConf := []config.ConsensusConfigByEpoch{
		{EnableEpoch: 0, NumRoundsToWaitBeforeSignalingChronologyStuck: 10},
		{EnableEpoch: 1, NumRoundsToWaitBeforeSignalingChronologyStuck: 11},
	}

	t.Run("get grace period rounds by epoch", func(t *testing.T) {
		t.Parallel()

		cc, _ := configs.NewCommonConfigsHandler(conf, confByRound, consensusConf)

		gracePeriodRounds := cc.GetGracePeriodRoundsByEpoch(0)
		require.Equal(t, uint32(10), gracePeriodRounds)

		gracePeriodRounds = cc.GetGracePeriodRoundsByEpoch(1)
		require.Equal(t, uint32(11), gracePeriodRounds)
	})

	t.Run("get extra delay for request block info", func(t *testing.T) {
		t.Parallel()

		cc, _ := configs.NewCommonConfigsHandler(conf, confByRound, consensusConf)

		extraDelayForRequests := cc.GetExtraDelayForRequestBlockInfoInMs(0)
		require.Equal(t, uint32(20), extraDelayForRequests)

		extraDelayForRequests = cc.GetExtraDelayForRequestBlockInfoInMs(1)
		require.Equal(t, uint32(21), extraDelayForRequests)
	})

	t.Run("get max rounds without commited start in epoch block by roud", func(t *testing.T) {
		t.Parallel()

		cc, _ := configs.NewCommonConfigsHandler(conf, confByRound, consensusConf)

		maxRoundsWithoutCommitedStartInEpochBlock := cc.GetMaxRoundsWithoutCommittedStartInEpochBlockInRound(0)
		require.Equal(t, uint32(30), maxRoundsWithoutCommitedStartInEpochBlock)

		maxRoundsWithoutCommitedStartInEpochBlock = cc.GetMaxRoundsWithoutCommittedStartInEpochBlockInRound(1)
		require.Equal(t, uint32(31), maxRoundsWithoutCommitedStartInEpochBlock)
	})
}
