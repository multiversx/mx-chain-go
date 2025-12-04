package v2

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/ntp"
)

const numRequiredMissedHeadersToForceResync = 4

type headerTracker struct {
	outOfRangeRounds *RoundRingBuffer
	deSyncedRounds   *RoundRingBuffer
	syncer           ntp.SyncTimer
	proofsPool       consensus.EquivalentProofsPool
}

func newHeaderTracker(proofsPool consensus.EquivalentProofsPool, syncer ntp.SyncTimer) (*headerTracker, error) {
	ht := &headerTracker{
		proofsPool:       proofsPool,
		outOfRangeRounds: NewRoundRingBuffer(10),
		deSyncedRounds:   NewRoundRingBuffer(10),
		syncer:           syncer,
	}

	proofsPool.RegisterHandler(ht.receivedProof)

	return ht, nil
}

func (ht *headerTracker) addOutOfRangeRound(round uint64) {
	ht.outOfRangeRounds.Add(round)
}

func (ht *headerTracker) receivedProof(headerProof data.HeaderProofHandler) {
	currRound := headerProof.GetHeaderRound()
	if ht.outOfRangeRounds.Contains(currRound) {
		ht.deSyncedRounds.Add(currRound)
		ht.tryResyncIfNeeded()
	}
}

func (ht *headerTracker) tryResyncIfNeeded() {
	if ht.deSyncedRounds.Size() < numRequiredMissedHeadersToForceResync {
		return
	}

	if areRoundsAscendingOrder(ht.deSyncedRounds.Last(numRequiredMissedHeadersToForceResync)) {
		ht.syncer.StartSyncingTime()
	}
}

func areRoundsAscendingOrder(rounds []uint64) bool {
	if len(rounds) == 0 {
		return false
	}

	for i := 1; i < len(rounds); i++ {
		if rounds[i] != rounds[i-1]+1 {
			return false
		}
	}

	return true
}
