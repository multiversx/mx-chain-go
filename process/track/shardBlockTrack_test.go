package track_test

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	processMocks "github.com/multiversx/mx-chain-go/testscommon/processMocks"
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

func TestShardBlockTrack_SourceAwareImmediatePublication(t *testing.T) {
	t.Parallel()

	type sourceAwareTracker interface {
		GetSelfHeadersWithSource(header data.HeaderHandler, hash []byte) []*track.SelfHeaderInfo
		PublishSelfNotarizedFromCrossHeaders(shardID uint32, headersInfo []*track.SelfHeaderInfo)
		RegisterSelfNotarizedFromCrossHeadersHandler(handler func(uint32, []data.HeaderHandler, [][]byte))
		AddCrossNotarizedHeader(shardID uint32, header data.HeaderHandler, hash []byte)
		AddTrackedHeader(header data.HeaderHandler, hash []byte)
		IsSettledCrossHeader(header data.HeaderHandler, hash []byte) bool
		ComputeLongestChain(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte)
		RemoveLastNotarizedHeaders()
		NumPendingSelfHeaders() int64
		NumPendingSources(hash []byte) int
		SetMetaFinalityView(view process.MetaFinalityView)
		Close() error
	}

	createTracker := func(t *testing.T) (
		sourceAwareTracker,
		*block.MetaBlockV3,
		[]byte,
		*atomic.Bool,
		*atomic.Bool,
		func(data.HeaderHandler, []byte),
		*requestHandlerWithIntervalHook,
	) {
		t.Helper()

		const sourceNonce = uint64(7)
		const sourceRound = uint64(11)
		parentHash := []byte("source-parent")
		sourceHash := []byte("source-meta")
		shardHash := []byte("referenced-shard")
		childHash := []byte("source-child")
		parent := &block.MetaBlockV3{Nonce: sourceNonce - 1, Round: sourceRound - 1}
		shardHeader := &block.HeaderV3{ShardID: 0, Nonce: 5, Epoch: 2}
		source := &block.MetaBlockV3{
			Nonce:    sourceNonce,
			Round:    sourceRound,
			PrevHash: parentHash,
			ShardInfoProposal: []block.ShardDataProposal{{
				HeaderHash: shardHash,
				ShardID:    0,
				Nonce:      shardHeader.Nonce,
				Epoch:      shardHeader.Epoch,
			}},
		}
		child := &block.MetaBlockV3{
			Nonce:    sourceNonce + 1,
			Round:    sourceRound + 1,
			PrevHash: sourceHash,
		}

		shardHeaderAvailable := &atomic.Bool{}
		shardHeaderAvailable.Store(true)
		var headerHandler func(data.HeaderHandler, []byte)
		headersPool := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				switch {
				case bytes.Equal(hash, parentHash):
					return parent, nil
				case bytes.Equal(hash, sourceHash):
					return source, nil
				case bytes.Equal(hash, childHash):
					return child, nil
				case bytes.Equal(hash, shardHash) && shardHeaderAvailable.Load():
					return shardHeader, nil
				default:
					return nil, errors.New("missing header")
				}
			},
			RegisterHandlerCalled: func(handler func(data.HeaderHandler, []byte)) {
				headerHandler = handler
			},
			GetHeaderByNonceAndShardIdCalled: func(nonce uint64, shardID uint32) ([]data.HeaderHandler, [][]byte, error) {
				if shardID == core.MetachainShardId && nonce == child.GetNonce() {
					return []data.HeaderHandler{child}, [][]byte{childHash}, nil
				}

				return nil, nil, errors.New("missing headers")
			},
		}
		hasSibling := &atomic.Bool{}
		proofsPool := &dataRetrieverMock.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, hash []byte) bool {
				return shardID == core.MetachainShardId &&
					(bytes.Equal(hash, parentHash) || bytes.Equal(hash, sourceHash) || bytes.Equal(hash, childHash))
			},
			GetProofsByNonceCalled: func(nonce uint64, shardID uint32) ([]data.HeaderProofHandler, error) {
				if nonce != sourceNonce || shardID != core.MetachainShardId {
					return nil, errors.New("missing proof")
				}

				proofs := []data.HeaderProofHandler{&block.HeaderProof{
					HeaderHash:    sourceHash,
					HeaderNonce:   sourceNonce,
					HeaderRound:   sourceRound,
					HeaderShardId: core.MetachainShardId,
				}}
				if hasSibling.Load() {
					proofs = append(proofs, &block.HeaderProof{
						HeaderHash:    []byte("missing-sibling"),
						HeaderNonce:   sourceNonce,
						HeaderRound:   sourceRound + 1,
						HeaderShardId: core.MetachainShardId,
					})
				}

				return proofs, nil
			},
		}

		arguments := CreateShardTrackerMockArguments()
		arguments.HeaderValidator = &processMocks.HeaderValidatorMock{}
		arguments.PoolsHolder = &dataRetrieverMock.PoolsHolderStub{
			HeadersCalled: func() dataRetriever.HeadersPool {
				return headersPool
			},
			ProofsCalled: func() dataRetriever.ProofsPool {
				return proofsPool
			},
		}
		arguments.ProofsPool = proofsPool
		requestHandler := &requestHandlerWithIntervalHook{}
		arguments.RequestHandler = requestHandler

		sbt, err := track.NewShardBlockTrack(arguments)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, sbt.Close())
		})
		sbt.AddCrossNotarizedHeader(core.MetachainShardId, parent, parentHash)
		sbt.AddTrackedHeader(source, sourceHash)
		sbt.AddTrackedHeader(child, childHash)
		require.True(t, sbt.IsSettledCrossHeader(source, sourceHash))
		continuation, continuationHashes := sbt.ComputeLongestChain(core.MetachainShardId, parent)
		require.NotEmpty(t, continuation)
		require.Equal(t, sourceHash, continuationHashes[0])

		return sbt, source, sourceHash, hasSibling, shardHeaderAvailable, headerHandler, requestHandler
	}

	t.Run("held final source is published", func(t *testing.T) {
		t.Parallel()

		sbt, source, sourceHash, _, _, _, _ := createTracker(t)
		notified := make(chan struct{}, 1)
		sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
			notified <- struct{}{}
		})

		headersInfo := sbt.GetSelfHeadersWithSource(source, sourceHash)
		require.Len(t, headersInfo, 1)
		sbt.PublishSelfNotarizedFromCrossHeaders(core.MetachainShardId, headersInfo)

		select {
		case <-notified:
		case <-time.After(time.Second):
			require.FailNow(t, "source-aware notification was not delivered")
		}

		headersInfo = sbt.GetSelfHeadersWithSource(source, sourceHash)
		require.Empty(t, headersInfo)
		sbt.PublishSelfNotarizedFromCrossHeaders(core.MetachainShardId, headersInfo)
		sbt.RemoveLastNotarizedHeaders()
		select {
		case <-notified:
			require.FailNow(t, "published authority was notified twice")
		default:
		}
	})

	t.Run("dead source is not published even when it also reports held final", func(t *testing.T) {
		t.Parallel()

		sbt, source, sourceHash, _, _, _, _ := createTracker(t)
		sbt.SetMetaFinalityView(&testscommon.MetaFinalityViewStub{
			IsMetaHeaderSettlementReadyCalled: func(_ data.HeaderHandler, _ []byte) bool {
				return true
			},
			IsDeadMetaBlockCalled: func(_ []byte, _ uint64) bool {
				return true
			},
		})
		var notifications atomic.Int32
		sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
			notifications.Add(1)
		})

		headersInfo := sbt.GetSelfHeadersWithSource(source, sourceHash)
		require.Len(t, headersInfo, 1)
		sbt.PublishSelfNotarizedFromCrossHeaders(core.MetachainShardId, headersInfo)
		sbt.RemoveLastNotarizedHeaders()

		require.Zero(t, notifications.Load())
	})

	t.Run("rollback before publication rejects the old view", func(t *testing.T) {
		t.Parallel()

		sbt, source, sourceHash, _, _, _, _ := createTracker(t)
		var notifications atomic.Int32
		sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
			notifications.Add(1)
		})

		headersInfo := sbt.GetSelfHeadersWithSource(source, sourceHash)
		require.Len(t, headersInfo, 1)
		sbt.RemoveLastNotarizedHeaders()
		sbt.PublishSelfNotarizedFromCrossHeaders(core.MetachainShardId, headersInfo)
		sbt.RemoveLastNotarizedHeaders()

		require.Zero(t, notifications.Load())
	})

	t.Run("new unresolved sibling evidence rejects publication", func(t *testing.T) {
		t.Parallel()

		sbt, source, sourceHash, hasSibling, _, _, _ := createTracker(t)
		var notifications atomic.Int32
		sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
			notifications.Add(1)
		})

		headersInfo := sbt.GetSelfHeadersWithSource(source, sourceHash)
		require.Len(t, headersInfo, 1)
		hasSibling.Store(true)
		sbt.PublishSelfNotarizedFromCrossHeaders(core.MetachainShardId, headersInfo)
		sbt.RemoveLastNotarizedHeaders()

		require.Zero(t, notifications.Load())
	})

	t.Run("legacy publication does not use the V3 source gate", func(t *testing.T) {
		t.Parallel()

		sbt, _, _, _, _, _, _ := createTracker(t)
		notified := make(chan struct{}, 1)
		sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
			notified <- struct{}{}
		})

		legacyMeta := &block.MetaBlock{
			ShardInfo: []block.ShardData{{
				HeaderHash: []byte("referenced-shard"),
				ShardID:    0,
				Nonce:      5,
				Epoch:      2,
			}},
		}
		headersInfo := sbt.GetSelfHeadersWithSource(legacyMeta, []byte("legacy-meta"))
		require.Len(t, headersInfo, 1)
		sbt.PublishSelfNotarizedFromCrossHeaders(core.MetachainShardId, headersInfo)

		select {
		case <-notified:
		case <-time.After(time.Second):
			require.FailNow(t, "legacy notification was not delivered")
		}
	})

	t.Run("V3 publication requires the source hash", func(t *testing.T) {
		t.Parallel()

		sbt, source, _, _, _, _, _ := createTracker(t)
		var notifications atomic.Int32
		sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
			notifications.Add(1)
		})

		headersInfo := sbt.GetSelfHeadersWithSource(source, nil)
		require.Len(t, headersInfo, 1)
		sbt.PublishSelfNotarizedFromCrossHeaders(core.MetachainShardId, headersInfo)
		sbt.RemoveLastNotarizedHeaders()

		require.Zero(t, notifications.Load())
	})

	t.Run("late header remains pending while its V3 source is unresolved", func(t *testing.T) {
		t.Parallel()

		sbt, source, sourceHash, hasSibling, shardHeaderAvailable, headerHandler, _ := createTracker(t)
		shardHeaderAvailable.Store(false)
		notified := make(chan struct{}, 1)
		sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
			notified <- struct{}{}
		})

		require.Empty(t, sbt.GetSelfHeadersWithSource(source, sourceHash))
		hasSibling.Store(true)
		shardHeaderAvailable.Store(true)
		headerHandler(&block.HeaderV3{ShardID: 0, Nonce: 5, Epoch: 2}, []byte("referenced-shard"))
		require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())

		hasSibling.Store(false)
		headerHandler(source, sourceHash)
		select {
		case <-notified:
		case <-time.After(time.Second):
			require.FailNow(t, "retained notification was not delivered after its source became final")
		}
	})

	t.Run("dead source discards a delivered pending header", func(t *testing.T) {
		t.Parallel()

		sbt, source, sourceHash, _, shardHeaderAvailable, headerHandler, _ := createTracker(t)
		sbt.SetMetaFinalityView(&testscommon.MetaFinalityViewStub{
			IsDeadMetaBlockCalled: func(hash []byte, nonce uint64) bool {
				return bytes.Equal(hash, sourceHash) && nonce == source.GetNonce()
			},
		})
		shardHeaderAvailable.Store(false)
		var notifications atomic.Int32
		sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
			notifications.Add(1)
		})

		require.Empty(t, sbt.GetSelfHeadersWithSource(source, sourceHash))
		shardHeaderAvailable.Store(true)
		headerHandler(&block.HeaderV3{ShardID: 0, Nonce: 5, Epoch: 2}, []byte("referenced-shard"))

		require.Zero(t, sbt.NumPendingSelfHeaders())
		require.Zero(t, notifications.Load())
	})

	t.Run("dead source is removed while unresolved source remains pending", func(t *testing.T) {
		t.Parallel()

		sbt, source, sourceHash, _, shardHeaderAvailable, headerHandler, _ := createTracker(t)
		deadSourceHash := []byte("dead-source-meta")
		deadSource := &block.MetaBlockV3{
			Nonce:             source.Nonce,
			Round:             source.Round + 1,
			PrevHash:          []byte("other-parent"),
			ShardInfoProposal: source.ShardInfoProposal,
		}
		sbt.SetMetaFinalityView(&testscommon.MetaFinalityViewStub{
			IsDeadMetaBlockCalled: func(hash []byte, _ uint64) bool {
				return bytes.Equal(hash, deadSourceHash)
			},
		})
		shardHeaderAvailable.Store(false)
		require.Empty(t, sbt.GetSelfHeadersWithSource(deadSource, deadSourceHash))
		require.Empty(t, sbt.GetSelfHeadersWithSource(source, sourceHash))

		shardHeaderAvailable.Store(true)
		headerHandler(&block.HeaderV3{ShardID: 0, Nonce: 5, Epoch: 2}, []byte("referenced-shard"))

		require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())
		require.Equal(t, 1, sbt.NumPendingSources([]byte("referenced-shard")))
	})

	for _, heldSourceFirst := range []bool{true, false} {
		t.Run(fmt.Sprintf("held source survives same-nonce dead source, held first %v", heldSourceFirst), func(t *testing.T) {
			t.Parallel()

			sbt, heldSource, heldSourceHash, _, shardHeaderAvailable, headerHandler, _ := createTracker(t)
			deadSourceHash := []byte("dead-source-meta")
			deadSource := &block.MetaBlockV3{
				Nonce:             heldSource.Nonce,
				Round:             heldSource.Round + 1,
				PrevHash:          []byte("other-parent"),
				ShardInfoProposal: heldSource.ShardInfoProposal,
			}
			sbt.SetMetaFinalityView(&testscommon.MetaFinalityViewStub{
				IsMetaHeaderSettlementReadyCalled: func(_ data.HeaderHandler, hash []byte) bool {
					return bytes.Equal(hash, heldSourceHash)
				},
				IsDeadMetaBlockCalled: func(hash []byte, nonce uint64) bool {
					return bytes.Equal(hash, deadSourceHash) && nonce == deadSource.Nonce
				},
			})
			shardHeaderAvailable.Store(false)
			notified := make(chan struct{}, 1)
			sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
				notified <- struct{}{}
			})

			if heldSourceFirst {
				require.Empty(t, sbt.GetSelfHeadersWithSource(heldSource, heldSourceHash))
				require.Empty(t, sbt.GetSelfHeadersWithSource(deadSource, deadSourceHash))
			} else {
				require.Empty(t, sbt.GetSelfHeadersWithSource(deadSource, deadSourceHash))
				require.Empty(t, sbt.GetSelfHeadersWithSource(heldSource, heldSourceHash))
			}

			shardHeaderAvailable.Store(true)
			headerHandler(&block.HeaderV3{ShardID: 0, Nonce: 5, Epoch: 2}, []byte("referenced-shard"))
			select {
			case <-notified:
			case <-time.After(time.Second):
				require.FailNow(t, "held-final source was lost after a same-nonce dead source update")
			}
			require.Zero(t, sbt.NumPendingSelfHeaders())
		})
	}

	t.Run("overflow source is resolved by the canonical meta scan", func(t *testing.T) {
		t.Parallel()

		sbt, canonicalSource, canonicalSourceHash, _, shardHeaderAvailable, headerHandler, _ := createTracker(t)
		shardHeaderAvailable.Store(false)
		for index := 0; index < 4; index++ {
			alternateSource := &block.MetaBlockV3{
				Nonce:             canonicalSource.Nonce,
				Round:             canonicalSource.Round + uint64(index+1),
				PrevHash:          []byte("alternate-parent"),
				ShardInfoProposal: canonicalSource.ShardInfoProposal,
			}
			require.Empty(t, sbt.GetSelfHeadersWithSource(alternateSource, []byte(fmt.Sprintf("alternate-%d", index))))
		}
		require.Empty(t, sbt.GetSelfHeadersWithSource(canonicalSource, canonicalSourceHash))

		notified := make(chan struct{}, 1)
		sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
			notified <- struct{}{}
		})
		shardHeaderAvailable.Store(true)
		headerHandler(&block.HeaderV3{ShardID: 0, Nonce: 5, Epoch: 2}, []byte("referenced-shard"))

		select {
		case <-notified:
		case <-time.After(time.Second):
			require.FailNow(t, "canonical overflow source did not resolve the pending header")
		}
		require.Zero(t, sbt.NumPendingSelfHeaders())
	})

	t.Run("overflow canonical scan follows the request interval", func(t *testing.T) {
		t.Parallel()

		sbt, canonicalSource, canonicalSourceHash, _, shardHeaderAvailable, headerHandler, _ := createTracker(t)
		var inclusionChecks atomic.Int32
		sbt.SetMetaFinalityView(&testscommon.MetaFinalityViewStub{
			IsMetaHeaderSettlementReadyCalled: func(_ data.HeaderHandler, _ []byte) bool {
				return true
			},
			IsShardHeaderIncludedCalled: func(_ data.MetaHeaderHandler, _ uint32, _ []byte, _ uint64) bool {
				inclusionChecks.Add(1)
				return false
			},
		})
		shardHeaderAvailable.Store(false)
		for index := 0; index < 5; index++ {
			alternateSource := &block.MetaBlockV3{
				Nonce:             canonicalSource.Nonce,
				Round:             canonicalSource.Round + uint64(index+1),
				PrevHash:          []byte("alternate-parent"),
				ShardInfoProposal: canonicalSource.ShardInfoProposal,
			}
			require.Empty(t, sbt.GetSelfHeadersWithSource(alternateSource, []byte(fmt.Sprintf("alternate-%d", index))))
		}

		shardHeaderAvailable.Store(true)
		headerHandler(&block.HeaderV3{ShardID: 0, Nonce: 5, Epoch: 2}, []byte("referenced-shard"))
		checksAfterFirstScan := inclusionChecks.Load()
		require.Positive(t, checksAfterFirstScan)
		require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())

		headerHandler(canonicalSource, canonicalSourceHash)
		require.Equal(t, checksAfterFirstScan, inclusionChecks.Load())
		require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())
	})

	t.Run("settled overflow source is evaluated without waiting for another scan", func(t *testing.T) {
		t.Parallel()

		sbt, canonicalSource, canonicalSourceHash, _, shardHeaderAvailable, headerHandler, _ := createTracker(t)
		var settlementReady atomic.Bool
		sbt.SetMetaFinalityView(&testscommon.MetaFinalityViewStub{
			IsMetaHeaderSettlementReadyCalled: func(_ data.HeaderHandler, hash []byte) bool {
				return settlementReady.Load() && bytes.Equal(hash, canonicalSourceHash)
			},
		})
		shardHeaderAvailable.Store(false)
		for index := 0; index < 4; index++ {
			alternateSource := &block.MetaBlockV3{
				Nonce:             canonicalSource.Nonce,
				Round:             canonicalSource.Round + uint64(index+1),
				PrevHash:          []byte("alternate-parent"),
				ShardInfoProposal: canonicalSource.ShardInfoProposal,
			}
			require.Empty(t, sbt.GetSelfHeadersWithSource(alternateSource, []byte(fmt.Sprintf("alternate-%d", index))))
		}
		require.Equal(t, 4, sbt.NumPendingSources([]byte("referenced-shard")))

		shardHeaderAvailable.Store(true)
		headerHandler(&block.HeaderV3{ShardID: 0, Nonce: 5, Epoch: 2}, []byte("referenced-shard"))
		require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())

		notified := make(chan struct{}, 1)
		sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
			notified <- struct{}{}
		})
		settlementReady.Store(true)
		require.Empty(t, sbt.GetSelfHeadersWithSource(canonicalSource, canonicalSourceHash))

		select {
		case <-notified:
		case <-time.After(time.Second):
			require.FailNow(t, "settled overflow source was delayed by the scan interval")
		}
		require.Zero(t, sbt.NumPendingSelfHeaders())
	})

	t.Run("overflow claim survives rollback of every retained source", func(t *testing.T) {
		t.Parallel()

		sbt, canonicalSource, canonicalSourceHash, _, shardHeaderAvailable, headerHandler, _ := createTracker(t)
		shardHeaderAvailable.Store(false)
		sbt.AddCrossNotarizedHeader(core.MetachainShardId, canonicalSource, canonicalSourceHash)

		previousHash := canonicalSourceHash
		for index := 0; index < 4; index++ {
			retainedSource := &block.MetaBlockV3{
				Nonce:             canonicalSource.Nonce + uint64(index+1),
				Round:             canonicalSource.Round + uint64(index+1),
				PrevHash:          previousHash,
				ShardInfoProposal: canonicalSource.ShardInfoProposal,
			}
			retainedHash := []byte(fmt.Sprintf("retained-%d", index))
			sbt.AddCrossNotarizedHeader(core.MetachainShardId, retainedSource, retainedHash)
			require.Empty(t, sbt.GetSelfHeadersWithSource(retainedSource, retainedHash))
			previousHash = retainedHash
		}
		require.Empty(t, sbt.GetSelfHeadersWithSource(canonicalSource, canonicalSourceHash))

		for index := 0; index < 4; index++ {
			sbt.RemoveLastNotarizedHeaders()
		}
		require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())

		notified := make(chan struct{}, 1)
		sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
			notified <- struct{}{}
		})
		shardHeaderAvailable.Store(true)
		headerHandler(&block.HeaderV3{ShardID: 0, Nonce: 5, Epoch: 2}, []byte("referenced-shard"))

		select {
		case <-notified:
		case <-time.After(time.Second):
			require.FailNow(t, "overflow source was lost after retained sources rolled back")
		}
		require.Zero(t, sbt.NumPendingSelfHeaders())
	})

	t.Run("rollback of a duplicate anchor retains its pending source", func(t *testing.T) {
		t.Parallel()

		sbt, source, sourceHash, _, shardHeaderAvailable, headerHandler, _ := createTracker(t)
		shardHeaderAvailable.Store(false)
		require.Empty(t, sbt.GetSelfHeadersWithSource(source, sourceHash))
		require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())

		sbt.AddCrossNotarizedHeader(core.MetachainShardId, source, sourceHash)
		sbt.AddCrossNotarizedHeader(core.MetachainShardId, source, sourceHash)
		sbt.RemoveLastNotarizedHeaders()

		notified := make(chan struct{}, 1)
		sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
			notified <- struct{}{}
		})
		shardHeaderAvailable.Store(true)
		headerHandler(&block.HeaderV3{ShardID: 0, Nonce: 5, Epoch: 2}, []byte("referenced-shard"))

		select {
		case <-notified:
		case <-time.After(time.Second):
			require.FailNow(t, "duplicate-anchor rollback removed the surviving source")
		}
		require.Zero(t, sbt.NumPendingSelfHeaders())
	})

	t.Run("source requests prefer canonical authority and are throttled per claim", func(t *testing.T) {
		t.Parallel()

		const maxSources = 4
		sbt, canonicalSource, canonicalSourceHash, _, shardHeaderAvailable, headerHandler, requestHandler := createTracker(t)
		sbt.SetMetaFinalityView(&testscommon.MetaFinalityViewStub{})
		shardHeaderAvailable.Store(false)
		for index := 0; index < maxSources-1; index++ {
			alternateSource := &block.MetaBlockV3{
				Nonce:             canonicalSource.Nonce,
				Round:             canonicalSource.Round + uint64(index+1),
				PrevHash:          []byte("alternate-parent"),
				ShardInfoProposal: canonicalSource.ShardInfoProposal,
			}
			require.Empty(t, sbt.GetSelfHeadersWithSource(alternateSource, []byte(fmt.Sprintf("alternate-%d", index))))
		}
		require.Empty(t, sbt.GetSelfHeadersWithSource(canonicalSource, canonicalSourceHash))

		var requestedHashes [][]byte
		var proofRequests int
		requestHandler.RequestMetaHeaderForEpochCalled = func(hash []byte, _ uint32) {
			requestedHashes = append(requestedHashes, append([]byte(nil), hash...))
		}
		requestHandler.RequestEquivalentProofByHashForEpochCalled = func(_ uint32, _ []byte, _ uint32) {
			proofRequests++
		}

		shardHeaderAvailable.Store(true)
		shardHeader := &block.HeaderV3{ShardID: 0, Nonce: 5, Epoch: 2}
		for index := 0; index < maxSources; index++ {
			headerHandler(shardHeader, []byte("referenced-shard"))
		}

		require.Equal(t, [][]byte{canonicalSourceHash}, requestedHashes)
		require.Equal(t, 1, proofRequests)
		require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())
	})

	t.Run("source requests do not retain the notification reset barrier", func(t *testing.T) {
		t.Parallel()

		sbt, source, sourceHash, _, shardHeaderAvailable, headerHandler, requestHandler := createTracker(t)
		sbt.SetMetaFinalityView(&testscommon.MetaFinalityViewStub{})
		shardHeaderAvailable.Store(false)
		require.Empty(t, sbt.GetSelfHeadersWithSource(source, sourceHash))

		requestStarted := make(chan struct{})
		releaseRequest := make(chan struct{})
		var releaseOnce sync.Once
		release := func() {
			releaseOnce.Do(func() {
				close(releaseRequest)
			})
		}
		t.Cleanup(release)
		requestHandler.RequestMetaHeaderForEpochCalled = func(_ []byte, _ uint32) {
			close(requestStarted)
			<-releaseRequest
		}

		shardHeaderAvailable.Store(true)
		deliveryDone := make(chan struct{})
		go func() {
			headerHandler(&block.HeaderV3{ShardID: 0, Nonce: 5, Epoch: 2}, []byte("referenced-shard"))
			close(deliveryDone)
		}()
		select {
		case <-requestStarted:
		case <-time.After(time.Second):
			require.FailNow(t, "pending source request did not start")
		}

		resetDone := make(chan struct{})
		go func() {
			sbt.RemoveLastNotarizedHeaders()
			close(resetDone)
		}()
		select {
		case <-resetDone:
		case <-time.After(time.Second):
			require.FailNow(t, "source request retained the notification reset barrier")
		}

		release()
		select {
		case <-deliveryDone:
		case <-time.After(time.Second):
			require.FailNow(t, "pending source request did not finish")
		}
	})
}

