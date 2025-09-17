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

		pce, err := configs.NewCommonConfigsHandler(nil)
		require.Nil(t, pce)
		require.Equal(t, configs.ErrEmptyProcessConfigsByEpoch, err)
	})

	t.Run("should return error for empty config by round", func(t *testing.T) {
		t.Parallel()

		pce, err := configs.NewCommonConfigsHandler([]config.EpochStartConfigByEpoch{})
		require.Nil(t, pce)
		require.Equal(t, configs.ErrEmptyProcessConfigsByEpoch, err)
	})

	t.Run("should return error for duplicated epoch configs", func(t *testing.T) {
		t.Parallel()

		conf := []config.EpochStartConfigByEpoch{
			{EnableEpoch: 0, GracePeriodRounds: 1},
			{EnableEpoch: 0, GracePeriodRounds: 2},
		}
		pce, err := configs.NewCommonConfigsHandler(conf)
		require.Nil(t, pce)
		require.Equal(t, configs.ErrDuplicatedEpochConfig, err)
	})

	t.Run("should return error for missing epoch 0 config", func(t *testing.T) {
		t.Parallel()

		conf := []config.EpochStartConfigByEpoch{
			{EnableEpoch: 1, GracePeriodRounds: 1},
			{EnableEpoch: 2, GracePeriodRounds: 2},
		}
		pce, err := configs.NewCommonConfigsHandler(conf)
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

		pce, err := configs.NewCommonConfigsHandler(conf)
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

	t.Run("get grace period rounds by epoch", func(t *testing.T) {
		t.Parallel()

		cc, _ := configs.NewCommonConfigsHandler(conf)

		gracePeriodRounds := cc.GetGracePeriodRoundsByEpoch(0)
		require.Equal(t, uint32(10), gracePeriodRounds)

		gracePeriodRounds = cc.GetGracePeriodRoundsByEpoch(1)
		require.Equal(t, uint32(11), gracePeriodRounds)
	})

	t.Run("get extra delay for request block info", func(t *testing.T) {
		t.Parallel()

		cc, _ := configs.NewCommonConfigsHandler(conf)

		extraDelayForRequests := cc.GetExtraDelayForRequestBlockInfoInMs(0)
		require.Equal(t, uint32(20), extraDelayForRequests)

		extraDelayForRequests = cc.GetExtraDelayForRequestBlockInfoInMs(1)
		require.Equal(t, uint32(21), extraDelayForRequests)
	})
}
