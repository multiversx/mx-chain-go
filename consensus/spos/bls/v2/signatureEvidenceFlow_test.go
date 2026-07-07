package v2

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

const flowRoundTimeDuration = 100 * time.Millisecond

var flowChainID = []byte("chain ID")

const flowCurrentPid = core.PeerID("pid")

func createFlowSubround(t *testing.T, container *spos.ConsensusCore, current int, name string) *spos.Subround {
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, err := spos.NewSubround(
		current-1,
		current,
		current+1,
		flowRoundTimeDuration,
		0.25,
		0.85,
		name,
		consensusState,
		ch,
		func() {},
		container,
		flowChainID,
		flowCurrentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	require.Nil(t, err)

	return sr
}

func createFlowWorker() *consensusMocks.SposWorkerMock {
	return &consensusMocks.SposWorkerMock{
		ConsensusMetricsCalled: func() spos.ConsensusMetricsHandler {
			consensusMetrics, _ := spos.NewConsensusMetrics(&statusHandler.AppStatusHandlerStub{})
			return consensusMetrics
		},
	}
}

func createFlowStartRound(t *testing.T, container *spos.ConsensusCore, store signatureEvidenceHandler) *subroundStartRound {
	sr := createFlowSubround(t, container, bls.SrStartRound, "(START_ROUND)")
	srStartRound, err := NewSubroundStartRound(
		sr,
		ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		createFlowWorker(),
		store,
	)
	require.Nil(t, err)

	return srStartRound
}

func createFlowSignatureSubround(t *testing.T, container *spos.ConsensusCore, store signatureEvidenceHandler) *subroundSignature {
	sr := createFlowSubround(t, container, bls.SrSignature, "(SIGNATURE)")
	srSignature, err := NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		createFlowWorker(),
		&dataRetrieverMock.ThrottlerStub{},
		store,
	)
	require.Nil(t, err)

	return srSignature
}

func TestSubroundStartRound_CaptureSignatureEvidence(t *testing.T) {
	t.Parallel()

	t.Run("snapshot reflects the pre-reset consensus state", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		container.SetSigningHandler(&consensusMocks.SigningHandlerStub{
			SignatureShareCalled: func(index uint16) ([]byte, error) {
				if index == 2 {
					return nil, errors.New("missing share")
				}
				return []byte(fmt.Sprintf("share_%d", index)), nil
			},
		})
		container.SetEquivalentProofsPool(createUnsettledProofsPool())

		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())
		srStartRound := createFlowStartRound(t, container, store)

		srStartRound.SetRoundIndex(10)
		srStartRound.SetHeader(&block.Header{Nonce: 5, Epoch: 2, Round: 20})
		srStartRound.SetData([]byte("header hash"))
		srStartRound.SetThreshold(bls.SrSignature, 7)
		group := srStartRound.ConsensusGroup()
		for i := 0; i < 3; i++ {
			_ = srStartRound.SetJobDone(group[i], bls.SrSignature, true)
		}

		ok := srStartRound.doStartRoundJob(context.Background())
		require.True(t, ok)

		ev, found := store.GetPreviousRoundEvidence(5, 11)
		require.True(t, found)
		assert.Equal(t, int64(10), ev.roundIndex)
		assert.Equal(t, []byte("header hash"), ev.headerHash)
		assert.Equal(t, uint32(2), ev.epoch)
		assert.Equal(t, uint64(20), ev.headerRound)
		assert.Equal(t, 7, ev.threshold)
		assert.Equal(t, group, ev.consensusGroup)
		// index 2 has job done but no share bytes: excluded from count and bitmap
		assert.Equal(t, 2, ev.getCount())
		bitmap, shares := ev.getAggregationData()
		assert.Equal(t, []byte("share_0"), shares[0])
		assert.Equal(t, []byte("share_1"), shares[1])
		assert.Nil(t, shares[2])
		assert.Equal(t, byte(0b00000011), bitmap[0])
	})

	t.Run("nil header clears the slot", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())
		srStartRound := createFlowStartRound(t, container, store)

		store.Capture(createTestEvidence(9, 5, 7, 3))
		srStartRound.SetRoundIndex(10)

		ok := srStartRound.doStartRoundJob(context.Background())
		require.True(t, ok)

		_, found := store.GetPreviousRoundEvidence(5, 10)
		assert.False(t, found)
	})

	t.Run("no signatures observed clears the slot", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())
		srStartRound := createFlowStartRound(t, container, store)

		srStartRound.SetRoundIndex(10)
		srStartRound.SetHeader(&block.Header{Nonce: 5})
		srStartRound.SetData([]byte("header hash"))

		srStartRound.captureSignatureEvidence()

		_, found := store.GetPreviousRoundEvidence(5, 11)
		assert.False(t, found)
	})
}

