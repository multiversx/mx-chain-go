package sync

import (
	"math"
	stdsync "sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
)

func newBranchAwareForkDetector(shardID uint32, finalNonce uint64, finalHash []byte) *baseForkDetector {
	finalCheckpoint := &checkpointInfo{nonce: finalNonce, round: finalNonce, hash: finalHash}
	bfd := &baseForkDetector{
		shardID: shardID,
		headers: make(map[uint64][]*headerInfo),
		fork: forkInfo{
			finalCheckpoint: finalCheckpoint,
		},
		enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, _ uint32) bool {
				return flag == common.AndromedaFlag || flag == common.SupernovaFlag
			},
		},
		enableRoundsHandler: &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, _ uint64) bool {
				return flag == common.SupernovaRoundFlag
			},
		},
	}
	bfd.headers[finalNonce] = []*headerInfo{{
		epoch:    1,
		nonce:    finalNonce,
		round:    finalNonce,
		hash:     finalHash,
		state:    process.BHProcessed,
		hasProof: true,
	}}

	return bfd
}

func TestShardForkDetector_SettlementRecomputesProbableHighestNonce(t *testing.T) {
	t.Parallel()

	parentHash := []byte("P")
	processedHash := []byte("A")
	settledHash := []byte("B")
	bfd := newBranchAwareForkDetector(0, 10, parentHash)
	bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: parentHash}
	bfd.fork.checkpoint = []*checkpointInfo{
		{nonce: 10, round: 10, hash: parentHash},
		{nonce: 11, round: 11, hash: processedHash},
	}
	addProvenHeaderInfo(bfd, 11, processedHash, parentHash, process.BHReceived)
	addProvenHeaderInfo(bfd, 11, settledHash, parentHash, process.BHReceived)
	addProvenHeaderInfo(bfd, 11, processedHash, parentHash, process.BHProcessed)
	addProvenHeaderInfo(bfd, 12, []byte("C"), settledHash, process.BHReceived)
	bfd.proofsPool = &dataRetrieverMock.ProofsPoolMock{}
	bfd.setProbableHighestNonce(11)

	sfd := &shardForkDetector{baseForkDetector: bfd}
	bfd.forkDetector = sfd
	sfd.ReceivedSelfNotarizedFromCrossHeaders(
		core.MetachainShardId,
		[]data.HeaderHandler{&block.HeaderV3{Epoch: 1, Nonce: 11, Round: 12, PrevHash: parentHash, ShardID: 0}},
		[][]byte{settledHash},
	)

	require.Equal(t, uint64(12), sfd.ProbableHighestNonce())
}

