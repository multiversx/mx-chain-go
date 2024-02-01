package block_test

import (
	"bytes"
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
	headerData []*headerData
}

func TestShardProcessor_RequestMissingFinalityAttestingHeaders(t *testing.T) {
	t.Parallel()

	t.Run("missing attesting meta header", func(t *testing.T) {
		t.Parallel()

		arguments, requestHandler := shardBlockRequestTestInit(t)
		testData := createShardProcessorTestData()
		metaChainData := testData[core.MetachainShardId]
		numCalls := atomic.Uint32{}
		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			require.Fail(t, fmt.Sprintf("should not request shard header by nonce, shardID: %d, nonce: %d", shardID, nonce))
		}
		requestHandler.RequestMetaHeaderByNonceCalled = func(nonce uint64) {
			attestationNonce := metaChainData.headerData[1].header.GetNonce()
			if nonce != attestationNonce {
				require.Fail(t, fmt.Sprintf("nonce should have been %d", attestationNonce))
			}
			numCalls.Add(1)
		}
		sp, _ := blproc.NewShardProcessor(arguments)

		metaBlockData := metaChainData.headerData[0]
		// not adding the confirmation metaBlock to the headers pool means it will be missing and requested
		sp.SetHighestHdrNonceForCurrentBlock(core.MetachainShardId, metaBlockData.header.GetNonce())
		res := sp.RequestMissingFinalityAttestingHeaders()
		time.Sleep(100 * time.Millisecond)

		require.Equal(t, uint32(1), res)
		require.Equal(t, uint32(1), numCalls.Load())
	})
	t.Run("no missing attesting meta header", func(t *testing.T) {
		t.Parallel()

		arguments, requestHandler := shardBlockRequestTestInit(t)
		testData := createShardProcessorTestData()
		metaChainData := testData[core.MetachainShardId]
		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			require.Fail(t, fmt.Sprintf("should not request shard header by nonce, shardID: %d, nonce: %d", shardID, nonce))
		}
		requestHandler.RequestMetaHeaderByNonceCalled = func(nonce uint64) {
			require.Fail(t, "should not request meta header by nonce")
		}
		sp, _ := blproc.NewShardProcessor(arguments)

		headersDataPool := arguments.DataComponents.Datapool().Headers()
		require.NotNil(t, headersDataPool)
		metaBlockData := metaChainData.headerData[0]
		confirmationMetaBlockData := metaChainData.headerData[1]
		headersDataPool.AddHeader(confirmationMetaBlockData.hash, confirmationMetaBlockData.header)
		sp.SetHighestHdrNonceForCurrentBlock(core.MetachainShardId, metaBlockData.header.GetNonce())
		res := sp.RequestMissingFinalityAttestingHeaders()
		time.Sleep(100 * time.Millisecond)

		require.Equal(t, uint32(0), res)
	})
}

