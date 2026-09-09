package sync

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process/mock"
	testscommonDataRetriever "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
)

type unprovenHeaderTrackerScenario struct {
	boot           *baseBootstrap
	headerInPool   *bool
	poolHash       *[]byte
	hasProof       *bool
	removedHashes  *[][]byte
	removedByFD    *[]uint64
	probableNonce  uint64
	nextBlockNonce uint64
}

func newUnprovenHeaderTrackerScenario() *unprovenHeaderTrackerScenario {
	roundHandler := &mock.RoundHandlerMock{RoundIndex: 10}
	probableNonce := uint64(8)
	currentHeader := &block.Header{ShardID: 1, Nonce: 5, Round: 5, Epoch: 6}
	header := &block.Header{ShardID: 1, Nonce: 6, Round: 6, Epoch: 6}

	headerInPool := false
	poolHash := []byte("hash of nonce 6")
	hasProof := false
	removedHashes := make([][]byte, 0)
	removedByFD := make([]uint64, 0)

	scenario := &unprovenHeaderTrackerScenario{
		headerInPool:   &headerInPool,
		poolHash:       &poolHash,
		hasProof:       &hasProof,
		removedHashes:  &removedHashes,
		removedByFD:    &removedByFD,
		probableNonce:  probableNonce,
		nextBlockNonce: 6,
	}

	boot := newRecoveryBootstrap(roundHandler, currentHeader, &probableNonce, &recoveryRequestHandlerStub{})
	boot.headers = &mock.HeadersCacherStub{
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
			if !headerInPool {
				return nil, nil, errors.New("header not found")
			}
			// a fresh copy each call, as the pool returns its internally stored slice
			hashCopy := append([]byte{}, poolHash...)
			return []data.HeaderHandler{header}, [][]byte{hashCopy}, nil
		},
		RemoveHeaderByHashCalled: func(headerHash []byte) {
			removedHashes = append(removedHashes, headerHash)
		},
	}
	boot.forkDetector = &mock.ForkDetectorMock{
		ProbableHighestNonceCalled: func() uint64 { return probableNonce },
		RemoveHeaderCalled: func(nonce uint64, hash []byte) {
			removedByFD = append(removedByFD, nonce)
		},
	}
	boot.proofs = &testscommonDataRetriever.ProofsPoolMock{
		HasProofCalled: func(_ uint32, _ []byte) bool { return hasProof },
	}
	scenario.boot = boot

	return scenario
}

func TestRemoveBlockingUnprovenNextHeader_FullWindowAfterAbsence(t *testing.T) {
	t.Parallel()

	scenario := newUnprovenHeaderTrackerScenario()

	// the incident shape: many failed iterations while the header is still absent
	for i := 0; i < 15; i++ {
		scenario.boot.removeBlockingUnprovenNextHeader()
	}
	require.Empty(t, *scenario.removedHashes)

	// the header arrives: it must survive present-failures 1 through 9
	*scenario.headerInPool = true
	for i := 0; i < 9; i++ {
		scenario.boot.removeBlockingUnprovenNextHeader()
		require.Empty(t, *scenario.removedHashes, "removed too early, at present-failure %d", i+1)
	}

	// and be removed exactly on present-failure 10
	scenario.boot.removeBlockingUnprovenNextHeader()
	require.Equal(t, 1, len(*scenario.removedHashes))
	require.Equal(t, *scenario.poolHash, (*scenario.removedHashes)[0])
	require.Equal(t, []uint64{scenario.nextBlockNonce}, *scenario.removedByFD)
}

func TestRemoveBlockingUnprovenNextHeader_DifferentHashRestartsCount(t *testing.T) {
	t.Parallel()

	scenario := newUnprovenHeaderTrackerScenario()
	*scenario.headerInPool = true

	for i := 0; i < 5; i++ {
		scenario.boot.removeBlockingUnprovenNextHeader()
	}
	require.Empty(t, *scenario.removedHashes)

	*scenario.poolHash = []byte("another hash of nonce 6")
	for i := 0; i < 9; i++ {
		scenario.boot.removeBlockingUnprovenNextHeader()
		require.Empty(t, *scenario.removedHashes)
	}
	scenario.boot.removeBlockingUnprovenNextHeader()
	require.Equal(t, 1, len(*scenario.removedHashes))
	require.Equal(t, *scenario.poolHash, (*scenario.removedHashes)[0])
}