func TestShardForkDetector_SettlementConflictChecksAreV3Only(t *testing.T) {
	t.Parallel()

	newDetector := func(supernovaRoundEnabled bool) *shardForkDetector {
		bfd := newBranchAwareForkDetector(0, 10, []byte("P"))
		bfd.enableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, _ uint64) bool {
				return flag == common.SupernovaRoundFlag && supernovaRoundEnabled
			},
		}

		return &shardForkDetector{baseForkDetector: bfd}
	}

	records := []*headerInfo{
		{epoch: 1, round: 11, hash: []byte("A"), state: process.BHProcessed},
		{epoch: 1, round: 11, hash: []byte("B"), state: process.BHProcessed},
		{epoch: 1, round: 11, hash: []byte("A"), state: process.BHNotarized},
		{epoch: 1, round: 11, hash: []byte("B"), state: process.BHNotarized},
	}

	drainDetector := newDetector(false)
	processedIndex, notarizedIndex := drainDetector.getProcessedAndNotarizedIndexes(records)
	require.Equal(t, 1, processedIndex)
	require.Equal(t, 3, notarizedIndex)

	v3Detector := newDetector(true)
	processedIndex, notarizedIndex = v3Detector.getProcessedAndNotarizedIndexes(records)
	require.Equal(t, -1, processedIndex)
	require.Equal(t, -1, notarizedIndex)

	processedChild := &headerInfo{
		epoch: 1, nonce: 11, round: 11, hash: []byte("A"), prevHash: []byte("P"),
		state: process.BHProcessed, hasProof: true,
	}
	conflictingNotarization := &headerInfo{
		epoch: 1, nonce: 11, round: 11, hash: []byte("B"), prevHash: []byte("P"),
		state: process.BHNotarized,
	}
	drainDetector.headers[11] = []*headerInfo{processedChild, conflictingNotarization}
	v3Detector.headers[11] = []*headerInfo{processedChild, conflictingNotarization}

	require.Same(t, processedChild, drainDetector.getCleanProcessedChild(&checkpointInfo{nonce: 10, round: 10, hash: []byte("P")}))
	require.Nil(t, v3Detector.getCleanProcessedChild(&checkpointInfo{nonce: 10, round: 10, hash: []byte("P")}))

	drainDetector.proofsPool = &dataRetrieverMock.ProofsPoolMock{}
	drainDetector.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: []byte("P")}
	drainDetector.setProbableHighestNonce(20)
	drainDetector.ReceivedSelfNotarizedFromCrossHeaders(
		core.MetachainShardId,
		[]data.HeaderHandler{&block.Header{Epoch: 1, Nonce: 12, Round: 12, PrevHash: []byte("A"), ShardID: 0}},
		[][]byte{[]byte("legacy")},
	)
	require.Equal(t, uint64(20), drainDetector.ProbableHighestNonce())
}

func TestBaseForkDetector_CompetingSiblingProofDefersV3Finality(t *testing.T) {
	t.Parallel()

	parentHash := []byte("P")
	headerHash := []byte("A")
	header := &block.HeaderV3{Epoch: 1, Nonce: 11, Round: 11, PrevHash: parentHash}
	bfd := newBranchAwareForkDetector(0, 10, parentHash)
	addProvenHeaderInfo(bfd, 11, []byte("B"), parentHash, process.BHReceived)

	require.False(t, bfd.canInstantlyFinalize(header, headerHash))

	bfd.headers[11][0].prevHash = []byte("other-parent")
	require.True(t, bfd.canInstantlyFinalize(header, headerHash))

	bfd.headers[11][0].prevHash = parentHash
	bfd.enableRoundsHandler = &testscommon.EnableRoundsHandlerStub{}
	require.True(t, bfd.canInstantlyFinalize(header, headerHash))
}

func TestShardForkDetector_CompetingSiblingProofStopsFinalityCascade(t *testing.T) {
	t.Parallel()

	parentHash := []byte("P")
	processedHash := []byte("A")
	bfd := newBranchAwareForkDetector(0, 10, parentHash)
	processedChild := &headerInfo{
		epoch: 1, nonce: 11, round: 11, hash: processedHash, prevHash: parentHash,
		state: process.BHProcessed, hasProof: true,
	}
	bfd.headers[11] = []*headerInfo{
		processedChild,
		{
			epoch: 1, nonce: 11, round: 12, hash: []byte("B"), prevHash: parentHash,
			state: process.BHReceived, hasProof: true,
		},
	}
	sfd := &shardForkDetector{baseForkDetector: bfd}

	require.Nil(t, sfd.getCleanProcessedChild(bfd.finalCheckpoint()))
}

func addProvenHeaderInfo(
	bfd *baseForkDetector,
	nonce uint64,
	hash []byte,
	prevHash []byte,
	state process.BlockHeaderState,
) {
	bfd.headers[nonce] = append(bfd.headers[nonce], &headerInfo{
		epoch:    1,
		nonce:    nonce,
		round:    nonce,
		hash:     hash,
		prevHash: prevHash,
		state:    state,
		hasProof: true,
	})
}

