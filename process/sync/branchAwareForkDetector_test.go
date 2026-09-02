package sync

import (
	"bytes"
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
		proofsPool: &dataRetrieverMock.ProofsPoolMock{},
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

	bfd.headers[11][0].prevHash = nil
	require.False(t, bfd.canInstantlyFinalize(header, headerHash))

	bfd.headers[11][0].hasProof = false
	bfd.headers[11][0].state = process.BHNotarized
	bfd.headers[11][0].prevHash = []byte("other-parent")
	require.False(t, bfd.canInstantlyFinalize(header, headerHash))

	bfd.headers[11][0].prevHash = parentHash
	bfd.enableRoundsHandler = &testscommon.EnableRoundsHandlerStub{}
	require.True(t, bfd.canInstantlyFinalize(header, headerHash))
}

func TestBaseForkDetector_SupernovaFinalityModeIsV3Only(t *testing.T) {
	t.Parallel()

	parentHash := []byte("P")
	bfd := newBranchAwareForkDetector(0, 10, parentHash)
	legacyHeader := &block.Header{Epoch: 1, Nonce: 11, Round: 12, PrevHash: parentHash}
	v3Header := &block.HeaderV3{Epoch: 1, Nonce: 11, Round: 12, PrevHash: parentHash}

	require.False(t, bfd.isSupernovaForHeader(legacyHeader))
	require.True(t, bfd.canInstantlyFinalize(legacyHeader, []byte("legacy")))
	require.True(t, bfd.isSupernovaForHeader(v3Header))
	require.False(t, bfd.canInstantlyFinalize(v3Header, []byte("V3")))

	bfd.enableRoundsHandler = &testscommon.EnableRoundsHandlerStub{}
	require.False(t, bfd.isSupernovaForHeader(v3Header))
	require.True(t, bfd.canInstantlyFinalize(v3Header, []byte("V3")))
}

func TestShardForkDetector_CompetingSiblingProofStopsFinalityCascade(t *testing.T) {
	t.Parallel()

	parentHash := []byte("P")
	processedHash := []byte("A")
	bfd := newBranchAwareForkDetector(0, 10, parentHash)
	bfd.fork.settledCheckpoint = bfd.fork.finalCheckpoint
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

	bfd.headers[11][1].prevHash = nil
	require.Nil(t, sfd.getCleanProcessedChild(bfd.finalCheckpoint()))

	bfd.headers[11][1].hasProof = false
	bfd.headers[11][1].state = process.BHNotarized
	bfd.headers[11][1].prevHash = []byte("other-parent")
	require.Nil(t, sfd.getCleanProcessedChild(bfd.finalCheckpoint()))

	bfd.headers[11][1].state = process.BHReceived
	require.Same(t, processedChild, sfd.getCleanProcessedChild(bfd.finalCheckpoint()))

	bfd.proofsPool = &dataRetrieverMock.ProofsPoolMock{
		HasProofForDifferentHashCalled: func(_ uint32, _ uint64, _ []byte) bool {
			return true
		},
	}
	require.Same(t, processedChild, sfd.getCleanProcessedChild(bfd.finalCheckpoint()))
	sfd.finalizeCleanProcessedDescendants()
	require.Equal(t, uint64(10), bfd.finalCheckpoint().nonce)

	bfd.enableRoundsHandler = &testscommon.EnableRoundsHandlerStub{}
	sfd.finalizeCleanProcessedDescendants()
	require.Equal(t, uint64(11), bfd.finalCheckpoint().nonce)
	require.Equal(t, uint64(11), bfd.settledCheckpoint().nonce)
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

func TestShardForkDetector_ApplyNotarizedHeaderSelectionPreservesEvidence(t *testing.T) {
	t.Parallel()

	parentHash := []byte("P")
	staleHash := []byte("A")
	selectedHash := []byte("B")
	bfd := newBranchAwareForkDetector(0, 10, parentHash)
	bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: parentHash}
	bfd.fork.checkpoint = []*checkpointInfo{{nonce: 10, round: 10, hash: parentHash}}
	bfd.proofsPool = &dataRetrieverMock.ProofsPoolMock{}
	bfd.headers[11] = []*headerInfo{
		{epoch: 1, nonce: 11, round: 11, hash: staleHash, prevHash: parentHash, state: process.BHNotarized},
		{epoch: 1, nonce: 11, round: 11, hash: staleHash, prevHash: parentHash, state: process.BHReceived, hasProof: true},
		{epoch: 1, nonce: 11, round: 12, hash: selectedHash, prevHash: parentHash, state: process.BHNotarized},
		{epoch: 1, nonce: 11, round: 12, hash: selectedHash, prevHash: parentHash, state: process.BHReceived, hasProof: true},
	}
	bfd.hasAmbiguousNotarization.Store(true)
	sfd := &shardForkDetector{baseForkDetector: bfd}
	bfd.forkDetector = sfd

	require.Equal(t, notarizedHeaderApplied, sfd.applyNotarizedHeaderSelection(11, selectedHash))
	require.False(t, sfd.hasUnresolvedNotarizedAmbiguity())

	numStaleNotarized := 0
	numStaleEvidence := 0
	numSelectedNotarized := 0
	for _, info := range sfd.headers[11] {
		if bytes.Equal(info.hash, staleHash) && info.state == process.BHNotarized {
			numStaleNotarized++
		}
		if bytes.Equal(info.hash, staleHash) && info.state == process.BHReceived {
			numStaleEvidence++
		}
		if bytes.Equal(info.hash, selectedHash) && info.state == process.BHNotarized {
			numSelectedNotarized++
		}
	}
	require.Zero(t, numStaleNotarized)
	require.Equal(t, 1, numStaleEvidence)
	require.Equal(t, 1, numSelectedNotarized)
	settledNonce, _ := sfd.GetHighestSettledBlockInfo()
	require.Equal(t, uint64(10), settledNonce)
}

