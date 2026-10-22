package track

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func newPendingSettlementTracker(provider func() uint64) *shardBlockTrack {
	tracker := &shardBlockTrack{
		baseBlockTrack: &baseBlockTrack{
			requestHandler:              &testscommon.RequestHandlerStub{},
			maxNumHeadersToKeepPerShard: 100,
		},
		pendingSelfHeaders:  make(map[string]*pendingSelfHeader),
		resolvedSelfHeaders: make(map[string]uint64),
	}
	tracker.SetSettledNonceProvider(provider)
	return tracker
}

func TestShardBlockTrack_PendingCleanupUsesOnlySettlement(t *testing.T) {
	t.Parallel()

	var settled atomic.Uint64
	tracker := newPendingSettlementTracker(settled.Load)
	for _, nonce := range []uint64{3, 4, 5, 6} {
		key := fmt.Sprint(nonce)
		pending := &pendingSelfHeader{
			nonce: nonce,
			hash:  []byte(key),
			sources: []pendingMetaSource{
				{hash: []byte("source-A")}, {hash: []byte("source-B")},
			},
		}
		if nonce%2 == 0 {
			pending.deliveredHeader = &block.HeaderV3{Nonce: nonce}
		}
		tracker.pendingSelfHeaders[key] = pending
		tracker.resolvedSelfHeaders["resolved-"+key] = nonce
	}
	tracker.numPendingSelfHeaders.Store(4)
	tracker.numResolvedSelfHeaders.Store(4)

	tracker.cleanupPendingSelfHeaders()
	require.Len(t, tracker.pendingSelfHeaders, 4)
	settled.Store(5)
	tracker.cleanupPendingSelfHeaders()
	require.Len(t, tracker.pendingSelfHeaders, 2)
	require.Contains(t, tracker.pendingSelfHeaders, "5")
	require.Contains(t, tracker.pendingSelfHeaders, "6")
	require.Equal(t, int64(2), tracker.numPendingSelfHeaders.Load())
	require.Equal(t, map[string]uint64{"resolved-5": 5, "resolved-6": 6}, tracker.resolvedSelfHeaders)
	require.Equal(t, int64(2), tracker.numResolvedSelfHeaders.Load())

	settled.Store(2)
	tracker.cleanupPendingSelfHeaders()
	require.Equal(t, uint64(5), tracker.settledNonce.Load())
	tracker.rememberResolvedSelfHeader("late-resolved", 4)
	require.NotContains(t, tracker.resolvedSelfHeaders, "late-resolved")
	tracker.SetSettledNonceProvider(nil)
	tracker.cleanupPendingSelfHeaders()
	require.Len(t, tracker.pendingSelfHeaders, 2)
}

func TestShardBlockTrack_SettlementSuppressesQueuedRequests(t *testing.T) {
	t.Parallel()

	for _, meta := range []bool{false, true} {
		for _, nonce := range []uint64{4, 5, 6} {
			t.Run(fmt.Sprintf("meta=%t/nonce=%d", meta, nonce), func(t *testing.T) {
				t.Parallel()
				var settled atomic.Uint64
				settled.Store(5)
				tracker := newPendingSettlementTracker(settled.Load)
				var headers, proofs int
				tracker.requestHandler = &testscommon.RequestHandlerStub{
					RequestShardHeaderForEpochCalled:           func(_ uint32, _ []byte, _ uint32) { headers++ },
					RequestMetaHeaderForEpochCalled:            func(_ []byte, _ uint32) { headers++ },
					RequestEquivalentProofByHashForEpochCalled: func(_ uint32, _ []byte, _ uint32) { proofs++ },
				}
				if meta {
					tracker.requestPendingSources([]pendingSourceRequest{{hash: []byte("meta"), epoch: 2242, nonce: nonce}})
				} else {
					tracker.requestPendingSelfHeader(pendingSelfHeaderRequest{hash: []byte("shard"), epoch: 2242, nonce: nonce})
				}
				if nonce < 5 {
					require.Zero(t, headers)
					require.Zero(t, proofs)
				} else {
					require.Equal(t, 1, headers)
					require.Equal(t, 1, proofs)
				}
				tracker.pendingSelfHeadersView.Add(2)
				tracker.requestPendingSources([]pendingSourceRequest{{hash: []byte("stale-view"), nonce: 6}})
				tracker.requestPendingSelfHeader(pendingSelfHeaderRequest{hash: []byte("stale-view"), nonce: 6})
				if nonce >= 5 {
					require.Equal(t, 1, headers)
					require.Equal(t, 1, proofs)
				} else {
					require.Zero(t, headers)
					require.Zero(t, proofs)
				}
			})
		}
	}
}

