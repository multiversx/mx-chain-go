package preprocess

import (
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewMiniBlocksCompaction_NilEconomicsFeeShouldErr(t *testing.T) {
	t.Parallel()

	msc := mock.NewMultiShardsCoordinatorMock(3)
	mbc, err := NewMiniBlocksCompaction(
		nil,
		msc,
		&mock.GasHandlerMock{},
	)

	assert.Nil(t, mbc)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewMiniBlocksCompaction_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	mbc, err := NewMiniBlocksCompaction(
		feeHandlerMock(),
		nil,
		&mock.GasHandlerMock{},
	)

	assert.Nil(t, mbc)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewMiniBlocksCompaction_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	msc := mock.NewMultiShardsCoordinatorMock(3)
	mbc, err := NewMiniBlocksCompaction(
		feeHandlerMock(),
		msc,
		&mock.GasHandlerMock{},
	)

	assert.NotNil(t, mbc)
	assert.Nil(t, err)
	assert.False(t, mbc.IsInterfaceNil())
}

func TestMiniBlocksCompaction_CompactSingleMiniblockWontCompact(t *testing.T) {
	t.Parallel()

	msc := mock.NewMultiShardsCoordinatorMock(3)
	mbc, _ := NewMiniBlocksCompaction(
		feeHandlerMock(),
		msc,
		&mock.GasHandlerMock{},
	)

	mbs := block.MiniBlockSlice{&block.MiniBlock{
		ReceiverShardID: 1,
		SenderShardID:   2,
	}}
	mapHashesAndTxs := make(map[string]data.TransactionHandler)
	newMbs := mbc.Compact(mbs, mapHashesAndTxs)

	assert.Equal(t, mbs, newMbs)
}

func TestMiniBlocksCompaction_CompactSingleDifferentReceiversWontCompact(t *testing.T) {
	t.Parallel()

	msc := mock.NewMultiShardsCoordinatorMock(3)
	mbc, _ := NewMiniBlocksCompaction(
		feeHandlerMock(),
		msc,
		&mock.GasHandlerMock{
			ComputeGasConsumedByMiniBlockCalled: func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
		},
	)

	mbs := block.MiniBlockSlice{
		&block.MiniBlock{
			ReceiverShardID: 1,
			SenderShardID:   2,
		},
		&block.MiniBlock{
			ReceiverShardID: 2,
			SenderShardID:   2,
		},
	}
	mapHashesAndTxs := make(map[string]data.TransactionHandler)
	newMbs := mbc.Compact(mbs, mapHashesAndTxs)

	assert.Equal(t, mbs, newMbs)
}

func TestMiniBlocksCompaction_CompactSingleDifferentSenderWontCompact(t *testing.T) {
	t.Parallel()

	msc := mock.NewMultiShardsCoordinatorMock(3)
	mbc, _ := NewMiniBlocksCompaction(
		feeHandlerMock(),
		msc,
		&mock.GasHandlerMock{
			ComputeGasConsumedByMiniBlockCalled: func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
		},
	)

	mbs := block.MiniBlockSlice{
		&block.MiniBlock{
			ReceiverShardID: 1,
			SenderShardID:   2,
		},
		&block.MiniBlock{
			ReceiverShardID: 1,
			SenderShardID:   1,
		},
	}
	mapHashesAndTxs := make(map[string]data.TransactionHandler)
	newMbs := mbc.Compact(mbs, mapHashesAndTxs)

	assert.Equal(t, mbs, newMbs)
}

func TestMiniBlocksCompaction_CompactSingleDifferentMbTypesWontCompact(t *testing.T) {
	t.Parallel()

	msc := mock.NewMultiShardsCoordinatorMock(3)
	mbc, _ := NewMiniBlocksCompaction(
		feeHandlerMock(),
		msc,
		&mock.GasHandlerMock{
			ComputeGasConsumedByMiniBlockCalled: func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
		},
	)

	mbs := block.MiniBlockSlice{
		&block.MiniBlock{
			Type: 2,
		},
		&block.MiniBlock{
			Type: 1,
		},
	}
	mapHashesAndTxs := make(map[string]data.TransactionHandler)
	newMbs := mbc.Compact(mbs, mapHashesAndTxs)

	assert.Equal(t, mbs, newMbs)
}

func TestMiniBlocksCompaction_CompactConditionsMetShouldCompact(t *testing.T) {
	t.Parallel()

	msc := mock.NewMultiShardsCoordinatorMock(3)
	mbc, _ := NewMiniBlocksCompaction(
		feeHandlerMock(),
		msc,
		&mock.GasHandlerMock{
			ComputeGasConsumedByMiniBlockCalled: func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
		},
	)

	mbs := block.MiniBlockSlice{
		&block.MiniBlock{
			SenderShardID:   1,
			ReceiverShardID: 2,
			Type:            1,
		},
		&block.MiniBlock{
			SenderShardID:   1,
			ReceiverShardID: 2,
			Type:            1,
		},
	}
	mapHashesAndTxs := make(map[string]data.TransactionHandler)
	newMbs := mbc.Compact(mbs, mapHashesAndTxs)

	assert.NotEqual(t, mbs, newMbs)
	assert.Equal(t, 1, len(newMbs))
}

func TestMiniBlocksCompaction_ExpandGasLimitExceededShouldNotCompact(t *testing.T) {
	t.Parallel()

	msc := mock.NewMultiShardsCoordinatorMock(3)
	mbc, _ := NewMiniBlocksCompaction(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit()
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
		},
		msc,
		&mock.GasHandlerMock{
			ComputeGasConsumedByMiniBlockCalled: func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
				return MaxGasLimitPerBlock, MaxGasLimitPerBlock, nil
			},
			ComputeGasConsumedByTxCalled: func(txSndShId uint32, txRcvShId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return MaxGasLimitPerBlock / 2, MaxGasLimitPerBlock / 2, nil
			},
		},
	)

	mbs := block.MiniBlockSlice{
		&block.MiniBlock{
			TxHashes:        [][]byte{[]byte("tx0"), []byte("tx1")},
			SenderShardID:   1,
			ReceiverShardID: 2,
			Type:            1,
		},
		&block.MiniBlock{
			TxHashes:        [][]byte{[]byte("tx2")},
			SenderShardID:   1,
			ReceiverShardID: 2,
			Type:            1,
		},
	}
	mapHashesAndTxs := make(map[string]data.TransactionHandler)
	mapHashesAndTxs["tx0"] = &transaction.Transaction{
		Nonce:    0,
		GasLimit: MaxGasLimitPerBlock / 2,
	}
	mapHashesAndTxs["tx1"] = &transaction.Transaction{
		Nonce:    1,
		GasLimit: MaxGasLimitPerBlock / 2,
	}
	mapHashesAndTxs["tx2"] = &transaction.Transaction{
		Nonce:    2,
		GasLimit: 1,
	}
	newMbs := mbc.Compact(mbs, mapHashesAndTxs)

	// all mbs should not be compacted in the same miniblock as the total gas limit of them exceeds max gas limit per
	// miniblock (100k). Therefore, the result miniblocks must be the same (no compaction)
	assert.Equal(t, mbs, newMbs)
	assert.NotEqual(t, 1, len(newMbs))
}