func TestShardForkDetector_ApplyNotarizedHeaderSelectionAcceptsProcessedCandidateBeforeNotarizedCallback(t *testing.T) {
	t.Parallel()

	parentHash := []byte("P")
	processedHash := []byte("A")
	conflictingHash := []byte("B")
	bfd := newBranchAwareForkDetector(0, 10, parentHash)
	bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: parentHash}
	bfd.headers[11] = []*headerInfo{
		{
			epoch: 1, nonce: 11, round: 11, hash: processedHash, prevHash: parentHash,
			state: process.BHProcessed, hasProof: true,
		},
		{
			epoch: 1, nonce: 11, round: 12, hash: conflictingHash, prevHash: parentHash,
			state: process.BHReceived, hasProof: true,
		},
		{
			epoch: 1, nonce: 11, round: 12, hash: conflictingHash, prevHash: parentHash,
			state: process.BHNotarized, hasProof: true,
		},
	}
	bfd.hasAmbiguousNotarization.Store(true)
	sfd := &shardForkDetector{baseForkDetector: bfd}
	bfd.forkDetector = sfd

	require.Equal(t, notarizedHeaderApplied, sfd.applyNotarizedHeaderSelection(11, processedHash))
	require.False(t, sfd.hasUnresolvedNotarizedAmbiguity())
	require.True(t, containsHeaderInfo(sfd.headers[11], process.BHProcessed, processedHash))
	require.True(t, containsHeaderInfo(sfd.headers[11], process.BHNotarized, processedHash))
	require.True(t, containsHeaderInfo(sfd.headers[11], process.BHReceived, conflictingHash))
	require.False(t, containsHeaderInfo(sfd.headers[11], process.BHNotarized, conflictingHash))
	require.Equal(t, uint64(11), sfd.GetHighestFinalBlockNonce())
	settledNonce, settledHash := sfd.GetHighestSettledBlockInfo()
	require.Equal(t, uint64(11), settledNonce)
	require.Equal(t, processedHash, settledHash)
}

