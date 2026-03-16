package v2_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	v2 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v2"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func createSubroundSignatureForCompetingBlockTests(
	sentSigTracker *testscommon.SentSignatureTrackerStub,
	proofsPool *dataRetriever.ProofsPoolMock,
	roundHandler *testscommon.RoundHandlerMock,
) v2.SubroundSignature {
	container := consensusMocks.InitConsensusCore()
	if proofsPool != nil {
		container.SetEquivalentProofsPool(proofsPool)
	}
	if roundHandler != nil {
		container.SetRoundHandler(roundHandler)
	}

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.25,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	if sentSigTracker == nil {
		sentSigTracker = &testscommon.SentSignatureTrackerStub{}
	}

	srSignature, _ := v2.NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		sentSigTracker,
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMock.ThrottlerStub{},
	)

	srSignature.SetHeader(&block.Header{Nonce: 100})
	srSignature.SetData([]byte("current_hash"))

	return srSignature
}

func TestWaitIfCompetingBlock_NoPreviousHashExists(t *testing.T) {
	t.Parallel()

	sr := createSubroundSignatureForCompetingBlockTests(
		&testscommon.SentSignatureTrackerStub{
			GetSignedHashCalled: func(pkBytes []byte, nonce uint64) ([]byte, bool) {
				return nil, false
			},
		},
		nil,
		nil,
	)

	result := sr.WaitIfCompetingBlock(context.Background(), []byte("pk"), 100, []byte("current_hash"))
	assert.False(t, result, "should return false when no previous hash exists")
}

func TestWaitIfCompetingBlock_PreviousHashEqualsCurrent(t *testing.T) {
	t.Parallel()

	currentHash := []byte("same_hash")
	sr := createSubroundSignatureForCompetingBlockTests(
		&testscommon.SentSignatureTrackerStub{
			GetSignedHashCalled: func(pkBytes []byte, nonce uint64) ([]byte, bool) {
				return currentHash, true
			},
		},
		nil,
		nil,
	)

	result := sr.WaitIfCompetingBlock(context.Background(), []byte("pk"), 100, currentHash)
	assert.False(t, result, "should return false when previous hash equals current hash")
}

func TestWaitIfCompetingBlock_AlreadyPastDelayDeadline(t *testing.T) {
	t.Parallel()

	sr := createSubroundSignatureForCompetingBlockTests(
		&testscommon.SentSignatureTrackerStub{
			GetSignedHashCalled: func(pkBytes []byte, nonce uint64) ([]byte, bool) {
				return []byte("previous_hash"), true
			},
		},
		nil,
		&testscommon.RoundHandlerMock{
			TimeDurationCalled: func() time.Duration {
				return 100 * time.Millisecond
			},
			RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
				// Already past the competing block delay deadline (and subround end)
				return 0
			},
		},
	)

	result := sr.WaitIfCompetingBlock(context.Background(), []byte("pk"), 100, []byte("current_hash"))
	assert.False(t, result, "should return false (proceed to sign) when already past delay deadline")
}

func TestWaitIfCompetingBlock_NoTimeRemainingInSubround(t *testing.T) {
	t.Parallel()

	sr := createSubroundSignatureForCompetingBlockTests(
		&testscommon.SentSignatureTrackerStub{
			GetSignedHashCalled: func(pkBytes []byte, nonce uint64) ([]byte, bool) {
				return []byte("previous_hash"), true
			},
		},
		nil,
		&testscommon.RoundHandlerMock{
			TimeDurationCalled: func() time.Duration {
				return 600 * time.Millisecond
			},
			RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
				// targetTime = 300ms: still has time to target
				// sigEndDuration (85ms): no time left
				if maxTime > 200*time.Millisecond {
					return 200 * time.Millisecond
				}
				// No time remaining in signature subround
				return 0
			},
		},
	)

	result := sr.WaitIfCompetingBlock(context.Background(), []byte("pk"), 100, []byte("current_hash"))
	assert.False(t, result, "should return false (proceed to sign) when no time remaining in subround")
}

func TestWaitIfCompetingBlock_ContextCancelled(t *testing.T) {
	t.Parallel()

	sr := createSubroundSignatureForCompetingBlockTests(
		&testscommon.SentSignatureTrackerStub{
			GetSignedHashCalled: func(pkBytes []byte, nonce uint64) ([]byte, bool) {
				return []byte("previous_hash"), true
			},
		},
		&dataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				return false
			},
		},
		&testscommon.RoundHandlerMock{
			TimeDurationCalled: func() time.Duration {
				return 600 * time.Millisecond
			},
			RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
				return 300 * time.Millisecond
			},
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	result := sr.WaitIfCompetingBlock(ctx, []byte("pk"), 100, []byte("current_hash"))
	assert.True(t, result, "should return true (abort) when context is cancelled")
}