func TestBaseForkDetector_RecomputeProbableHighestNonceSerializesComputeAndPublish(t *testing.T) {
	t.Parallel()

	computeEntered := make(chan struct{})
	resumeCompute := make(chan struct{})
	var resumeComputeOnce stdsync.Once
	resumeBlockedCompute := func() {
		resumeComputeOnce.Do(func() {
			close(resumeCompute)
		})
	}
	t.Cleanup(resumeBlockedCompute)
	var blockComputeOnce stdsync.Once

	bfd := newBranchAwareForkDetector(0, 10, []byte("P"))
	bfd.enableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			if flag != common.SupernovaFlag {
				return false
			}
			if epoch == 2 {
				blockComputeOnce.Do(func() {
					close(computeEntered)
					<-resumeCompute
				})
			}

			return true
		},
	}
	bfd.headers[11] = []*headerInfo{{
		epoch: 2, nonce: 11, round: 11, hash: []byte("A"), prevHash: []byte("P"),
		state: process.BHReceived, hasProof: true,
	}}

	result := make(chan uint64, 1)
	go func() {
		result <- bfd.recomputeProbableHighestNonce()
	}()

	select {
	case <-computeEntered:
	case <-time.After(time.Second):
		require.FailNow(t, "probable-highest computation did not start")
	}
	lockAcquiredDuringCompute := bfd.mutProbableHighestNonceUpdate.TryLock()
	if lockAcquiredDuringCompute {
		bfd.mutProbableHighestNonceUpdate.Unlock()
	}
	resumeBlockedCompute()
	select {
	case probableHighestNonce := <-result:
		require.Equal(t, uint64(11), probableHighestNonce)
	case <-time.After(time.Second):
		require.FailNow(t, "probable-highest computation did not finish")
	}
	require.False(t, lockAcquiredDuringCompute)
	require.Equal(t, uint64(11), bfd.probableHighestNonce())

	require.True(t, bfd.mutProbableHighestNonceUpdate.TryLock())
	bfd.mutProbableHighestNonceUpdate.Unlock()
}