func TestShardForkDetector_ApplyNotarizedHeaderSelectionDoesNotSettleUnprovenProcessedCandidate(t *testing.T) {
	t.Parallel()

	parentHash := []byte("P")
	processedHash := []byte("A")
	conflictingHash := []byte("B")
	bfd := newBranchAwareForkDetector(0, 10, parentHash)
	bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: parentHash}
	bfd.headers[11] = []*headerInfo{
		{
			epoch: 1, nonce: 11, round: 11, hash: processedHash, prevHash: parentHash,
			state: process.BHProcessed,
		},
		{
			epoch: 1, nonce: 11, round: 12, hash: conflictingHash, prevHash: parentHash,
			state: process.BHNotarized, hasProof: true,
		},
	}
	bfd.hasAmbiguousNotarization.Store(true)
	sfd := &shardForkDetector{baseForkDetector: bfd}
	bfd.forkDetector = sfd

	require.Equal(t, notarizedHeaderApplied, sfd.applyNotarizedHeaderSelection(11, processedHash))
	require.True(t, containsHeaderInfo(sfd.headers[11], process.BHNotarized, processedHash))
	require.Equal(t, uint64(10), sfd.GetHighestFinalBlockNonce())
	settledNonce, settledHash := sfd.GetHighestSettledBlockInfo()
	require.Equal(t, uint64(10), settledNonce)
	require.Equal(t, parentHash, settledHash)
}

func TestShardForkDetector_DoesNotReplaceAuthorityAtSettledNonce(t *testing.T) {
	t.Parallel()

	bfd := newBranchAwareForkDetector(0, 10, []byte("A"))
	bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: []byte("A")}
	bfd.headers[10] = []*headerInfo{
		{epoch: 1, nonce: 10, round: 10, hash: []byte("A"), state: process.BHNotarized},
		{epoch: 1, nonce: 10, round: 11, hash: []byte("B"), state: process.BHNotarized},
	}
	bfd.hasAmbiguousNotarization.Store(true)
	sfd := &shardForkDetector{baseForkDetector: bfd}
	bfd.forkDetector = sfd

	require.Equal(t, notarizedHeaderUnresolved, sfd.applyNotarizedHeaderSelection(10, []byte("B")))
	require.Len(t, sfd.headers[10], 2)
	require.True(t, sfd.hasUnresolvedNotarizedAmbiguity())
	settledNonce, settledHash := sfd.GetHighestSettledBlockInfo()
	require.Equal(t, uint64(10), settledNonce)
	require.Equal(t, []byte("A"), settledHash)
}

func TestShardForkDetector_HigherSettlementResolvesLowerAuthorityOnProcessedAncestry(t *testing.T) {
	t.Parallel()

	parentHash := []byte("P")
	localHash := []byte("A")
	conflictingHash := []byte("B")
	descendantHash := []byte("C")
	bfd := newBranchAwareForkDetector(0, 10, parentHash)
	bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: parentHash}
	bfd.headers[11] = []*headerInfo{
		{epoch: 1, nonce: 11, round: 11, hash: localHash, prevHash: parentHash, state: process.BHProcessed, hasProof: true},
		{epoch: 1, nonce: 11, round: 12, hash: conflictingHash, prevHash: parentHash, state: process.BHNotarized, hasProof: true},
	}
	bfd.headers[12] = []*headerInfo{
		{epoch: 1, nonce: 12, round: 13, hash: descendantHash, prevHash: localHash, state: process.BHProcessed, hasProof: true},
		{epoch: 1, nonce: 12, round: 13, hash: descendantHash, prevHash: localHash, state: process.BHNotarized, hasProof: true},
	}
	bfd.hasAmbiguousNotarization.Store(true)
	sfd := &shardForkDetector{baseForkDetector: bfd}

	sfd.computeFinalCheckpoint()

	settledNonce, settledHash := sfd.GetHighestSettledBlockInfo()
	require.Equal(t, uint64(12), settledNonce)
	require.Equal(t, descendantHash, settledHash)
	require.False(t, sfd.hasUnresolvedNotarizedAmbiguity())
	for _, info := range sfd.headers[11] {
		require.False(t, info.state == process.BHNotarized && bytes.Equal(info.hash, conflictingHash))
	}
}

