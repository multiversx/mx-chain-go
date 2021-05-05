package factory

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
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
}

func TestNewDataPoolFromConfig_BadConfigShouldErr(t *testing.T) {
	// We test one (arbitrary and trivial) erroneous config for each component that needs to be created

	args := getGoodArgs()
	args.Config.TxDataPool.Capacity = 0
	holder, err := NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.NotNil(t, err)

	args = getGoodArgs()
	args.Config.UnsignedTransactionDataPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.NotNil(t, err)

	args = getGoodArgs()
	args.Config.RewardTransactionDataPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.NotNil(t, err)

	args = getGoodArgs()
	args.Config.HeadersPoolConfig.MaxHeadersPerShard = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.NotNil(t, err)

	args = getGoodArgs()
	args.Config.TxBlockBodyDataPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.NotNil(t, err)

	args = getGoodArgs()
	args.Config.PeerBlockBodyDataPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.NotNil(t, err)

	args = getGoodArgs()
	args.Config.TrieSyncStorage.Cache.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.NotNil(t, err)
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
	}
}
