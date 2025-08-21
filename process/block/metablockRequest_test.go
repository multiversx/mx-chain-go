package block_test

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	blockProcess "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/headerForBlock"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaProcessor_computeExistingAndRequestMissingShardHeaders(t *testing.T) {
	t.Parallel()

	noOfShards := uint32(2)
	td := createTestData()

	t.Run("all referenced shard headers missing", func(t *testing.T) {
		t.Parallel()
		referencedHeaders := []*shardHeaderData{td[0].referencedHeaderData, td[1].referencedHeaderData}
		shardInfo := createShardInfo(referencedHeaders)
		metaBlock := &block.MetaBlock{
			ShardInfo: shardInfo,
		}

		numCallsMissingAttestation := atomic.Uint32{}
		numCallsMissingHeaders := atomic.Uint32{}
		arguments := createMetaProcessorArguments(t, noOfShards)
		updateRequestsHandlerForCountingRequests(t, arguments, td, metaBlock, &numCallsMissingHeaders, &numCallsMissingAttestation)

		mp, err := blockProcess.NewMetaProcessor(*arguments)
		require.Nil(t, err)
		require.NotNil(t, mp)

		arguments.HeadersForBlock.RequestShardHeaders(metaBlock)
		numMissing, missingProofs, numAttestationMissing := arguments.HeadersForBlock.GetMissingData()
		time.Sleep(100 * time.Millisecond)
		require.Equal(t, uint32(2), numMissing)
		// before receiving all missing headers referenced in metaBlock, the number of missing attestations is not updated
		require.Equal(t, uint32(0), numAttestationMissing)
		require.Equal(t, uint32(0), missingProofs)
		require.Len(t, arguments.HeadersForBlock.GetHeadersInfoMap(), 2)
		require.Equal(t, uint32(0), numCallsMissingAttestation.Load())
		require.Equal(t, uint32(2), numCallsMissingHeaders.Load())
	})
	t.Run("one referenced shard header present and one missing", func(t *testing.T) {
		t.Parallel()
		referencedHeaders := []*shardHeaderData{td[0].referencedHeaderData, td[1].referencedHeaderData}
		shardInfo := createShardInfo(referencedHeaders)
		metaBlock := &block.MetaBlock{
			ShardInfo: shardInfo,
		}

		numCallsMissingAttestation := atomic.Uint32{}
		numCallsMissingHeaders := atomic.Uint32{}
		arguments := createMetaProcessorArguments(t, noOfShards)
		poolsHolder, ok := arguments.DataComponents.Datapool().(*dataRetrieverMock.PoolsHolderMock)
		require.True(t, ok)

		headersPoolStub := createPoolsHolderForHeaderRequests()
		poolsHolder.SetHeadersPool(headersPoolStub)
		updateRequestsHandlerForCountingRequests(t, arguments, td, metaBlock, &numCallsMissingHeaders, &numCallsMissingAttestation)

		mp, err := blockProcess.NewMetaProcessor(*arguments)
		require.Nil(t, err)
		require.NotNil(t, mp)

		headersPool := mp.GetDataPool().Headers()
		// adding the existing header
		headersPool.AddHeader(td[0].referencedHeaderData.headerHash, td[0].referencedHeaderData.header)
		arguments.HeadersForBlock.RequestShardHeaders(metaBlock)

		time.Sleep(100 * time.Millisecond)
		headersForBlock := mp.GetHdrForBlock()
		missingHdrs, missingProofs, missingAttesting := headersForBlock.GetMissingData()
		require.Equal(t, uint32(1), missingHdrs)
		// before receiving all missing headers referenced in metaBlock, the number of missing attestations is not updated
		require.Equal(t, uint32(0), missingAttesting)
		require.Equal(t, uint32(0), missingProofs)
		require.Len(t, headersForBlock.GetHeadersInfoMap(), 2)
		require.Equal(t, uint32(0), numCallsMissingAttestation.Load())
		require.Equal(t, uint32(1), numCallsMissingHeaders.Load())
	})
	t.Run("all referenced shard headers present, all attestation headers missing", func(t *testing.T) {
		t.Parallel()
		referencedHeaders := []*shardHeaderData{td[0].referencedHeaderData, td[1].referencedHeaderData}
		shardInfo := createShardInfo(referencedHeaders)
		metaBlock := &block.MetaBlock{
			ShardInfo: shardInfo,
		}

		numCallsMissingAttestation := atomic.Uint32{}
		numCallsMissingHeaders := atomic.Uint32{}
		arguments := createMetaProcessorArguments(t, noOfShards)
		poolsHolder, ok := arguments.DataComponents.Datapool().(*dataRetrieverMock.PoolsHolderMock)
		require.True(t, ok)

		headersPoolStub := createPoolsHolderForHeaderRequests()
		poolsHolder.SetHeadersPool(headersPoolStub)
		updateRequestsHandlerForCountingRequests(t, arguments, td, metaBlock, &numCallsMissingHeaders, &numCallsMissingAttestation)

		mp, err := blockProcess.NewMetaProcessor(*arguments)
		require.Nil(t, err)
		require.NotNil(t, mp)

		headersPool := mp.GetDataPool().Headers()
		// adding the existing headers
		headersPool.AddHeader(td[0].referencedHeaderData.headerHash, td[0].referencedHeaderData.header)
		headersPool.AddHeader(td[1].referencedHeaderData.headerHash, td[1].referencedHeaderData.header)
		arguments.HeadersForBlock.RequestShardHeaders(metaBlock)
		time.Sleep(100 * time.Millisecond)
		headersForBlock := mp.GetHdrForBlock()
		missingHdrs, missingProofs, missingAttesting := headersForBlock.GetMissingData()
		require.Equal(t, uint32(0), missingHdrs)
		require.Equal(t, uint32(0), missingProofs)
		require.Equal(t, uint32(2), missingAttesting)
		require.Len(t, headersForBlock.GetHeadersInfoMap(), 2)
		require.Equal(t, uint32(2), numCallsMissingAttestation.Load())
		require.Equal(t, uint32(0), numCallsMissingHeaders.Load())
	})
	t.Run("all referenced shard headers present, one attestation header missing", func(t *testing.T) {
		t.Parallel()
		referencedHeaders := []*shardHeaderData{td[0].referencedHeaderData, td[1].referencedHeaderData}
		shardInfo := createShardInfo(referencedHeaders)
		metaBlock := &block.MetaBlock{
			ShardInfo: shardInfo,
		}

		numCallsMissingAttestation := atomic.Uint32{}
		numCallsMissingHeaders := atomic.Uint32{}
		arguments := createMetaProcessorArguments(t, noOfShards)
		poolsHolder, ok := arguments.DataComponents.Datapool().(*dataRetrieverMock.PoolsHolderMock)
		require.True(t, ok)

		headersPoolStub := createPoolsHolderForHeaderRequests()
		poolsHolder.SetHeadersPool(headersPoolStub)
		updateRequestsHandlerForCountingRequests(t, arguments, td, metaBlock, &numCallsMissingHeaders, &numCallsMissingAttestation)

		mp, err := blockProcess.NewMetaProcessor(*arguments)
		require.Nil(t, err)
		require.NotNil(t, mp)

		headersPool := mp.GetDataPool().Headers()
		// adding the existing headers
		headersPool.AddHeader(td[0].referencedHeaderData.headerHash, td[0].referencedHeaderData.header)
		headersPool.AddHeader(td[1].referencedHeaderData.headerHash, td[1].referencedHeaderData.header)
		headersPool.AddHeader(td[0].attestationHeaderData.headerHash, td[0].attestationHeaderData.header)
		arguments.HeadersForBlock.RequestShardHeaders(metaBlock)
		time.Sleep(100 * time.Millisecond)
		headersForBlock := mp.GetHdrForBlock()
		missingHdrs, missingProofs, missingAttesting := headersForBlock.GetMissingData()
		require.Equal(t, uint32(0), missingHdrs)
		require.Equal(t, uint32(0), missingProofs)
		require.Equal(t, uint32(1), missingAttesting)
		require.Len(t, headersForBlock.GetHeadersInfoMap(), 3)
		require.Equal(t, uint32(1), numCallsMissingAttestation.Load())
		require.Equal(t, uint32(0), numCallsMissingHeaders.Load())
	})
	t.Run("all referenced shard headers present, all attestation headers present", func(t *testing.T) {
		t.Parallel()
		referencedHeaders := []*shardHeaderData{td[0].referencedHeaderData, td[1].referencedHeaderData}
		shardInfo := createShardInfo(referencedHeaders)
		metaBlock := &block.MetaBlock{
			ShardInfo: shardInfo,
		}

		numCallsMissingAttestation := atomic.Uint32{}
		numCallsMissingHeaders := atomic.Uint32{}
		arguments := createMetaProcessorArguments(t, noOfShards)
		poolsHolder, ok := arguments.DataComponents.Datapool().(*dataRetrieverMock.PoolsHolderMock)
		require.True(t, ok)

		headersPoolStub := createPoolsHolderForHeaderRequests()
		poolsHolder.SetHeadersPool(headersPoolStub)
		updateRequestsHandlerForCountingRequests(t, arguments, td, metaBlock, &numCallsMissingHeaders, &numCallsMissingAttestation)

		mp, err := blockProcess.NewMetaProcessor(*arguments)
		require.Nil(t, err)
		require.NotNil(t, mp)

		headersPool := mp.GetDataPool().Headers()
		// adding the existing headers
		headersPool.AddHeader(td[0].referencedHeaderData.headerHash, td[0].referencedHeaderData.header)
		headersPool.AddHeader(td[1].referencedHeaderData.headerHash, td[1].referencedHeaderData.header)
		headersPool.AddHeader(td[0].attestationHeaderData.headerHash, td[0].attestationHeaderData.header)
		headersPool.AddHeader(td[1].attestationHeaderData.headerHash, td[1].attestationHeaderData.header)
		arguments.HeadersForBlock.RequestShardHeaders(metaBlock)
		time.Sleep(100 * time.Millisecond)
		headersForBlock := mp.GetHdrForBlock()
		missingHdrs, missingProofs, missingAttesting := headersForBlock.GetMissingData()
		require.Equal(t, uint32(0), missingHdrs)
		require.Equal(t, uint32(0), missingProofs)
		require.Equal(t, uint32(0), missingAttesting)
		require.Len(t, headersForBlock.GetHeadersInfoMap(), 4)
		require.Equal(t, uint32(0), numCallsMissingAttestation.Load())
		require.Equal(t, uint32(0), numCallsMissingHeaders.Load())
	})
}