func TestShardForkDetector_HigherSettlementDoesNotResolveOffBranchAuthority(t *testing.T) {
	t.Parallel()

	parentHash := []byte("P")
	localHash := []byte("A")
	conflictingHash := []byte("B")
	descendantHash := []byte("C")
	bfd := newBranchAwareForkDetector(0, 10, parentHash)
	bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: parentHash}
	bfd.headers[11] = []*headerInfo{
		{epoch: 1, nonce: 11, round: 11, hash: localHash, prevHash: parentHash, state: process.BHProcessed, hasProof: true},
		{epoch: 1, nonce: 11, round: 12, hash: conflictingHash, prevHash: parentHash, state: process.BHNotarized, hasProof: true},
	}
	bfd.headers[12] = []*headerInfo{
		{epoch: 1, nonce: 12, round: 13, hash: descendantHash, prevHash: conflictingHash, state: process.BHProcessed, hasProof: true},
		{epoch: 1, nonce: 12, round: 13, hash: descendantHash, prevHash: conflictingHash, state: process.BHNotarized, hasProof: true},
	}
	bfd.hasAmbiguousNotarization.Store(true)
	sfd := &shardForkDetector{baseForkDetector: bfd}

	sfd.computeFinalCheckpoint()

	settledNonce, settledHash := sfd.GetHighestSettledBlockInfo()
	require.Equal(t, uint64(10), settledNonce)
	require.Equal(t, parentHash, settledHash)
	require.True(t, sfd.hasUnresolvedNotarizedAmbiguity())
	require.True(t, containsHeaderInfo(sfd.headers[11], process.BHNotarized, conflictingHash))
}

func containsHeaderInfo(infos []*headerInfo, state process.BlockHeaderState, hash []byte) bool {
	for _, info := range infos {
		if info.state == state && bytes.Equal(info.hash, hash) {
			return true
		}
	}

	return false
}

func TestShardForkDetector_ApplyNotarizedHeaderSelectionDoesNotChangePureLegacyRecords(t *testing.T) {
	t.Parallel()

	bfd := newBranchAwareForkDetector(0, 10, []byte("P"))
	bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: []byte("P")}
	bfd.enableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, _ uint32) bool {
			return flag == common.AndromedaFlag
		},
	}
	bfd.headers[11] = []*headerInfo{
		{epoch: 1, nonce: 11, round: 11, hash: []byte("A"), state: process.BHNotarized},
		{epoch: 1, nonce: 11, round: 12, hash: []byte("B"), state: process.BHNotarized},
	}
	bfd.hasAmbiguousNotarization.Store(true)
	sfd := &shardForkDetector{baseForkDetector: bfd}
	bfd.forkDetector = sfd

	require.Equal(t, notarizedHeaderUnresolved, sfd.applyNotarizedHeaderSelection(11, []byte("B")))
	require.Len(t, sfd.headers[11], 2)
}

func TestShardForkDetector_ApplyNotarizedHeaderSelectionResolvesMixedLegacyV3Records(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		selectedHash []byte
		reverseOrder bool
	}{
		{name: "legacy selected", selectedHash: []byte("legacy")},
		{name: "V3 selected", selectedHash: []byte("V3")},
		{name: "legacy selected reverse order", selectedHash: []byte("legacy"), reverseOrder: true},
		{name: "V3 selected reverse order", selectedHash: []byte("V3"), reverseOrder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			parentHash := []byte("P")
			bfd := newBranchAwareForkDetector(0, 10, parentHash)
			bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: parentHash}
			bfd.enableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
					return flag == common.AndromedaFlag || flag == common.SupernovaFlag && epoch >= 2
				},
			}
			records := []*headerInfo{
				{epoch: 1, nonce: 11, round: 11, hash: []byte("legacy"), prevHash: parentHash, state: process.BHNotarized},
				{epoch: 2, nonce: 11, round: 12, hash: []byte("V3"), prevHash: parentHash, state: process.BHNotarized},
			}
			if testCase.reverseOrder {
				records[0], records[1] = records[1], records[0]
			}
			bfd.headers[11] = records
			bfd.hasAmbiguousNotarization.Store(true)
			sfd := &shardForkDetector{baseForkDetector: bfd}
			bfd.forkDetector = sfd

			require.Equal(t, notarizedHeaderApplied, sfd.applyNotarizedHeaderSelection(11, testCase.selectedHash))
			require.Len(t, sfd.headers[11], 1)
			require.Equal(t, testCase.selectedHash, sfd.headers[11][0].hash)
			require.False(t, sfd.hasUnresolvedNotarizedAmbiguity())
		})
	}
}

