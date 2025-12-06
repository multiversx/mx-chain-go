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

	proofsPool := &dataRetriever.ProofsPoolMock{}
	wasSyncCalled := false
	syncer := &facadeMock.SyncTimerMock{
		ForceSyncCalled: func() {
			wasSyncCalled = true
		},
	}

	tracker, _ := NewRoundSyncController(proofsPool, syncer, 0)

	wg := sync.WaitGroup{}
	wg.Add(240)
	for i := 1; i <= 504; i++ {
		r := rand.Int() % 1000

		if i%3 == 0 {
			go func() {
				tracker.AddOutOfRangeRound(uint64(r))
				wg.Done()
			}()
		}

		if i%7 == 0 {
			go func() {
				tracker.receivedProof(&block.HeaderProof{HeaderRound: uint64(r)})
				wg.Done()
			}()

		}
	}
	wg.Wait()

	tracker.AddOutOfRangeRound(1)
	tracker.AddOutOfRangeRound(0)
	tracker.AddOutOfRangeRound(2)
	tracker.AddOutOfRangeRound(3)

	// receive proofs for different shard in order, which are not relevant
	for i := 0; i <= 3; i++ {
		tracker.receivedProof(&block.HeaderProof{HeaderRound: uint64(i), HeaderShardId: 1})
	}

	tracker.receivedProof(&block.HeaderProof{HeaderRound: 0})
	tracker.receivedProof(&block.HeaderProof{HeaderRound: 2})
	tracker.receivedProof(&block.HeaderProof{HeaderRound: 3})

	tracker.AddOutOfRangeRound(4)
	tracker.AddOutOfRangeRound(5)

	tracker.receivedProof(&block.HeaderProof{HeaderRound: 4})
	require.False(t, wasSyncCalled)

	// Rounds 2,3,4,5 are out of range with received proofs, should force sync
	tracker.receivedProof(&block.HeaderProof{HeaderRound: 5})
	require.True(t, wasSyncCalled)

	// Receive proof for the same header round, should not force resync again
	wasSyncCalled = false
	tracker.receivedProof(&block.HeaderProof{HeaderRound: 5})
	require.False(t, wasSyncCalled)

	// Receive proof, but no out of range round, should not force resync
	tracker.receivedProof(&block.HeaderProof{HeaderRound: 6})
	require.False(t, wasSyncCalled)
}