func TestBaseForkDetector_CheckForkUsesAncestryForSupernovaHeaders(t *testing.T) {
	t.Parallel()

	finalHash := []byte("A")
	processedHash := []byte("D")
	competitorHash := []byte("C")
	offBranchParentHash := []byte("B")

	checkFork := func(
		shardID uint32,
		supernovaRoundEnabled bool,
		competitorState process.BlockHeaderState,
		competitorParentHash []byte,
	) *process.ForkInfo {
		bfd := newBranchAwareForkDetector(shardID, 10, finalHash)
		bfd.enableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, _ uint64) bool {
				return flag == common.SupernovaRoundFlag && supernovaRoundEnabled
			},
		}
		bfd.roundHandler = &testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 14
			},
		}
		bfd.processConfigsHandler = testscommon.GetDefaultProcessConfigsHandler()
		bfd.proofsPool = &dataRetrieverMock.ProofsPoolMock{
			HasProofCalled: func(_ uint32, _ []byte) bool {
				return true
			},
		}
		bfd.fork.rollBackNonce = math.MaxUint64
		bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: finalHash}
		bfd.fork.checkpoint = []*checkpointInfo{
			{nonce: 10, round: 10, hash: finalHash},
			{nonce: 11, round: 14, hash: processedHash},
		}
		bfd.headers[11] = []*headerInfo{
			{
				epoch: 1, nonce: 11, round: 14, hash: processedHash, prevHash: finalHash,
				state: process.BHProcessed, hasProof: true,
			},
			{
				epoch: 1, nonce: 11, round: 13, hash: competitorHash, prevHash: competitorParentHash,
				state: competitorState, hasProof: true,
			},
		}
		bfd.setProbableHighestNonce(11)

		return bfd.CheckFork()
	}

	t.Run("off-branch V3 proof does not roll back the processed branch", func(t *testing.T) {
		t.Parallel()

		forkInfo := checkFork(0, true, process.BHReceived, offBranchParentHash)
		require.False(t, forkInfo.IsDetected)
	})

	t.Run("unknown V3 ancestry waits for the header", func(t *testing.T) {
		t.Parallel()

		forkInfo := checkFork(0, true, process.BHReceived, nil)
		require.False(t, forkInfo.IsDetected)
	})

	t.Run("same-parent V3 competitor retains the round rule", func(t *testing.T) {
		t.Parallel()

		forkInfo := checkFork(0, true, process.BHReceived, finalHash)
		require.True(t, forkInfo.IsDetected)
		require.Equal(t, uint64(11), forkInfo.Nonce)
		require.Equal(t, competitorHash, forkInfo.Hash)
	})

	t.Run("metachain-notarized V3 competitor remains authoritative", func(t *testing.T) {
		t.Parallel()

		forkInfo := checkFork(0, true, process.BHNotarized, offBranchParentHash)
		require.True(t, forkInfo.IsDetected)
		require.Equal(t, uint64(11), forkInfo.Nonce)
		require.Equal(t, competitorHash, forkInfo.Hash)
	})

	t.Run("pre-Supernova shard behavior is unchanged", func(t *testing.T) {
		t.Parallel()

		forkInfo := checkFork(0, false, process.BHReceived, offBranchParentHash)
		require.True(t, forkInfo.IsDetected)
		require.Equal(t, uint64(11), forkInfo.Nonce)
		require.Equal(t, competitorHash, forkInfo.Hash)
	})

	t.Run("off-branch V3 metachain proof does not roll back the processed branch", func(t *testing.T) {
		t.Parallel()

		forkInfo := checkFork(core.MetachainShardId, true, process.BHReceived, offBranchParentHash)
		require.False(t, forkInfo.IsDetected)
	})

	t.Run("same-parent V3 metachain competitor rolls back only the child", func(t *testing.T) {
		t.Parallel()

		forkInfo := checkFork(core.MetachainShardId, true, process.BHReceived, finalHash)
		require.True(t, forkInfo.IsDetected)
		require.Equal(t, uint64(11), forkInfo.Nonce)
		require.Equal(t, competitorHash, forkInfo.Hash)
	})

	t.Run("pre-Supernova metachain behavior is unchanged", func(t *testing.T) {
		t.Parallel()

		forkInfo := checkFork(core.MetachainShardId, false, process.BHReceived, offBranchParentHash)
		require.True(t, forkInfo.IsDetected)
		require.Equal(t, uint64(11), forkInfo.Nonce)
		require.Equal(t, competitorHash, forkInfo.Hash)
	})
}

func TestBaseForkDetector_CheckForkIgnoresOffBranchEpochWhenSelectingV3Sibling(t *testing.T) {
	t.Parallel()

	parentHash := []byte("P")
	processedHash := []byte("B")
	siblingHash := []byte("A")
	bfd := newBranchAwareForkDetector(core.MetachainShardId, 10, parentHash)
	bfd.roundHandler = &testscommon.RoundHandlerMock{IndexCalled: func() int64 { return 20 }}
	bfd.processConfigsHandler = testscommon.GetDefaultProcessConfigsHandler()
	bfd.proofsPool = &dataRetrieverMock.ProofsPoolMock{HasProofCalled: func(_ uint32, _ []byte) bool { return true }}
	bfd.fork.rollBackNonce = math.MaxUint64
	bfd.fork.checkpoint = []*checkpointInfo{
		{nonce: 10, round: 10, hash: parentHash},
		{nonce: 11, round: 14, hash: processedHash},
	}
	bfd.headers[11] = []*headerInfo{
		{epoch: 1, nonce: 11, round: 14, hash: processedHash, prevHash: parentHash, state: process.BHProcessed, hasProof: true},
		{epoch: 1, nonce: 11, round: 13, hash: siblingHash, prevHash: parentHash, state: process.BHReceived, hasProof: true},
		{epoch: 2, nonce: 11, round: 12, hash: []byte("off_branch"), prevHash: []byte("Q"), state: process.BHReceived, hasProof: true},
	}

	forkInfo := bfd.CheckFork()
	require.True(t, forkInfo.IsDetected)
	require.Equal(t, uint64(11), forkInfo.Nonce)
	require.Equal(t, siblingHash, forkInfo.Hash)
}

