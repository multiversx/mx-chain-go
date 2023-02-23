package processing_test

import (
	"strings"
	"sync"
	"testing"

	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	outportCore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/factory/mock"
	processComp "github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/process"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessComponentsFactory_CloseShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
	pcf, err := processComp.NewProcessComponentsFactory(processArgs)
	require.Nil(t, err)

	pc, err := pcf.Create()
	require.Nil(t, err)

	err = pc.Close()
	require.NoError(t, err)
}

func TestProcessComponentsFactory_CreateWithInvalidTxAccumulatorTimeExpectError(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
	processArgs.Config.Antiflood.TxAccumulator.MaxAllowedTimeInMilliseconds = 0
	pcf, err := processComp.NewProcessComponentsFactory(processArgs)
	require.Nil(t, err)

	instance, err := pcf.Create()
	require.Nil(t, instance)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), process.ErrInvalidValue.Error()))
}

func TestProcessComponentsFactory_IndexGenesisBlocks(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
	processArgs.Data = &mock.DataComponentsMock{
		Storage: &storageStubs.ChainStorerStub{},
	}

	saveBlockCalledMutex := sync.Mutex{}

	outportHandler := &outport.OutportStub{
		HasDriversCalled: func() bool {
			return true
		},
		SaveBlockCalled: func(args *outportCore.ArgsSaveBlockData) {
			saveBlockCalledMutex.Lock()
			require.NotNil(t, args)

			bodyRequired := &dataBlock.Body{
				MiniBlocks: make([]*block.MiniBlock, 4),
			}

			txsPoolRequired := &outportCore.Pool{}

			assert.Equal(t, txsPoolRequired, args.TransactionsPool)
			assert.Equal(t, bodyRequired, args.Body)
			saveBlockCalledMutex.Unlock()
		},
	}

	processArgs.StatusComponents = &mainFactoryMocks.StatusComponentsStub{
		Outport: outportHandler,
	}

	pcf, err := processComp.NewProcessComponentsFactory(processArgs)
	require.Nil(t, err)

	genesisBlocks := make(map[uint32]coreData.HeaderHandler)
	indexingData := make(map[uint32]*genesis.IndexingData)

	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		genesisBlocks[i] = &block.Header{}
	}

	err = pcf.IndexGenesisBlocks(genesisBlocks, indexingData)
	require.Nil(t, err)
}

func TestProcessComponentsFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("creating process components factory in regular chain should work", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
		pcf, _ := processComp.NewProcessComponentsFactory(processArgs)

		require.NotNil(t, pcf)

		pcf.SetChainRunType(common.ChainRunTypeRegular)

		pc, err := pcf.Create()
		require.NotNil(t, pc)
		require.Nil(t, err)
	})

	t.Run("creating process components factory in sovereign chain should work", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
		pcf, _ := processComp.NewProcessComponentsFactory(processArgs)

		require.NotNil(t, pcf)

		pcf.SetChainRunType(common.ChainRunTypeSovereign)

		pc, err := pcf.Create()
		require.NotNil(t, pc)
		require.Nil(t, err)
	})

	t.Run("creating process components factory in sovereign observer chain should work", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
		pcf, _ := processComp.NewProcessComponentsFactory(processArgs)

		require.NotNil(t, pcf)

		pcf.SetChainRunType(common.ChainRunTypeSovereign)

		pc, err := pcf.Create()
		require.NotNil(t, pc)
		require.Nil(t, err)
	})
}
