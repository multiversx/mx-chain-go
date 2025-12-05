package roundSync

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/ntp"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("roundSyncController")

const (
	roundBufferSize                       = 10
	numRequiredMissedHeadersToForceResync = 4
)

type roundSyncController struct {
	outOfRangeRounds *roundRingBuffer
	deSyncedRounds   *roundRingBuffer
	syncer           ntp.SyncTimer
	selfShardID      uint32
}

func NewRoundSyncController(proofsPool consensus.EquivalentProofsPool, syncer ntp.SyncTimer, selfShardID uint32) (*roundSyncController, error) {
	if check.IfNil(proofsPool) {
		return nil, spos.ErrNilEquivalentProofPool
	}
	if check.IfNil(syncer) {
		return nil, spos.ErrNilSyncTimer
	}

	rsc := &roundSyncController{
		outOfRangeRounds: newRoundRingBuffer(roundBufferSize),
		deSyncedRounds:   newRoundRingBuffer(roundBufferSize),
		syncer:           syncer,
		selfShardID:      selfShardID,
	}

	proofsPool.RegisterHandler(rsc.receivedProof)

	return rsc, nil
}

func (rsc *roundSyncController) AddOutOfRangeRound(round uint64) {
	rsc.outOfRangeRounds.add(round)
}

func (rsc *roundSyncController) receivedProof(headerProof data.HeaderProofHandler) {
	if headerProof.GetHeaderShardId() != rsc.selfShardID {
		return
	}

	currRound := headerProof.GetHeaderRound()
	// this should probably not happen, but return early if we receive the same proof for this round so we don't trigger resync multiple times
	if rsc.deSyncedRounds.contains(currRound) {
		return
	}

	if rsc.outOfRangeRounds.contains(currRound) {
		rsc.deSyncedRounds.add(currRound)
		rsc.tryResyncIfNeeded()
	}
}

func (rsc *roundSyncController) tryResyncIfNeeded() {
	if rsc.deSyncedRounds.len() < numRequiredMissedHeadersToForceResync {
		return
	}

	lastDeSyncedRounds := rsc.deSyncedRounds.last(numRequiredMissedHeadersToForceResync)
	if areRoundsInAscendingOrder(lastDeSyncedRounds) {
		log.Debug("roundSyncController: force ntp synchronization")
		rsc.syncer.ForceSync()
	}
}

func areRoundsInAscendingOrder(rounds []uint64) bool {
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