func createPoolsHolderForHeaderRequests() dataRetriever.HeadersPool {
	headersInPool := make(map[string]data.HeaderHandler)
	mutHeadersInPool := sync.RWMutex{}
	errNotFound := errors.New("header not found")

	return &pool.HeadersPoolStub{
		AddCalled: func(headerHash []byte, header data.HeaderHandler) {
			mutHeadersInPool.Lock()
			headersInPool[string(headerHash)] = header
			mutHeadersInPool.Unlock()
		},
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			mutHeadersInPool.RLock()
			defer mutHeadersInPool.RUnlock()
			if h, ok := headersInPool[string(hash)]; ok {
				return h, nil
			}
			return nil, errNotFound
		},
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
			mutHeadersInPool.RLock()
			defer mutHeadersInPool.RUnlock()
			for hash, h := range headersInPool {
				if h.GetNonce() == hdrNonce && h.GetShardID() == shardId {
					return []data.HeaderHandler{h}, [][]byte{[]byte(hash)}, nil
				}
			}
			return nil, nil, errNotFound
		},
	}
}

func createMetaProcessorArguments(t *testing.T, noOfShards uint32) *blockProcess.ArgMetaProcessor {
	poolMock := dataRetrieverMock.NewPoolsHolderMock()
	poolMock.Headers()
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
	coreComponents.Hash = &hashingMocks.HasherMock{}
	dataComponents.DataPool = poolMock
	dataComponents.Storage = initStore()
	bootstrapComponents.Coordinator = mock.NewMultiShardsCoordinatorMock(noOfShards)
	arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}

	startHeaders := createGenesisBlocks(bootstrapComponents.ShardCoordinator())
	arguments.BlockTracker = mock.NewBlockTrackerMock(bootstrapComponents.ShardCoordinator(), startHeaders)
	arguments.ArgBaseProcessor.RequestHandler = &testscommon.RequestHandlerStub{
		RequestShardHeaderByNonceCalled: func(shardID uint32, nonce uint64) {
			require.Fail(t, "should not have been called")
		},
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			require.Fail(t, "should not have been called")
		},

		RequestShardHeaderCalled: func(shardID uint32, hash []byte) {
			require.Fail(t, "should not have been called")
		},
		RequestMetaHeaderCalled: func(hash []byte) {
			require.Fail(t, "should not have been called")
		},
	}
	arguments.HeadersForBlock, _ = headerForBlock.NewHeadersForBlock(headerForBlock.ArgHeadersForBlock{
		DataPool:            arguments.DataComponents.Datapool(),
		RequestHandler:      arguments.RequestHandler,
		EnableEpochsHandler: arguments.CoreComponents.EnableEpochsHandler(),
		ShardCoordinator:    arguments.BootstrapComponents.ShardCoordinator(),
		BlockTracker:        arguments.BlockTracker,
		TxCoordinator:       arguments.TxCoordinator,
		RoundHandler:        arguments.CoreComponents.RoundHandler(),
		ExtraDelayForRequestBlockInfoInMilliseconds: 100,
		GenesisNonce: 0,
	})

	return &arguments
}

