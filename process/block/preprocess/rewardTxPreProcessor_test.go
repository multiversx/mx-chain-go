package preprocess

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	"github.com/multiversx/mx-chain-go/testscommon/common"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

const testTxHash = "tx1_hash"

func TestNewRewardTxPreprocessor_NilRewardTxDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.DataPool = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

func TestNewRewardTxPreprocessor_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.Store = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewRewardTxPreprocessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.Hasher = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewRewardTxPreprocessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.Marshalizer = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewRewardTxPreprocessor_NilRewardTxProcessorShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.RewardProcessor = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilRewardsTxProcessor, err)
}

func TestNewRewardTxPreprocessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.ShardCoordinator = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewRewardTxPreprocessor_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.Accounts = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewRewardTxPreprocessor_NilAccountsProposalShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.AccountsProposal = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.ErrorIs(t, err, process.ErrNilAccountsAdapter)
}

func TestNewRewardTxPreprocessor_NilRequestHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.OnRequestTransaction = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilRequestHandler, err)
}

func TestNewRewardTxPreprocessor_NilGasHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.GasHandler = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilGasHandler, err)
}

func TestNewRewardTxPreprocessor_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.PubkeyConverter = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewRewardTxPreprocessor_NilBlockSizeComputationHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.BlockSizeComputation = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
}

func TestNewRewardTxPreprocessor_NilBalanceComputationHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.BalanceComputation = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilBalanceComputationHandler, err)
}

func TestNewRewardTxPreprocessor_NilProcessedMiniBlocksTrackerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.ProcessedMiniBlocksTracker = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilProcessedMiniBlocksTracker, err)
}

func TestNewRewardTxPreprocessor_NilTxExecutionOrderHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.TxExecutionOrderHandler = nil
	rtp, err := NewRewardTxPreprocessor(args)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilTxExecutionOrderHandler, err)
}

func TestNewRewardTxPreprocessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	rtp, err := NewRewardTxPreprocessor(args)
	assert.Nil(t, err)
	assert.NotNil(t, rtp)
}

func TestRewardTxPreprocessor_CreateMarshalizedDataShouldWork(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	rtp, _ := NewRewardTxPreprocessor(args)

	txHashes := [][]byte{[]byte(txHash)}
	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	res, err := rtp.CreateMarshalledData(txHashes)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
}

func TestRewardTxPreprocessor_ProcessMiniBlockInvalidMiniBlockTypeShouldErr(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	rtp, _ := NewRewardTxPreprocessor(args)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            0,
	}

	preProcessorExecutionInfoHandlerMock := &testscommon.PreProcessorExecutionInfoHandlerMock{
		GetNumOfCrossInterMbsAndTxsCalled: getNumOfCrossInterMbsAndTxsZero,
	}

	_, _, _, err := rtp.ProcessMiniBlock(&mb1, haveTimeTrue, haveAdditionalTimeFalse, false, false, -1, preProcessorExecutionInfoHandlerMock)
	assert.Equal(t, process.ErrWrongTypeInMiniBlock, err)
}

func TestRewardTxPreprocessor_ProcessMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	calledCount := 0
	txHash := testTxHash
	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.TxExecutionOrderHandler = &common.TxExecutionOrderHandlerStub{
		AddCalled: func(txHash []byte) {
			calledCount++
		},
	}
	rtp, _ := NewRewardTxPreprocessor(args)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   core.MetachainShardId,
		Type:            block.RewardsBlock,
	}

	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	preProcessorExecutionInfoHandlerMock := &testscommon.PreProcessorExecutionInfoHandlerMock{
		GetNumOfCrossInterMbsAndTxsCalled: getNumOfCrossInterMbsAndTxsZero,
	}

	_, _, _, err := rtp.ProcessMiniBlock(&mb1, haveTimeTrue, haveAdditionalTimeFalse, false, false, -1, preProcessorExecutionInfoHandlerMock)
	assert.Nil(t, err)
	assert.Equal(t, 1, calledCount)

	txsMap := rtp.GetAllCurrentUsedTxs()
	if _, ok := txsMap[txHash]; !ok {
		assert.Fail(t, "miniblock was not added")
	}
}