func TestShardForkDetector_AmbiguousNotarizationDoesNotChooseForkByOrder(t *testing.T) {
	t.Parallel()

	parentHash := []byte("P")
	bfd := newBranchAwareForkDetector(0, 10, parentHash)
	bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: parentHash}
	bfd.headers[11] = []*headerInfo{
		{epoch: 1, nonce: 11, round: 12, hash: []byte("B"), prevHash: parentHash, state: process.BHNotarized},
		{epoch: 1, nonce: 11, round: 11, hash: []byte("A"), prevHash: parentHash, state: process.BHNotarized},
		{epoch: 1, nonce: 11, round: 12, hash: []byte("B"), prevHash: parentHash, state: process.BHProcessed},
	}
	bfd.hasAmbiguousNotarization.Store(true)
	bfd.roundHandler = &testscommon.RoundHandlerMock{
		IndexCalled: func() int64 {
			return 10
		},
	}
	bfd.processConfigsHandler = &testscommon.ProcessConfigsHandlerStub{
		GetMaxRoundsWithoutCommittedBlockCalled: func(_ uint64) uint32 {
			return 100
		},
	}
	bfd.SetRollBackNonce(math.MaxUint64)

	require.False(t, bfd.CheckFork().IsDetected)
	selection, found := bfd.getLowestAmbiguousNotarizedHeaderSelection()
	require.True(t, found)
	require.Len(t, selection.candidates, 2)
}

func TestShardForkDetector_UniqueV3AuthorityConflictNeedsReconciliation(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		finalNonce uint64
	}{
		{name: "conflict above final", finalNonce: 10},
		{name: "conflict below final", finalNonce: 13},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			parentHash := []byte("P")
			localHash := []byte("A")
			selectedHash := []byte("B")
			selectedChildHash := []byte("E")
			finalHash := []byte("tip")
			if testCase.finalNonce == 10 {
				finalHash = parentHash
			}
			bfd := newBranchAwareForkDetector(0, testCase.finalNonce, finalHash)
			bfd.headers = map[uint64][]*headerInfo{
				10: {{epoch: 1, nonce: 10, round: 10, hash: parentHash, state: process.BHProcessed, hasProof: true}},
				11: {{epoch: 1, nonce: 11, round: 11, hash: localHash, prevHash: parentHash, state: process.BHProcessed, hasProof: true}},
				12: {{epoch: 1, nonce: 12, round: 12, hash: []byte("D"), prevHash: localHash, state: process.BHProcessed, hasProof: true}},
				13: {{epoch: 1, nonce: 13, round: 13, hash: []byte("tip"), prevHash: []byte("D"), state: process.BHProcessed, hasProof: true}},
			}
			bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: parentHash}
			bfd.fork.finalCheckpoint = &checkpointInfo{nonce: testCase.finalNonce, round: testCase.finalNonce, hash: finalHash}
			bfd.fork.checkpoint = []*checkpointInfo{
				{nonce: 10, round: 10, hash: parentHash},
				{nonce: 11, round: 11, hash: localHash},
				{nonce: 12, round: 12, hash: []byte("D")},
				{nonce: 13, round: 13, hash: []byte("tip")},
			}
			sfd := &shardForkDetector{baseForkDetector: bfd}
			bfd.forkDetector = sfd

			require.True(t, bfd.append(&headerInfo{
				epoch: 1, nonce: 11, round: 12, hash: selectedHash, prevHash: parentHash,
				state: process.BHNotarized,
			}))
			require.True(t, bfd.append(&headerInfo{
				epoch: 1, nonce: 12, round: 13, hash: selectedChildHash, prevHash: selectedHash,
				state: process.BHReceived, hasProof: true,
			}))
			require.True(t, sfd.hasUnresolvedNotarizedAmbiguity())

			selection, found := sfd.getLowestAmbiguousNotarizedHeaderSelection()
			require.True(t, found)
			require.Len(t, selection.candidates, 2)
			require.Equal(t, localHash, selection.candidates[0].hash)
			require.Equal(t, selectedHash, selection.candidates[1].hash)
			require.Equal(t, notarizedHeaderNeedsReconciliation, sfd.applyNotarizedHeaderSelection(11, selectedHash))

			require.True(t, sfd.ReconcileFinalCheckpointFromAuthority(11, selectedHash))
			require.Equal(t, uint64(10), sfd.GetHighestFinalBlockNonce())
			require.Equal(t, parentHash, sfd.finalCheckpoint().hash)
			settledNonce, settledHash := sfd.GetHighestSettledBlockInfo()
			require.Equal(t, uint64(10), settledNonce)
			require.Equal(t, parentHash, settledHash)
			require.False(t, sfd.hasUnresolvedNotarizedAmbiguity())
			require.Len(t, sfd.headers[11], 1)
			require.Equal(t, selectedHash, sfd.headers[11][0].hash)
			require.Equal(t, process.BHNotarized, sfd.headers[11][0].state)
			for _, info := range sfd.headers[12] {
				require.True(t, info.hasProof)
				require.Equal(t, process.BHReceived, info.state)
			}
			require.True(t, containsHeaderInfo(sfd.headers[12], process.BHReceived, selectedChildHash))
			require.Len(t, sfd.headers[13], 1)
			require.True(t, sfd.headers[13][0].hasProof)
			require.Equal(t, process.BHReceived, sfd.headers[13][0].state)
			require.Equal(t, uint64(12), sfd.ProbableHighestNonce())
		})
	}
}

