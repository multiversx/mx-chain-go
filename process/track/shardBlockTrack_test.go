package track_test

import (
	"bytes"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
)

type requestHandlerWithIntervalHook struct {
	testscommon.RequestHandlerStub
	intervalHook func()
}

func (handler *requestHandlerWithIntervalHook) RequestInterval() time.Duration {
	if handler.intervalHook != nil {
		handler.intervalHook()
	}

	return time.Second
}

func TestShardBlockTrack_ComputeCrossInfo(t *testing.T) {
	t.Parallel()

	t.Run("legacy meta block reads pending miniblocks from shard info", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateShardTrackerMockArguments()
		sbt, err := track.NewShardBlockTrack(shardArguments)
		require.Nil(t, err)

		metaBlock := &block.MetaBlock{
			ShardInfo: []block.ShardData{
				{ShardID: 0, NumPendingMiniBlocks: 3, LastIncludedMetaNonce: 11},
				{ShardID: 1, NumPendingMiniBlocks: 5, LastIncludedMetaNonce: 22},
			},
		}

		sbt.ComputeCrossInfo([]data.HeaderHandler{metaBlock})

		require.Equal(t, uint32(3), sbt.GetNumPendingMiniBlocks(0))
		require.Equal(t, uint64(11), sbt.GetLastShardProcessedMetaNonce(0))
		require.Equal(t, uint32(5), sbt.GetNumPendingMiniBlocks(1))
		require.Equal(t, uint64(22), sbt.GetLastShardProcessedMetaNonce(1))
	})

	t.Run("V3 meta block reads pending miniblocks from shard info proposal", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateShardTrackerMockArguments()
		sbt, err := track.NewShardBlockTrack(shardArguments)
		require.Nil(t, err)

		metaBlock := &block.MetaBlockV3{
			ShardInfo: []block.ShardData{
				{ShardID: 0, NumPendingMiniBlocks: 0, LastIncludedMetaNonce: 11},
				{ShardID: 1, NumPendingMiniBlocks: 0, LastIncludedMetaNonce: 22},
			},
			ShardInfoProposal: []block.ShardDataProposal{
				{ShardID: 0, NumPendingMiniBlocks: 3},
				{ShardID: 1, NumPendingMiniBlocks: 5},
			},
		}

		sbt.ComputeCrossInfo([]data.HeaderHandler{metaBlock})

		require.Equal(t, uint32(3), sbt.GetNumPendingMiniBlocks(0))
		require.Equal(t, uint64(11), sbt.GetLastShardProcessedMetaNonce(0))
		require.Equal(t, uint32(5), sbt.GetNumPendingMiniBlocks(1))
		require.Equal(t, uint64(22), sbt.GetLastShardProcessedMetaNonce(1))
	})
}

