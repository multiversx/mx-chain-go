package sync

import (
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
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

func TestBaseForkDetector_ComputeProbableHighestNonceUsesFinalV3Branch(t *testing.T) {
	t.Parallel()

	finalHash := []byte("A")

	t.Run("known losing child does not advance probable", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(0, 10, finalHash)
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
		addProvenHeaderInfo(bfd, 11, []byte("C"), []byte("B"), process.BHReceived)
		addProvenHeaderInfo(bfd, 11, []byte("D"), finalHash, process.BHReceived)
		addProvenHeaderInfo(bfd, 12, []byte("F"), []byte("C"), process.BHReceived)
		addProvenHeaderInfo(bfd, 12, []byte("E"), []byte("D"), process.BHReceived)

		require.Equal(t, uint64(12), bfd.computeProbableHighestNonce())
	})

	t.Run("complete losing suffix does not advance probable", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(0, 10, finalHash)
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

func TestBaseForkDetector_ComputeProbableHighestNoncePreservesOtherModes(t *testing.T) {
	t.Parallel()

	finalHash := []byte("A")

	t.Run("metachain remains raw", func(t *testing.T) {
		t.Parallel()

		bfd := newBranchAwareForkDetector(core.MetachainShardId, 10, finalHash)
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