func TestShardForkDetector_UniqueLegacyAuthorityConflictKeepsPriorSelection(t *testing.T) {
	t.Parallel()

	bfd := newBranchAwareForkDetector(0, 10, []byte("P"))
	bfd.enableRoundsHandler = &testscommon.EnableRoundsHandlerStub{}
	bfd.headers[11] = []*headerInfo{
		{epoch: 1, nonce: 11, round: 11, hash: []byte("A"), state: process.BHProcessed},
		{epoch: 1, nonce: 11, round: 12, hash: []byte("B"), state: process.BHNotarized},
	}

	selection := bfd.getNotarizedHeaderSelection(11)
	require.Equal(t, []byte("B"), selection.hash)
	require.Empty(t, selection.candidates)
	require.False(t, bfd.hasUnresolvedNotarizedAmbiguity())
}

func TestShardForkDetector_UniqueLegacyAuthorityConflictingWithV3ProcessedNeedsReconciliation(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name           string
		processedFirst bool
	}{
		{name: "processed first", processedFirst: true},
		{name: "authority first"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			parentHash := []byte("P")
			localHash := []byte("V3")
			selectedHash := []byte("legacy")
			bfd := newBranchAwareForkDetector(0, 10, parentHash)
			bfd.enableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
					return flag == common.AndromedaFlag || flag == common.SupernovaFlag && epoch >= 2
				},
			}
			bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: parentHash}

			processed := &headerInfo{
				epoch: 2, nonce: 11, round: 12, hash: localHash, prevHash: parentHash,
				state: process.BHProcessed,
			}
			authority := &headerInfo{
				epoch: 1, nonce: 11, round: 11, hash: selectedHash, prevHash: parentHash,
				state: process.BHNotarized,
			}
			if testCase.processedFirst {
				require.True(t, bfd.append(processed))
				require.True(t, bfd.append(authority))
			} else {
				require.True(t, bfd.append(authority))
				require.True(t, bfd.append(processed))
			}

			selection, found := bfd.getLowestAmbiguousNotarizedHeaderSelection()
			require.True(t, found)
			require.Len(t, selection.candidates, 2)
			require.Equal(t, notarizedHeaderNeedsReconciliation,
				bfd.applyNotarizedHeaderSelection(11, selectedHash))
		})
	}
}

func TestBaseForkDetector_LegacyNotarizationsKeepPriorSelection(t *testing.T) {
	t.Parallel()

	bfd := newBranchAwareForkDetector(0, 10, []byte("P"))
	bfd.enableRoundsHandler = &testscommon.EnableRoundsHandlerStub{}
	bfd.headers[11] = []*headerInfo{
		{epoch: 1, nonce: 11, round: 11, hash: []byte("A"), state: process.BHNotarized},
		{epoch: 1, nonce: 11, round: 12, hash: []byte("B"), state: process.BHNotarized},
	}

	selection := bfd.getNotarizedHeaderSelection(11)
	require.Equal(t, []byte("A"), selection.hash)
	require.False(t, selection.isV3)
	require.Empty(t, selection.candidates)
	require.False(t, bfd.hasUnresolvedNotarizedAmbiguity())
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
