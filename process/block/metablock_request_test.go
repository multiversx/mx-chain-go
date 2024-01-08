package block_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/block"
	blockProcess "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMetaProcessorArguments(t *testing.T, noOfShards uint32) *blockProcess.ArgMetaProcessor {
	pool := dataRetrieverMock.NewPoolsHolderMock()
	pool.Headers()
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
	coreComponents.Hash = &hashingMocks.HasherMock{}
	dataComponents.DataPool = pool
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

func TestMetaProcessor_receivedShardHeader(t *testing.T) {
	noOfShards := uint32(5)
	header1Hash := []byte("testHash1")
	header2Hash := []byte("testHash2")

	header1 := &block.HeaderV2{
		Header: &block.Header{
			ShardID: 0,
			Round:   100,
			Nonce:   100,
		},
	}

	header2 := &block.HeaderV2{
		Header: &block.Header{
			ShardID:  0,
			Round:    101,
			Nonce:    101,
			PrevHash: header1Hash,
		},
	}

	t.Run("receiving the last used in block shard header", func(t *testing.T) {
		numCalls := atomic.Uint32{}
		arguments := createMetaProcessorArguments(t, noOfShards)
		requestHandler, ok := arguments.ArgBaseProcessor.RequestHandler.(*testscommon.RequestHandlerStub)
		require.True(t, ok)

		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			if nonce != 101 {
				require.Fail(t, "nonce should have been 101")
			}
			numCalls.Add(1)
		}

		mp, err := blockProcess.NewMetaProcessor(*arguments)
		require.Nil(t, err)
		require.NotNil(t, mp)

		hdrsForBlock := mp.GetHdrForBlock()
		hdrsForBlock.SetNumMissingHdrs(1)
		hdrsForBlock.SetNumMissingFinalityAttestingHdrs(0)
		hdrsForBlock.SetHighestHdrNonce(0, 99)
		hdrsForBlock.SetHdrHashAndInfo(string(header1Hash), &blockProcess.HdrInfo{
			UsedInBlock: true,
			Hdr:         nil,
		})

		mp.ReceivedShardHeader(header1, header1Hash)

		time.Sleep(100 * time.Millisecond)
		require.Nil(t, err)
		require.NotNil(t, mp)
		require.Equal(t, uint32(1), numCalls.Load())
		require.Equal(t, uint32(1), hdrsForBlock.GetMissingFinalityAttestingHdrs())
	})

	t.Run("shard header used in block received, not latest", func(t *testing.T) {
		numCalls := atomic.Uint32{}
		arguments := createMetaProcessorArguments(t, noOfShards)
		requestHandler, ok := arguments.ArgBaseProcessor.RequestHandler.(*testscommon.RequestHandlerStub)
		require.True(t, ok)

		// for requesting attestation header
		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			if nonce != 101 {
				require.Fail(t, "nonce should have been 101")
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
		hdrsForBlock.SetHdrHashAndInfo(string(header1Hash), &blockProcess.HdrInfo{
			UsedInBlock: true,
			Hdr:         nil,
		})

		mp.ReceivedShardHeader(header1, header1Hash)

		time.Sleep(100 * time.Millisecond)
		require.Nil(t, err)
		require.NotNil(t, mp)
		// not yet requested attestation blocks as still missing one header
		require.Equal(t, uint32(0), numCalls.Load())
		// not yet computed
		require.Equal(t, uint32(0), hdrsForBlock.GetMissingFinalityAttestingHdrs())
	})
	t.Run("shard attestation header received", func(t *testing.T) {
		numCalls := atomic.Uint32{}
		arguments := createMetaProcessorArguments(t, noOfShards)
		arguments.DataComponents
		requestHandler, ok := arguments.ArgBaseProcessor.RequestHandler.(*testscommon.RequestHandlerStub)
		require.True(t, ok)

		// for requesting attestation header
		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
			if nonce != 101 {
				require.Fail(t, "nonce should have been 101")
			}
			numCalls.Add(1)
		}

		mp, err := blockProcess.NewMetaProcessor(*arguments)
		require.Nil(t, err)
		require.NotNil(t, mp)

		hdrsForBlock := mp.GetHdrForBlock()
		hdrsForBlock.SetNumMissingHdrs(1)
		hdrsForBlock.SetNumMissingFinalityAttestingHdrs(0)
		hdrsForBlock.SetHighestHdrNonce(0, 99)
		hdrsForBlock.SetHdrHashAndInfo(string(header1Hash), &blockProcess.HdrInfo{
			UsedInBlock: true,
			Hdr:         nil,
		})

		headersPool := mp.GetDataPool().Headers()
		// mp.ReceivedShardHeader(header1, header1Hash) is called through the headersPool.AddHeader callback

		time.Sleep(100 * time.Millisecond)
		require.Nil(t, err)
		require.NotNil(t, mp)
		require.Equal(t, uint32(1), numCalls.Load())
		require.Equal(t, uint32(1), hdrsForBlock.GetMissingFinalityAttestingHdrs())

		// receive also the attestation header
		headersPool.AddHeader(header2Hash, header2)
		// mp.ReceivedShardHeader(header2, header2Hash) is called through the headersPool.AddHeader callback
		require.Equal(t, uint32(1), numCalls.Load())
		require.Equal(t, uint32(0), hdrsForBlock.GetMissingFinalityAttestingHdrs())
	})
}