func TestWaitIfCompetingBlock_ProofArrivesForPreviousBlock(t *testing.T) {
	t.Parallel()

	previousHash := []byte("previous_hash")
	var proofAvailable atomic.Int32

	sr := createSubroundSignatureForCompetingBlockTests(
		&testscommon.SentSignatureTrackerStub{
			GetSignedHashCalled: func(pkBytes []byte, nonce uint64) ([]byte, bool) {
				return previousHash, true
			},
		},
		&dataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				if string(headerHash) == string(previousHash) {
					return proofAvailable.Load() == 1
				}
				return false
			},
		},
		&testscommon.RoundHandlerMock{
			TimeDurationCalled: func() time.Duration {
				return 600 * time.Millisecond
			},
			RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
				return 500 * time.Millisecond
			},
		},
	)

	// Make proof available after a short delay
	go func() {
		time.Sleep(15 * time.Millisecond)
		proofAvailable.Store(1)
	}()

	start := time.Now()
	result := sr.WaitIfCompetingBlock(context.Background(), []byte("pk"), 100, []byte("current_hash"))
	elapsed := time.Since(start)

	assert.True(t, result, "should return true (abort) when proof arrives for previous block")
	assert.Less(t, elapsed, 200*time.Millisecond, "should return quickly after proof arrives, not wait full delay")
}

func TestWaitIfCompetingBlock_DeadlineExpiresNoProof(t *testing.T) {
	t.Parallel()

	sr := createSubroundSignatureForCompetingBlockTests(
		&testscommon.SentSignatureTrackerStub{
			GetSignedHashCalled: func(pkBytes []byte, nonce uint64) ([]byte, bool) {
				return []byte("previous_hash"), true
			},
		},
		&dataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				return false
			},
		},
		&testscommon.RoundHandlerMock{
			TimeDurationCalled: func() time.Duration {
				return 100 * time.Millisecond
			},
			RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
				// Simulate round just started: remaining = maxTime
				return maxTime
			},
		},
	)

	start := time.Now()
	result := sr.WaitIfCompetingBlock(context.Background(), []byte("pk"), 100, []byte("current_hash"))
	elapsed := time.Since(start)

	assert.False(t, result, "should return false (proceed to sign) when deadline expires")
	// competingBlockSignDelay = 0.5, roundDuration = 100ms
	// targetTime = 50ms, sigEndDuration = 85ms (0.85 * 100ms)
	// delay = min(50ms, 85ms - 10ms) = 50ms
	assert.GreaterOrEqual(t, elapsed, 40*time.Millisecond, "should have waited at least ~50ms")
}

func TestWaitIfCompetingBlock_DelayCappedBySubroundRemaining(t *testing.T) {
	t.Parallel()

	sr := createSubroundSignatureForCompetingBlockTests(
		&testscommon.SentSignatureTrackerStub{
			GetSignedHashCalled: func(pkBytes []byte, nonce uint64) ([]byte, bool) {
				return []byte("previous_hash"), true
			},
		},
		&dataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				return false
			},
		},
		&testscommon.RoundHandlerMock{
			TimeDurationCalled: func() time.Duration {
				return 600 * time.Millisecond // targetTime = 300ms
			},
			RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
				// Simulate round just started: remaining = maxTime
				// targetTime = 300ms, sigEndDuration = 85ms (0.85 * roundTimeDuration=100ms)
				// delay = min(300ms, 85ms - 10ms) = 75ms
				return maxTime
			},
		},
	)

	start := time.Now()
	result := sr.WaitIfCompetingBlock(context.Background(), []byte("pk"), 100, []byte("current_hash"))
	elapsed := time.Since(start)

	assert.False(t, result, "should return false (proceed to sign) after capped delay expires")
	// delay should be capped to 75ms (sigEndDuration 85ms - 10ms safety), not full 300ms
	assert.Less(t, elapsed, 150*time.Millisecond, "delay should be capped, not full 300ms")
}

