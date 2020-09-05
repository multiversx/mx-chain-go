package preprocess

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

const testTxHash = "tx1_hash"

func TestNewRewardTxPreprocessor_NilRewardTxDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	rtp, err := NewRewardTxPreprocessor(
		nil,
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilRewardTxDataPool, err)
}

func TestNewRewardTxPreprocessor_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.Transactions(),
		nil,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewRewardTxPreprocessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		nil,
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewRewardTxPreprocessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		nil,
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewRewardTxPreprocessor_NilRewardTxProcessorShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		nil,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilRewardsTxProcessor, err)
}

func TestNewRewardTxPreprocessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		nil,
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewRewardTxPreprocessor_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewRewardTxPreprocessor_NilRequestHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		nil,
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilRequestHandler, err)
}

func TestNewRewardTxPreprocessor_NilGasHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		nil,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilGasHandler, err)
}

func TestNewRewardTxPreprocessor_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		nil,
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewRewardTxPreprocessor_NilBlockSizeComputationHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		nil,
		&mock.BalanceComputationStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
}

func TestNewRewardTxPreprocessor_NilBalanceComputationHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		nil,
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilBalanceComputationHandler, err)
}

func TestNewRewardTxPreprocessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, rtp)
}

func TestRewardTxPreprocessor_CreateMarshalizedDataShouldWork(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	txHashes := [][]byte{[]byte(txHash)}
	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	res, err := rtp.CreateMarshalizedData(txHashes)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
}

func TestRewardTxPreprocessor_ProcessMiniBlockInvalidMiniBlockTypeShouldErr(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            0,
	}

	_, err := rtp.ProcessMiniBlock(&mb1, haveTimeTrue, getNumOfCrossInterMbsAndTxsZero)
	assert.Equal(t, process.ErrWrongTypeInMiniBlock, err)
}

func TestRewardTxPreprocessor_ProcessMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   core.MetachainShardId,
		Type:            block.RewardsBlock,
	}

	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	_, err := rtp.ProcessMiniBlock(&mb1, haveTimeTrue, getNumOfCrossInterMbsAndTxsZero)
	assert.Nil(t, err)

	txsMap := rtp.GetAllCurrentUsedTxs()
	if _, ok := txsMap[txHash]; !ok {
		assert.Fail(t, "miniblock was not added")
	}
}

func TestRewardTxPreprocessor_ProcessMiniBlockNotFromMeta(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}

	txs := []data.TransactionHandler{&rewardTx.RewardTx{}}
	rtp.AddTxs(txHashes, txs)

	_, err := rtp.ProcessMiniBlock(&mb1, haveTimeTrue, getNumOfCrossInterMbsAndTxsZero)
	assert.Equal(t, process.ErrRewardMiniBlockNotFromMeta, err)
}

func TestRewardTxPreprocessor_SaveTxBlockToStorageShouldWork(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

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
	err := rtp.SaveTxBlockToStorage(blockBody)

	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_RequestBlockTransactionsNoMissingTxsShouldWork(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

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

	_ = rtp.SaveTxBlockToStorage(blockBody)

	res := rtp.RequestBlockTransactions(blockBody)
	assert.Equal(t, 0, res)
}

func TestRewardTxPreprocessor_RequestTransactionsForMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	txHash := testTxHash
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

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
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

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

	var blockBody block.Body
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1, &mb2)

	err := rtp.ProcessBlockTransactions(&blockBody, haveTimeTrue)
	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_IsDataPreparedShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	err := rtp.IsDataPrepared(1, haveTime)

	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestRewardTxPreprocessor_IsDataPrepared(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	go func() {
		time.Sleep(50 * time.Millisecond)
		rtp.chReceivedAllRewardTxs <- true
	}()

	err := rtp.IsDataPrepared(1, haveTime)

	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_RestoreTxBlockIntoPools(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	storer := mock.ChainStorerMock{
		GetAllCalled: func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
			retMap := map[string][]byte{
				"tx_hash1": []byte(`{"Round": 0}`),
			}

			return retMap, nil
		},
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				RemoveCalled: func(key []byte) error {
					return nil
				},
			}
		},
	}
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storer,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	txHashes := [][]byte{[]byte("tx_hash1")}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}

	blockBody := &block.Body{}
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1)
	miniBlockPool := testscommon.NewCacherMock()

	numRestoredTxs, err := rtp.RestoreTxBlockIntoPools(blockBody, miniBlockPool)
	assert.Equal(t, 1, numRestoredTxs)
	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_CreateAndProcessMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	totalGasConsumed := uint64(0)
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{
			InitCalled: func() {
				totalGasConsumed = 0
			},
			TotalGasConsumedCalled: func() uint64 {
				return totalGasConsumed
			},
		},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	mBlocksSlice, err := rtp.CreateAndProcessMiniBlocks(haveTimeTrue)
	assert.NotNil(t, mBlocksSlice)
	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_CreateBlockStartedShouldCleanMap(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
		&mock.GasHandlerMock{},
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
	)

	rtp.CreateBlockStarted()
	assert.Equal(t, 0, len(rtp.rewardTxsForBlock.txHashAndInfo))
}