func TestMiniBlocksCompaction_ExpandConditionsMetShouldExpand(t *testing.T) {
	t.Parallel()

	msc := mock.NewMultiShardsCoordinatorMock(3)
	mbc, _ := NewMiniBlocksCompaction(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit()
			},
		},
		msc,
		&mock.GasHandlerMock{},
	)

	mbs := block.MiniBlockSlice{
		&block.MiniBlock{
			TxHashes:        [][]byte{[]byte("tx0"), []byte("tx1"), []byte("tx2")},
			SenderShardID:   0,
			ReceiverShardID: 2,
			Type:            1,
		},
		&block.MiniBlock{
			TxHashes:        [][]byte{[]byte("tx4")},
			SenderShardID:   0,
			ReceiverShardID: 1,
			Type:            1,
		},
	}
	mapHashesAndTxs := make(map[string]data.TransactionHandler)
	mapHashesAndTxs["tx0"] = &transaction.Transaction{
		SndAddr:  []byte("sndr1"),
		Nonce:    0,
		GasLimit: 15000,
	}
	mapHashesAndTxs["tx1"] = &transaction.Transaction{
		SndAddr:  []byte("sndr1"),
		Nonce:    1,
		GasLimit: 6000,
	}
	mapHashesAndTxs["tx2"] = &transaction.Transaction{
		SndAddr:  []byte("sndr1"),
		Nonce:    3,
		GasLimit: 8000,
	}
	mapHashesAndTxs["tx4"] = &transaction.Transaction{
		SndAddr:  []byte("sndr1"),
		Nonce:    2,
		GasLimit: 1000,
	}
	newMbs, err := mbc.Expand(mbs, mapHashesAndTxs)
	assert.Nil(t, err)

	// we should get 3 miniblocks as there are 2 continuous txs (with consecutive nonces) then a tx for other shard
	// and also a tx for the same shard as the first 2
	assert.Equal(t, 3, len(newMbs))
}

func TestMiniBlocksCompaction_ExpandConditionsNotMetShouldNotExpand(t *testing.T) {
	t.Parallel()

	msc := mock.NewMultiShardsCoordinatorMock(3)
	mbc, _ := NewMiniBlocksCompaction(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit()
			},
		},
		msc,
		&mock.GasHandlerMock{},
	)

	mbs := block.MiniBlockSlice{
		&block.MiniBlock{
			TxHashes:        [][]byte{[]byte("tx0"), []byte("tx1")},
			SenderShardID:   0,
			ReceiverShardID: 2,
			Type:            block.TxBlock,
		},
		&block.MiniBlock{
			TxHashes:        [][]byte{[]byte("tx2")},
			SenderShardID:   0,
			ReceiverShardID: 1,
			Type:            block.TxBlock,
		},
	}
	mapHashesAndTxs := make(map[string]data.TransactionHandler)
	mapHashesAndTxs["tx0"] = &transaction.Transaction{
		SndAddr:  []byte("sndr1"),
		Nonce:    0,
		GasLimit: 15000,
	}
	mapHashesAndTxs["tx1"] = &transaction.Transaction{
		SndAddr:  []byte("sndr1"),
		Nonce:    1,
		GasLimit: 6000,
	}
	mapHashesAndTxs["tx2"] = &transaction.Transaction{
		SndAddr:  []byte("sndr2"),
		Nonce:    3,
		GasLimit: 8000,
	}
	newMbs, err := mbc.Expand(mbs, mapHashesAndTxs)
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(mbs, newMbs))
}