func TestRewardTxPreprocessor_ProcessMiniBlockNotFromMeta(t *testing.T) {
	t.Parallel()

	calledCount := 0
	txHash := testTxHash
	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.TxExecutionOrderHandler = &common.TxExecutionOrderHandlerStub{
		AddCalled: func(txHash []byte) {
			calledCount++
		},
	}
	rtp, _ := NewRewardTxPreprocessor(args)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}

	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	preProcessorExecutionInfoHandlerMock := &testscommon.PreProcessorExecutionInfoHandlerMock{
		GetNumOfCrossInterMbsAndTxsCalled: getNumOfCrossInterMbsAndTxsZero,
	}

	_, _, _, err := rtp.ProcessMiniBlock(&mb1, haveTimeTrue, haveAdditionalTimeFalse, false, false, -1, preProcessorExecutionInfoHandlerMock)
	assert.Equal(t, 0, calledCount)
	assert.Equal(t, process.ErrRewardMiniBlockNotFromMeta, err)
}

func TestRewardTxPreprocessor_SaveTxsToStorageShouldWork(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	rtp, _ := NewRewardTxPreprocessor(args)

	txHashes := [][]byte{[]byte(txHash)}
	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}
	mb2 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 0,
		SenderShardID:   1,
		Type:            block.RewardsBlock,
	}

	blockBody := &block.Body{}
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1, &mb2)
	err := rtp.SaveTxsToStorage(blockBody)

	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_RequestBlockTransactionsNoMissingTxsShouldWork(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	rtp, _ := NewRewardTxPreprocessor(args)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}
	mb2 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 0,
		SenderShardID:   1,
		Type:            block.RewardsBlock,
	}

	blockBody := &block.Body{}
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1, &mb2)

	_ = rtp.SaveTxsToStorage(blockBody)

	res := rtp.RequestBlockTransactions(blockBody)
	assert.Equal(t, 0, res)
}

func TestRewardTxPreprocessor_RequestTransactionsForMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	rtp, _ := NewRewardTxPreprocessor(args)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := &block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}

	res := rtp.RequestTransactionsForMiniBlock(mb1)
	assert.Equal(t, 0, res)
}

func TestRewardTxPreprocessor_ProcessBlockTransactions(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	calledCount := 0
	args := createDefaultRewardsProcessorArgs(tdp)
	args.TxExecutionOrderHandler = &common.TxExecutionOrderHandlerStub{
		AddCalled: func(txHash []byte) {
			calledCount++
		},
	}
	rtp, _ := NewRewardTxPreprocessor(args)

	txHashes := [][]byte{[]byte(txHash)}
	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}
	mb2 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 0,
		SenderShardID:   1,
		Type:            block.RewardsBlock,
	}

	mbHash1, _ := core.CalculateHash(rtp.marshalizer, rtp.hasher, &mb1)
	mbHash2, _ := core.CalculateHash(rtp.marshalizer, rtp.hasher, &mb2)

	var blockBody block.Body
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1, &mb2)

	err := rtp.ProcessBlockTransactions(&block.Header{MiniBlockHeaders: []block.MiniBlockHeader{{TxCount: 1, Hash: mbHash1}, {TxCount: 1, Hash: mbHash2}}}, &blockBody, haveTimeTrue)
	assert.Equal(t, 2, calledCount)
	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_ProcessBlockTransactionsMissingTrieNode(t *testing.T) {
	t.Parallel()

	missingNodeErr := fmt.Errorf(core.GetNodeFromDBErrorString)
	txHash := testTxHash
	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.Accounts = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return nil, missingNodeErr
		},
	}
	rtp, _ := NewRewardTxPreprocessor(args)

	txHashes := [][]byte{[]byte(txHash)}
	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}
	mb2 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 0,
		SenderShardID:   1,
		Type:            block.RewardsBlock,
	}

	mbHash1, _ := core.CalculateHash(rtp.marshalizer, rtp.hasher, &mb1)
	mbHash2, _ := core.CalculateHash(rtp.marshalizer, rtp.hasher, &mb2)

	var blockBody block.Body
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1, &mb2)

	err := rtp.ProcessBlockTransactions(&block.Header{MiniBlockHeaders: []block.MiniBlockHeader{{TxCount: 1, Hash: mbHash1}, {TxCount: 1, Hash: mbHash2}}}, &blockBody, haveTimeTrue)
	assert.Equal(t, missingNodeErr, err)
}