func TestShardBlockTrack_NoOpRollbackRetainsPendingSourceWithoutHash(t *testing.T) {
	t.Parallel()

	arguments := CreateShardTrackerMockArguments()
	sbt, err := track.NewShardBlockTrack(arguments)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sbt.Close())
	})

	metaBlock := &block.MetaBlockV3{
		ShardInfoProposal: []block.ShardDataProposal{{
			HeaderHash: []byte("missing-shard-header"),
			ShardID:    0,
			Nonce:      7,
			Epoch:      2,
		}},
	}
	require.Empty(t, sbt.GetSelfHeadersWithSource(metaBlock, nil))
	require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())

	sbt.RemoveLastNotarizedHeaders()

	require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())
	require.Equal(t, 1, sbt.NumPendingSources([]byte("missing-shard-header")))
}

func TestShardBlockTrack_ResolvedMarkersRetainNewestBoundedSet(t *testing.T) {
	t.Parallel()

	arguments := CreateShardTrackerMockArguments()
	sbt, err := track.NewShardBlockTrack(arguments)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sbt.Close())
	})

	limit := sbt.GetMaxNumHeadersToKeepPerShard()
	require.Positive(t, limit)
	for index := 0; index < limit; index++ {
		sbt.RememberResolvedSelfHeader([]byte{byte(index >> 8), byte(index)}, uint64(index+10))
	}
	require.Equal(t, int64(limit), sbt.NumResolvedSelfHeaders())

	newestHash := []byte("newest")
	sbt.RememberResolvedSelfHeader(newestHash, uint64(limit+10))
	markers := sbt.ResolvedSelfHeaders()
	require.Len(t, markers, limit)
	require.NotContains(t, markers, string([]byte{0, 0}))
	require.Equal(t, uint64(limit+10), markers[string(newestHash)])

	sbt.RememberResolvedSelfHeader([]byte("older"), 1)
	require.Equal(t, markers, sbt.ResolvedSelfHeaders())

	sbt.RememberResolvedSelfHeader(newestHash, uint64(limit+11))
	markers = sbt.ResolvedSelfHeaders()
	require.Len(t, markers, limit)
	require.Equal(t, uint64(limit+11), markers[string(newestHash)])

	sbt.RestoreToGenesis()
	require.Zero(t, sbt.NumResolvedSelfHeaders())
	require.Empty(t, sbt.ResolvedSelfHeaders())
}