type shardHeaderData struct {
	header     *block.HeaderV2
	headerHash []byte
}

type shardTestData struct {
	referencedHeaderData  *shardHeaderData
	attestationHeaderData *shardHeaderData
}

func createTestData() map[uint32]*shardTestData {
	shard0Header1Hash := []byte("sh0TestHash1")
	shard0header2Hash := []byte("sh0TestHash2")
	shard1Header1Hash := []byte("sh1TestHash1")
	shard1header2Hash := []byte("sh1TestHash2")
	shard0ReferencedNonce := uint64(100)
	shard1ReferencedNonce := uint64(98)
	shard0AttestationNonce := shard0ReferencedNonce + 1
	shard1AttestationNonce := shard1ReferencedNonce + 1

	shardsTestData := map[uint32]*shardTestData{
		0: {
			referencedHeaderData: &shardHeaderData{
				header: &block.HeaderV2{
					Header: &block.Header{
						ShardID: 0,
						Round:   100,
						Nonce:   shard0ReferencedNonce,
					},
				},
				headerHash: shard0Header1Hash,
			},
			attestationHeaderData: &shardHeaderData{
				header: &block.HeaderV2{
					Header: &block.Header{
						ShardID:  0,
						Round:    101,
						Nonce:    shard0AttestationNonce,
						PrevHash: shard0Header1Hash,
					},
				},
				headerHash: shard0header2Hash,
			},
		},
		1: {
			referencedHeaderData: &shardHeaderData{
				header: &block.HeaderV2{
					Header: &block.Header{
						ShardID: 1,
						Round:   100,
						Nonce:   shard1ReferencedNonce,
					},
				},
				headerHash: shard1Header1Hash,
			},
			attestationHeaderData: &shardHeaderData{
				header: &block.HeaderV2{
					Header: &block.Header{
						ShardID:  1,
						Round:    101,
						Nonce:    shard1AttestationNonce,
						PrevHash: shard1Header1Hash,
					},
				},
				headerHash: shard1header2Hash,
			},
		},
	}

	return shardsTestData
}

