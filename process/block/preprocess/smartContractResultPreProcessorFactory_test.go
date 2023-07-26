package preprocess_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewSmartContractResultPreProcessorFactory(t *testing.T) {
	fact, err := preprocess.NewSmartContractResultPreProcessorFactory()
	require.Nil(t, err)
	require.NotNil(t, fact)
}

func TestSmartContractResultPreProcessorFactory_CreateSmartContractResultPreProcessor(t *testing.T) {
	fact, _ := preprocess.NewSmartContractResultPreProcessorFactory()

	args := preprocess.SmartContractResultPreProcessorCreatorArgs{}
	preProcessor, err := fact.CreateSmartContractResultPreProcessor(args)
	require.NotNil(t, err)
	require.Nil(t, preProcessor)

	args = getDefaultSmartContractResultPreProcessorCreatorArgs()
	preProcessor, err = fact.CreateSmartContractResultPreProcessor(args)
	require.Nil(t, err)
	require.NotNil(t, preProcessor)
}

func TestSmartContractResultPreProcessorFactory_IsInterfaceNil(t *testing.T) {
	fact, _ := preprocess.NewSmartContractResultPreProcessorFactory()
	require.False(t, fact.IsInterfaceNil())
}

func getDefaultSmartContractResultPreProcessorCreatorArgs() preprocess.SmartContractResultPreProcessorCreatorArgs {
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	args := preprocess.SmartContractResultPreProcessorCreatorArgs{
		ScrDataPool: &testscommon.ShardedDataStub{},
		Store:       &storageStubs.ChainStorerStub{},
		Hasher:      &hashingMocks.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
		ScrProcessor: &testscommon.TxProcessorMock{
			ProcessSmartContractResultCalled: func(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		ShardCoordinator:             mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                     &stateMock.AccountsStub{},
		OnRequestSmartContractResult: requestTransaction,
		GasHandler:                   &testscommon.GasHandlerStub{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerStub{},
		PubkeyConverter:              testscommon.NewPubkeyConverterMock(32),
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	return args
}
