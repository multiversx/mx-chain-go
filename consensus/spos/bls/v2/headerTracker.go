package v2

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/ntp"
)

const numRequiredMissedHeadersToForceResync = 4

type headerTracker struct {
	outOfRangeRounds []uint64
	deSyncedRounds   []uint64
	latestRound      uint64
	syncer           ntp.SyncTimer
	proofsPool       consensus.EquivalentProofsPool
}

func newHeaderTracker(proofsPool consensus.EquivalentProofsPool) *headerTracker {
	ht := &headerTracker{
		proofsPool:       proofsPool,
		outOfRangeRounds: make([]uint64, 0),
	}

	go ht.checkIfForceResyncIsNeeded()

	proofsPool.RegisterHandler(ht.receivedProof)

	return ht
}

func (ht *headerTracker) addOutOfRangeRound(round uint64) {
	ht.outOfRangeRounds = append(ht.outOfRangeRounds, round)
}

func (ht *headerTracker) receivedProof(headerProof data.HeaderProofHandler) {
	currRound := headerProof.GetHeaderRound()
	if currRound > ht.latestRound {
		ht.latestRound = currRound
	}

	if contains(ht.outOfRangeRounds, currRound) {
		ht.deSyncedRounds = append(ht.deSyncedRounds, currRound)
	}
}

func contains(rounds []uint64, round uint64) bool {
	for _, r := range rounds {
		if round == r {
			return true
		}
	}
	return false
}

func (ht *headerTracker) checkIfForceResyncIsNeeded() {
	for {
		time.Sleep(time.Second) // mby check every N rounds

		if len(ht.deSyncedRounds) >= numRequiredMissedHeadersToForceResync && areRoundsAscendingOrder(ht.deSyncedRounds[len(ht.deSyncedRounds)-numRequiredMissedHeadersToForceResync:]) {
			ht.syncer.StartSyncingTime()
		}

		ht.cleanOlderData(ht.outOfRangeRounds)
		ht.cleanOlderData(ht.deSyncedRounds)
	}
}

func (ht *headerTracker) cleanOlderData(rounds []uint64) {
	for idx, r := range rounds {
		if ht.latestRound > r+100 {
			rounds = append(rounds[:idx], rounds[idx+1:]...)
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
