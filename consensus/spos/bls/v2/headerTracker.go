package v2

import (
	"time"

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

func newHeaderTracker(proofsPool consensus.EquivalentProofsPool) *headerTracker {
	ht := &headerTracker{
		proofsPool:       proofsPool,
		outOfRangeRounds: NewRoundRingBuffer(100),
		deSyncedRounds:   NewRoundRingBuffer(100),
	}

	go ht.checkIfForceResyncIsNeeded()

	proofsPool.RegisterHandler(ht.receivedProof)

	return ht
}

func (ht *headerTracker) addOutOfRangeRound(round uint64) {
	ht.outOfRangeRounds.Add(round)
}

func (ht *headerTracker) receivedProof(headerProof data.HeaderProofHandler) {
	currRound := headerProof.GetHeaderRound()
	if ht.outOfRangeRounds.Contains(currRound) {
		ht.deSyncedRounds.Add(currRound)
	}
}

func (ht *headerTracker) checkIfForceResyncIsNeeded() {
	for {
		time.Sleep(time.Second) // mby check every N rounds

		if ht.deSyncedRounds.Size() >= numRequiredMissedHeadersToForceResync && areRoundsAscendingOrder(ht.deSyncedRounds.Last(numRequiredMissedHeadersToForceResync)) {
			ht.syncer.StartSyncingTime()
		}

	}
}

func areRoundsAscendingOrder(rounds []uint64) bool {
	prevRound := rounds[0]
	for i := 1; i < len(rounds); i++ {
		if rounds[i] != prevRound+1 {
			return false
		}
	}

	return true
}