func TestBaseForkDetector_ComputeProbableHighestNonceUsesFinalV3Branch(t *testing.T) {
	t.Parallel()

	finalHash := []byte("A")

	t.Run("known losing child does not advance probable", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(0, 10, finalHash)
		addProvenHeaderInfo(bfd, 10, []byte("B"), []byte("P"), process.BHReceived)
		addProvenHeaderInfo(bfd, 11, []byte("C"), []byte("B"), process.BHReceived)

		require.Equal(t, uint64(10), bfd.computeProbableHighestNonce())
	})

	t.Run("canonical descendants advance probable", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(0, 10, finalHash)
		addProvenHeaderInfo(bfd, 11, []byte("D"), finalHash, process.BHReceived)
		addProvenHeaderInfo(bfd, 12, []byte("E"), []byte("D"), process.BHReceived)

		require.Equal(t, uint64(12), bfd.computeProbableHighestNonce())
	})

	t.Run("canonical descendants advance while a losing suffix remains", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(0, 10, finalHash)
		addProvenHeaderInfo(bfd, 10, []byte("B"), []byte("P"), process.BHReceived)
		addProvenHeaderInfo(bfd, 11, []byte("C"), []byte("B"), process.BHReceived)
		addProvenHeaderInfo(bfd, 11, []byte("D"), finalHash, process.BHReceived)
		addProvenHeaderInfo(bfd, 12, []byte("F"), []byte("C"), process.BHReceived)
		addProvenHeaderInfo(bfd, 12, []byte("E"), []byte("D"), process.BHReceived)

		require.Equal(t, uint64(12), bfd.computeProbableHighestNonce())
	})

	t.Run("complete losing suffix does not advance probable", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(0, 10, finalHash)
		addProvenHeaderInfo(bfd, 10, []byte("B"), []byte("P"), process.BHReceived)
		addProvenHeaderInfo(bfd, 11, []byte("C"), []byte("B"), process.BHReceived)
		addProvenHeaderInfo(bfd, 12, []byte("F"), []byte("C"), process.BHReceived)

		require.Equal(t, uint64(10), bfd.computeProbableHighestNonce())
	})

	t.Run("unknown higher lineage keeps raw probable", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(0, 10, finalHash)
		addProvenHeaderInfo(bfd, 11, []byte("C"), []byte("B"), process.BHReceived)
		addProvenHeaderInfo(bfd, 12, []byte("E"), []byte("D"), process.BHReceived)

		require.Equal(t, uint64(12), bfd.computeProbableHighestNonce())
	})

	t.Run("missing ancestry keeps raw probable", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(0, 10, finalHash)
		addProvenHeaderInfo(bfd, 11, []byte("C"), nil, process.BHReceived)

		require.Equal(t, uint64(11), bfd.computeProbableHighestNonce())
	})

	t.Run("gap keeps raw probable", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(0, 10, finalHash)
		addProvenHeaderInfo(bfd, 12, []byte("E"), []byte("D"), process.BHReceived)

		require.Equal(t, uint64(12), bfd.computeProbableHighestNonce())
	})

	t.Run("unique notarized matching child resolves contention", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(0, 10, finalHash)
		addProvenHeaderInfo(bfd, 11, []byte("D1"), finalHash, process.BHReceived)
		addProvenHeaderInfo(bfd, 11, []byte("D2"), finalHash, process.BHReceived)
		addProvenHeaderInfo(bfd, 11, []byte("D2"), finalHash, process.BHNotarized)
		addProvenHeaderInfo(bfd, 12, []byte("E"), []byte("D2"), process.BHReceived)

		require.Equal(t, uint64(12), bfd.computeProbableHighestNonce())
	})
}

