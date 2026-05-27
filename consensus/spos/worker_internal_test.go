package spos

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	consensusMock "github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
)

func TestWorker_AddBlockToPoolSkipsNonWhitelistedCrossShardMiniBlocks(t *testing.T) {
	t.Parallel()

	miniBlock := &block.MiniBlock{
		SenderShardID:   1,
		ReceiverShardID: 0,
		Type:            block.TxBlock,
		TxHashes:        [][]byte{[]byte("tx-hash")},
	}

	putCalled := false
	worker := &Worker{
		blockProcessor: &testscommon.BlockProcessorStub{
			DecodeBlockBodyCalled: func(_ []byte) data.BodyHandler {
				return &block.Body{MiniBlocks: []*block.MiniBlock{miniBlock}}
			},
		},
		marshalizer:      &consensusMock.MarshalizerMock{},
		hasher:           &hashingMocks.HasherMock{},
		shardCoordinator: testscommon.NewMultiShardsCoordinatorMock(2),
		whiteListHandler: &testscommon.WhiteListHandlerStub{},
		poolAdder: &cache.CacherStub{
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				putCalled = true
				return false
			},
		},
	}

	_ = worker.addBlockToPool([]byte("body"))

	require.False(t, putCalled)
}

func TestWorker_AddBlockToPoolAcceptsWhitelistedCrossShardMiniBlocks(t *testing.T) {
	t.Parallel()

	miniBlock := &block.MiniBlock{
		SenderShardID:   1,
		ReceiverShardID: 0,
		Type:            block.TxBlock,
		TxHashes:        [][]byte{[]byte("tx-hash")},
	}
	marshalizer := &consensusMock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	expectedHash, err := core.CalculateHash(marshalizer, hasher, miniBlock)
	require.NoError(t, err)

	putCalled := false
	worker := &Worker{
		blockProcessor: &testscommon.BlockProcessorStub{
			DecodeBlockBodyCalled: func(_ []byte) data.BodyHandler {
				return &block.Body{MiniBlocks: []*block.MiniBlock{miniBlock}}
			},
		},
		marshalizer:      marshalizer,
		hasher:           hasher,
		shardCoordinator: testscommon.NewMultiShardsCoordinatorMock(2),
		whiteListHandler: &testscommon.WhiteListHandlerStub{
			IsWhiteListedAtLeastOneCalled: func(identifiers [][]byte) bool {
				return len(identifiers) == 1 && bytes.Equal(identifiers[0], expectedHash)
			},
		},
		poolAdder: &cache.CacherStub{
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				putCalled = true
				require.True(t, bytes.Equal(expectedHash, key))
				return false
			},
		},
	}

	_ = worker.addBlockToPool([]byte("body"))

	require.True(t, putCalled)
}