func createShardInfo(referencedHeaders []*shardHeaderData) []block.ShardData {
	shardData := make([]block.ShardData, len(referencedHeaders))
	for i, h := range referencedHeaders {
		shardData[i] = block.ShardData{
			HeaderHash: h.headerHash,
			Round:      h.header.GetRound(),
			PrevHash:   h.header.GetPrevHash(),
			Nonce:      h.header.GetNonce(),
			ShardID:    h.header.GetShardID(),
		}
	}

	return shardData
}

func updateRequestsHandlerForCountingRequests(
	t *testing.T,
	arguments *blockProcess.ArgMetaProcessor,
	td map[uint32]*shardTestData,
	metaBlock *block.MetaBlock,
	numCallsMissingHeaders, numCallsMissingAttestation *atomic.Uint32,
) {
	requestHandler, ok := arguments.ArgBaseProcessor.RequestHandler.(*testscommon.RequestHandlerStub)
	require.True(t, ok)

	requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
		attestationNonce := td[shardID].attestationHeaderData.header.GetNonce()
		if nonce != attestationNonce {
			require.Fail(t, fmt.Sprintf("nonce should have been %d", attestationNonce))
		}
		numCallsMissingAttestation.Add(1)
	}
	requestHandler.RequestShardHeaderCalled = func(shardID uint32, hash []byte) {
		for _, sh := range metaBlock.ShardInfo {
			if bytes.Equal(sh.HeaderHash, hash) && sh.ShardID == shardID {
				numCallsMissingHeaders.Add(1)
				return
			}
		}

		require.Fail(t, fmt.Sprintf("header hash %s not found in meta block", hash))
	}
}
