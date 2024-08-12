package preprocess

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process/factory/containers"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/common"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func createArgsRewardsPreProc() ArgsRewardTxPreProcessor {
	return ArgsRewardTxPreProcessor{
		initDataPool().Transactions(),
		&storageStubs.ChainStorerStub{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&testscommon.GasHandlerStub{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&common.TxExecutionOrderHandlerStub{},
	}
}

func TestRewardsTxPreProcFactory_CreateRewardsTxPreProcessor(t *testing.T) {
	t.Parallel()

	f := NewRewardsTxPreProcFactory()
	require.False(t, f.IsInterfaceNil())

	args := createArgsRewardsPreProc()
	container := containers.NewPreProcessorsContainer()
	err := f.CreateRewardsTxPreProcessorAndAddToContainer(args, container)
	require.Nil(t, err)

	preProc, err := container.Get(block.RewardsBlock)
	require.Nil(t, err)
	require.Equal(t, fmt.Sprintf("%T", preProc), "*preprocess.rewardTxPreprocessor")
}
