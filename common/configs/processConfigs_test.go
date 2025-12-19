package configs_test

import (
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/common/configs"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/stretchr/testify/require"
)

func getConfigsByRound() []config.ProcessConfigByRound {
	return []config.ProcessConfigByRound{
		{
			EnableRound:                            0,
			MaxRoundsWithoutNewBlockReceived:       10,
			MaxRoundsToKeepUnprocessedMiniBlocks:   1,
			MaxRoundsToKeepUnprocessedTransactions: 1,
			NumFloodingRoundsSlowReacting:          2,
			NumFloodingRoundsFastReacting:          3,
			NumFloodingRoundsOutOfSpecs:            4,
		},
		{
			EnableRound:                            1,
			MaxRoundsWithoutNewBlockReceived:       11,
			MaxRoundsToKeepUnprocessedMiniBlocks:   1,
			MaxRoundsToKeepUnprocessedTransactions: 1,
			NumFloodingRoundsSlowReacting:          20,
			NumFloodingRoundsFastReacting:          30,
			NumFloodingRoundsOutOfSpecs:            40,
		},
	}
}

func TestNewProcessConfigsByEpoch(t *testing.T) {
	t.Parallel()

	t.Run("should return error for empty config by epoch", func(t *testing.T) {
		t.Parallel()

		pce, err := configs.NewProcessConfigsHandler(nil, nil, &epochNotifier.RoundNotifierStub{})
		require.Nil(t, pce)
		require.Equal(t, configs.ErrEmptyProcessConfigsByEpoch, err)
	})

	t.Run("should return error for empty config by round", func(t *testing.T) {
		t.Parallel()

		pce, err := configs.NewProcessConfigsHandler([]config.ProcessConfigByEpoch{}, nil, &epochNotifier.RoundNotifierStub{})
		require.Nil(t, pce)
		require.Equal(t, configs.ErrEmptyProcessConfigsByEpoch, err)
	})

	t.Run("should return error for duplicated epoch configs", func(t *testing.T) {
		t.Parallel()

		conf := []config.ProcessConfigByEpoch{
			{EnableEpoch: 0, MaxMetaNoncesBehind: 15},
			{EnableEpoch: 0, MaxMetaNoncesBehind: 30},
		}
		pce, err := configs.NewProcessConfigsHandler(conf, []config.ProcessConfigByRound{}, &epochNotifier.RoundNotifierStub{})
		require.Nil(t, pce)
		require.Equal(t, configs.ErrDuplicatedEpochConfig, err)
	})

	t.Run("should return error for missing epoch 0 config", func(t *testing.T) {
		t.Parallel()

		conf := []config.ProcessConfigByEpoch{
			{EnableEpoch: 1, MaxMetaNoncesBehind: 15},
			{EnableEpoch: 2, MaxMetaNoncesBehind: 30},
		}
		pce, err := configs.NewProcessConfigsHandler(conf, []config.ProcessConfigByRound{}, &epochNotifier.RoundNotifierStub{})
		require.Nil(t, pce)
		require.Equal(t, configs.ErrMissingEpochZeroConfig, err)
	})

	t.Run("should return error for zero MaxRoundsToKeepUnprocessedTransactions value", func(t *testing.T) {
		t.Parallel()

		confByEpoch := []config.ProcessConfigByEpoch{
			{EnableEpoch: 0, MaxMetaNoncesBehind: 15},
		}
		confByRound := getConfigsByRound()
		confByRound[0].MaxRoundsToKeepUnprocessedTransactions = 0

		pce, err := configs.NewProcessConfigsHandler(confByEpoch, confByRound, &epochNotifier.RoundNotifierStub{})
		require.Nil(t, pce)
		require.ErrorIs(t, err, process.ErrInvalidValue)
		require.True(t, strings.Contains(err.Error(), "MaxRoundsToKeepUnprocessedTransactions"))
	})

	t.Run("should return error for zero MaxRoundsToKeepUnprocessedMiniBlocks value", func(t *testing.T) {
		t.Parallel()

		confByEpoch := []config.ProcessConfigByEpoch{
			{EnableEpoch: 0, MaxMetaNoncesBehind: 15},
		}
		confByRound := getConfigsByRound()
		confByRound[0].MaxRoundsToKeepUnprocessedMiniBlocks = 0

		pce, err := configs.NewProcessConfigsHandler(confByEpoch, confByRound, &epochNotifier.RoundNotifierStub{})
		require.Nil(t, pce)
		require.ErrorIs(t, err, process.ErrInvalidValue)
		require.True(t, strings.Contains(err.Error(), "MaxRoundsToKeepUnprocessedMiniBlocks"))
	})

	t.Run("should return error for invalid num flooding rounds fast reacting value", func(t *testing.T) {
		t.Parallel()

		confByEpoch := []config.ProcessConfigByEpoch{
			{EnableEpoch: 0, MaxMetaNoncesBehind: 15},
		}
		confByRound := getConfigsByRound()
		confByRound[0].NumFloodingRoundsFastReacting = 0

		pce, err := configs.NewProcessConfigsHandler(confByEpoch, confByRound, &epochNotifier.RoundNotifierStub{})
		require.Nil(t, pce)
		require.ErrorIs(t, err, process.ErrInvalidValue)
		require.True(t, strings.Contains(err.Error(), "NumFloodingRoundsFastReacting"))
	})

	t.Run("should return error for invalid num flooding rounds slow reacting value", func(t *testing.T) {
		t.Parallel()

		confByEpoch := []config.ProcessConfigByEpoch{
			{EnableEpoch: 0, MaxMetaNoncesBehind: 15},
		}
		confByRound := getConfigsByRound()
		confByRound[0].NumFloodingRoundsSlowReacting = 0

		pce, err := configs.NewProcessConfigsHandler(confByEpoch, confByRound, &epochNotifier.RoundNotifierStub{})
		require.Nil(t, pce)
		require.ErrorIs(t, err, process.ErrInvalidValue)
		require.True(t, strings.Contains(err.Error(), "NumFloodingRoundsSlowReacting"))
	})

	t.Run("should return error for invalid num flooding rounds out of specs value", func(t *testing.T) {
		t.Parallel()

		confByEpoch := []config.ProcessConfigByEpoch{
			{EnableEpoch: 0, MaxMetaNoncesBehind: 15},
		}
		confByRound := getConfigsByRound()
		confByRound[0].NumFloodingRoundsOutOfSpecs = 0

		pce, err := configs.NewProcessConfigsHandler(confByEpoch, confByRound, &epochNotifier.RoundNotifierStub{})
		require.Nil(t, pce)
		require.ErrorIs(t, err, process.ErrInvalidValue)
		require.True(t, strings.Contains(err.Error(), "NumFloodingRoundsOutOfSpecs"))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		conf := []config.ProcessConfigByEpoch{
			{EnableEpoch: 0, MaxMetaNoncesBehind: 15},
			{EnableEpoch: 2, MaxMetaNoncesBehind: 30},
			{EnableEpoch: 1, MaxMetaNoncesBehind: 45},
		}
		confByRound := getConfigsByRound()

		pce, err := configs.NewProcessConfigsHandler(conf, confByRound, &epochNotifier.RoundNotifierStub{})
		require.NotNil(t, pce)
		require.NoError(t, err)
		require.False(t, pce.IsInterfaceNil())

		require.Equal(t, uint32(15), pce.GetOrderedConfigsByEpoch(0).MaxMetaNoncesBehind)
		require.Equal(t, uint32(45), pce.GetOrderedConfigsByEpoch(1).MaxMetaNoncesBehind)
		require.Equal(t, uint32(30), pce.GetOrderedConfigsByEpoch(2).MaxMetaNoncesBehind)
	})
}