func TestSubroundStartRound_CaptureTriggersSelfAssemblyOnQuorum(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	container.SetSigningHandler(&consensusMocks.SigningHandlerStub{})
	container.SetEquivalentProofsPool(createUnsettledProofsPool())

	broadcastChan := make(chan data.HeaderProofHandler, 2)
	container.SetBroadcastMessenger(&consensusMocks.BroadcastMessengerMock{
		BroadcastEquivalentProofCalled: func(proof data.HeaderProofHandler, pkBytes []byte) error {
			broadcastChan <- proof
			return nil
		},
	})

	store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())
	srStartRound := createFlowStartRound(t, container, store)

	srStartRound.SetRoundIndex(10)
	srStartRound.SetHeader(&block.Header{Nonce: 5, Round: 20})
	srStartRound.SetData([]byte("header hash"))
	srStartRound.SetThreshold(bls.SrSignature, 7)
	group := srStartRound.ConsensusGroup()
	for i := 0; i < 8; i++ {
		_ = srStartRound.SetJobDone(group[i], bls.SrSignature, true)
	}

	srStartRound.captureSignatureEvidence()

	select {
	case proof := <-broadcastChan:
		assert.Equal(t, []byte("header hash"), proof.GetHeaderHash())
		assert.Equal(t, uint64(5), proof.GetHeaderNonce())
	case <-time.After(time.Second):
		require.Fail(t, "self-assembly was not triggered at round start")
	}

	// retained evidence re-fires on the next round while the nonce stays unsettled
	srStartRound.SetHeader(nil)
	srStartRound.SetData(nil)
	srStartRound.captureSignatureEvidence()

	select {
	case <-broadcastChan:
	case <-time.After(time.Second):
		require.Fail(t, "self-assembly was not re-triggered for retained evidence")
	}
}

func TestSubroundSignature_ShouldAbortOnSignatureEvidence(t *testing.T) {
	t.Parallel()

	currentHash := []byte("current hash")
	currentRound := int64(11)

	createContainerWithRoundIndex := func() *spos.ConsensusCore {
		container := consensusMocks.InitConsensusCore()
		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return currentRound
			},
		})
		return container
	}

	t.Run("no evidence falls through", func(t *testing.T) {
		t.Parallel()

		container := createContainerWithRoundIndex()
		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())
		srSignature := createFlowSignatureSubround(t, container, store)

		assert.False(t, srSignature.shouldAbortOnSignatureEvidence(context.Background(), 5, currentHash))
	})

	t.Run("evidence for the same hash falls through", func(t *testing.T) {
		t.Parallel()

		container := createContainerWithRoundIndex()
		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())
		srSignature := createFlowSignatureSubround(t, container, store)

		ev := createTestEvidence(10, 5, 7, 8)
		store.Capture(ev)

		assert.False(t, srSignature.shouldAbortOnSignatureEvidence(context.Background(), 5, ev.headerHash))
	})

	t.Run("evidence from an older round falls through", func(t *testing.T) {
		t.Parallel()

		container := createContainerWithRoundIndex()
		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())
		srSignature := createFlowSignatureSubround(t, container, store)

		store.Capture(createTestEvidence(9, 5, 7, 6))

		assert.False(t, srSignature.shouldAbortOnSignatureEvidence(context.Background(), 5, currentHash))
	})

	t.Run("quorum evidence refuses and triggers assembly exactly once", func(t *testing.T) {
		t.Parallel()

		var numAggregations atomic.Int32
		aggregationChan := make(chan struct{}, 2)

		container := createContainerWithRoundIndex()
		container.SetSigningHandler(&consensusMocks.SigningHandlerStub{
			AggregateSigsWithKeysCalled: func(pubKeys []string, bitmap []byte, sigShares [][]byte, epoch uint32) ([]byte, error) {
				numAggregations.Add(1)
				aggregationChan <- struct{}{}
				return []byte("agg"), nil
			},
		})
		container.SetEquivalentProofsPool(createUnsettledProofsPool())

		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())
		srSignature := createFlowSignatureSubround(t, container, store)

		store.Capture(createTestEvidence(10, 5, 7, 7))

		assert.True(t, srSignature.shouldAbortOnSignatureEvidence(context.Background(), 5, currentHash))

		select {
		case <-aggregationChan:
		case <-time.After(time.Second):
			require.Fail(t, "self-assembly was not triggered")
		}

		// second evaluation still refuses but must not start a second assembly
		assert.True(t, srSignature.shouldAbortOnSignatureEvidence(context.Background(), 5, currentHash))
		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, int32(1), numAggregations.Load())
	})

	t.Run("significant evidence waits and aborts when the proof arrives", func(t *testing.T) {
		t.Parallel()

		ev := createTestEvidence(10, 5, 7, 4)

		container := createContainerWithRoundIndex()
		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return currentRound
			},
			TimeDurationCalled: func() time.Duration {
				return 600 * time.Millisecond
			},
			RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
				return maxTime
			},
		})
		proofsPool := createUnsettledProofsPool()
		proofsPool.HasProofCalled = func(shardID uint32, headerHash []byte) bool {
			return string(headerHash) == string(ev.headerHash)
		}
		container.SetEquivalentProofsPool(proofsPool)

		store, _ := newSignatureEvidenceStore(proofsPool)
		srSignature := createFlowSignatureSubround(t, container, store)

		store.Capture(ev)

		start := time.Now()
		aborted := srSignature.shouldAbortOnSignatureEvidence(context.Background(), 5, currentHash)
		assert.True(t, aborted)
		assert.Less(t, time.Since(start), 200*time.Millisecond)
	})

	t.Run("significant evidence proceeds after the wait expires with no proof", func(t *testing.T) {
		t.Parallel()

		container := createContainerWithRoundIndex()
		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return currentRound
			},
			TimeDurationCalled: func() time.Duration {
				return 100 * time.Millisecond
			},
			RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
				return maxTime
			},
		})
		proofsPool := createUnsettledProofsPool()
		proofsPool.HasProofCalled = func(shardID uint32, headerHash []byte) bool {
			return false
		}
		container.SetEquivalentProofsPool(proofsPool)

		store, _ := newSignatureEvidenceStore(proofsPool)
		srSignature := createFlowSignatureSubround(t, container, store)

		store.Capture(createTestEvidence(10, 5, 7, 4))

		assert.False(t, srSignature.shouldAbortOnSignatureEvidence(context.Background(), 5, currentHash))
	})

	t.Run("few observed shares fall through", func(t *testing.T) {
		t.Parallel()

		container := createContainerWithRoundIndex()
		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())
		srSignature := createFlowSignatureSubround(t, container, store)

		store.Capture(createTestEvidence(10, 5, 7, 3))

		assert.False(t, srSignature.shouldAbortOnSignatureEvidence(context.Background(), 5, currentHash))
	})
}