func TestShardProcessor_computeExistingAndRequestMissingMetaHeaders(t *testing.T) {
	t.Parallel()

	shard1ID := uint32(1)
	t.Run("one referenced metaBlock missing will be requested", func(t *testing.T) {
		t.Parallel()

		arguments, requestHandler := shardBlockRequestTestInit(t)
		testData := createShardProcessorTestData()
		metaChainData := testData[core.MetachainShardId]
		shard1Data := testData[shard1ID]
		numCalls := atomic.Uint32{}
		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			require.Fail(t, fmt.Sprintf("should not request shard header by nonce, shardID: %d, nonce: %d", shardID, nonce))
		}
		requestHandler.RequestMetaHeaderByNonceCalled = func(nonce uint64) {
			// should only be called when requesting attestation meta header block
			require.Fail(t, "should not request meta header by nonce")
		}
		requestHandler.RequestMetaHeaderCalled = func(hash []byte) {
			require.Equal(t, metaChainData.headerData[1].hash, hash)
			numCalls.Add(1)
		}
		sp, _ := blproc.NewShardProcessor(arguments)

		metaBlockData := metaChainData.headerData[0]
		sp.SetHighestHdrNonceForCurrentBlock(core.MetachainShardId, metaBlockData.header.GetNonce())
		// not adding the referenced metaBlock to the headers pool means it will be missing and requested
		// first of the 2 referenced headers is added, the other will be missing
		headersDataPool := arguments.DataComponents.Datapool().Headers()
		headersDataPool.AddHeader(metaBlockData.hash, metaBlockData.header)

		blockBeingProcessed := shard1Data.headerData[1].header
		shardBlockBeingProcessed := blockBeingProcessed.(*block.Header)
		missingHeaders, missingFinalityAttestingHeaders := sp.ComputeExistingAndRequestMissingMetaHeaders(shardBlockBeingProcessed)
		time.Sleep(100 * time.Millisecond)

		require.Equal(t, uint32(1), missingHeaders)
		require.Equal(t, uint32(0), missingFinalityAttestingHeaders)
		require.Equal(t, uint32(1), numCalls.Load())
	})
	t.Run("multiple referenced metaBlocks missing will be requested", func(t *testing.T) {
		t.Parallel()

		arguments, requestHandler := shardBlockRequestTestInit(t)
		testData := createShardProcessorTestData()
		numCalls := atomic.Uint32{}
		metaChainData := testData[core.MetachainShardId]
		shard1Data := testData[shard1ID]
		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			require.Fail(t, fmt.Sprintf("should not request shard header by nonce, shardID: %d, nonce: %d", shardID, nonce))
		}
		requestHandler.RequestMetaHeaderByNonceCalled = func(nonce uint64) {
			// not yet requesting the attestation metaBlock
			require.Fail(t, "should not request meta header by nonce")
		}
		requestHandler.RequestMetaHeaderCalled = func(hash []byte) {
			if !(bytes.Equal(hash, metaChainData.headerData[0].hash) || bytes.Equal(hash, metaChainData.headerData[1].hash)) {
				require.Fail(t, "other requests than the expected 2 metaBlocks are not expected")
			}

			numCalls.Add(1)
		}
		sp, _ := blproc.NewShardProcessor(arguments)
		metaBlockData := testData[core.MetachainShardId].headerData[0]
		// not adding the referenced metaBlock to the headers pool means it will be missing and requested
		sp.SetHighestHdrNonceForCurrentBlock(core.MetachainShardId, metaBlockData.header.GetNonce())

		blockBeingProcessed := shard1Data.headerData[1].header
		shardBlockBeingProcessed := blockBeingProcessed.(*block.Header)
		missingHeaders, missingFinalityAttestingHeaders := sp.ComputeExistingAndRequestMissingMetaHeaders(shardBlockBeingProcessed)
		time.Sleep(100 * time.Millisecond)

		require.Equal(t, uint32(2), missingHeaders)
		require.Equal(t, uint32(0), missingFinalityAttestingHeaders)
		require.Equal(t, uint32(2), numCalls.Load())
	})
	t.Run("all referenced metaBlocks existing with missing attestation, will request the attestation metaBlock", func(t *testing.T) {
		t.Parallel()

		arguments, requestHandler := shardBlockRequestTestInit(t)
		testData := createShardProcessorTestData()
		numCallsMissing := atomic.Uint32{}
		numCallsAttestation := atomic.Uint32{}
		metaChainData := testData[core.MetachainShardId]
		shard1Data := testData[shard1ID]
		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			require.Fail(t, fmt.Sprintf("should not request shard header by nonce, shardID: %d, nonce: %d", shardID, nonce))
		}
		requestHandler.RequestMetaHeaderByNonceCalled = func(nonce uint64) {
			// not yet requesting the attestation metaBlock
			require.Equal(t, metaChainData.headerData[1].header.GetNonce()+1, nonce)
			numCallsAttestation.Add(1)
		}
		requestHandler.RequestMetaHeaderCalled = func(hash []byte) {
			if !(bytes.Equal(hash, metaChainData.headerData[0].hash) || bytes.Equal(hash, metaChainData.headerData[1].hash)) {
				require.Fail(t, "other requests than the expected 2 metaBlocks are not expected")
			}

			numCallsMissing.Add(1)
		}
		sp, _ := blproc.NewShardProcessor(arguments)
		// not adding the referenced metaBlock to the headers pool means it will be missing and requested
		headersDataPool := arguments.DataComponents.Datapool().Headers()
		headersDataPool.AddHeader(metaChainData.headerData[0].hash, metaChainData.headerData[0].header)
		headersDataPool.AddHeader(metaChainData.headerData[1].hash, metaChainData.headerData[1].header)

		blockBeingProcessed := shard1Data.headerData[1].header
		shardBlockBeingProcessed := blockBeingProcessed.(*block.Header)
		missingHeaders, missingFinalityAttestingHeaders := sp.ComputeExistingAndRequestMissingMetaHeaders(shardBlockBeingProcessed)
		time.Sleep(100 * time.Millisecond)

		require.Equal(t, uint32(0), missingHeaders)
		require.Equal(t, uint32(1), missingFinalityAttestingHeaders)
		require.Equal(t, uint32(0), numCallsMissing.Load())
		require.Equal(t, uint32(1), numCallsAttestation.Load())
	})
	t.Run("all referenced metaBlocks existing and existing attestation metaBlock will not request", func(t *testing.T) {
		t.Parallel()

		arguments, requestHandler := shardBlockRequestTestInit(t)
		testData := createShardProcessorTestData()
		numCallsMissing := atomic.Uint32{}
		numCallsAttestation := atomic.Uint32{}
		shard1Data := testData[shard1ID]
		metaChainData := testData[core.MetachainShardId]
		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			require.Fail(t, fmt.Sprintf("should not request shard header by nonce, shardID: %d, nonce: %d", shardID, nonce))
		}
		requestHandler.RequestMetaHeaderByNonceCalled = func(nonce uint64) {
			numCallsAttestation.Add(1)
		}
		requestHandler.RequestMetaHeaderCalled = func(hash []byte) {
			numCallsMissing.Add(1)
		}
		sp, _ := blproc.NewShardProcessor(arguments)
		// not adding the referenced metaBlock to the headers pool means it will be missing and requested
		headersDataPool := arguments.DataComponents.Datapool().Headers()
		headersDataPool.AddHeader(metaChainData.headerData[0].hash, metaChainData.headerData[0].header)
		headersDataPool.AddHeader(metaChainData.headerData[1].hash, metaChainData.headerData[1].header)
		attestationMetaBlock := &block.MetaBlock{
			Nonce:     102,
			Round:     102,
			PrevHash:  metaChainData.headerData[1].hash,
			ShardInfo: []block.ShardData{},
		}
		attestationMetaBlockHash := []byte("attestationHash")

		headersDataPool.AddHeader(attestationMetaBlockHash, attestationMetaBlock)

		blockBeingProcessed := shard1Data.headerData[1].header
		shardBlockBeingProcessed := blockBeingProcessed.(*block.Header)
		missingHeaders, missingFinalityAttestingHeaders := sp.ComputeExistingAndRequestMissingMetaHeaders(shardBlockBeingProcessed)
		time.Sleep(100 * time.Millisecond)

		require.Equal(t, uint32(0), missingHeaders)
		require.Equal(t, uint32(0), missingFinalityAttestingHeaders)
		require.Equal(t, uint32(0), numCallsMissing.Load())
		require.Equal(t, uint32(0), numCallsAttestation.Load())
	})
}

