package roundSync

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/ntp"
)

const (
	roundBufferSize                       = 10
	numRequiredMissedHeadersToForceResync = 4
)

type roundSyncController struct {
	outOfRangeRounds *roundRingBuffer
	deSyncedRounds   *roundRingBuffer
	syncer           ntp.SyncTimer
	proofsPool       consensus.EquivalentProofsPool
}

func NewRoundSyncController(proofsPool consensus.EquivalentProofsPool, syncer ntp.SyncTimer) (*roundSyncController, error) {
	rsc := &roundSyncController{
		proofsPool:       proofsPool,
		outOfRangeRounds: newRoundRingBuffer(roundBufferSize),
		deSyncedRounds:   newRoundRingBuffer(roundBufferSize),
		syncer:           syncer,
	}

	proofsPool.RegisterHandler(rsc.receivedProof)

	return rsc, nil
}

func (rsc *roundSyncController) AddOutOfRangeRound(round uint64) {
	rsc.outOfRangeRounds.add(round)
}

func (rsc *roundSyncController) receivedProof(headerProof data.HeaderProofHandler) {
	currRound := headerProof.GetHeaderRound()
	if rsc.outOfRangeRounds.contains(currRound) {
		rsc.deSyncedRounds.add(currRound)
		rsc.tryResyncIfNeeded()
	}
}

func (rsc *roundSyncController) tryResyncIfNeeded() {
	if rsc.deSyncedRounds.len() < numRequiredMissedHeadersToForceResync {
		return
	}

	if areRoundsAscendingOrder(rsc.deSyncedRounds.last(numRequiredMissedHeadersToForceResync)) {
		rsc.syncer.ForceSync()
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

// IsInterfaceNil checks if the underlying pointer is nil
func (rsc *roundSyncController) IsInterfaceNil() bool {
	return rsc == nil
}
