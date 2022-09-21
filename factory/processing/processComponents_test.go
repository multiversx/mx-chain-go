package processing_test

import (
	"strings"
	"sync"
	"testing"

	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	processComp "github.com/ElrondNetwork/elrond-go/factory/processing"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/ElrondNetwork/elrond-go/testscommon/mainFactoryMocks"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ------------ Test TestProcessComponents --------------------
func TestProcessComponents_CloseShouldWork(t *testing.T) {
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

func TestProcessComponents_IndexGenesisBlocks(t *testing.T) {
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

	outportHandler := &testscommon.OutportStub{
		HasDriversCalled: func() bool {
			return true
		},
		SaveBlockCalled: func(args *indexer.ArgsSaveBlockData) {
			saveBlockCalledMutex.Lock()
			require.NotNil(t, args)

			bodyRequired := &dataBlock.Body{
				MiniBlocks: make([]*block.MiniBlock, 4),
			}

			txsPoolRequired := &indexer.Pool{}

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
