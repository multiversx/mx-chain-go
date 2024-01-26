package block_test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
)

type headerData struct {
	hash   []byte
	header data.HeaderHandler
}

type shardBlockTestData struct {
	headerData             *headerData
	confirmationHeaderData *headerData
}

func TestShardProcessor_RequestMissingFinalityAttestingHeaders(t *testing.T) {
	t.Parallel()

	t.Run("missing attesting meta header", func(t *testing.T) {
		arguments, requestHandler := shardBlockRequestTestInit(t)
		testData := createShardProcessorTestData()
		numCalls := atomic.Uint32{}
		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			require.Fail(t, fmt.Sprintf("should not request shard header by nonce, shardID: %d, nonce: %d", shardID, nonce))
		}
		requestHandler.RequestMetaHeaderByNonceCalled = func(nonce uint64) {
			attestationNonce := testData[core.MetachainShardId].confirmationHeaderData.header.GetNonce()
			if nonce != attestationNonce {
				require.Fail(t, fmt.Sprintf("nonce should have been %d", attestationNonce))
			}
			numCalls.Add(1)
		}
		sp, _ := blproc.NewShardProcessor(arguments)

		metaBlockData := testData[core.MetachainShardId].headerData
		// not adding the confirmation metaBlock to the headers pool means it will be missing and requested
		sp.SetHighestHdrNonceForCurrentBlock(core.MetachainShardId, metaBlockData.header.GetNonce())
		res := sp.RequestMissingFinalityAttestingHeaders()
		time.Sleep(100 * time.Millisecond)

		require.Equal(t, uint32(1), res)
		require.Equal(t, uint32(1), numCalls.Load())
	})
	t.Run("no missing attesting meta header", func(t *testing.T) {
		arguments, requestHandler := shardBlockRequestTestInit(t)
		testData := createShardProcessorTestData()
		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			require.Fail(t, fmt.Sprintf("should not request shard header by nonce, shardID: %d, nonce: %d", shardID, nonce))
		}
		requestHandler.RequestMetaHeaderByNonceCalled = func(nonce uint64) {
			require.Fail(t, "should not request meta header by nonce")
		}
		sp, _ := blproc.NewShardProcessor(arguments)

		headersDataPool := arguments.DataComponents.Datapool().Headers()
		require.NotNil(t, headersDataPool)
		metaBlockData := testData[core.MetachainShardId].headerData
		confirmationMetaBlockData := testData[core.MetachainShardId].confirmationHeaderData
		headersDataPool.AddHeader(confirmationMetaBlockData.hash, confirmationMetaBlockData.header)
		sp.SetHighestHdrNonceForCurrentBlock(core.MetachainShardId, metaBlockData.header.GetNonce())
		res := sp.RequestMissingFinalityAttestingHeaders()
		time.Sleep(100 * time.Millisecond)
		require.Equal(t, uint32(0), res)
	})
}

func TestShardProcessor_computeExistingAndRequestMissingMetaHeaders(t *testing.T) {

}

func TestShardProcessor_receivedMetaBlock(t *testing.T) {

}

func shardBlockRequestTestInit(t *testing.T) (blproc.ArgShardProcessor, *testscommon.RequestHandlerStub) {
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	poolMock := dataRetrieverMock.NewPoolsHolderMock()
	dataComponents.DataPool = poolMock
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	poolsHolderAsInterface := arguments.DataComponents.Datapool()
	poolsHolder, ok := poolsHolderAsInterface.(*dataRetrieverMock.PoolsHolderMock)
	require.True(t, ok)

	headersPoolStub := createPoolsHolderForHeaderRequests()
	poolsHolder.SetHeadersPool(headersPoolStub)

	requestHandler, ok := arguments.ArgBaseProcessor.RequestHandler.(*testscommon.RequestHandlerStub)
	require.True(t, ok)
	return arguments, requestHandler
}

