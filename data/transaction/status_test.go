package transaction

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericmocks"
	"github.com/stretchr/testify/require"
)

func TestStatusComputer_ComputeStatusWhenInStorageKnowingMiniblock(t *testing.T) {
	statusComputer := &StatusComputer{
		SelfShardId: 12,
	}

	// Invalid miniblock
	tx := &ApiTransactionResult{
		SourceShard: 12,
		DestinationShard: 12,
		Tx: &Transaction{},
	}
	responseStatus, err := statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.InvalidBlock,tx)
	require.Equal(t, TxStatusInvalid, responseStatus)
	require.Nil(t, err)

	// Intra-shard
	responseStatus, err = statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.TxBlock,tx)
	require.Equal(t, TxStatusSuccess, responseStatus)
	require.Nil(t, err)

	// Cross, at source
	tx.DestinationShard = 13
	responseStatus, err = statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.TxBlock,tx)
	require.Equal(t, TxStatusPending, responseStatus)
	require.Nil(t, err)

	// Cross, at source, but knowing that it has been fully notarized (through DatabaseLookupExtensions)
	tx.DestinationShard = 13
	tx.SourceShard = 12
	tx.NotarizedAtDestinationInMetaNonce = 1
	responseStatus, err = statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.TxBlock,tx)
	require.Equal(t, TxStatusSuccess, responseStatus)
	require.Nil(t, err)


	// Cross, destination me
	tx.DestinationShard = 12
	tx.SourceShard = 13
	tx.NotarizedAtDestinationInMetaNonce = 0
	responseStatus, err = statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.TxBlock,tx)
	require.Equal(t, TxStatusSuccess, responseStatus)
	require.Nil(t, err)

	tx.SourceShard = core.MetachainShardId
	tx.DestinationShard = 12
	tx.NotarizedAtDestinationInMetaNonce = 0
	responseStatus, err = statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.RewardsBlock,tx)
	require.Equal(t, TxStatusSuccess, responseStatus)
	require.Nil(t, err)

	// Contract deploy
	statusComputer.SelfShardId = 13
	tx.SourceShard = 12
	tx.DestinationShard = 13
	tx.Tx.SetRcvAddr(make([]byte, 32))
	tx.Data = []byte("deployingAContract")
	responseStatus, err = statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.TxBlock,tx)
	require.Equal(t, TxStatusSuccess, responseStatus)
	require.Nil(t, err)

	// Nil parameters
	responseStatus, err = statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.TxBlock,nil)
	require.Equal(t, ErrNilApiTransactionResult, err)
}

func TestStatusComputer_ComputeStatusWhenInStorageNotKnowingMiniblock(t *testing.T) {
	statusComputer := &StatusComputer{
		SelfShardId: 12,
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
	responseStatus, err := statusComputer.ComputeStatusWhenInStorageNotKnowingMiniblock(tx.DestinationShard,tx)
	require.Equal(t, TxStatusSuccess, responseStatus)
	require.Nil(t, err)

	// Cross, at source
	tx.SourceShard = 12
	tx.DestinationShard = 13
	responseStatus, err = statusComputer.ComputeStatusWhenInStorageNotKnowingMiniblock(tx.DestinationShard,tx)
	require.Equal(t, TxStatusPending, responseStatus)
	require.Nil(t, err)

	// Cross, destination me
	tx.SourceShard = 13
	tx.DestinationShard = 12
	responseStatus, err = statusComputer.ComputeStatusWhenInStorageNotKnowingMiniblock(tx.DestinationShard,tx)
	require.Equal(t, TxStatusSuccess, responseStatus)
	require.Nil(t, err)

	tx.SourceShard = core.MetachainShardId
	tx.DestinationShard = 12
	responseStatus, err = statusComputer.ComputeStatusWhenInStorageNotKnowingMiniblock(tx.DestinationShard,tx)
	require.Equal(t, TxStatusSuccess, responseStatus)
	require.Nil(t, err)

	// Contract deploy
	statusComputer.SelfShardId = 13
	tx.SourceShard = 12
	tx.DestinationShard = 13
	tx.Tx.SetRcvAddr(make([]byte, 32))
	tx.Data = []byte("deployingAContract")
	responseStatus, err = statusComputer.ComputeStatusWhenInStorageNotKnowingMiniblock(tx.DestinationShard,tx)
	require.Equal(t, TxStatusSuccess, responseStatus)
	require.Nil(t, err)

	// Nil parameters
	responseStatus, err = statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(block.TxBlock,nil)
	require.Equal(t, ErrNilApiTransactionResult, err)
}


func TestStatusComputer_SetStatusIfIsRewardReverted(t *testing.T) {
	t.Parallel()

	chainStorer := genericmocks.NewChainStorerMock(0)
	uint64Converter := mock.NewNonceHashConverterMock()

	// not reward transaction should not set status
	txA := &ApiTransactionResult{Status: TxStatusSuccess}

	isRewardReverted := (&StatusComputer{
		Uint64ByteSliceConverter: uint64Converter,
		Store: chainStorer,
	}).SetStatusIfIsRewardReverted(txA,block.TxBlock,0,nil)
	require.False(t, isRewardReverted)
	require.Equal(t, TxStatusSuccess, txA.Status)

	// shard - meta : reward transaction
	txB := &ApiTransactionResult{Status: TxStatusSuccess}

	headerHash := []byte("hash-1")
	headerNonce := uint64(10)
	nonceBytes := uint64Converter.ToByteSlice(headerNonce)
	_ = chainStorer.HdrNonce.Put(nonceBytes, headerHash)

	isRewardReverted = (&StatusComputer{
		SelfShardId: 			  core.MetachainShardId,
		Uint64ByteSliceConverter: uint64Converter,
		Store: 					  chainStorer,

	}).SetStatusIfIsRewardReverted(txB,block.RewardsBlock,headerNonce,headerHash)
	require.False(t, isRewardReverted)
	require.Equal(t, TxStatusSuccess, txB.Status)

	// shard 0: reward transaction
	txC := &ApiTransactionResult{Status: TxStatusSuccess}

	headerHash = []byte("hash-2")
	headerNonce = uint64(12)
	nonceBytes = uint64Converter.ToByteSlice(headerNonce)
	_ = chainStorer.HdrNonce.Put(nonceBytes, headerHash)

	isRewardReverted = (&StatusComputer{
		SelfShardId: 			  0,
		Uint64ByteSliceConverter: uint64Converter,
		Store: 					  chainStorer,
	}).SetStatusIfIsRewardReverted(txC,block.RewardsBlock,headerNonce,headerHash)
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
		SelfShardId: 			  core.MetachainShardId,
		Uint64ByteSliceConverter: uint64Converter,
		Store: 					  chainStorer,
	}).SetStatusIfIsRewardReverted(txD,block.RewardsBlock,headerNonce,wrongHash)
	require.True(t, isRewardReverted)
	require.Equal(t, TxStatusRewardReverted, txD.Status)
}
