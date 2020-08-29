package transaction

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/require"
)

func TestStatusComputer_ComputeStatusWhenInPool(t *testing.T) {
	computer := &StatusComputer{
		SourceShard:      12,
		DestinationShard: 13,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusReceived, computer.ComputeStatusWhenInPool())

	// Cross-shard, destination me
	computer = &StatusComputer{
		SourceShard:      12,
		DestinationShard: 13,
		SelfShard:        13,
	}
	require.Equal(t, TxStatusPartiallyExecuted, computer.ComputeStatusWhenInPool())

	// Contract deploy
	computer = &StatusComputer{
		SourceShard:      12,
		DestinationShard: 13,
		Receiver:         make([]byte, 32),
		TransactionData:  []byte("deployingAContract"),
		SelfShard:        13,
	}
	require.Equal(t, TxStatusReceived, computer.ComputeStatusWhenInPool())
}

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
	require.Equal(t, TxStatusExecuted, computer.ComputeStatusWhenInStorageKnowingMiniblock())

	// Cross, at source
	computer = &StatusComputer{
		MiniblockType:    block.TxBlock,
		SourceShard:      12,
		DestinationShard: 13,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusPartiallyExecuted, computer.ComputeStatusWhenInStorageKnowingMiniblock())

	// Cross, at source, but knowing that it has been fully notarized (through DatabaseLookupExtensions)
	computer = &StatusComputer{
		MiniblockType:           block.TxBlock,
		MiniblockFullyNotarized: true,
		SourceShard:             12,
		DestinationShard:        13,
		SelfShard:               12,
	}
	require.Equal(t, TxStatusExecuted, computer.ComputeStatusWhenInStorageKnowingMiniblock())

	// Cross, destination me
	computer = &StatusComputer{
		MiniblockType:    block.TxBlock,
		SourceShard:      13,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusExecuted, computer.ComputeStatusWhenInStorageKnowingMiniblock())

	computer = &StatusComputer{
		MiniblockType:    block.RewardsBlock,
		SourceShard:      core.MetachainShardId,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusExecuted, computer.ComputeStatusWhenInStorageKnowingMiniblock())

	// Contract deploy
	computer = &StatusComputer{
		SourceShard:      12,
		DestinationShard: 13,
		Receiver:         make([]byte, 32),
		TransactionData:  []byte("deployingAContract"),
		SelfShard:        13,
	}
	require.Equal(t, TxStatusExecuted, computer.ComputeStatusWhenInStorageKnowingMiniblock())
}

func TestStatusComputer_ComputeStatusWhenInStorageNotKnowingMiniblock(t *testing.T) {
	// Intra shard
	computer := &StatusComputer{
		SourceShard:      12,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusExecuted, computer.ComputeStatusWhenInStorageNotKnowingMiniblock())

	// Cross, at source
	computer = &StatusComputer{
		SourceShard:      12,
		DestinationShard: 13,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusPartiallyExecuted, computer.ComputeStatusWhenInStorageNotKnowingMiniblock())

	// Cross, destination me
	computer = &StatusComputer{
		SourceShard:      13,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusExecuted, computer.ComputeStatusWhenInStorageNotKnowingMiniblock())

	computer = &StatusComputer{
		SourceShard:      core.MetachainShardId,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusExecuted, computer.ComputeStatusWhenInStorageNotKnowingMiniblock())

	// Contract deploy
	computer = &StatusComputer{
		SourceShard:      12,
		DestinationShard: 13,
		Receiver:         make([]byte, 32),
		TransactionData:  []byte("deployingAContract"),
		SelfShard:        13,
	}
	require.Equal(t, TxStatusExecuted, computer.ComputeStatusWhenInStorageNotKnowingMiniblock())
}
