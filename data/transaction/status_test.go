package transaction

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/stretchr/testify/require"
)

func TestStatusComputer_ComputeStatusWhenInStorageKnowingMiniblock(t *testing.T) {
	// Invalid miniblock
	computer := &StatusComputer{
		MiniblockType:    block.InvalidBlock,
		SourceShard:      12,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusInvalid, computer.ComputeStatusWhenInStorageKnowingMiniblock())

	// Intra-shard
	computer = &StatusComputer{
		MiniblockType:    block.TxBlock,
		SourceShard:      12,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusSuccess, computer.ComputeStatusWhenInStorageKnowingMiniblock())

	// Cross, at source
	computer = &StatusComputer{
		MiniblockType:    block.TxBlock,
		SourceShard:      12,
		DestinationShard: 13,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusPending, computer.ComputeStatusWhenInStorageKnowingMiniblock())

	// Cross, at source, but knowing that it has been fully notarized (through DatabaseLookupExtensions)
	computer = &StatusComputer{
		MiniblockType:        block.TxBlock,
		IsMiniblockFinalized: true,
		SourceShard:          12,
		DestinationShard:     13,
		SelfShard:            12,
	}
	require.Equal(t, TxStatusSuccess, computer.ComputeStatusWhenInStorageKnowingMiniblock())

	// Cross, destination me
	computer = &StatusComputer{
		MiniblockType:    block.TxBlock,
		SourceShard:      13,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusSuccess, computer.ComputeStatusWhenInStorageKnowingMiniblock())

	computer = &StatusComputer{
		MiniblockType:    block.RewardsBlock,
		SourceShard:      core.MetachainShardId,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusSuccess, computer.ComputeStatusWhenInStorageKnowingMiniblock())

	// Contract deploy
	computer = &StatusComputer{
		SourceShard:      12,
		DestinationShard: 13,
		Receiver:         make([]byte, 32),
		TransactionData:  []byte("deployingAContract"),
		SelfShard:        13,
	}
	require.Equal(t, TxStatusSuccess, computer.ComputeStatusWhenInStorageKnowingMiniblock())
}

func TestStatusComputer_ComputeStatusWhenInStorageNotKnowingMiniblock(t *testing.T) {
	// Intra shard
	computer := &StatusComputer{
		SourceShard:      12,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusSuccess, computer.ComputeStatusWhenInStorageNotKnowingMiniblock())

	// Cross, at source
	computer = &StatusComputer{
		SourceShard:      12,
		DestinationShard: 13,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusPending, computer.ComputeStatusWhenInStorageNotKnowingMiniblock())

	// Cross, destination me
	computer = &StatusComputer{
		SourceShard:      13,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusSuccess, computer.ComputeStatusWhenInStorageNotKnowingMiniblock())

	computer = &StatusComputer{
		SourceShard:      core.MetachainShardId,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusSuccess, computer.ComputeStatusWhenInStorageNotKnowingMiniblock())

	// Contract deploy
	computer = &StatusComputer{
		SourceShard:      12,
		DestinationShard: 13,
		Receiver:         make([]byte, 32),
		TransactionData:  []byte("deployingAContract"),
		SelfShard:        13,
	}
	require.Equal(t, TxStatusSuccess, computer.ComputeStatusWhenInStorageNotKnowingMiniblock())
}

func TestStatusComputer_SetStatusIfIsRewardReverted(t *testing.T) {
	t.Parallel()

	chainStorer := genericMocks.NewChainStorerMock(0)
	uint64Converter := mock.NewNonceHashConverterMock()

	// not reward transaction should not set status
	txA := &ApiTransactionResult{Status: TxStatusSuccess}

	isRewardReverted := (&StatusComputer{
		MiniblockType: block.TxBlock,
	}).SetStatusIfIsRewardReverted(txA)
	require.False(t, isRewardReverted)
	require.Equal(t, TxStatusSuccess, txA.Status)

	// shard - meta : reward transaction
	txB := &ApiTransactionResult{Status: TxStatusSuccess}

	headerHash := []byte("hash-1")
	headerNonce := uint64(10)
	nonceBytes := uint64Converter.ToByteSlice(headerNonce)
	_ = chainStorer.HdrNonce.Put(nonceBytes, headerHash)

	isRewardReverted = (&StatusComputer{
		HeaderHash:               headerHash,
		HeaderNonce:              headerNonce,
		MiniblockType:            block.RewardsBlock,
		SelfShard:                core.MetachainShardId,
		Store:                    chainStorer,
		Uint64ByteSliceConverter: uint64Converter,
	}).SetStatusIfIsRewardReverted(txB)
	require.False(t, isRewardReverted)
	require.Equal(t, TxStatusSuccess, txB.Status)

	// shard 0: reward transaction
	txC := &ApiTransactionResult{Status: TxStatusSuccess}

	headerHash = []byte("hash-2")
	headerNonce = uint64(12)
	nonceBytes = uint64Converter.ToByteSlice(headerNonce)
	_ = chainStorer.HdrNonce.Put(nonceBytes, headerHash)

	isRewardReverted = (&StatusComputer{
		HeaderHash:               headerHash,
		HeaderNonce:              headerNonce,
		MiniblockType:            block.RewardsBlock,
		SelfShard:                0,
		Store:                    chainStorer,
		Uint64ByteSliceConverter: uint64Converter,
	}).SetStatusIfIsRewardReverted(txC)
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
		HeaderHash:               wrongHash,
		HeaderNonce:              headerNonce,
		MiniblockType:            block.RewardsBlock,
		SelfShard:                core.MetachainShardId,
		Store:                    chainStorer,
		Uint64ByteSliceConverter: uint64Converter,
	}).SetStatusIfIsRewardReverted(txD)
	require.True(t, isRewardReverted)
	require.Equal(t, TxStatusRewardReverted, txD.Status)
}