func TestWaitIfCompetingBlock_RecordSignedNonceCalledBeforeBroadcast(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	container.SetSigningHandler(&consensusMocks.SigningHandlerStub{
		CreateSignatureShareForPublicKeyCalled: func(msg []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
			return []byte("SIG"), nil
		},
	})
	container.SetBroadcastMessenger(&consensusMocks.BroadcastMessengerMock{
		BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
			return expectedErr // broadcast fails
		},
	})

	consensusState := initializers.InitConsensusStateWithKeysHandler(
		&testscommon.KeysHandlerStub{
			IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
				return true
			},
		},
	)
	ch := make(chan bool, 1)
	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.25,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	sr.SetHeader(&block.Header{Nonce: 100})

	recordCalled := false
	srSignature, _ := v2.NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{
			RecordSignedNonceCalled: func(pkBytes []byte, nonce uint64, headerHash []byte) {
				recordCalled = true
			},
		},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMock.ThrottlerStub{},
	)

	// broadcast will fail but RecordSignedNonce should still be called
	result := srSignature.SendSignatureForManagedKey(context.Background(), 0, "A")
	assert.False(t, result, "should return false because broadcast failed")
	assert.True(t, recordCalled, "RecordSignedNonce should be called before broadcast")
}

func TestWaitIfCompetingBlockForNode_NoCompetingBlockForAnyKey(t *testing.T) {
	t.Parallel()

	sr := createSubroundSignatureForCompetingBlockTests(
		&testscommon.SentSignatureTrackerStub{
			GetSignedHashCalled: func(pkBytes []byte, nonce uint64) ([]byte, bool) {
				return nil, false // no key has previously signed
			},
		},
		nil,
		nil,
	)

	result := sr.WaitIfCompetingBlockForNode(context.Background(), 100, []byte("current_hash"))
	assert.False(t, result, "should return false when no key has a competing block")
}

func TestWaitIfCompetingBlockForNode_SameHashForAllKeys(t *testing.T) {
	t.Parallel()

	currentHash := []byte("current_hash")
	sr := createSubroundSignatureForCompetingBlockTests(
		&testscommon.SentSignatureTrackerStub{
			GetSignedHashCalled: func(pkBytes []byte, nonce uint64) ([]byte, bool) {
				return currentHash, true // all keys signed the same hash
			},
		},
		nil,
		nil,
	)

	result := sr.WaitIfCompetingBlockForNode(context.Background(), 100, currentHash)
	assert.False(t, result, "should return false when all keys signed the same hash")
}

func TestWaitIfCompetingBlockForNode_SelfKeyHasCompetingBlock(t *testing.T) {
	container := consensusMocks.InitConsensusCore()
	container.SetRoundHandler(&testscommon.RoundHandlerMock{
		TimeDurationCalled: func() time.Duration {
			return 100 * time.Millisecond
		},
		RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
			return maxTime
		},
	})

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.25,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	selfPk := sr.SelfPubKey()

	srSignature, _ := v2.NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{
			GetSignedHashCalled: func(pkBytes []byte, nonce uint64) ([]byte, bool) {
				if string(pkBytes) == selfPk {
					return []byte("different_hash"), true
				}
				return nil, false
			},
		},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMock.ThrottlerStub{},
	)

	srSignature.SetHeader(&block.Header{Nonce: 100})
	srSignature.SetData([]byte("current_hash"))

	start := time.Now()
	result := srSignature.WaitIfCompetingBlockForNode(context.Background(), 100, []byte("current_hash"))
	elapsed := time.Since(start)

	// Should have waited (delay from round start) and returned false (no proof arrived)
	assert.False(t, result, "should return false after delay expires")
	assert.GreaterOrEqual(t, elapsed, 40*time.Millisecond, "should have waited for competing block delay")
}