func TestShardBlockTrack_PendingSourceRemainsCanonicalAfterAnchorAdvances(t *testing.T) {
	t.Parallel()

	var headerHandler func(data.HeaderHandler, []byte)
	var availableHeaders sync.Map
	headersPool := &pool.HeadersPoolStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			header, exists := availableHeaders.Load(string(hash))
			if !exists {
				return nil, errors.New("missing header")
			}

			return header.(data.HeaderHandler), nil
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

	sbt, err := track.NewShardBlockTrack(arguments)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sbt.Close())
	})
	require.NotNil(t, headerHandler)
	var latestAnchorHash []byte
	var deadAnchor atomic.Bool
	sbt.SetMetaFinalityView(&testscommon.MetaFinalityViewStub{
		IsMetaHeaderSettlementReadyCalled: func(_ data.HeaderHandler, _ []byte) bool {
			return true
		},
		IsDeadMetaBlockCalled: func(hash []byte, _ uint64) bool {
			return deadAnchor.Load() && bytes.Equal(hash, latestAnchorHash)
		},
	})

	anchor, anchorHash, err := sbt.GetLastCrossNotarizedHeader(core.MetachainShardId)
	require.NoError(t, err)
	shardHash := []byte("late-shard-header")
	source := &block.MetaBlockV3{
		Nonce:    anchor.GetNonce() + 1,
		Round:    anchor.GetRound() + 1,
		Epoch:    2,
		PrevHash: anchorHash,
		ShardInfoProposal: []block.ShardDataProposal{{
			HeaderHash: shardHash,
			ShardID:    0,
			Nonce:      7,
			Epoch:      2,
		}},
	}
	sourceHash := []byte("canonical-source")
	availableHeaders.Store(string(sourceHash), source)
	sbt.AddTrackedHeader(source, sourceHash)
	sbt.AddCrossNotarizedHeader(core.MetachainShardId, source, sourceHash)
	require.Empty(t, sbt.GetSelfHeadersWithSource(source, sourceHash))
	require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())

	previous := data.HeaderHandler(source)
	previousHash := sourceHash
	for index := 0; index < track.MaxMetaBlocksScannedForInclusion+4; index++ {
		next := &block.MetaBlockV3{
			Nonce:    previous.GetNonce() + 1,
			Round:    previous.GetRound() + 1,
			Epoch:    2,
			PrevHash: previousHash,
		}
		nextHash := []byte(fmt.Sprintf("meta-%d", index))
		availableHeaders.Store(string(nextHash), next)
		sbt.AddTrackedHeader(next, nextHash)
		sbt.AddCrossNotarizedHeader(core.MetachainShardId, next, nextHash)
		previous = next
		previousHash = nextHash
	}
	latestAnchorHash = previousHash

	notified := make(chan struct{}, 1)
	sbt.RegisterSelfNotarizedFromCrossHeadersHandler(func(
		_ uint32,
		_ []data.HeaderHandler,
		_ [][]byte,
	) {
		notified <- struct{}{}
	})
	deadAnchor.Store(true)
	headerHandler(&block.HeaderV3{ShardID: 0, Nonce: 7, Epoch: 2}, shardHash)
	require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())
	select {
	case <-notified:
		require.FailNow(t, "dead current anchor authorized a crossed source")
	default:
	}

	deadAnchor.Store(false)
	headerHandler(&block.HeaderV3{ShardID: 0, Nonce: 7, Epoch: 2}, shardHash)

	select {
	case <-notified:
	case <-time.After(time.Second):
		require.FailNow(t, "canonical source was lost after the anchor advanced")
	}
	require.Zero(t, sbt.NumPendingSelfHeaders())
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
	sbt.SetMetaFinalityView(&testscommon.MetaFinalityViewStub{
		IsMetaHeaderSettlementReadyCalled: func(_ data.HeaderHandler, _ []byte) bool {
			return true
		},
	})

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
	require.Empty(t, sbt.GetSelfHeadersWithSource(metaBlock, metaHash))
	require.Equal(t, int32(1), headerRequests.Load())
	require.Equal(t, int32(1), proofRequests.Load())
	require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())
	require.Empty(t, sbt.GetSelfHeadersWithSource(metaBlock, metaHash))
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

	require.Empty(t, sbt.GetSelfHeadersWithSource(metaBlock, metaHash))
	require.Equal(t, int32(1), headerRequests.Load())
	require.Equal(t, int32(1), proofRequests.Load())

	sbt.RestoreToGenesis()
	require.Zero(t, sbt.NumPendingSelfHeaders())
	availableHeaders.Store(string(headerHash), header)
	sbt.AddCrossNotarizedHeader(core.MetachainShardId, metaBlock, metaHash)
	restoredHeaders := sbt.GetSelfHeadersWithSource(metaBlock, metaHash)
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
	require.Empty(t, sbt.GetSelfHeadersWithSource(secondMetaBlock, secondMetaHash))
	secondHeader := &block.HeaderV3{ShardID: 0, Nonce: 8, Epoch: 2}
	availableHeaders.Store(string(secondHash), secondHeader)

	var returned []*track.SelfHeaderInfo
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	go func() {
		defer waitGroup.Done()
		headerHandlers[0](secondHeader, secondHash)
	}()
	go func() {
		defer waitGroup.Done()
		returned = sbt.GetSelfHeadersWithSource(secondMetaBlock, secondMetaHash)
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
	serializedMetaHash, err := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, serializedMetaBlock)
	require.NoError(t, err)
	sbt.AddCrossNotarizedHeader(core.MetachainShardId, serializedMetaBlock, serializedMetaHash)
	require.Empty(t, sbt.GetSelfHeadersWithSource(serializedMetaBlock, serializedMetaHash))
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
	retainedMetaHash, err := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, retainedMetaBlock)
	require.NoError(t, err)
	sbt.AddCrossNotarizedHeader(core.MetachainShardId, retainedMetaBlock, retainedMetaHash)
	require.Empty(t, sbt.GetSelfHeadersWithSource(retainedMetaBlock, retainedMetaHash))

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
	rolledBackMetaHash, err := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, rolledBackMetaBlock)
	require.NoError(t, err)
	sbt.AddCrossNotarizedHeader(core.MetachainShardId, rolledBackMetaBlock, rolledBackMetaHash)
	require.Empty(t, sbt.GetSelfHeadersWithSource(rolledBackMetaBlock, rolledBackMetaHash))
	notificationsBeforeRollback := notifications.Load()
	sbt.RemoveLastNotarizedHeaders()
	headerHandlers[0](&block.HeaderV3{ShardID: 0, Nonce: 10, Epoch: 2}, rolledBackHeaderHash)
	require.Equal(t, notificationsBeforeRollback, notifications.Load())
	require.Equal(t, int64(1), sbt.NumPendingSelfHeaders())
	headerHandlers[0](&block.HeaderV3{ShardID: 0, Nonce: 9, Epoch: 2}, retainedHeaderHash)
	require.Zero(t, sbt.NumPendingSelfHeaders())
	require.Eventually(t, func() bool {
		return notifications.Load() == notificationsBeforeRollback+1
	}, time.Second, 5*time.Millisecond)

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
	sbt.SetMetaFinalityView(&testscommon.MetaFinalityViewStub{
		IsMetaHeaderSettlementReadyCalled: func(_ data.HeaderHandler, _ []byte) bool {
			return true
		},
	})

	headerHash := []byte("header-arriving-before-pending-insertion")
	header := &block.HeaderV3{ShardID: 0, Nonce: 7, Epoch: 2}
	metaBlock := &block.MetaBlockV3{
		Nonce: 1,
		ShardInfoProposal: []block.ShardDataProposal{{
			HeaderHash: headerHash,
			ShardID:    0,
			Nonce:      header.Nonce,
			Epoch:      header.Epoch,
		}},
	}
	metaHash := []byte("source-meta-hash")
	sbt.AddCrossNotarizedHeader(core.MetachainShardId, metaBlock, metaHash)

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

	selfHeaders := sbt.GetSelfHeadersWithSource(metaBlock, metaHash)
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

	require.Empty(t, sbt.GetSelfHeadersWithSource(metaBlock, metaHash))
	require.Zero(t, sbt.NumPendingSelfHeaders())
	require.Equal(t, int32(1), headerRequests.Load())
	require.Equal(t, int32(1), proofRequests.Load())
}