func TestShardBlockTrack_V3HeldFinalReferenceResolvedAfterHeaderArrival(t *testing.T) {
	t.Parallel()

	var headerHandlers []func(data.HeaderHandler, []byte)
	var availableHeaders sync.Map
	headersPool := &pool.HeadersPoolStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			if header, ok := availableHeaders.Load(string(hash)); ok {
				return header.(data.HeaderHandler), nil
			}
			return nil, errors.New("missing header")
		},
		RegisterHandlerCalled: func(handler func(data.HeaderHandler, []byte)) {
			headerHandlers = append(headerHandlers, handler)
		},
	}
	var headerRequests atomic.Int32
	var proofRequests atomic.Int32
	arguments := CreateShardTrackerMockArguments()
	arguments.PoolsHolder = &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return headersPool
		},
		ProofsCalled: func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{}
		},
	}
	arguments.RequestHandler = &testscommon.RequestHandlerStub{
		RequestShardHeaderForEpochCalled: func(_ uint32, _ []byte, _ uint32) {
			headerRequests.Add(1)
		},
		RequestEquivalentProofByHashForEpochCalled: func(_ uint32, _ []byte, _ uint32) {
			proofRequests.Add(1)
		},
	}

	sbt, err := track.NewShardBlockTrack(arguments)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sbt.Close())
	})
	require.Len(t, headerHandlers, 1)

	headerHash := []byte("held-final-shard-header")
	metaBlock := &block.MetaBlockV3{
		Nonce: 1,
		ShardInfoProposal: []block.ShardDataProposal{{
			HeaderHash: headerHash,
			ShardID:    0,
			Nonce:      7,
			Epoch:      2,
		}},
	}
	metaHash, err := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, metaBlock)
	require.NoError(t, err)
	sbt.AddCrossNotarizedHeader(core.MetachainShardId, metaBlock, metaHash)
	require.Empty(t, sbt.GetSelfHeaders(metaBlock))
	require.Equal(t, int32(1), headerRequests.Load())
	require.Equal(t, int32(1), proofRequests.Load())
	require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())
	require.Empty(t, sbt.GetSelfHeaders(metaBlock))
	require.Equal(t, int32(1), headerRequests.Load())
	require.Equal(t, int32(1), proofRequests.Load())

	type notification struct {
		shardID    uint32
		numHeaders int
		hash       []byte
	}
	var notifications atomic.Int32
	notified := make(chan notification, 2)
	sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(
		shardID uint32,
		headers []data.HeaderHandler,
		hashes [][]byte,
	) {
		notifications.Add(1)
		notified <- notification{shardID: shardID, numHeaders: len(headers), hash: hashes[0]}
	})

	header := &block.HeaderV3{ShardID: 0, Nonce: 7, Epoch: 2}
	headerHandlers[0](header, headerHash)
	headerHandlers[0](header, headerHash)
	select {
	case received := <-notified:
		require.Equal(t, uint32(core.MetachainShardId), received.shardID)
		require.Equal(t, 1, received.numHeaders)
		require.Equal(t, headerHash, received.hash)
	case <-time.After(time.Second):
		require.FailNow(t, "late held-final notification was not delivered")
	}
	require.Equal(t, int32(1), notifications.Load())

	require.Empty(t, sbt.GetSelfHeaders(metaBlock))
	require.Equal(t, int32(1), headerRequests.Load())
	require.Equal(t, int32(1), proofRequests.Load())

	sbt.RestoreToGenesis()
	require.Zero(t, sbt.NumPendingSelfHeaders())
	availableHeaders.Store(string(headerHash), header)
	restoredHeaders := sbt.GetSelfHeaders(metaBlock)
	require.Len(t, restoredHeaders, 1)
	require.Equal(t, headerHash, restoredHeaders[0].Hash)
	require.Same(t, header, restoredHeaders[0].Header)

	secondHash := []byte("second-held-final-shard-header")
	secondMetaBlock := &block.MetaBlockV3{
		Nonce: 2,
		ShardInfoProposal: []block.ShardDataProposal{{
			HeaderHash: secondHash,
			ShardID:    0,
			Nonce:      8,
			Epoch:      2,
		}},
	}
	secondMetaHash, err := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, secondMetaBlock)
	require.NoError(t, err)
	sbt.AddCrossNotarizedHeader(core.MetachainShardId, secondMetaBlock, secondMetaHash)
	require.Empty(t, sbt.GetSelfHeaders(secondMetaBlock))
	secondHeader := &block.HeaderV3{ShardID: 0, Nonce: 8, Epoch: 2}
	availableHeaders.Store(string(secondHash), secondHeader)

	var returned []*track.HeaderInfo
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	go func() {
		defer waitGroup.Done()
		headerHandlers[0](secondHeader, secondHash)
	}()
	go func() {
		defer waitGroup.Done()
		returned = sbt.GetSelfHeaders(secondMetaBlock)
	}()
	waitGroup.Wait()
	if len(returned) == 0 {
		select {
		case received := <-notified:
			require.Equal(t, uint32(core.MetachainShardId), received.shardID)
			require.Equal(t, 1, received.numHeaders)
			require.Equal(t, secondHash, received.hash)
		case <-time.After(time.Second):
			require.FailNow(t, "concurrent late held-final notification was not delivered")
		}
	}
	require.Equal(t, 1, len(returned)+int(notifications.Load()-1))

	serializedHash := []byte("serialized-held-final-shard-header")
	serializedMetaBlock := &block.MetaBlockV3{
		Nonce: 3,
		ShardInfoProposal: []block.ShardDataProposal{{
			HeaderHash: serializedHash,
			ShardID:    0,
			Nonce:      9,
			Epoch:      2,
		}},
	}
	require.Empty(t, sbt.GetSelfHeaders(serializedMetaBlock))
	handlerStarted := make(chan struct{})
	releaseHandler := make(chan struct{})
	var releaseHandlerOnce sync.Once
	releasePendingHandler := func() {
		releaseHandlerOnce.Do(func() {
			close(releaseHandler)
		})
	}
	t.Cleanup(releasePendingHandler)
	sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(
		_ uint32,
		_ []data.HeaderHandler,
		hashes [][]byte,
	) {
		if !bytes.Equal(hashes[0], serializedHash) {
			return
		}

		close(handlerStarted)
		<-releaseHandler
	})
	headerDeliveryDone := make(chan struct{})
	go func() {
		headerHandlers[0](&block.HeaderV3{ShardID: 0, Nonce: 9, Epoch: 2}, serializedHash)
		close(headerDeliveryDone)
	}()
	select {
	case <-handlerStarted:
	case <-time.After(time.Second):
		require.FailNow(t, "late held-final notification did not start")
	}
	select {
	case <-headerDeliveryDone:
	case <-time.After(time.Second):
		require.FailNow(t, "late held-final delivery blocked on notification")
	}
	rollbackDone := make(chan struct{})
	go func() {
		sbt.RemoveLastNotarizedHeaders()
		close(rollbackDone)
	}()
	releasePendingHandler()
	select {
	case <-rollbackDone:
	case <-time.After(time.Second):
		require.FailNow(t, "rollback did not wait for the admitted notification")
	}

	retainedHeaderHash := []byte("retained-held-final-shard-header")
	retainedMetaBlock := &block.MetaBlockV3{
		Nonce: 4,
		ShardInfoProposal: []block.ShardDataProposal{{
			HeaderHash: retainedHeaderHash,
			ShardID:    0,
			Nonce:      9,
			Epoch:      2,
		}},
	}
	require.Empty(t, sbt.GetSelfHeaders(retainedMetaBlock))

	rolledBackHeaderHash := []byte("rolled-back-held-final-shard-header")
	rolledBackMetaBlock := &block.MetaBlockV3{
		Nonce: 5,
		ShardInfoProposal: []block.ShardDataProposal{
			{
				HeaderHash: retainedHeaderHash,
				ShardID:    0,
				Nonce:      9,
				Epoch:      2,
			},
			{
				HeaderHash: rolledBackHeaderHash,
				ShardID:    0,
				Nonce:      10,
				Epoch:      2,
			},
		},
	}
	require.Empty(t, sbt.GetSelfHeaders(rolledBackMetaBlock))
	notificationsBeforeRollback := notifications.Load()
	sbt.RemoveLastNotarizedHeaders()
	headerHandlers[0](&block.HeaderV3{ShardID: 0, Nonce: 10, Epoch: 2}, rolledBackHeaderHash)
	require.Equal(t, notificationsBeforeRollback, notifications.Load())
	require.Zero(t, sbt.NumPendingSelfHeaders())
	headerHandlers[0](&block.HeaderV3{ShardID: 0, Nonce: 9, Epoch: 2}, retainedHeaderHash)
	require.Equal(t, notificationsBeforeRollback, notifications.Load())

	limit := sbt.GetMaxNumHeadersToKeepPerShard()
	for index := 0; index <= limit; index++ {
		hash := []byte{byte(index), byte(index >> 8)}
		unresolvedMetaBlock := &block.MetaBlockV3{
			ShardInfoProposal: []block.ShardDataProposal{{
				HeaderHash: hash,
				ShardID:    0,
				Nonce:      uint64(index + 100),
				Epoch:      2,
			}},
		}
		require.Empty(t, sbt.GetSelfHeaders(unresolvedMetaBlock))
	}
	require.Equal(t, int64(limit), sbt.NumPendingSelfHeaders())

	sbt.RestoreToGenesis()
	require.Zero(t, sbt.NumPendingSelfHeaders())
}