func TestProcessConfigsByEpoch_Getters(t *testing.T) {
	t.Parallel()

	conf := []config.ProcessConfigByEpoch{
		{EnableEpoch: 0,
			MaxMetaNoncesBehind:               10,
			MaxMetaNoncesBehindForGlobalStuck: 11,
			MaxShardNoncesBehind:              12,
		},
		{EnableEpoch: 1,
			MaxMetaNoncesBehind:               20,
			MaxMetaNoncesBehindForGlobalStuck: 21,
			MaxShardNoncesBehind:              22,
		},
	}

	confByRound := []config.ProcessConfigByRound{
		{EnableRound: 0,
			MaxRoundsWithoutNewBlockReceived:       10,
			MaxRoundsWithoutCommittedBlock:         20,
			MaxSyncWithErrorsAllowed:               30,
			MaxRoundsToKeepUnprocessedTransactions: 50,
			MaxRoundsToKeepUnprocessedMiniBlocks:   60,
			NumFloodingRoundsSlowReacting:          2,
			NumFloodingRoundsFastReacting:          3,
			NumFloodingRoundsOutOfSpecs:            4,
		},
		{EnableRound: 1,
			MaxRoundsWithoutNewBlockReceived:       11,
			MaxRoundsWithoutCommittedBlock:         21,
			MaxSyncWithErrorsAllowed:               31,
			MaxRoundsToKeepUnprocessedTransactions: 500,
			MaxRoundsToKeepUnprocessedMiniBlocks:   600,
			NumFloodingRoundsSlowReacting:          20,
			NumFloodingRoundsFastReacting:          30,
			NumFloodingRoundsOutOfSpecs:            40,
		},
	}

	t.Run("get max meta nonces behind", func(t *testing.T) {
		t.Parallel()

		pce, _ := configs.NewProcessConfigsHandler(conf, confByRound, &epochNotifier.RoundNotifierStub{})

		maxMetaNoncesBehind := pce.GetMaxMetaNoncesBehindByEpoch(0)
		require.Equal(t, uint32(10), maxMetaNoncesBehind)

		maxMetaNoncesBehind = pce.GetMaxMetaNoncesBehindByEpoch(1)
		require.Equal(t, uint32(20), maxMetaNoncesBehind)
	})

	t.Run("get max meta nonces behind for global stuck", func(t *testing.T) {
		t.Parallel()

		pce, _ := configs.NewProcessConfigsHandler(conf, confByRound, &epochNotifier.RoundNotifierStub{})

		maxMetaNoncesBehindForGlobalStuck := pce.GetMaxMetaNoncesBehindForGlobalStuckByEpoch(0)
		require.Equal(t, uint32(11), maxMetaNoncesBehindForGlobalStuck)

		maxMetaNoncesBehindForGlobalStuck = pce.GetMaxMetaNoncesBehindForGlobalStuckByEpoch(1)
		require.Equal(t, uint32(21), maxMetaNoncesBehindForGlobalStuck)
	})

	t.Run("get max shard nonces behind", func(t *testing.T) {
		t.Parallel()

		pce, _ := configs.NewProcessConfigsHandler(conf, confByRound, &epochNotifier.RoundNotifierStub{})

		maxShardNoncesBehind := pce.GetMaxShardNoncesBehindByEpoch(0)
		require.Equal(t, uint32(12), maxShardNoncesBehind)

		maxShardNoncesBehind = pce.GetMaxShardNoncesBehindByEpoch(1)
		require.Equal(t, uint32(22), maxShardNoncesBehind)
	})

	t.Run("get max rounds without new block received", func(t *testing.T) {
		t.Parallel()

		pce, _ := configs.NewProcessConfigsHandler(conf, confByRound, &epochNotifier.RoundNotifierStub{})

		maxRoundsWithoutNewBlockReceived := pce.GetMaxRoundsWithoutNewBlockReceivedByRound(0)
		require.Equal(t, uint32(10), maxRoundsWithoutNewBlockReceived)

		maxRoundsWithoutNewBlockReceived = pce.GetMaxRoundsWithoutNewBlockReceivedByRound(1)
		require.Equal(t, uint32(11), maxRoundsWithoutNewBlockReceived)
	})

	t.Run("get max rounds without committed block", func(t *testing.T) {
		t.Parallel()

		pce, _ := configs.NewProcessConfigsHandler(conf, confByRound, &epochNotifier.RoundNotifierStub{})

		maxRoundsWithoutCommittedBlock := pce.GetMaxRoundsWithoutCommittedBlock(0)
		require.Equal(t, uint32(20), maxRoundsWithoutCommittedBlock)

		maxRoundsWithoutCommittedBlock = pce.GetMaxRoundsWithoutCommittedBlock(1)
		require.Equal(t, uint32(21), maxRoundsWithoutCommittedBlock)
	})

	t.Run("get max allowed sync with errors", func(t *testing.T) {
		t.Parallel()

		pce, _ := configs.NewProcessConfigsHandler(conf, confByRound, &epochNotifier.RoundNotifierStub{})

		maxSyncWithErrorsAllowed := pce.GetMaxSyncWithErrorsAllowed(0)
		require.Equal(t, uint32(30), maxSyncWithErrorsAllowed)

		maxSyncWithErrorsAllowed = pce.GetMaxSyncWithErrorsAllowed(1)
		require.Equal(t, uint32(31), maxSyncWithErrorsAllowed)
	})

	t.Run("get max rounds to keep unprocessed transactions", func(t *testing.T) {
		t.Parallel()

		pce, _ := configs.NewProcessConfigsHandler(conf, confByRound, &epochNotifier.RoundNotifierStub{})

		res := pce.GetMaxRoundsToKeepUnprocessedTransactions(0)
		require.Equal(t, uint64(50), res)

		res = pce.GetMaxRoundsToKeepUnprocessedTransactions(1)
		require.Equal(t, uint64(500), res)
	})

	t.Run("get max rounds to keep unprocessed mini blocks", func(t *testing.T) {
		t.Parallel()

		pce, _ := configs.NewProcessConfigsHandler(conf, confByRound, &epochNotifier.RoundNotifierStub{})

		res := pce.GetMaxRoundsToKeepUnprocessedMiniBlocks(0)
		require.Equal(t, uint64(60), res)

		res = pce.GetMaxRoundsToKeepUnprocessedMiniBlocks(1)
		require.Equal(t, uint64(600), res)
	})
}