func TestRemoveBlockingUnprovenNextHeader_AbsenceResetsTracker(t *testing.T) {
	t.Parallel()

	scenario := newUnprovenHeaderTrackerScenario()
	*scenario.headerInPool = true

	for i := 0; i < 5; i++ {
		scenario.boot.removeBlockingUnprovenNextHeader()
	}

	*scenario.headerInPool = false
	scenario.boot.removeBlockingUnprovenNextHeader()

	// same hash returns: it gets a full window again
	*scenario.headerInPool = true
	for i := 0; i < 9; i++ {
		scenario.boot.removeBlockingUnprovenNextHeader()
		require.Empty(t, *scenario.removedHashes)
	}
	scenario.boot.removeBlockingUnprovenNextHeader()
	require.Equal(t, 1, len(*scenario.removedHashes))
}

func TestRemoveBlockingUnprovenNextHeader_ProofClearsTracker(t *testing.T) {
	t.Parallel()

	scenario := newUnprovenHeaderTrackerScenario()
	*scenario.headerInPool = true

	for i := 0; i < 9; i++ {
		scenario.boot.removeBlockingUnprovenNextHeader()
	}

	*scenario.hasProof = true
	scenario.boot.removeBlockingUnprovenNextHeader()
	require.Empty(t, *scenario.removedHashes)
	require.Zero(t, scenario.boot.blockingUnprovenHdrFailures)

	// proof gone again (e.g. different fork header at the same nonce): full window required
	*scenario.hasProof = false
	for i := 0; i < 9; i++ {
		scenario.boot.removeBlockingUnprovenNextHeader()
		require.Empty(t, *scenario.removedHashes)
	}
	scenario.boot.removeBlockingUnprovenNextHeader()
	require.Equal(t, 1, len(*scenario.removedHashes))
}

func TestRemoveBlockingUnprovenNextHeader_TrackedHashIsOwnedCopy(t *testing.T) {
	t.Parallel()

	scenario := newUnprovenHeaderTrackerScenario()
	*scenario.headerInPool = true

	originalHash := []byte("hash of nonce 6")
	firstReturnedSlice := append([]byte{}, originalHash...)
	numCalls := 0
	scenario.boot.headers = &mock.HeadersCacherStub{
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
			numCalls++
			if numCalls == 1 {
				return []data.HeaderHandler{&block.Header{ShardID: 1, Nonce: 6}}, [][]byte{firstReturnedSlice}, nil
			}
			return []data.HeaderHandler{&block.Header{ShardID: 1, Nonce: 6}}, [][]byte{append([]byte{}, originalHash...)}, nil
		},
		RemoveHeaderByHashCalled: func(headerHash []byte) {
			*scenario.removedHashes = append(*scenario.removedHashes, headerHash)
		},
	}

	scenario.boot.removeBlockingUnprovenNextHeader()

	// mutating the slice the tracker saw must not corrupt its stored copy: subsequent calls
	// return the original value and the count must keep progressing to removal
	firstReturnedSlice[0] = 'X'
	for i := 0; i < 8; i++ {
		scenario.boot.removeBlockingUnprovenNextHeader()
	}
	require.Empty(t, *scenario.removedHashes)
	scenario.boot.removeBlockingUnprovenNextHeader()
	require.Equal(t, 1, len(*scenario.removedHashes))
}

func TestRemoveBlockingUnprovenNextHeader_ClearedOnSuccessfulCommitCleanup(t *testing.T) {
	t.Parallel()

	scenario := newUnprovenHeaderTrackerScenario()
	*scenario.headerInPool = true

	for i := 0; i < 9; i++ {
		scenario.boot.removeBlockingUnprovenNextHeader()
	}
	require.Equal(t, uint32(9), scenario.boot.blockingUnprovenHdrFailures)

	scenario.boot.clearBlockingUnprovenHdrTracker()
	require.Zero(t, scenario.boot.blockingUnprovenHdrFailures)

	for i := 0; i < 9; i++ {
		scenario.boot.removeBlockingUnprovenNextHeader()
		require.Empty(t, *scenario.removedHashes)
	}
	scenario.boot.removeBlockingUnprovenNextHeader()
	require.Equal(t, 1, len(*scenario.removedHashes))
}

func TestRemoveBlockingUnprovenNextHeader_DoesNotResetRollbackLimitCounter(t *testing.T) {
	t.Parallel()

	scenario := newUnprovenHeaderTrackerScenario()
	*scenario.headerInPool = true
	scenario.boot.mapNonceSyncedWithErrors = map[uint64]uint32{scenario.nextBlockNonce: 15}

	for i := 0; i < 10; i++ {
		scenario.boot.removeBlockingUnprovenNextHeader()
	}
	require.Equal(t, 1, len(*scenario.removedHashes))
	require.Equal(t, uint32(15), scenario.boot.mapNonceSyncedWithErrors[scenario.nextBlockNonce])
}
