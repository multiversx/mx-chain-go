package preprocess

import (
	"math/big"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/stretchr/testify/require"
)

func createDefaultMiniBlockBuilderArgs() miniBlocksBuilderArgs {
	return miniBlocksBuilderArgs{
		gasTracker: gasTracker{
			shardCoordinator: &testscommon.ShardsCoordinatorMock{},
			economicsFee:     &economicsmocks.EconomicsHandlerMock{},
			gasHandler:       &testscommon.GasHandlerStub{},
		},
		accounts: &testscommon.AccountsStub{},
		accountTxsShards: &accountTxsShards{
			accountsInfo: make(map[string]*txShardInfo),
			RWMutex:      sync.RWMutex{},
		},
		blockSizeComputation:      &testscommon.BlockSizeComputationStub{},
		balanceComputationHandler: &testscommon.BalanceComputationStub{},
		haveTime:                  haveTimeTrue,
		isShardStuck:              isShardStuckFalse,
		isMaxBlockSizeReached:     isMaxBlockSizeReachedFalse,
		getTxMaxTotalCost: func(txHandler data.TransactionHandler) *big.Int {
			txMaxTotalCost, _ := big.NewInt(0).SetString("1500000000", 0)
			return txMaxTotalCost
		},
	}
}

func Test_newMiniBlockBuilderWithError(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()
	args.accounts = nil

	mbb, err := newMiniBlockBuilder(args)
	require.Equal(t, process.ErrNilAccountsAdapter, err)
	require.Nil(t, mbb)
}

func Test_newMiniBlockBuilderOK(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()

	mbb, err := newMiniBlockBuilder(args)
	require.Nil(t, err)
	require.NotNil(t, mbb)
}

func Test_checkMiniBlocksBuilderArgsNilShardCoordinatorShouldErr(t *testing.T) {

}

func Test_checkMiniBlocksBuilderArgsNilGasHandlerShouldErr(t *testing.T) {

}

func Test_checkMiniBlocksBuilderArgsNilBalanceComputationHandlerShouldErr(t *testing.T) {

}

func Test_checkMiniBlocksBuilderArgsNilHaveTimeHandlerShouldErr(t *testing.T) {

}

func Test_checkMiniBlocksBuilderArgsNilShardStuckHandlerShouldErr(t *testing.T) {

}

func Test_checkMiniBlocksBuilderArgsNilMaxBlockSizeReachedHandlerShouldErr(t *testing.T) {

}

func Test_checkMiniBlocksBuilderArgsOK(t *testing.T) {

}