func TestShardBlockTrack_SettlementDuringHeaderRequestSkipsProof(t *testing.T) {
	t.Parallel()

	for _, meta := range []bool{false, true} {
		t.Run(fmt.Sprintf("meta=%t", meta), func(t *testing.T) {
			t.Parallel()
			var settled atomic.Uint64
			tracker := newPendingSettlementTracker(settled.Load)
			var headers, proofs int
			requestHeader := func() {
				headers++
				settled.Store(8)
				tracker.cleanupPendingSelfHeaders()
			}
			tracker.requestHandler = &testscommon.RequestHandlerStub{
				RequestShardHeaderForEpochCalled:           func(_ uint32, _ []byte, _ uint32) { requestHeader() },
				RequestMetaHeaderForEpochCalled:            func(_ []byte, _ uint32) { requestHeader() },
				RequestEquivalentProofByHashForEpochCalled: func(_ uint32, _ []byte, _ uint32) { proofs++ },
			}
			if meta {
				tracker.requestPendingSources([]pendingSourceRequest{{hash: []byte("source"), nonce: 7}})
			} else {
				tracker.requestPendingSelfHeader(pendingSelfHeaderRequest{hash: []byte("shard"), nonce: 7})
			}
			require.Equal(t, 1, headers)
			require.Zero(t, proofs)
		})
	}
}

func TestShardBlockTrack_SettlementRejectsLateInsertion(t *testing.T) {
	t.Parallel()

	tracker := newPendingSettlementTracker(func() uint64 { return 5 })
	var requests atomic.Int64
	tracker.requestHandler = &testscommon.RequestHandlerStub{
		RequestShardHeaderForEpochCalled:           func(_ uint32, _ []byte, _ uint32) { requests.Add(1) },
		RequestEquivalentProofByHashForEpochCalled: func(_ uint32, _ []byte, _ uint32) { requests.Add(1) },
	}
	source := &block.MetaBlockV3{Nonce: 100, Epoch: 2242}
	ref := &block.ShardDataProposal{ShardID: 0, HeaderHash: []byte("late"), Nonce: 4, Epoch: 2242}
	tracker.addOrRefreshPendingSelfHeader(ref, 0, source, []byte("source"))
	require.Empty(t, tracker.pendingSelfHeaders)
	require.Zero(t, requests.Load())
	info := &selfHeaderInfo{
		Header: &block.HeaderV3{ShardID: 0, Nonce: 4}, Hash: ref.HeaderHash,
		sourceMetaHeader: source, sourceMetaHash: []byte("source"),
	}
	require.Empty(t, tracker.retainDeliveredSelfHeader(info))
	require.False(t, tracker.prepareSelfHeaderForReturn(info.Header, info.Hash, 4, 0, true))
	_, handled := tracker.getHeaderFromPendingState(ref, 0, source, []byte("source"))
	require.True(t, handled)
}

func TestShardBlockTrack_SettlementReadCannotCrossReset(t *testing.T) {
	t.Parallel()

	started, resume, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	tracker := newPendingSettlementTracker(func() uint64 {
		close(started)
		<-resume
		return 100
	})
	go func() {
		tracker.cleanupPendingSelfHeaders()
		close(done)
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("settlement read did not start")
	}
	tracker.mutPendingSelfHeaders.Lock()
	tracker.pendingSelfHeadersView.Add(2)
	tracker.pendingSelfHeaders["new-history"] = &pendingSelfHeader{nonce: 1, view: 2}
	tracker.numPendingSelfHeaders.Store(1)
	tracker.mutPendingSelfHeaders.Unlock()
	close(resume)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("cleanup did not finish")
	}
	require.Zero(t, tracker.settledNonce.Load())
	require.Contains(t, tracker.pendingSelfHeaders, "new-history")
}

func TestShardBlockTrack_ConcurrentSettlementAndPendingInsertion(t *testing.T) {
	t.Parallel()

	var settled atomic.Uint64
	tracker := newPendingSettlementTracker(settled.Load)
	source := &block.MetaBlockV3{Nonce: 100, Epoch: 2242}
	var workers sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for nonce := uint64(1); nonce <= 100; nonce++ {
				ref := &block.ShardDataProposal{ShardID: 0, HeaderHash: []byte(fmt.Sprint(nonce)), Nonce: nonce}
				tracker.addOrRefreshPendingSelfHeader(ref, 0, source, []byte("source"))
				tracker.cleanupPendingSelfHeaders()
			}
		}()
	}
	for nonce := uint64(1); nonce <= 100; nonce++ {
		settled.Store(nonce)
		tracker.cleanupPendingSelfHeaders()
	}
	workers.Wait()
	tracker.cleanupPendingSelfHeaders()
	for _, pending := range tracker.pendingSelfHeaders {
		require.GreaterOrEqual(t, pending.nonce, uint64(100))
	}
	require.Equal(t, int64(len(tracker.pendingSelfHeaders)), tracker.numPendingSelfHeaders.Load())
	tracker.pendingSelfHeadersView.Add(2)
	tracker.requestPendingSources([]pendingSourceRequest{{hash: []byte("old-meta"), epoch: 2242, nonce: 1}})
	require.False(t, tracker.canRequestPendingHeader(1, 0))
	require.False(t, tracker.canRequestPendingHeader(100, 0))
	require.True(t, tracker.canRequestPendingHeader(100, 2))
}