func TestShardBlockTrack_ResetRejectsInFlightPendingReference(t *testing.T) {
	t.Parallel()

	for _, testName := range []string{"rollback", "restore to genesis"} {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			availableHeaderHash := []byte("available-held-final-reference")
			availableHeader := &block.HeaderV3{ShardID: 0, Nonce: 6, Epoch: 2}
			var headerHandler func(data.HeaderHandler, []byte)
			headersPool := &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					if bytes.Equal(hash, availableHeaderHash) {
						return availableHeader, nil
					}

					return nil, errors.New("missing header")
				},
				RegisterHandlerCalled: func(handler func(data.HeaderHandler, []byte)) {
					headerHandler = handler
				},
			}
			arguments := CreateShardTrackerMockArguments()
			arguments.PoolsHolder = &dataRetrieverMock.PoolsHolderStub{
				HeadersCalled: func() dataRetriever.HeadersPool {
					return headersPool
				},
				ProofsCalled: func() dataRetriever.ProofsPool {
					return &dataRetrieverMock.ProofsPoolMock{}
				},
			}

			scanStarted := make(chan struct{})
			resumeScan := make(chan struct{})
			var resumeScanOnce sync.Once
			resumePendingScan := func() {
				resumeScanOnce.Do(func() {
					close(resumeScan)
				})
			}
			t.Cleanup(resumePendingScan)
			var startOnce sync.Once
			requestHandler := &requestHandlerWithIntervalHook{}
			requestHandler.intervalHook = func() {
				startOnce.Do(func() {
					close(scanStarted)
					<-resumeScan
				})
			}
			var headerRequests atomic.Int32
			var proofRequests atomic.Int32
			requestHandler.RequestShardHeaderForEpochCalled = func(_ uint32, _ []byte, _ uint32) {
				headerRequests.Add(1)
			}
			requestHandler.RequestEquivalentProofByHashForEpochCalled = func(_ uint32, _ []byte, _ uint32) {
				proofRequests.Add(1)
			}
			arguments.RequestHandler = requestHandler

			sbt, err := track.NewShardBlockTrack(arguments)
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, sbt.Close())
			})
			require.NotNil(t, headerHandler)

			headerHash := []byte("in-flight-held-final-reference")
			header := &block.HeaderV3{ShardID: 0, Nonce: 7, Epoch: 2}
			metaBlock := &block.MetaBlockV3{
				ShardInfoProposal: []block.ShardDataProposal{
					{
						HeaderHash: availableHeaderHash,
						ShardID:    availableHeader.ShardID,
						Nonce:      availableHeader.Nonce,
						Epoch:      availableHeader.Epoch,
					},
					{
						HeaderHash: headerHash,
						ShardID:    header.ShardID,
						Nonce:      header.Nonce,
						Epoch:      header.Epoch,
					},
				},
			}

			notified := make(chan struct{}, 1)
			sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(
				_ uint32,
				_ []data.HeaderHandler,
				_ [][]byte,
			) {
				notified <- struct{}{}
			})

			scanDone := make(chan []*track.HeaderInfo, 1)
			go func() {
				scanDone <- sbt.GetSelfHeaders(metaBlock)
			}()
			select {
			case <-scanStarted:
			case <-time.After(time.Second):
				require.FailNow(t, "pending-reference scan did not start")
			}

			resetDone := make(chan struct{})
			go func() {
				if testName == "rollback" {
					sbt.RemoveLastNotarizedHeaders()
				} else {
					sbt.RestoreToGenesis()
				}
				close(resetDone)
			}()
			select {
			case <-resetDone:
			case <-time.After(time.Second):
				require.FailNow(t, "reset was blocked by a scan that had not admitted pending state")
			}
			resumePendingScan()

			select {
			case selfHeaders := <-scanDone:
				require.Empty(t, selfHeaders)
			case <-time.After(time.Second):
				require.FailNow(t, "pending-reference scan did not finish")
			}
			require.Zero(t, sbt.NumPendingSelfHeaders())
			require.Zero(t, headerRequests.Load())
			require.Zero(t, proofRequests.Load())

			headerHandler(header, headerHash)
			select {
			case <-notified:
				require.FailNow(t, "stale pending reference was notified after reset")
			case <-time.After(50 * time.Millisecond):
			}
		})
	}
}

