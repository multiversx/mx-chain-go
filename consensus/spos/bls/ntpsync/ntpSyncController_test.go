package ntpsync

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

func TestNewNonceSyncController(t *testing.T) {
	t.Parallel()

	t.Run("nil proofs pool", func(t *testing.T) {
		rsc, err := NewNtpSyncController(nil, &facadeMock.SyncTimerMock{}, 0)
		require.Nil(t, rsc)
		require.Equal(t, spos.ErrNilEquivalentProofPool, err)
	})
	t.Run("nil sync timer", func(t *testing.T) {
		rsc, err := NewNtpSyncController(&dataRetriever.ProofsPoolMock{}, nil, 0)
		require.Nil(t, rsc)
		require.Equal(t, spos.ErrNilSyncTimer, err)
	})
	t.Run("should work", func(t *testing.T) {
		rsc, err := NewNtpSyncController(&dataRetriever.ProofsPoolMock{}, &facadeMock.SyncTimerMock{}, 0)
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

		tracker, _ := NewNtpSyncController(proofsPool, syncer, 0)

		wg := sync.WaitGroup{}

		numIterations := 500

		handlers := make([]func(), 0, numIterations)

		for i := 0; i < numIterations; i++ {
			r := rand.Intn(1000) + 50 // to avoid interfering with the other samples

			if i%3 == 0 {
				handlers = append(handlers, func() {
					tracker.AddOutOfRangeNonce(uint64(r), "")
					wg.Done()
				})
			}

			if i%7 == 0 {
				handlers = append(handlers, func() {
					tracker.receivedProof(&block.HeaderProof{HeaderNonce: uint64(r)})
					wg.Done()
				})
			}
		}

		wg.Add(len(handlers))

		for _, handler := range handlers {
			go handler()
		}

		wg.Wait()

		tracker.AddOutOfRangeNonce(1, "")
		tracker.AddOutOfRangeNonce(0, "")
		tracker.AddOutOfRangeNonce(2, "")
		tracker.AddOutOfRangeNonce(3, "")

		// receive nonce ordered proofs for different shard and hash, which are not relevant
		for i := 0; i <= 9; i++ {
			tracker.receivedProof(&block.HeaderProof{HeaderNonce: uint64(i), HeaderShardId: 1})
			tracker.receivedProof(&block.HeaderProof{HeaderNonce: uint64(i), HeaderHash: []byte("h")})
		}

		tracker.receivedProof(&block.HeaderProof{HeaderNonce: 0})
		tracker.receivedProof(&block.HeaderProof{HeaderNonce: 1})
		tracker.receivedProof(&block.HeaderProof{HeaderNonce: 2})
		tracker.receivedProof(&block.HeaderProof{HeaderNonce: 3})

		tracker.AddOutOfRangeNonce(4, "")
		tracker.AddOutOfRangeNonce(5, "")

		tracker.receivedProof(&block.HeaderProof{HeaderNonce: 4})
		require.False(t, wasSyncCalled)

		tracker.AddOutOfRangeNonce(6, "")
		tracker.AddOutOfRangeNonce(7, "")
		tracker.AddOutOfRangeNonce(8, "")
		tracker.AddOutOfRangeNonce(9, "")
		tracker.AddOutOfRangeNonce(10, "")

		tracker.receivedProof(&block.HeaderProof{HeaderNonce: 5})
		tracker.receivedProof(&block.HeaderProof{HeaderNonce: 6})
		tracker.receivedProof(&block.HeaderProof{HeaderNonce: 7})
		tracker.receivedProof(&block.HeaderProof{HeaderNonce: 8})
		require.False(t, wasSyncCalled)

		tracker.receivedProof(&block.HeaderProof{HeaderNonce: 9})
		require.True(t, wasSyncCalled)

		// Nonces 1-10 are out of range with received proofs, should force sync
		tracker.receivedProof(&block.HeaderProof{HeaderNonce: 10})
		require.True(t, wasSyncCalled)

		// Receive proof for the same header nonce, should not force resync again
		wasSyncCalled = false
		tracker.receivedProof(&block.HeaderProof{HeaderNonce: 10})
		require.False(t, wasSyncCalled)

		// Receive proof, but no out of range nonce, should not force resync
		tracker.receivedProof(&block.HeaderProof{HeaderNonce: 11})
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

		tracker, _ := NewNtpSyncController(proofsPool, syncer, 0)

		for i := 0; i <= 7; i++ {
			tracker.AddOutOfRangeNonce(uint64(i), "")
			tracker.receivedProof(&block.HeaderProof{HeaderNonce: uint64(i)})
		}

		tracker.AddOutOfRangeNonce(9, "")
		require.False(t, wasSyncCalled)

		// add for self leader, without receiving proof
		tracker.AddLeaderNonceAsOutOfRange(8, "")
		require.False(t, wasSyncCalled)

		tracker.receivedProof(&block.HeaderProof{HeaderNonce: 9})
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

		tracker, _ := NewNtpSyncController(proofsPool, syncer, 0)

		for i := 0; i <= 9; i++ {
			tracker.AddLeaderNonceAsOutOfRange(uint64(i), "")
		}

		// 10 consecutive leader proposed headers will trigger ntp force sync
		// this a very unlikely scenario

		require.True(t, wasSyncCalled)
	})
}
