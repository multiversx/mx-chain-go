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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	blockProcess "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
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

		headersForBlock := mp.GetHdrForBlock()
		numMissing, numAttestationMissing := mp.ComputeExistingAndRequestMissingShardHeaders(metaBlock)
		time.Sleep(100 * time.Millisecond)
		require.Equal(t, uint32(2), numMissing)
		require.Equal(t, uint32(2), headersForBlock.GetMissingHdrs())
		// before receiving all missing headers referenced in metaBlock, the number of missing attestations is not updated
		require.Equal(t, uint32(0), numAttestationMissing)
		require.Equal(t, uint32(0), headersForBlock.GetMissingFinalityAttestingHdrs())
		require.Len(t, headersForBlock.GetHdrHashAndInfo(), 2)
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
		numMissing, numAttestationMissing := mp.ComputeExistingAndRequestMissingShardHeaders(metaBlock)
		time.Sleep(100 * time.Millisecond)
		headersForBlock := mp.GetHdrForBlock()
		require.Equal(t, uint32(1), numMissing)
		require.Equal(t, uint32(1), headersForBlock.GetMissingHdrs())
		// before receiving all missing headers referenced in metaBlock, the number of missing attestations is not updated
		require.Equal(t, uint32(0), numAttestationMissing)
		require.Equal(t, uint32(0), headersForBlock.GetMissingFinalityAttestingHdrs())
		require.Len(t, headersForBlock.GetHdrHashAndInfo(), 2)
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
		numMissing, numAttestationMissing := mp.ComputeExistingAndRequestMissingShardHeaders(metaBlock)
		time.Sleep(100 * time.Millisecond)
		headersForBlock := mp.GetHdrForBlock()
		require.Equal(t, uint32(0), numMissing)
		require.Equal(t, uint32(0), headersForBlock.GetMissingHdrs())
		require.Equal(t, uint32(2), numAttestationMissing)
		require.Equal(t, uint32(2), headersForBlock.GetMissingFinalityAttestingHdrs())
		require.Len(t, headersForBlock.GetHdrHashAndInfo(), 2)
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
		numMissing, numAttestationMissing := mp.ComputeExistingAndRequestMissingShardHeaders(metaBlock)
		time.Sleep(100 * time.Millisecond)
		headersForBlock := mp.GetHdrForBlock()
		require.Equal(t, uint32(0), numMissing)
		require.Equal(t, uint32(0), headersForBlock.GetMissingHdrs())
		require.Equal(t, uint32(1), numAttestationMissing)
		require.Equal(t, uint32(1), headersForBlock.GetMissingFinalityAttestingHdrs())
		require.Len(t, headersForBlock.GetHdrHashAndInfo(), 3)
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
		numMissing, numAttestationMissing := mp.ComputeExistingAndRequestMissingShardHeaders(metaBlock)
		time.Sleep(100 * time.Millisecond)
		headersForBlock := mp.GetHdrForBlock()
		require.Equal(t, uint32(0), numMissing)
		require.Equal(t, uint32(0), headersForBlock.GetMissingHdrs())
		require.Equal(t, uint32(0), numAttestationMissing)
		require.Equal(t, uint32(0), headersForBlock.GetMissingFinalityAttestingHdrs())
		require.Len(t, headersForBlock.GetHdrHashAndInfo(), 4)
		require.Equal(t, uint32(0), numCallsMissingAttestation.Load())
		require.Equal(t, uint32(0), numCallsMissingHeaders.Load())
	})
}