func TestBaseForkDetector_ComputeProbableHighestNonceResolvesV3Equivocation(t *testing.T) {
	t.Parallel()

	parentHash := []byte("P")
	preferredHash := []byte("A")
	laterSiblingHash := []byte("B")
	laterSiblingChildHash := []byte("C")

	newDetector := func(processedHash []byte) *baseForkDetector {
		bfd := newBranchAwareForkDetector(0, 10, parentHash)
		bfd.fork.checkpoint = []*checkpointInfo{{nonce: 10, round: 10, hash: parentHash}}
		addProvenHeaderInfo(bfd, 11, preferredHash, parentHash, process.BHReceived)
		bfd.headers[11][0].round = 11
		addProvenHeaderInfo(bfd, 11, laterSiblingHash, parentHash, process.BHReceived)
		bfd.headers[11][1].round = 12
		addProvenHeaderInfo(bfd, 12, laterSiblingChildHash, laterSiblingHash, process.BHReceived)
		if processedHash != nil {
			addProvenHeaderInfo(bfd, 11, processedHash, parentHash, process.BHProcessed)
			bfd.fork.checkpoint = append(bfd.fork.checkpoint, &checkpointInfo{nonce: 11, round: 11, hash: processedHash})
		}

		return bfd
	}

	t.Run("preferred processed sibling makes losing child suffix non-actionable", func(t *testing.T) {
		t.Parallel()

		bfd := newDetector(preferredHash)

		require.Equal(t, uint64(11), bfd.computeProbableHighestNonce())
	})

	t.Run("losing processed sibling preserves raw frontier for rollback", func(t *testing.T) {
		t.Parallel()

		bfd := newDetector(laterSiblingHash)

		require.Equal(t, uint64(12), bfd.computeProbableHighestNonce())
	})

	t.Run("metachain notarization overrides proof ordering", func(t *testing.T) {
		t.Parallel()

		bfd := newDetector(nil)
		bfd.headers[11] = append(bfd.headers[11], &headerInfo{
			epoch:    1,
			nonce:    11,
			round:    12,
			hash:     laterSiblingHash,
			prevHash: parentHash,
			state:    process.BHNotarized,
		})

		require.Equal(t, uint64(12), bfd.computeProbableHighestNonce())
	})

	t.Run("notarization conflicting with processed sibling preserves rollback signal", func(t *testing.T) {
		t.Parallel()

		bfd := newDetector(preferredHash)
		bfd.headers[11] = append(bfd.headers[11], &headerInfo{
			epoch:    1,
			nonce:    11,
			round:    12,
			hash:     laterSiblingHash,
			prevHash: parentHash,
			state:    process.BHNotarized,
		})

		require.Equal(t, uint64(12), bfd.computeProbableHighestNonce())
	})

	t.Run("first V3 contention above V2 final is classified", func(t *testing.T) {
		t.Parallel()

		bfd := newDetector(preferredHash)
		bfd.headers[10][0].epoch = 0
		bfd.headers[10][0].round = 0

		require.Equal(t, uint64(11), bfd.computeProbableHighestNonce())
	})

	t.Run("siblings from different epochs remain conservative", func(t *testing.T) {
		t.Parallel()

		bfd := newDetector(preferredHash)
		bfd.headers[11][1].epoch = 2

		require.Equal(t, uint64(12), bfd.computeProbableHighestNonce())
	})
}