func TestShardProcessor_receivedMetaBlock(t *testing.T) {
	t.Parallel()

	t.Run("received non referenced metaBlock, while still having missing referenced metaBlocks", func(t *testing.T) {
		t.Parallel()

		arguments, requestHandler := shardBlockRequestTestInit(t)
		testData := createShardProcessorTestData()
		sp, _ := blproc.NewShardProcessor(arguments)
		hdrsForBlock := sp.GetHdrForBlock()

		firstMissingMetaBlockData := testData[core.MetachainShardId].headerData[0]
		secondMissingMetaBlockData := testData[core.MetachainShardId].headerData[1]

		requestHandler.RequestMetaHeaderCalled = func(hash []byte) {
			require.Fail(t, "no requests expected")
		}
		requestHandler.RequestMetaHeaderByNonceCalled = func(nonce uint64) {
			require.Fail(t, "no requests expected")
		}

		highestHeaderNonce := firstMissingMetaBlockData.header.GetNonce() - 1
		hdrsForBlock.SetNumMissingHdrs(2)
		hdrsForBlock.SetNumMissingFinalityAttestingHdrs(0)
		hdrsForBlock.SetHighestHdrNonce(core.MetachainShardId, highestHeaderNonce)
		hdrsForBlock.SetHdrHashAndInfo(string(firstMissingMetaBlockData.hash),
			&blproc.HdrInfo{
				UsedInBlock: true,
				Hdr:         nil,
			})
		hdrsForBlock.SetHdrHashAndInfo(string(secondMissingMetaBlockData.hash),
			&blproc.HdrInfo{
				UsedInBlock: true,
				Hdr:         nil,
			})
		otherMetaBlock := &block.MetaBlock{
			Nonce:    102,
			Round:    102,
			PrevHash: []byte("other meta block prev hash"),
		}

		otherMetaBlockHash := []byte("other meta block hash")
		sp.ReceivedMetaBlock(otherMetaBlock, otherMetaBlockHash)
		time.Sleep(100 * time.Millisecond)

		require.Equal(t, uint32(2), hdrsForBlock.GetMissingHdrs())
		require.Equal(t, uint32(0), hdrsForBlock.GetMissingFinalityAttestingHdrs())
		highestHeaderNonces := hdrsForBlock.GetHighestHdrNonce()
		require.Equal(t, highestHeaderNonce, highestHeaderNonces[core.MetachainShardId])
	})
	t.Run("received missing referenced metaBlock, other referenced metaBlock still missing", func(t *testing.T) {
		t.Parallel()

		arguments, requestHandler := shardBlockRequestTestInit(t)
		testData := createShardProcessorTestData()
		sp, _ := blproc.NewShardProcessor(arguments)
		hdrsForBlock := sp.GetHdrForBlock()

		firstMissingMetaBlockData := testData[core.MetachainShardId].headerData[0]
		secondMissingMetaBlockData := testData[core.MetachainShardId].headerData[1]

		requestHandler.RequestMetaHeaderCalled = func(hash []byte) {
			require.Fail(t, "no requests expected")
		}
		requestHandler.RequestMetaHeaderByNonceCalled = func(nonce uint64) {
			require.Fail(t, "no requests expected")
		}

		highestHeaderNonce := firstMissingMetaBlockData.header.GetNonce() - 1
		hdrsForBlock.SetNumMissingHdrs(2)
		hdrsForBlock.SetNumMissingFinalityAttestingHdrs(0)
		hdrsForBlock.SetHighestHdrNonce(core.MetachainShardId, highestHeaderNonce)
		hdrsForBlock.SetHdrHashAndInfo(string(firstMissingMetaBlockData.hash),
			&blproc.HdrInfo{
				UsedInBlock: true,
				Hdr:         nil,
			})
		hdrsForBlock.SetHdrHashAndInfo(string(secondMissingMetaBlockData.hash),
			&blproc.HdrInfo{
				UsedInBlock: true,
				Hdr:         nil,
			})

		sp.ReceivedMetaBlock(firstMissingMetaBlockData.header, firstMissingMetaBlockData.hash)
		time.Sleep(100 * time.Millisecond)

		require.Equal(t, uint32(1), hdrsForBlock.GetMissingHdrs())
		require.Equal(t, uint32(0), hdrsForBlock.GetMissingFinalityAttestingHdrs())
		highestHeaderNonces := hdrsForBlock.GetHighestHdrNonce()
		require.Equal(t, firstMissingMetaBlockData.header.GetNonce(), highestHeaderNonces[core.MetachainShardId])
	})
	t.Run("received non missing referenced metaBlock", func(t *testing.T) {
		t.Parallel()

		arguments, requestHandler := shardBlockRequestTestInit(t)
		testData := createShardProcessorTestData()
		sp, _ := blproc.NewShardProcessor(arguments)
		hdrsForBlock := sp.GetHdrForBlock()

		notMissingReferencedMetaBlockData := testData[core.MetachainShardId].headerData[0]
		missingMetaBlockData := testData[core.MetachainShardId].headerData[1]

		requestHandler.RequestMetaHeaderCalled = func(hash []byte) {
			require.Fail(t, "no requests expected")
		}
		requestHandler.RequestMetaHeaderByNonceCalled = func(nonce uint64) {
			require.Fail(t, "no requests expected")
		}

		highestHeaderNonce := notMissingReferencedMetaBlockData.header.GetNonce() - 1
		hdrsForBlock.SetNumMissingHdrs(1)
		hdrsForBlock.SetNumMissingFinalityAttestingHdrs(0)
		hdrsForBlock.SetHighestHdrNonce(core.MetachainShardId, highestHeaderNonce)
		hdrsForBlock.SetHdrHashAndInfo(string(notMissingReferencedMetaBlockData.hash),
			&blproc.HdrInfo{
				UsedInBlock: true,
				Hdr:         notMissingReferencedMetaBlockData.header,
			})
		hdrsForBlock.SetHdrHashAndInfo(string(missingMetaBlockData.hash),
			&blproc.HdrInfo{
				UsedInBlock: true,
				Hdr:         nil,
			})

		headersDataPool := arguments.DataComponents.Datapool().Headers()
		require.NotNil(t, headersDataPool)
		headersDataPool.AddHeader(notMissingReferencedMetaBlockData.hash, notMissingReferencedMetaBlockData.header)

		sp.ReceivedMetaBlock(notMissingReferencedMetaBlockData.header, notMissingReferencedMetaBlockData.hash)
		time.Sleep(100 * time.Millisecond)

		require.Equal(t, uint32(1), hdrsForBlock.GetMissingHdrs())
		require.Equal(t, uint32(0), hdrsForBlock.GetMissingFinalityAttestingHdrs())
		hdrsForBlockHighestNonces := hdrsForBlock.GetHighestHdrNonce()
		require.Equal(t, highestHeaderNonce, hdrsForBlockHighestNonces[core.MetachainShardId])
	})
	t.Run("received missing attestation metaBlock", func(t *testing.T) {
		t.Parallel()

		arguments, requestHandler := shardBlockRequestTestInit(t)
		testData := createShardProcessorTestData()
		sp, _ := blproc.NewShardProcessor(arguments)
		hdrsForBlock := sp.GetHdrForBlock()

		referencedMetaBlock := testData[core.MetachainShardId].headerData[0]
		lastReferencedMetaBlock := testData[core.MetachainShardId].headerData[1]
		attestationMetaBlockHash := []byte("attestation meta block hash")
		attestationMetaBlock := &block.MetaBlock{
			Nonce:    lastReferencedMetaBlock.header.GetNonce() + 1,
			Round:    lastReferencedMetaBlock.header.GetRound() + 1,
			PrevHash: lastReferencedMetaBlock.hash,
		}

		requestHandler.RequestMetaHeaderCalled = func(hash []byte) {
			require.Fail(t, "no requests expected")
		}
		requestHandler.RequestMetaHeaderByNonceCalled = func(nonce uint64) {
			require.Fail(t, "no requests expected")
		}

		hdrsForBlock.SetNumMissingHdrs(0)
		hdrsForBlock.SetNumMissingFinalityAttestingHdrs(1)
		hdrsForBlock.SetHighestHdrNonce(core.MetachainShardId, lastReferencedMetaBlock.header.GetNonce())
		hdrsForBlock.SetHdrHashAndInfo(string(referencedMetaBlock.hash),
			&blproc.HdrInfo{
				UsedInBlock: true,
				Hdr:         referencedMetaBlock.header,
			})
		hdrsForBlock.SetHdrHashAndInfo(string(lastReferencedMetaBlock.hash),
			&blproc.HdrInfo{
				UsedInBlock: true,
				Hdr:         lastReferencedMetaBlock.header,
			})

		headersDataPool := arguments.DataComponents.Datapool().Headers()
		require.NotNil(t, headersDataPool)
		headersDataPool.AddHeader(referencedMetaBlock.hash, referencedMetaBlock.header)
		headersDataPool.AddHeader(lastReferencedMetaBlock.hash, lastReferencedMetaBlock.header)
		headersDataPool.AddHeader(attestationMetaBlockHash, attestationMetaBlock)
		wg := startWaitingForAllHeadersReceivedSignal(t, sp)

		sp.ReceivedMetaBlock(attestationMetaBlock, attestationMetaBlockHash)
		wg.Wait()

		require.Equal(t, uint32(0), hdrsForBlock.GetMissingHdrs())
		require.Equal(t, uint32(0), hdrsForBlock.GetMissingFinalityAttestingHdrs())
		hdrsForBlockHighestNonces := hdrsForBlock.GetHighestHdrNonce()
		require.Equal(t, lastReferencedMetaBlock.header.GetNonce(), hdrsForBlockHighestNonces[core.MetachainShardId])
	})
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

	prevMetaBlockHash := []byte("prev meta block hash")
	metaBlockHash := []byte("meta block hash")
	metaConfirmationHash := []byte("confirmation meta block hash")

	shard0Block0Hash := []byte("shard 0 block 0 hash")
	shard0Block1Hash := []byte("shard 0 block 1 hash")
	shard0Block2Hash := []byte("shard 0 block 2 hash")

	shard1Block0Hash := []byte("shard 1 block 0 hash")
	shard1Block1Hash := []byte("shard 1 block 1 hash")
	shard1Block2Hash := []byte("shard 1 block 2 hash")

	metaBlock := &block.MetaBlock{
		Nonce:    100,
		Round:    100,
		PrevHash: prevMetaBlockHash,
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
		ShardID:         1,
		PrevHash:        shard1Block0Hash,
		MetaBlockHashes: [][]byte{prevMetaBlockHash},
		Nonce:           102,
		Round:           102,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: mbHash4, SenderShardID: 0, ReceiverShardID: 1},
			{Hash: mbHash5, SenderShardID: 0, ReceiverShardID: 1},
			{Hash: mbHash6, SenderShardID: 0, ReceiverShardID: 1},
		},
	}

	shard1Block2 := &block.Header{
		ShardID:          1,
		PrevHash:         shard1Block1Hash,
		MetaBlockHashes:  [][]byte{metaBlockHash, metaConfirmationHash},
		Nonce:            103,
		Round:            103,
		MiniBlockHeaders: []block.MiniBlockHeader{},
	}

	sbd := map[uint32]*shardBlockTestData{
		0: {
			headerData: []*headerData{
				{
					hash:   shard0Block1Hash,
					header: shard0Block1,
				},
				{
					hash:   shard0Block2Hash,
					header: shard0Block2,
				},
			},
		},
		1: {
			headerData: []*headerData{
				{
					hash:   shard1Block1Hash,
					header: shar1Block1,
				},
				{
					hash:   shard1Block2Hash,
					header: shard1Block2,
				},
			},
		},
		core.MetachainShardId: {
			headerData: []*headerData{
				{
					hash:   metaBlockHash,
					header: metaBlock,
				},
				{
					hash:   metaConfirmationHash,
					header: metaConfirmationBlock,
				},
			},
		},
	}

	return sbd
}
