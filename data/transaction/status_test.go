package transaction

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericmocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStatusComputer_ComputeStatusWhenInStorageKnowingMiniblock(t *testing.T) {
	statusComputer := &StatusComputer{
		SelfShard:12,
	}

	// Invalid miniblock
	tx := &ApiTransactionResult{
		SourceShard: 12,
		DestinationShard: 12,
		Tx: &Transaction{},
	}
	require.Equal(t, TxStatusInvalid, statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.InvalidBlock,tx))

	// Intra-shard
	require.Equal(t, TxStatusSuccess, statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.TxBlock,tx))


	// Cross, at source
	tx.DestinationShard = 13
	require.Equal(t, TxStatusPending, statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.TxBlock,tx))

	// Cross, at source, but knowing that it has been fully notarized (through DatabaseLookupExtensions)
	tx.DestinationShard = 13
	tx.SourceShard = 12
	tx.NotarizedAtDestinationInMetaNonce = 1
	require.Equal(t, TxStatusSuccess, statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.TxBlock,tx))


	// Cross, destination me
	tx.DestinationShard = 12
	tx.SourceShard = 13
	tx.NotarizedAtDestinationInMetaNonce = 0
	require.Equal(t, TxStatusSuccess, statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.TxBlock,tx))

	tx.SourceShard = core.MetachainShardId
	tx.DestinationShard = 12
	tx.NotarizedAtDestinationInMetaNonce = 0
	require.Equal(t, TxStatusSuccess, statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.RewardsBlock,tx))

	// Contract deploy
	statusComputer.SelfShard = 13
	tx.SourceShard = 12
	tx.DestinationShard = 13
	tx.Tx.SetRcvAddr(make([]byte, 32))
	tx.Data = []byte("deployingAContract")
	require.Equal(t, TxStatusSuccess, statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.TxBlock,tx))
}

func TestStatusComputer_ComputeStatusWhenInStorageNotKnowingMiniblock(t *testing.T) {
	statusComputer := &StatusComputer{
		SelfShard:12,
	}

	// Invalid miniblock
	tx := &ApiTransactionResult{
		SourceShard: 12,
		DestinationShard: 12,
		Tx: &Transaction{},
	}

	// Intra shard
	tx.DestinationShard = 12
	tx.SourceShard =12
	require.Equal(t, TxStatusSuccess, statusComputer.ComputeStatusWhenInStorageNotKnowingMiniblock(tx.DestinationShard,tx))

	// Cross, at source
	tx.SourceShard = 12
	tx.DestinationShard = 13
	require.Equal(t, TxStatusPending, statusComputer.ComputeStatusWhenInStorageNotKnowingMiniblock(tx.DestinationShard,tx))

	// Cross, destination me
	tx.SourceShard = 13
	tx.DestinationShard = 12
	require.Equal(t, TxStatusSuccess, statusComputer.ComputeStatusWhenInStorageNotKnowingMiniblock(tx.DestinationShard,tx))

	tx.SourceShard = core.MetachainShardId
	tx.DestinationShard = 12
	require.Equal(t, TxStatusSuccess, statusComputer.ComputeStatusWhenInStorageNotKnowingMiniblock(tx.DestinationShard,tx))

	// Contract deploy
	statusComputer.SelfShard = 13
	tx.SourceShard = 12
	tx.DestinationShard = 13
	tx.Tx.SetRcvAddr(make([]byte, 32))
	tx.Data = []byte("deployingAContract")

	require.Equal(t, TxStatusSuccess, statusComputer.ComputeStatusWhenInStorageNotKnowingMiniblock(tx.DestinationShard,tx))
}


func TestStatusComputer_SetStatusIfIsRewardReverted(t *testing.T) {
	t.Parallel()

	chainStorer := genericmocks.NewChainStorerMock(0)
	uint64Converter := mock.NewNonceHashConverterMock()

	// not reward transaction should not set status
	txA := &ApiTransactionResult{Status: TxStatusSuccess}

	isRewardReverted := (&StatusComputer{
		//MiniblockType: block.TxBlock,
	}).SetStatusIfIsRewardReverted(txA,block.TxBlock,0,nil,nil,nil)
	require.False(t, isRewardReverted)
	require.Equal(t, TxStatusSuccess, txA.Status)

	// shard - meta : reward transaction
	txB := &ApiTransactionResult{Status: TxStatusSuccess}

	headerHash := []byte("hash-1")
	headerNonce := uint64(10)
	nonceBytes := uint64Converter.ToByteSlice(headerNonce)
	_ = chainStorer.HdrNonce.Put(nonceBytes, headerHash)

	isRewardReverted = (&StatusComputer{
		SelfShard:	core.MetachainShardId,

	}).SetStatusIfIsRewardReverted(txB,block.RewardsBlock,headerNonce,headerHash,uint64Converter,chainStorer)
	require.False(t, isRewardReverted)
	require.Equal(t, TxStatusSuccess, txB.Status)

	// shard 0: reward transaction
	txC := &ApiTransactionResult{Status: TxStatusSuccess}

	headerHash = []byte("hash-2")
	headerNonce = uint64(12)
	nonceBytes = uint64Converter.ToByteSlice(headerNonce)
	_ = chainStorer.HdrNonce.Put(nonceBytes, headerHash)

	isRewardReverted = (&StatusComputer{
		SelfShard:                0,
	}).SetStatusIfIsRewardReverted(txC,block.RewardsBlock,headerNonce,headerHash,uint64Converter,chainStorer)
	require.False(t, isRewardReverted)
	require.Equal(t, TxStatusSuccess, txC.Status)

	// shard meta: reverted reward transaction
	txD := &ApiTransactionResult{Status: TxStatusSuccess}

	headerHash = []byte("hash-3")
	headerNonce = uint64(12)
	nonceBytes = uint64Converter.ToByteSlice(headerNonce)
	_ = chainStorer.HdrNonce.Put(nonceBytes, headerHash)

	wrongHash := []byte("wrong")
	isRewardReverted = (&StatusComputer{
		SelfShard:                core.MetachainShardId,
	}).SetStatusIfIsRewardReverted(txD,block.RewardsBlock,headerNonce,wrongHash,uint64Converter,chainStorer)
	require.True(t, isRewardReverted)
	require.Equal(t, TxStatusRewardReverted, txD.Status)
}
