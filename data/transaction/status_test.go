package transaction

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
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
	require.Equal(t, TxStatusSuccessful, computer.ComputeStatusWhenInStorageKnowingMiniblock())

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
	require.Equal(t, TxStatusSuccessful, computer.ComputeStatusWhenInStorageKnowingMiniblock())

	// Cross, destination me
	computer = &StatusComputer{
		MiniblockType:    block.TxBlock,
		SourceShard:      13,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusSuccessful, computer.ComputeStatusWhenInStorageKnowingMiniblock())

	computer = &StatusComputer{
		MiniblockType:    block.RewardsBlock,
		SourceShard:      core.MetachainShardId,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusSuccessful, computer.ComputeStatusWhenInStorageKnowingMiniblock())

	// Contract deploy
	computer = &StatusComputer{
		SourceShard:      12,
		DestinationShard: 13,
		Receiver:         make([]byte, 32),
		TransactionData:  []byte("deployingAContract"),
		SelfShard:        13,
	}
	require.Equal(t, TxStatusSuccessful, computer.ComputeStatusWhenInStorageKnowingMiniblock())
}

func TestStatusComputer_ComputeStatusWhenInStorageNotKnowingMiniblock(t *testing.T) {
	// Intra shard
	computer := &StatusComputer{
		SourceShard:      12,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusSuccessful, computer.ComputeStatusWhenInStorageNotKnowingMiniblock())

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
	require.Equal(t, TxStatusSuccessful, computer.ComputeStatusWhenInStorageNotKnowingMiniblock())

	computer = &StatusComputer{
		SourceShard:      core.MetachainShardId,
		DestinationShard: 12,
		SelfShard:        12,
	}
	require.Equal(t, TxStatusSuccessful, computer.ComputeStatusWhenInStorageNotKnowingMiniblock())

	// Contract deploy
	computer = &StatusComputer{
		SourceShard:      12,
		DestinationShard: 13,
		Receiver:         make([]byte, 32),
		TransactionData:  []byte("deployingAContract"),
		SelfShard:        13,
	}
	require.Equal(t, TxStatusSuccessful, computer.ComputeStatusWhenInStorageNotKnowingMiniblock())
}
