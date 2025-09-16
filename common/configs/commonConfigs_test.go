package configs_test

// import (
// 	"testing"

// 	"github.com/multiversx/mx-chain-go/common/configs"
// 	"github.com/multiversx/mx-chain-go/config"
// 	"github.com/stretchr/testify/require"
// )

// func TestNewProcessConfigsByEpoch(t *testing.T) {
// 	t.Parallel()

// 	t.Run("should return error for empty config by epoch", func(t *testing.T) {
// 		t.Parallel()

// 		pce, err := configs.NewProcessConfigsHandler(nil, nil)
// 		require.Nil(t, pce)
// 		require.Equal(t, configs.ErrEmptyProcessConfigsByEpoch, err)
// 	})

// 	t.Run("should return error for empty config by round", func(t *testing.T) {
// 		t.Parallel()

// 		pce, err := configs.NewProcessConfigsHandler([]config.ProcessConfigByEpoch{}, nil)
// 		require.Nil(t, pce)
// 		require.Equal(t, configs.ErrEmptyProcessConfigsByEpoch, err)
// 	})

// 	t.Run("should return error for duplicated epoch configs", func(t *testing.T) {
// 		t.Parallel()

// 		conf := []config.ProcessConfigByEpoch{
// 			{EnableEpoch: 0, MaxMetaNoncesBehind: 15},
// 			{EnableEpoch: 0, MaxMetaNoncesBehind: 30},
// 		}
// 		pce, err := configs.NewProcessConfigsHandler(conf, []config.ProcessConfigByRound{})
// 		require.Nil(t, pce)
// 		require.Equal(t, configs.ErrDuplicatedEpochConfig, err)
// 	})

// 	t.Run("should return error for missing epoch 0 config", func(t *testing.T) {
// 		t.Parallel()

// 		conf := []config.ProcessConfigByEpoch{
// 			{EnableEpoch: 1, MaxMetaNoncesBehind: 15},
// 			{EnableEpoch: 2, MaxMetaNoncesBehind: 30},
// 		}
// 		pce, err := configs.NewProcessConfigsHandler(conf, []config.ProcessConfigByRound{})
// 		require.Nil(t, pce)
// 		require.Equal(t, configs.ErrMissingEpochZeroConfig, err)
// 	})

// 	t.Run("should work", func(t *testing.T) {
// 		t.Parallel()

// 		conf := []config.ProcessConfigByEpoch{
// 			{EnableEpoch: 0, MaxMetaNoncesBehind: 15},
// 			{EnableEpoch: 2, MaxMetaNoncesBehind: 30},
// 			{EnableEpoch: 1, MaxMetaNoncesBehind: 45},
// 		}
// 		confByRound := []config.ProcessConfigByRound{
// 			{EnableRound: 0, MaxRoundsWithoutNewBlockReceived: 10},
// 			{EnableRound: 1, MaxRoundsWithoutNewBlockReceived: 11},
// 		}

// 		pce, err := configs.NewProcessConfigsHandler(conf, confByRound)
// 		require.NotNil(t, pce)
// 		require.NoError(t, err)
// 		require.False(t, pce.IsInterfaceNil())

// 		require.Equal(t, uint32(15), pce.GetOrderedConfigsByEpoch(0).MaxMetaNoncesBehind)
// 		require.Equal(t, uint32(45), pce.GetOrderedConfigsByEpoch(1).MaxMetaNoncesBehind)
// 		require.Equal(t, uint32(30), pce.GetOrderedConfigsByEpoch(2).MaxMetaNoncesBehind)
// 	})
// }

// func TestProcessConfigsByEpoch_Getters(t *testing.T) {
// 	t.Parallel()

// 	conf := []config.ProcessConfigByEpoch{
// 		{EnableEpoch: 0, MaxMetaNoncesBehind: 10, MaxMetaNoncesBehindForGlobalStuck: 11, MaxShardNoncesBehind: 12},
// 		{EnableEpoch: 1, MaxMetaNoncesBehind: 20, MaxMetaNoncesBehindForGlobalStuck: 21, MaxShardNoncesBehind: 22},
// 	}

// 	confByRound := []config.ProcessConfigByRound{
// 		{EnableRound: 0, MaxRoundsWithoutNewBlockReceived: 10, MaxRoundsWithoutCommittedBlock: 20},
// 		{EnableRound: 1, MaxRoundsWithoutNewBlockReceived: 11, MaxRoundsWithoutCommittedBlock: 21},
// 	}

// 	t.Run("get max meta nonces behind", func(t *testing.T) {
// 		t.Parallel()

// 		pce, _ := configs.NewProcessConfigsHandler(conf, confByRound)

// 		maxMetaNoncesBehind := pce.GetMaxMetaNoncesBehindByEpoch(0)
// 		require.Equal(t, uint32(10), maxMetaNoncesBehind)

// 		maxMetaNoncesBehind = pce.GetMaxMetaNoncesBehindByEpoch(1)
// 		require.Equal(t, uint32(20), maxMetaNoncesBehind)
// 	})

// 	t.Run("get max meta nonces behind for global stuck", func(t *testing.T) {
// 		t.Parallel()

// 		pce, _ := configs.NewProcessConfigsHandler(conf, confByRound)

// 		maxMetaNoncesBehindForGlobalStuck := pce.GetMaxMetaNoncesBehindForBlobalStuckByEpoch(0)
// 		require.Equal(t, uint32(11), maxMetaNoncesBehindForGlobalStuck)

// 		maxMetaNoncesBehindForGlobalStuck = pce.GetMaxMetaNoncesBehindForBlobalStuckByEpoch(1)
// 		require.Equal(t, uint32(21), maxMetaNoncesBehindForGlobalStuck)
// 	})

// 	t.Run("get max shard nonces behind", func(t *testing.T) {
// 		t.Parallel()

// 		pce, _ := configs.NewProcessConfigsHandler(conf, confByRound)

// 		maxShardNoncesBehind := pce.GetMaxShardNoncesBehindByEpoch(0)
// 		require.Equal(t, uint32(12), maxShardNoncesBehind)

// 		maxShardNoncesBehind = pce.GetMaxShardNoncesBehindByEpoch(1)
// 		require.Equal(t, uint32(22), maxShardNoncesBehind)
// 	})

// 	t.Run("get max rounds without new block received", func(t *testing.T) {
// 		t.Parallel()

// 		pce, _ := configs.NewProcessConfigsHandler(conf, confByRound)

// 		maxRoundsWithoutNewBlockReceived := pce.GetMaxRoundsWithoutNewBlockReceivedByRound(0)
// 		require.Equal(t, uint32(10), maxRoundsWithoutNewBlockReceived)

// 		maxRoundsWithoutNewBlockReceived = pce.GetMaxRoundsWithoutNewBlockReceivedByRound(1)
// 		require.Equal(t, uint32(11), maxRoundsWithoutNewBlockReceived)
// 	})

// 	t.Run("get max rounds without committed block", func(t *testing.T) {
// 		t.Parallel()

// 		pce, _ := configs.NewProcessConfigsHandler(conf, confByRound)

// 		maxRoundsWithoutCommittedBlock := pce.GetMaxRoundsWithoutCommittedBlock(0)
// 		require.Equal(t, uint32(20), maxRoundsWithoutCommittedBlock)

// 		maxRoundsWithoutCommittedBlock = pce.GetMaxRoundsWithoutCommittedBlock(1)
// 		require.Equal(t, uint32(21), maxRoundsWithoutCommittedBlock)
// 	})
// }