func TestShardBlockTrack_HeaderArrivesBeforePendingReferenceInsertion(t *testing.T) {
	t.Parallel()

	var headerHandler func(data.HeaderHandler, []byte)
	var availableHeaders sync.Map
	headersPool := &pool.HeadersPoolStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			if header, ok := availableHeaders.Load(string(hash)); ok {
				return header.(data.HeaderHandler), nil
			}

			return nil, errors.New("missing header")
		},
		RegisterHandlerCalled: func(handler func(data.HeaderHandler, []byte)) {
			headerHandler = handler
		},
	}
	arguments := CreateShardTrackerMockArguments()
	arguments.PoolsHolder = &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return headersPool
		},
		ProofsCalled: func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{}
		},
	}

	var headerRequests atomic.Int32
	var proofRequests atomic.Int32
	requestHandler := &requestHandlerWithIntervalHook{}
	requestHandler.RequestShardHeaderForEpochCalled = func(_ uint32, _ []byte, _ uint32) {
		headerRequests.Add(1)
	}
	requestHandler.RequestEquivalentProofByHashForEpochCalled = func(_ uint32, _ []byte, _ uint32) {
		proofRequests.Add(1)
	}
	arguments.RequestHandler = requestHandler

	sbt, err := track.NewShardBlockTrack(arguments)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sbt.Close())
	})
	require.NotNil(t, headerHandler)

	headerHash := []byte("header-arriving-before-pending-insertion")
	header := &block.HeaderV3{ShardID: 0, Nonce: 7, Epoch: 2}
	metaBlock := &block.MetaBlockV3{
		ShardInfoProposal: []block.ShardDataProposal{{
			HeaderHash: headerHash,
			ShardID:    0,
			Nonce:      header.Nonce,
			Epoch:      header.Epoch,
		}},
	}

	var delivered atomic.Bool
	requestHandler.intervalHook = func() {
		if !delivered.CompareAndSwap(false, true) {
			return
		}

		availableHeaders.Store(string(headerHash), header)
		headerHandler(header, headerHash)
	}

	var notifications atomic.Int32
	notificationDone := make(chan struct{}, 1)
	sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(
		_ uint32,
		_ []data.HeaderHandler,
		_ [][]byte,
	) {
		notifications.Add(1)
		notificationDone <- struct{}{}
	})

	selfHeaders := sbt.GetSelfHeaders(metaBlock)
	require.Empty(t, selfHeaders)
	select {
	case <-notificationDone:
	case <-time.After(time.Second):
		require.FailNow(t, "header was not notified after concurrent pending insertion")
	}
	require.Zero(t, sbt.NumPendingSelfHeaders())
	require.Equal(t, int32(1), notifications.Load())
	require.Equal(t, int32(1), headerRequests.Load())
	require.Equal(t, int32(1), proofRequests.Load())

	require.Empty(t, sbt.GetSelfHeaders(metaBlock))
	require.Zero(t, sbt.NumPendingSelfHeaders())
	require.Equal(t, int32(1), headerRequests.Load())
	require.Equal(t, int32(1), proofRequests.Load())
}