func TestRewardTxPreprocessor_IsDataPreparedShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	rtp, _ := NewRewardTxPreprocessor(args)

	txHashesMissing := [][]byte{[]byte("missing_tx_hash")}

	rtp.SetMissingRewardTxs(len(txHashesMissing))
	err := rtp.IsDataPrepared(len(txHashesMissing), haveTime)

	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestRewardTxPreprocessor_IsDataPrepared(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	rtp, _ := NewRewardTxPreprocessor(args)

	go func() {
		time.Sleep(50 * time.Millisecond)
		rtp.SetMissingRewardTxs(0)
	}()

	err := rtp.IsDataPrepared(1, haveTime)

	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_RestoreBlockDataIntoPools(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	storer := storageStubs.ChainStorerStub{
		GetAllCalled: func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
			retMap := map[string][]byte{
				"tx_hash1": []byte(`{"Round": 0}`),
			}

			return retMap, nil
		},
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				RemoveCalled: func(key []byte) error {
					return nil
				},
			}, nil
		},
	}
	args := createDefaultRewardsProcessorArgs(tdp)
	args.Store = &storer
	rtp, _ := NewRewardTxPreprocessor(args)

	txHashes := [][]byte{[]byte("tx_hash1")}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}

	blockBody := &block.Body{}
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1)
	miniBlockPool := cache.NewCacherMock()

	numRestoredTxs, err := rtp.RestoreBlockDataIntoPools(blockBody, miniBlockPool)
	assert.Equal(t, 1, numRestoredTxs)
	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_CreateAndProcessMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	totalGasProvided := uint64(0)
	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	args.GasHandler = &testscommon.GasHandlerStub{
		InitCalled: func() {
			totalGasProvided = 0
		},
		TotalGasProvidedCalled: func() uint64 {
			return totalGasProvided
		},
	}
	rtp, _ := NewRewardTxPreprocessor(args)

	mBlocksSlice, err := rtp.CreateAndProcessMiniBlocks(haveTimeTrue, []byte("randomness"))
	assert.NotNil(t, mBlocksSlice)
	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_CreateBlockStartedShouldCleanMap(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	args := createDefaultRewardsProcessorArgs(tdp)
	rtp, _ := NewRewardTxPreprocessor(args)

	rtp.CreateBlockStarted()
	rewardsForBlock := rtp.rewardTxsForBlock.(*txsForBlock)
	assert.Equal(t, 0, len(rewardsForBlock.txHashAndInfo))
}

func createDefaultRewardsProcessorArgs(tdp dataRetriever.PoolsHolder) RewardsPreProcessorArgs {
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	return RewardsPreProcessorArgs{
		BasePreProcessorArgs: BasePreProcessorArgs{
			DataPool:                   tdp.RewardTransactions(),
			Store:                      &storageStubs.ChainStorerStub{},
			Hasher:                     &hashingMocks.HasherMock{},
			Marshalizer:                &mock.MarshalizerMock{},
			ShardCoordinator:           mock.NewMultiShardsCoordinatorMock(3),
			Accounts:                   &stateMock.AccountsStub{},
			AccountsProposal:           &stateMock.AccountsStub{},
			OnRequestTransaction:       requestTransaction,
			GasHandler:                 &testscommon.GasHandlerStub{},
			PubkeyConverter:            createMockPubkeyConverter(),
			BlockSizeComputation:       &testscommon.BlockSizeComputationStub{},
			BalanceComputation:         &testscommon.BalanceComputationStub{},
			ProcessedMiniBlocksTracker: &testscommon.ProcessedMiniBlocksTrackerStub{},
			TxExecutionOrderHandler:    &common.TxExecutionOrderHandlerStub{},
			EconomicsFee:               feeHandlerMock(),
			EnableEpochsHandler:        enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
		},
		RewardProcessor: &testscommon.RewardTxProcessorMock{},
	}
}