func createShardProcessorTestData() map[uint32]*shardBlockTestData {
	// shard 0 miniblocks
	mbHash1 := []byte("mb hash 1")
	mbHash2 := []byte("mb hash 2")
	mbHash3 := []byte("mb hash 3")

	// shard 1 miniblocks
	mbHash4 := []byte("mb hash 4")
	mbHash5 := []byte("mb hash 5")
	mbHash6 := []byte("mb hash 6")

	metaBlockHash := []byte("meta block hash")
	metaConfirmationHash := []byte("confirmation meta block hash")

	shard0Block0Hash := []byte("shard 0 block 0 hash")
	shard0Block1Hash := []byte("shard 0 block 1 hash")
	shard0Block2Hash := []byte("shard 0 block 2 hash")

	shard1Block0Hash := []byte("shard 1 block 0 hash")
	shard1Block1Hash := []byte("shard 1 block 1 hash")
	shard1Block2Hash := []byte("shard 1 block 2 hash")

	metaBlock := &block.MetaBlock{
		Nonce: 100,
		Round: 100,
		ShardInfo: []block.ShardData{
			{
				ShardID:    0,
				HeaderHash: shard0Block1Hash,
				PrevHash:   shard0Block0Hash,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{Hash: mbHash1, SenderShardID: 0, ReceiverShardID: 1},
					{Hash: mbHash2, SenderShardID: 0, ReceiverShardID: 1},
					{Hash: mbHash3, SenderShardID: 0, ReceiverShardID: 1},
				},
			},
			{
				ShardID:    1,
				HeaderHash: shard1Block1Hash,
				PrevHash:   shard1Block0Hash,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{Hash: mbHash4, SenderShardID: 1, ReceiverShardID: 0},
					{Hash: mbHash5, SenderShardID: 1, ReceiverShardID: 0},
					{Hash: mbHash6, SenderShardID: 1, ReceiverShardID: 0},
				},
			},
		},
	}
	metaConfirmationBlock := &block.MetaBlock{
		Nonce:     101,
		Round:     101,
		PrevHash:  metaBlockHash,
		ShardInfo: []block.ShardData{},
	}

	shard0Block1 := &block.Header{
		ShardID:  0,
		PrevHash: shard0Block0Hash,
		Nonce:    98,
		Round:    98,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: mbHash1, SenderShardID: 0, ReceiverShardID: 1},
			{Hash: mbHash2, SenderShardID: 0, ReceiverShardID: 1},
			{Hash: mbHash3, SenderShardID: 0, ReceiverShardID: 1},
		},
	}

	shard0Block2 := &block.Header{
		ShardID:          0,
		PrevHash:         shard0Block1Hash,
		Nonce:            99,
		Round:            99,
		MiniBlockHeaders: []block.MiniBlockHeader{},
	}

	shar1Block1 := &block.Header{
		ShardID:  1,
		PrevHash: shard1Block0Hash,
		Nonce:    98,
		Round:    98,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: mbHash4, SenderShardID: 0, ReceiverShardID: 1},
			{Hash: mbHash5, SenderShardID: 0, ReceiverShardID: 1},
			{Hash: mbHash6, SenderShardID: 0, ReceiverShardID: 1},
		},
	}

	shard1Block2 := &block.Header{
		ShardID:          1,
		PrevHash:         shard1Block1Hash,
		Nonce:            99,
		Round:            99,
		MiniBlockHeaders: []block.MiniBlockHeader{},
	}

	sbd := map[uint32]*shardBlockTestData{
		0: {
			headerData: &headerData{
				hash:   shard0Block1Hash,
				header: shard0Block1,
			},
			confirmationHeaderData: &headerData{
				hash:   shard0Block2Hash,
				header: shard0Block2,
			},
		},
		1: {
			headerData: &headerData{
				hash:   shard1Block1Hash,
				header: shar1Block1,
			},
			confirmationHeaderData: &headerData{
				hash:   shard1Block2Hash,
				header: shard1Block2,
			},
		},
		core.MetachainShardId: {
			headerData: &headerData{
				hash:   metaBlockHash,
				header: metaBlock,
			},
			confirmationHeaderData: &headerData{
				hash:   metaConfirmationHash,
				header: metaConfirmationBlock,
			},
		},
	}

	return sbd
}
