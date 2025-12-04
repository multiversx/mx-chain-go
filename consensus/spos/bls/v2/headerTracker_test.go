package v2

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	facadeMock "github.com/multiversx/mx-chain-go/facade/mock"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/require"
)

func TestHeaderTracker_ShouldForceNTPResync(t *testing.T) {
	t.Parallel()

	proofsPool := &dataRetriever.ProofsPoolMock{}
	wasSyncCalled := false
	syncer := &facadeMock.SyncTimerMock{
		StartSyncingTimeCalled: func() {
			wasSyncCalled = true
		},
	}

	tracker, _ := newHeaderTracker(proofsPool, syncer)

	wg := sync.WaitGroup{}
	wg.Add(240)
	for i := 1; i <= 504; i++ {
		r := rand.Int() % 1000

		if i%3 == 0 {
			go func() {
				tracker.addOutOfRangeRound(uint64(r))
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

	tracker.addOutOfRangeRound(2)
	tracker.addOutOfRangeRound(3)

	tracker.receivedProof(&block.HeaderProof{HeaderRound: 2})
	tracker.receivedProof(&block.HeaderProof{HeaderRound: 3})

	tracker.addOutOfRangeRound(4)
	tracker.addOutOfRangeRound(5)

	tracker.receivedProof(&block.HeaderProof{HeaderRound: 4})
	require.False(t, wasSyncCalled)

	tracker.receivedProof(&block.HeaderProof{HeaderRound: 5})
	require.True(t, wasSyncCalled)
}
