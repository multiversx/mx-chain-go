package ntpsync

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/ntp"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("nonceSyncController")

const (
	nonceBufferSize                       = 20
	numRequiredMissedHeadersToForceResync = 10
)

type nonceSyncController struct {
	outOfRangeNonces *ringBuffer
	deSyncedNonces   *ringBuffer
	syncer           ntp.SyncTimer
	selfShardID      uint32
}

// NewNtpSyncController creates a new nonce sync controller which detects nonce desynchronization and triggers a forced NTP resync.
func NewNtpSyncController(proofsPool consensus.EquivalentProofsPool, syncer ntp.SyncTimer, selfShardID uint32) (*nonceSyncController, error) {
	if check.IfNil(proofsPool) {
		return nil, spos.ErrNilEquivalentProofPool
	}
	if check.IfNil(syncer) {
		return nil, spos.ErrNilSyncTimer
	}

	rsc := &nonceSyncController{
		outOfRangeNonces: newNonceRingBuffer(nonceBufferSize),
		deSyncedNonces:   newNonceRingBuffer(nonceBufferSize),
		syncer:           syncer,
		selfShardID:      selfShardID,
	}

	proofsPool.RegisterHandler(rsc.receivedProof)

	return rsc, nil
}

// AddOutOfRangeNonce records a consensus nonce that was outside the expected range.
// These nonces are later correlated with received valid proofs to detect NTP desynchronization
// and potentially trigger a forced time resync.
func (rsc *nonceSyncController) AddOutOfRangeNonce(nonce uint64, hash string) {
	rsc.outOfRangeNonces.add(nonce, hash)
}

// AddLeaderNonceAsOutOfRange will forcely add nonce as out of sync when self node as leader
// this will make sure that self proposed headers will not broke consecutivity while out of sync
func (rsc *nonceSyncController) AddLeaderNonceAsOutOfRange(nonce uint64, hash string) {
	rsc.outOfRangeNonces.add(nonce, hash)
	rsc.deSyncedNonces.add(nonce, hash)
	rsc.tryResyncIfNeeded()
}

func (rsc *nonceSyncController) receivedProof(headerProof data.HeaderProofHandler) {
	if headerProof.GetHeaderShardId() != rsc.selfShardID {
		return
	}

	currNonce := headerProof.GetHeaderNonce()
	hash := string(headerProof.GetHeaderHash())
	// this should probably not happen, but return early if we receive the same proof for this nonce so we don't trigger resync multiple times
	if rsc.deSyncedNonces.contains(currNonce, hash) {
		return
	}

	if rsc.outOfRangeNonces.contains(currNonce, hash) {
		rsc.deSyncedNonces.add(currNonce, hash)
		rsc.tryResyncIfNeeded()
	}
}

func (rsc *nonceSyncController) tryResyncIfNeeded() {
	if rsc.deSyncedNonces.len() < numRequiredMissedHeadersToForceResync {
		return
	}

	lastDeSyncedNonces := rsc.deSyncedNonces.last(numRequiredMissedHeadersToForceResync)

	if areNoncesInAscendingOrder(lastDeSyncedNonces) {
		log.Debug("nonceSyncController: force ntp synchronization")
		rsc.syncer.ForceSync()
	}
}

func areNoncesInAscendingOrder(nonces []uint64) bool {
	if len(nonces) == 0 {
		return false
	}

	for i := 1; i < len(nonces); i++ {
		if nonces[i] != nonces[i-1]+1 {
			return false
		}
	}

	return true
}

// IsInterfaceNil checks if the underlying pointer is nil
func (rsc *nonceSyncController) IsInterfaceNil() bool {
	return rsc == nil
}
