package roundSync

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	facadeMock "github.com/multiversx/mx-chain-go/facade/mock"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/require"
)

func TestNewRoundSyncController(t *testing.T) {
	t.Parallel()

	t.Run("nil proofs pool", func(t *testing.T) {
		rsc, err := NewRoundSyncController(nil, &facadeMock.SyncTimerMock{}, 0)
		require.Nil(t, rsc)
		require.Equal(t, spos.ErrNilEquivalentProofPool, err)
	})
	t.Run("nil sync timer", func(t *testing.T) {
		rsc, err := NewRoundSyncController(&dataRetriever.ProofsPoolMock{}, nil, 0)
		require.Nil(t, rsc)
		require.Equal(t, spos.ErrNilSyncTimer, err)
	})
	t.Run("should work", func(t *testing.T) {
		rsc, err := NewRoundSyncController(&dataRetriever.ProofsPoolMock{}, &facadeMock.SyncTimerMock{}, 0)
		require.NotNil(t, rsc)
		require.False(t, rsc.IsInterfaceNil())
		require.Nil(t, err)
	})
}

func TestHeaderTracker_ShouldForceNTPResync(t *testing.T) {
	t.Parallel()

	t.Run("with multiple random calls", func(t *testing.T) {
		t.Parallel()

		proofsPool := &dataRetriever.ProofsPoolMock{}
		wasSyncCalled := false
		syncer := &facadeMock.SyncTimerMock{
			ForceSyncCalled: func() {
				wasSyncCalled = true
			},
		}

		tracker, _ := NewRoundSyncController(proofsPool, syncer, 0)

		wg := sync.WaitGroup{}

		numIterations := 500

		handlers := make([]func(), 0, numIterations)

		for i := 0; i < numIterations; i++ {
			r := rand.Intn(1000) + 50 // to avoid interfering with the other samples

			if i%3 == 0 {
				handlers = append(handlers, func() {
					tracker.AddOutOfRangeRound(uint64(r), "")
					wg.Done()
				})
			}

			if i%7 == 0 {
				handlers = append(handlers, func() {
					tracker.receivedProof(&block.HeaderProof{HeaderRound: uint64(r)})
					wg.Done()
				})
			}
		}

		wg.Add(len(handlers))

		for _, handler := range handlers {
			go handler()
		}

		wg.Wait()

		tracker.AddOutOfRangeRound(1, "")
		tracker.AddOutOfRangeRound(0, "")
		tracker.AddOutOfRangeRound(2, "")
		tracker.AddOutOfRangeRound(3, "")

		// receive nonce ordered proofs for different shard and hash, which are not relevant
		for i := 0; i <= 9; i++ {
			tracker.receivedProof(&block.HeaderProof{HeaderRound: uint64(i), HeaderShardId: 1})
			tracker.receivedProof(&block.HeaderProof{HeaderRound: uint64(i), HeaderHash: []byte("h")})
		}

		tracker.receivedProof(&block.HeaderProof{HeaderRound: 0})
		tracker.receivedProof(&block.HeaderProof{HeaderRound: 1})
		tracker.receivedProof(&block.HeaderProof{HeaderRound: 2})
		tracker.receivedProof(&block.HeaderProof{HeaderRound: 3})

		tracker.AddOutOfRangeRound(4, "")
		tracker.AddOutOfRangeRound(5, "")

		tracker.receivedProof(&block.HeaderProof{HeaderRound: 4})
		require.False(t, wasSyncCalled)

		tracker.AddOutOfRangeRound(6, "")
		tracker.AddOutOfRangeRound(7, "")
		tracker.AddOutOfRangeRound(8, "")
		tracker.AddOutOfRangeRound(9, "")
		tracker.AddOutOfRangeRound(10, "")

		tracker.receivedProof(&block.HeaderProof{HeaderRound: 5})
		tracker.receivedProof(&block.HeaderProof{HeaderRound: 6})
		tracker.receivedProof(&block.HeaderProof{HeaderRound: 7})
		tracker.receivedProof(&block.HeaderProof{HeaderRound: 8})
		require.False(t, wasSyncCalled)

		tracker.receivedProof(&block.HeaderProof{HeaderRound: 9})
		require.True(t, wasSyncCalled)

		// Rounds 1-10 are out of range with received proofs, should force sync
		tracker.receivedProof(&block.HeaderProof{HeaderRound: 10})
		require.True(t, wasSyncCalled)

		// Receive proof for the same header round, should not force resync again
		wasSyncCalled = false
		tracker.receivedProof(&block.HeaderProof{HeaderRound: 10})
		require.False(t, wasSyncCalled)

		// Receive proof, but no out of range round, should not force resync
		tracker.receivedProof(&block.HeaderProof{HeaderRound: 11})
		require.False(t, wasSyncCalled)
	})

	t.Run("with self leader calls", func(t *testing.T) {
		t.Parallel()

		proofsPool := &dataRetriever.ProofsPoolMock{}
		wasSyncCalled := false
		syncer := &facadeMock.SyncTimerMock{
			ForceSyncCalled: func() {
				wasSyncCalled = true
			},
		}

		tracker, _ := NewRoundSyncController(proofsPool, syncer, 0)

		for i := 0; i <= 7; i++ {
			tracker.AddOutOfRangeRound(uint64(i), "")
			tracker.receivedProof(&block.HeaderProof{HeaderRound: uint64(i)})
		}

		tracker.AddOutOfRangeRound(9, "")
		require.False(t, wasSyncCalled)

		// add for self leader, without receiving proof
		tracker.AddLeaderRoundAsOutOfRange(8, "")
		require.False(t, wasSyncCalled)

		tracker.receivedProof(&block.HeaderProof{HeaderRound: 9})
		require.True(t, wasSyncCalled)
	})

	t.Run("only self leader calls", func(t *testing.T) {
		t.Parallel()

		proofsPool := &dataRetriever.ProofsPoolMock{}
		wasSyncCalled := false
		syncer := &facadeMock.SyncTimerMock{
			ForceSyncCalled: func() {
				wasSyncCalled = true
			},
		}

		tracker, _ := NewRoundSyncController(proofsPool, syncer, 0)

		for i := 0; i <= 9; i++ {
			tracker.AddLeaderRoundAsOutOfRange(uint64(i), "")
		}

		// 10 consecutive leader proposed headers will trigger ntp force sync
		// this a very unlikely scenario

		require.True(t, wasSyncCalled)
	})
}