func TestMetaProcessor_receivedShardHeader(t *testing.T) {
	t.Parallel()
	noOfShards := uint32(2)
	td := createTestData()

	t.Run("receiving the last used in block shard header", func(t *testing.T) {
		t.Parallel()

		numCalls := atomic.Uint32{}
		arguments := createMetaProcessorArguments(t, noOfShards)
		requestHandler, ok := arguments.ArgBaseProcessor.RequestHandler.(*testscommon.RequestHandlerStub)
		require.True(t, ok)

		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			attestationNonce := td[shardID].attestationHeaderData.header.GetNonce()
			if nonce != attestationNonce {
				require.Fail(t, fmt.Sprintf("nonce should have been %d", attestationNonce))
			}
			numCalls.Add(1)
		}

		mp, err := blockProcess.NewMetaProcessor(*arguments)
		require.Nil(t, err)
		require.NotNil(t, mp)

		hdrsForBlock := mp.GetHdrForBlock()
		hdrsForBlock.SetNumMissingHdrs(1)
		hdrsForBlock.SetNumMissingFinalityAttestingHdrs(0)
		hdrsForBlock.SetHighestHdrNonce(0, td[0].referencedHeaderData.header.GetNonce()-1)
		hdrsForBlock.SetHdrHashAndInfo(string(td[0].referencedHeaderData.headerHash), &blockProcess.HdrInfo{
			UsedInBlock: true,
			Hdr:         nil,
		})

		mp.ReceivedShardHeader(td[0].referencedHeaderData.header, td[0].referencedHeaderData.headerHash)

		time.Sleep(100 * time.Millisecond)
		require.Nil(t, err)
		require.NotNil(t, mp)
		require.Equal(t, uint32(1), numCalls.Load())
		require.Equal(t, uint32(1), hdrsForBlock.GetMissingFinalityAttestingHdrs())
	})

	t.Run("shard header used in block received, not latest", func(t *testing.T) {
		t.Parallel()

		numCalls := atomic.Uint32{}
		arguments := createMetaProcessorArguments(t, noOfShards)
		requestHandler, ok := arguments.ArgBaseProcessor.RequestHandler.(*testscommon.RequestHandlerStub)
		require.True(t, ok)

		// for requesting attestation header
		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			attestationNonce := td[shardID].attestationHeaderData.header.GetNonce()
			if nonce != attestationNonce {
				require.Fail(t, fmt.Sprintf("nonce should have been %d", attestationNonce))
			}
			numCalls.Add(1)
		}

		mp, err := blockProcess.NewMetaProcessor(*arguments)
		require.Nil(t, err)
		require.NotNil(t, mp)

		hdrsForBlock := mp.GetHdrForBlock()
		hdrsForBlock.SetNumMissingHdrs(2)
		hdrsForBlock.SetNumMissingFinalityAttestingHdrs(0)
		hdrsForBlock.SetHighestHdrNonce(0, td[1].referencedHeaderData.header.GetNonce()-1)
		hdrsForBlock.SetHdrHashAndInfo(string(td[1].referencedHeaderData.headerHash), &blockProcess.HdrInfo{
			UsedInBlock: true,
			Hdr:         nil,
		})

		mp.ReceivedShardHeader(td[1].referencedHeaderData.header, td[1].referencedHeaderData.headerHash)

		time.Sleep(100 * time.Millisecond)
		require.Nil(t, err)
		require.NotNil(t, mp)
		// not yet requested attestation blocks as still missing one header
		require.Equal(t, uint32(0), numCalls.Load())
		// not yet computed
		require.Equal(t, uint32(0), hdrsForBlock.GetMissingFinalityAttestingHdrs())
	})
	t.Run("all needed shard attestation headers received", func(t *testing.T) {
		t.Parallel()

		numCalls := atomic.Uint32{}
		arguments := createMetaProcessorArguments(t, noOfShards)

		poolsHolder, ok := arguments.DataComponents.Datapool().(*dataRetrieverMock.PoolsHolderMock)
		require.True(t, ok)

		headersPoolStub := createPoolsHolderForHeaderRequests()
		poolsHolder.SetHeadersPool(headersPoolStub)
		requestHandler, ok := arguments.ArgBaseProcessor.RequestHandler.(*testscommon.RequestHandlerStub)
		require.True(t, ok)

		// for requesting attestation header
		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			attestationNonce := td[shardID].attestationHeaderData.header.GetNonce()
			if nonce != attestationNonce {
				require.Fail(t, "nonce should have been %d", attestationNonce)
			}
			numCalls.Add(1)
		}

		mp, err := blockProcess.NewMetaProcessor(*arguments)
		require.Nil(t, err)
		require.NotNil(t, mp)

		hdrsForBlock := mp.GetHdrForBlock()
		hdrsForBlock.SetNumMissingHdrs(1)
		hdrsForBlock.SetNumMissingFinalityAttestingHdrs(0)
		hdrsForBlock.SetHighestHdrNonce(0, td[0].referencedHeaderData.header.GetNonce()-1)
		hdrsForBlock.SetHdrHashAndInfo(string(td[0].referencedHeaderData.headerHash), &blockProcess.HdrInfo{
			UsedInBlock: true,
			Hdr:         nil,
		})

		// receive the missing header
		headersPool := mp.GetDataPool().Headers()
		headersPool.AddHeader(td[0].referencedHeaderData.headerHash, td[0].referencedHeaderData.header)
		mp.ReceivedShardHeader(td[0].referencedHeaderData.header, td[0].referencedHeaderData.headerHash)

		time.Sleep(100 * time.Millisecond)
		require.Nil(t, err)
		require.NotNil(t, mp)
		require.Equal(t, uint32(1), numCalls.Load())
		require.Equal(t, uint32(1), hdrsForBlock.GetMissingFinalityAttestingHdrs())

		// needs to be done before receiving the last header otherwise it will
		// be blocked waiting on writing to the channel
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func(w *sync.WaitGroup) {
			receivedAllHeaders := checkReceivedAllHeaders(mp.ChannelReceiveAllHeaders())
			require.True(t, receivedAllHeaders)
			wg.Done()
		}(wg)

		// receive also the attestation header
		headersPool.AddHeader(td[0].attestationHeaderData.headerHash, td[0].attestationHeaderData.header)
		mp.ReceivedShardHeader(td[0].attestationHeaderData.header, td[0].attestationHeaderData.headerHash)
		wg.Wait()

		require.Equal(t, uint32(1), numCalls.Load())
		require.Equal(t, uint32(0), hdrsForBlock.GetMissingFinalityAttestingHdrs())
	})
	t.Run("all needed shard attestation headers received, when multiple shards headers missing", func(t *testing.T) {
		t.Parallel()

		numCalls := atomic.Uint32{}
		arguments := createMetaProcessorArguments(t, noOfShards)

		poolsHolder, ok := arguments.DataComponents.Datapool().(*dataRetrieverMock.PoolsHolderMock)
		require.True(t, ok)

		headersPoolStub := createPoolsHolderForHeaderRequests()
		poolsHolder.SetHeadersPool(headersPoolStub)
		requestHandler, ok := arguments.ArgBaseProcessor.RequestHandler.(*testscommon.RequestHandlerStub)
		require.True(t, ok)

		// for requesting attestation header
		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			attestationNonce := td[shardID].attestationHeaderData.header.GetNonce()
			if nonce != td[shardID].attestationHeaderData.header.GetNonce() {
				require.Fail(t, fmt.Sprintf("requested nonce for shard %d should have been %d", shardID, attestationNonce))
			}
			numCalls.Add(1)
		}

		mp, err := blockProcess.NewMetaProcessor(*arguments)
		require.Nil(t, err)
		require.NotNil(t, mp)

		hdrsForBlock := mp.GetHdrForBlock()
		hdrsForBlock.SetNumMissingHdrs(2)
		hdrsForBlock.SetNumMissingFinalityAttestingHdrs(0)
		hdrsForBlock.SetHighestHdrNonce(0, 99)
		hdrsForBlock.SetHighestHdrNonce(1, 97)
		hdrsForBlock.SetHdrHashAndInfo(string(td[0].referencedHeaderData.headerHash), &blockProcess.HdrInfo{
			UsedInBlock: true,
			Hdr:         nil,
		})
		hdrsForBlock.SetHdrHashAndInfo(string(td[1].referencedHeaderData.headerHash), &blockProcess.HdrInfo{
			UsedInBlock: true,
			Hdr:         nil,
		})

		// receive the missing header for shard 0
		headersPool := mp.GetDataPool().Headers()
		headersPool.AddHeader(td[0].referencedHeaderData.headerHash, td[0].referencedHeaderData.header)
		mp.ReceivedShardHeader(td[0].referencedHeaderData.header, td[0].referencedHeaderData.headerHash)

		time.Sleep(100 * time.Millisecond)
		require.Nil(t, err)
		require.NotNil(t, mp)
		// the attestation header for shard 0 is not requested as the attestation header for shard 1 is missing
		// TODO: refactor request logic to request missing attestation headers as soon as possible
		require.Equal(t, uint32(0), numCalls.Load())
		require.Equal(t, uint32(0), hdrsForBlock.GetMissingFinalityAttestingHdrs())

		// receive the missing header for shard 1
		headersPool.AddHeader(td[1].referencedHeaderData.headerHash, td[1].referencedHeaderData.header)
		mp.ReceivedShardHeader(td[1].referencedHeaderData.header, td[1].referencedHeaderData.headerHash)

		time.Sleep(100 * time.Millisecond)
		require.Nil(t, err)
		require.NotNil(t, mp)
		require.Equal(t, uint32(2), numCalls.Load())
		require.Equal(t, uint32(2), hdrsForBlock.GetMissingFinalityAttestingHdrs())

		// needs to be done before receiving the last header otherwise it will
		// be blocked writing to a channel no one is reading from
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func(w *sync.WaitGroup) {
			receivedAllHeaders := checkReceivedAllHeaders(mp.ChannelReceiveAllHeaders())
			require.True(t, receivedAllHeaders)
			wg.Done()
		}(wg)

		// receive also the attestation header
		headersPool.AddHeader(td[0].attestationHeaderData.headerHash, td[0].attestationHeaderData.header)
		mp.ReceivedShardHeader(td[0].attestationHeaderData.header, td[0].attestationHeaderData.headerHash)

		headersPool.AddHeader(td[1].attestationHeaderData.headerHash, td[1].attestationHeaderData.header)
		mp.ReceivedShardHeader(td[1].attestationHeaderData.header, td[1].attestationHeaderData.headerHash)
		wg.Wait()

		time.Sleep(100 * time.Millisecond)
		// the receive of an attestation header, if not the last one, will trigger a new request of missing attestation headers
		// TODO: refactor request logic to not request recently already requested headers
		require.Equal(t, uint32(3), numCalls.Load())
		require.Equal(t, uint32(0), hdrsForBlock.GetMissingFinalityAttestingHdrs())
	})
}

func checkReceivedAllHeaders(channelReceiveAllHeaders chan bool) bool {
	select {
	case <-time.After(100 * time.Millisecond):
		return false
	case <-channelReceiveAllHeaders:
		return true
	}
}

func createPoolsHolderForHeaderRequests() dataRetriever.HeadersPool {
	headersInPool := make(map[string]data.HeaderHandler)
	mutHeadersInPool := sync.RWMutex{}
	errNotFound := errors.New("header not found")

	return &pool.HeadersCacherStub{
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