func TestSubroundBlock_HasQuorumEvidenceForCompetingBlock(t *testing.T) {
	t.Parallel()

	currentRound := int64(11)

	createBlockSubround := func(t *testing.T, container *spos.ConsensusCore, store signatureEvidenceHandler) *subroundBlock {
		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return currentRound
			},
		})
		container.SetBlockchain(&testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.Header{Nonce: 4}
			},
		})

		sr := createFlowSubround(t, container, bls.SrBlock, "(BLOCK)")
		srBlock, err := NewSubroundBlock(
			sr,
			ProcessingThresholdPercent,
			createFlowWorker(),
			&consensusMocks.NtpSyncControllerMock{},
			&dataRetrieverMock.ThrottlerStub{},
			store,
		)
		require.Nil(t, err)

		return srBlock
	}

	t.Run("quorum evidence skips the proposal and triggers assembly", func(t *testing.T) {
		t.Parallel()

		aggregationChan := make(chan struct{}, 1)
		container := consensusMocks.InitConsensusCore()
		container.SetSigningHandler(&consensusMocks.SigningHandlerStub{
			AggregateSigsWithKeysCalled: func(pubKeys []string, bitmap []byte, sigShares [][]byte, epoch uint32) ([]byte, error) {
				aggregationChan <- struct{}{}
				return []byte("agg"), nil
			},
		})
		container.SetEquivalentProofsPool(createUnsettledProofsPool())

		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())
		srBlock := createBlockSubround(t, container, store)

		store.Capture(createTestEvidence(10, 5, 7, 7))

		assert.True(t, srBlock.hasQuorumEvidenceForCompetingBlock())

		select {
		case <-aggregationChan:
		case <-time.After(time.Second):
			require.Fail(t, "self-assembly was not triggered")
		}
	})

	t.Run("sub-quorum evidence does not skip the proposal", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())
		srBlock := createBlockSubround(t, container, store)

		store.Capture(createTestEvidence(10, 5, 7, 6))

		assert.False(t, srBlock.hasQuorumEvidenceForCompetingBlock())
	})

	t.Run("no evidence does not skip the proposal", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())
		srBlock := createBlockSubround(t, container, store)

		assert.False(t, srBlock.hasQuorumEvidenceForCompetingBlock())
	})
}