func TestWaitIfCompetingBlockForNode_ManagedKeyHasCompetingBlock(t *testing.T) {
	container := consensusMocks.InitConsensusCore()
	container.SetRoundHandler(&testscommon.RoundHandlerMock{
		TimeDurationCalled: func() time.Duration {
			return 100 * time.Millisecond
		},
		RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
			return maxTime
		},
	})

	// Self key has no competing block, but a managed key does
	consensusState := initializers.InitConsensusStateWithKeysHandler(
		&testscommon.KeysHandlerStub{
			IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
				// Mark the first consensus group member as managed
				return string(pkBytes) == "A"
			},
		},
	)
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.25,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	selfPk := sr.SelfPubKey()

	srSignature, _ := v2.NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{
			GetSignedHashCalled: func(pkBytes []byte, nonce uint64) ([]byte, bool) {
				if string(pkBytes) == selfPk {
					// Self key: no competing block
					return nil, false
				}
				if string(pkBytes) == "A" {
					// Managed key "A": has competing block
					return []byte("old_hash"), true
				}
				return nil, false
			},
		},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMock.ThrottlerStub{},
	)

	srSignature.SetHeader(&block.Header{Nonce: 100})
	srSignature.SetData([]byte("current_hash"))

	start := time.Now()
	result := srSignature.WaitIfCompetingBlockForNode(context.Background(), 100, []byte("current_hash"))
	elapsed := time.Since(start)

	// Managed key "A" has a competing block, so the node should wait
	assert.False(t, result, "should return false after delay expires (no proof arrived)")
	assert.GreaterOrEqual(t, elapsed, 40*time.Millisecond, "should have waited for competing block delay")
}

func TestWaitIfCompetingBlockForNode_WaitsOnceNotPerKey(t *testing.T) {
	t.Parallel()

	// This test verifies that waitIfCompetingBlockForNode returns after a single wait
	// even when multiple keys have competing blocks - it should not wait per-key.
	container := consensusMocks.InitConsensusCore()
	container.SetRoundHandler(&testscommon.RoundHandlerMock{
		TimeDurationCalled: func() time.Duration {
			return 100 * time.Millisecond
		},
		RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
			return maxTime
		},
	})

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.25,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	srSignature, _ := v2.NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{
			GetSignedHashCalled: func(pkBytes []byte, nonce uint64) ([]byte, bool) {
				// ALL keys have signed a different hash
				return []byte("old_hash"), true
			},
		},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMock.ThrottlerStub{},
	)

	srSignature.SetHeader(&block.Header{Nonce: 100})
	srSignature.SetData([]byte("current_hash"))

	start := time.Now()
	result := srSignature.WaitIfCompetingBlockForNode(context.Background(), 100, []byte("current_hash"))
	elapsed := time.Since(start)

	// Should return after ONE wait, not multiple
	assert.False(t, result)
	// targetTime = 50ms, sigEndDuration = 85ms, delay = min(50ms, 75ms) = 50ms
	// Should only wait once (~50ms), not per-key
	assert.Less(t, elapsed, 120*time.Millisecond, "should have waited only once, not per-key")
}

func TestShouldSendProof_GracePeriodNotExpired(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	container.SetEquivalentProofsPool(&dataRetriever.ProofsPoolMock{
		HasProofCalled: func(shardID uint32, headerHash []byte) bool {
			return false
		},
	})
	container.SetRoundHandler(&testscommon.RoundHandlerMock{
		TimeDurationCalled: func() time.Duration {
			return 600 * time.Millisecond
		},
		RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
			// positive remaining: grace period not expired
			return 100 * time.Millisecond
		},
		IndexCalled: func() int64 {
			return 1
		},
	})

	srEndRound := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

	// Set self as consensus member so the node is eligible to send proof
	leader, err := srEndRound.GetLeader()
	require.NoError(t, err)
	srEndRound.SetSelfPubKey(leader)

	result := srEndRound.ShouldSendProof()
	assert.True(t, result, "should return true when grace period has not expired and node is in consensus")
}

func TestShouldSendProof_GracePeriodExpired(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	container.SetRoundHandler(&testscommon.RoundHandlerMock{
		TimeDurationCalled: func() time.Duration {
			return 600 * time.Millisecond
		},
		RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
			// negative remaining: grace period expired
			return -100 * time.Millisecond
		},
		IndexCalled: func() int64 {
			return 5
		},
	})

	srEndRound := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

	result := srEndRound.ShouldSendProof()
	assert.False(t, result, "should return false when grace period has expired")
}

func TestShouldSendProof_ProofAlreadyExists(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	container.SetEquivalentProofsPool(&dataRetriever.ProofsPoolMock{
		HasProofCalled: func(shardID uint32, headerHash []byte) bool {
			return true // proof already in pool
		},
	})
	container.SetRoundHandler(&testscommon.RoundHandlerMock{
		TimeDurationCalled: func() time.Duration {
			return 600 * time.Millisecond
		},
		RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
			return 500 * time.Millisecond // grace period not expired
		},
	})

	srEndRound := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

	result := srEndRound.ShouldSendProof()
	assert.False(t, result, "should return false when proof already exists in pool")
}
