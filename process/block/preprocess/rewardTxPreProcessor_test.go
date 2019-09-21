package preprocess

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewRewardTxPreprocessor_NilRewardTxDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	rtp, err := NewRewardTxPreprocessor(
		nil,
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
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
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
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
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
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
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
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
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilRewardsTxProcessor, err)
}

func TestNewRewardTxPreprocessor_NilRewardProducerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		nil,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilInternalTransactionProducer, err)
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
		&mock.IntermediateTransactionHandlerMock{},
		nil,
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewRewardTxPreprocessor_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
		func(shardID uint32, txHashes [][]byte) {},
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
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		nil,
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilRequestHandler, err)
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
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
	)
	assert.Nil(t, err)
	assert.NotNil(t, rtp)
}

func TestRewardTxPreprocessor_AddComputedRewardMiniBlocksShouldAddMiniBlock(t *testing.T) {
	t.Parallel()

	txHash := "tx1_hash"

	tdp := initDataPool()

	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
	)

	assert.NotNil(t, rtp)

	txHashes := [][]byte{[]byte(txHash)}

	var rewardMiniBlocks block.MiniBlockSlice
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            0,
	}
	rewardMiniBlocks = append(rewardMiniBlocks, &mb1)

	rtp.AddComputedRewardMiniBlocks(rewardMiniBlocks)

	res := rtp.GetAllCurrentUsedTxs()

	if _, ok := res[txHash]; !ok {
		assert.Fail(t, "miniblock was not added")
	}
}

func TestRewardTxPreprocessor_CreateMarshalizedDataShouldWork(t *testing.T) {
	t.Parallel()

	txHash := "tx1_hash"
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
	)

	txHashes := [][]byte{[]byte(txHash)}
	var rewardMiniBlocks block.MiniBlockSlice
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}

	rewardMiniBlocks = append(rewardMiniBlocks, &mb1)
	rtp.AddComputedRewardMiniBlocks(rewardMiniBlocks)

	res, err := rtp.CreateMarshalizedData(txHashes)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
}

func TestRewardTxPreprocessor_ProcessMiniBlockInvalidMiniBlockTypeShouldErr(t *testing.T) {
	t.Parallel()

	txHash := "tx1_hash"
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
	)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            0,
	}

	err := rtp.ProcessMiniBlock(&mb1, haveTimeTrue, 0)
	assert.Equal(t, process.ErrWrongTypeInMiniBlock, err)
}

func TestRewardTxPreprocessor_ProcessMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	txHash := "tx1_hash"
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
	)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}

	err := rtp.ProcessMiniBlock(&mb1, haveTimeTrue, 0)
	assert.Nil(t, err)

	txsMap := rtp.GetAllCurrentUsedTxs()
	if _, ok := txsMap[txHash]; !ok {
		assert.Fail(t, "miniblock was not added")
	}
}

func TestRewardTxPreprocessor_SaveTxBlockToStorageShouldWork(t *testing.T) {
	t.Parallel()

	txHash := "tx1_hash"
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
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

	var rewardMiniBlocks block.MiniBlockSlice
	rewardMiniBlocks = append(rewardMiniBlocks, &mb1)

	rtp.AddComputedRewardMiniBlocks(rewardMiniBlocks)

	var blockBody block.Body
	blockBody = append(blockBody, &mb1, &mb2)
	err := rtp.SaveTxBlockToStorage(blockBody)

	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_RequestBlockTransactionsNoMissingTxsShouldWork(t *testing.T) {
	t.Parallel()

	txHash := "tx1_hash"
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
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

	var rewardMiniBlocks block.MiniBlockSlice
	rewardMiniBlocks = append(rewardMiniBlocks, &mb1)

	rtp.AddComputedRewardMiniBlocks(rewardMiniBlocks)

	var blockBody block.Body
	blockBody = append(blockBody, &mb1, &mb2)

	_ = rtp.SaveTxBlockToStorage(blockBody)

	res := rtp.RequestBlockTransactions(blockBody)
	assert.Equal(t, 0, res)
}

func TestRewardTxPreprocessor_RequestTransactionsForMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	txHash := "tx1_hash"
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
	)

	txHashes := [][]byte{[]byte(txHash)}
	mb1 := block.MiniBlock{
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

	txHash := "tx1_hash"
	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
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

	var rewardMiniBlocks block.MiniBlockSlice
	rewardMiniBlocks = append(rewardMiniBlocks, &mb1)

	rtp.AddComputedRewardMiniBlocks(rewardMiniBlocks)

	var blockBody block.Body
	blockBody = append(blockBody, &mb1, &mb2)

	err := rtp.ProcessBlockTransactions(blockBody, 0, haveTimeTrue)
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
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
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
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
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
	}
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&storer,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
	)

	txHashes := [][]byte{[]byte("tx_hash1")}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.RewardsBlock,
	}

	var blockBody block.Body
	blockBody = append(blockBody, &mb1)
	miniBlockPool := mock.NewCacherMock()

	numRestoredTxs, resMap, err := rtp.RestoreTxBlockIntoPools(blockBody, miniBlockPool)
	assert.Equal(t, 1, numRestoredTxs)
	assert.NotNil(t, resMap)
	assert.Nil(t, err)
}

func TestRewardTxPreprocessor_CreateAndProcessMiniBlocksTxForMiniBlockNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{
			CreateAllInterMiniBlocksCalled: func() map[uint32]*block.MiniBlock {
				txHashes := [][]byte{[]byte("hash_unavailable")}
				mb1 := block.MiniBlock{
					TxHashes:        txHashes,
					ReceiverShardID: 1,
					SenderShardID:   0,
					Type:            block.RewardsBlock,
				}

				return map[uint32]*block.MiniBlock{
					0: &mb1,
				}
			},
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
	)

	mBlocksSlice, err := rtp.CreateAndProcessMiniBlocks(1, 1, 0, haveTimeTrue)
	assert.Nil(t, mBlocksSlice)
	assert.Equal(t, process.ErrNilRewardTransaction, err)
}

func TestRewardTxPreprocessor_CreateAndProcessMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, _ := NewRewardTxPreprocessor(
		tdp.RewardTransactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{
			CreateAllInterMiniBlocksCalled: func() map[uint32]*block.MiniBlock {
				txHashes := [][]byte{[]byte("tx1_hash")}
				mb1 := block.MiniBlock{
					TxHashes:        txHashes,
					ReceiverShardID: 1,
					SenderShardID:   0,
					Type:            block.RewardsBlock,
				}

				return map[uint32]*block.MiniBlock{
					0: &mb1,
				}
			},
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
	)

	mBlocksSlice, err := rtp.CreateAndProcessMiniBlocks(1, 1, 0, haveTimeTrue)
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
		&mock.IntermediateTransactionHandlerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		func(shardID uint32, txHashes [][]byte) {},
	)

	rtp.CreateBlockStarted()
	assert.Equal(t, 0, len(rtp.rewardTxsForBlock.txHashAndInfo))
}
