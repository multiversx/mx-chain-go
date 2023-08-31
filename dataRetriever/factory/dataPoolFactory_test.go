package factory

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/dataPool/headersCache"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/stretchr/testify/require"
)

func TestNewDataPoolFromConfig(t *testing.T) {
	args := getGoodArgs()
	holder, err := NewDataPoolFromConfig(args)
	require.Nil(t, err)
	require.NotNil(t, holder)
}

func TestNewDataPoolFromConfig_MissingDependencyShouldErr(t *testing.T) {
	args := getGoodArgs()
	args.Config = nil
	holder, err := NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.Equal(t, dataRetriever.ErrNilConfig, err)

	args = getGoodArgs()
	args.EconomicsData = nil
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.Equal(t, dataRetriever.ErrNilEconomicsData, err)

	args = getGoodArgs()
	args.ShardCoordinator = nil
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.Equal(t, dataRetriever.ErrNilShardCoordinator, err)

	args = getGoodArgs()
	args.PathManager = nil
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.Equal(t, dataRetriever.ErrNilPathManager, err)
}

func TestNewDataPoolFromConfig_BadConfigShouldErr(t *testing.T) {
	// We test one (arbitrary and trivial) erroneous config for each component that needs to be created

	args := getGoodArgs()
	args.Config.TxDataPool.Capacity = 0
	holder, err := NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.True(t, errors.Is(err, dataRetriever.ErrCacheConfigInvalidSize))
	require.True(t, strings.Contains(err.Error(), "the cache for the transactions"))

	args = getGoodArgs()
	args.Config.UserTxDataPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.True(t, errors.Is(err, dataRetriever.ErrCacheConfigInvalidSize))
	require.True(t, strings.Contains(err.Error(), "the cache for the user transactions"))

	args = getGoodArgs()
	args.Config.UnsignedTransactionDataPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.True(t, errors.Is(err, storage.ErrInvalidConfig))
	require.True(t, strings.Contains(err.Error(), "the cache for the unsigned transactions"))

	args = getGoodArgs()
	args.Config.RewardTransactionDataPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.True(t, errors.Is(err, storage.ErrInvalidConfig))
	require.True(t, strings.Contains(err.Error(), "the cache for the rewards"))

	args = getGoodArgs()
	args.Config.HeadersPoolConfig.MaxHeadersPerShard = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.True(t, errors.Is(err, headersCache.ErrInvalidHeadersCacheParameter))
	require.True(t, strings.Contains(err.Error(), "the cache for the headers"))

	args = getGoodArgs()
	args.Config.TxBlockBodyDataPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.NotNil(t, err)
	require.True(t, strings.Contains(err.Error(), "must provide a positive size while creating the cache for the miniblocks"))

	args = getGoodArgs()
	args.Config.PeerBlockBodyDataPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.NotNil(t, err)
	require.True(t, strings.Contains(err.Error(), "must provide a positive size while creating the cache for the peer mini block body"))

	args = getGoodArgs()
	args.Config.TrieSyncStorage.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.True(t, errors.Is(err, storage.ErrCacheSizeInvalid))
	require.True(t, strings.Contains(err.Error(), "the cache for the trie nodes"))

	args = getGoodArgs()
	args.Config.TrieSyncStorage.EnableDB = true
	args.Config.TrieSyncStorage.DB.Type = "invalid DB type"
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.True(t, errors.Is(err, storage.ErrNotSupportedDBType))
	require.True(t, strings.Contains(err.Error(), "the db for the trie nodes"))

	args = getGoodArgs()
	args.Config.TrieNodesChunksDataPool.Type = "invalid cache type"
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.True(t, errors.Is(err, storage.ErrNotSupportedCacheType))
	require.True(t, strings.Contains(err.Error(), "the cache for the trie chunks"))

	args = getGoodArgs()
	args.Config.SmartContractDataPool.Type = "invalid cache type"
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.True(t, errors.Is(err, storage.ErrNotSupportedCacheType))
	require.True(t, strings.Contains(err.Error(), "the cache for the smartcontract results"))

	args = getGoodArgs()
	args.Config.HeartbeatV2.HeartbeatPool.Type = "invalid cache type"
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.True(t, errors.Is(err, storage.ErrNotSupportedCacheType))
	require.True(t, strings.Contains(err.Error(), "the cache for the heartbeat messages"))

	args = getGoodArgs()
	args.Config.ValidatorInfoPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.True(t, errors.Is(err, storage.ErrInvalidConfig))
	require.True(t, strings.Contains(err.Error(), "the cache for the validator info results"))
}

func getGoodArgs() ArgsDataPool {
	testEconomics := &economicsmocks.EconomicsHandlerStub{
		MinGasPriceCalled: func() uint64 {
			return 200000000000
		},
	}
	config := testscommon.GetGeneralConfig()

	return ArgsDataPool{
		Config:           &config,
		EconomicsData:    testEconomics,
		ShardCoordinator: mock.NewMultipleShardsCoordinatorMock(),
		Marshalizer:      &mock.MarshalizerMock{},
		PathManager:      &testscommon.PathManagerStub{},
	}
}