func TestBaseForkDetector_ComputeProbableHighestNoncePreservesOtherModes(t *testing.T) {
	t.Parallel()

	finalHash := []byte("A")

	t.Run("V3 metachain ignores a known losing child", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(core.MetachainShardId, 10, finalHash)
		addProvenHeaderInfo(bfd, 10, []byte("B"), []byte("P"), process.BHReceived)
		addProvenHeaderInfo(bfd, 11, []byte("C"), []byte("B"), process.BHReceived)

		require.Equal(t, uint64(10), bfd.computeProbableHighestNonce())
	})

	t.Run("V3 ignores a proven losing child without its parent proof", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(0, 10, finalHash)
		addProvenHeaderInfo(bfd, 11, []byte("C"), []byte("B"), process.BHReceived)

		require.Equal(t, uint64(10), bfd.computeProbableHighestNonce())
	})

	t.Run("V3 metachain advances on canonical descendants while ignoring a losing suffix", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(core.MetachainShardId, 10, finalHash)
		addProvenHeaderInfo(bfd, 10, []byte("B"), []byte("P"), process.BHReceived)
		addProvenHeaderInfo(bfd, 11, []byte("C"), []byte("B"), process.BHReceived)
		addProvenHeaderInfo(bfd, 11, []byte("D"), finalHash, process.BHReceived)
		addProvenHeaderInfo(bfd, 12, []byte("F"), []byte("C"), process.BHReceived)
		addProvenHeaderInfo(bfd, 12, []byte("E"), []byte("D"), process.BHReceived)

		require.Equal(t, uint64(12), bfd.computeProbableHighestNonce())
	})

	t.Run("legacy metachain remains raw", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(core.MetachainShardId, 10, finalHash)
		bfd.enableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
		addProvenHeaderInfo(bfd, 11, []byte("C"), []byte("B"), process.BHReceived)

		require.Equal(t, uint64(11), bfd.computeProbableHighestNonce())
	})

	t.Run("legacy final anchor remains raw", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(0, 10, finalHash)
		bfd.headers[10][0].epoch = 0
		bfd.enableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
		addProvenHeaderInfo(bfd, 11, []byte("C"), []byte("B"), process.BHReceived)

		require.Equal(t, uint64(11), bfd.computeProbableHighestNonce())
	})

	t.Run("maximum nonce does not wrap", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(0, math.MaxUint64-1, finalHash)
		addProvenHeaderInfo(bfd, math.MaxUint64-1, []byte("B"), []byte("P"), process.BHReceived)
		addProvenHeaderInfo(bfd, math.MaxUint64, []byte("C"), []byte("B"), process.BHReceived)

		require.Equal(t, uint64(math.MaxUint64-1), bfd.computeProbableHighestNonce())
	})
}

func TestBaseForkDetector_AppendHeaderInfoEnrichesWithoutInsertion(t *testing.T) {
	t.Parallel()

	bfd := newBranchAwareForkDetector(0, 10, []byte("A"))
	proofInfo := &headerInfo{
		epoch:    1,
		nonce:    11,
		round:    11,
		hash:     []byte("D"),
		state:    process.BHReceived,
		hasProof: true,
	}
	require.True(t, bfd.appendHeaderInfo(proofInfo).inserted)

	result := bfd.appendHeaderInfo(&headerInfo{
		epoch:    1,
		nonce:    11,
		round:    11,
		hash:     []byte("D"),
		prevHash: []byte("A"),
		state:    process.BHReceived,
		hasProof: true,
	})

	require.False(t, result.inserted)
	require.True(t, result.enriched)
	require.Len(t, bfd.headers[11], 1)
	require.Equal(t, []byte("A"), bfd.headers[11][0].prevHash)
}

func TestBaseForkDetector_SetProbableHighestNonceDoesNotRegressShardFinal(t *testing.T) {
	t.Parallel()

	bfd := newBranchAwareForkDetector(0, 10, []byte("A"))
	bfd.setProbableHighestNonce(9)
	require.Equal(t, uint64(10), bfd.probableHighestNonce())

	metaBFD := newBranchAwareForkDetector(core.MetachainShardId, 10, []byte("A"))
	metaBFD.setProbableHighestNonce(9)
	require.Equal(t, uint64(9), metaBFD.probableHighestNonce())
}
