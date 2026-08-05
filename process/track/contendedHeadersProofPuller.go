package track

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
)

// proofPullCheckInterval is shorter than any round duration; requests still go out at most once
// per round, gated by the round index
const proofPullCheckInterval = 200 * time.Millisecond

const maxProofPullBackoffRounds = 8

type proofPullKey struct {
	shardID uint32
	nonce   uint64
}

type proofPullState struct {
	nextPullRound int64
	backoffRounds int64
}

type contendedHeaderCandidate struct {
	shardID uint32
	header  data.HeaderHandler
	hash    []byte
}

func (bbt *baseBlockTrack) pullProofsForContendedNoncesLoop(ctx context.Context) {
	ticker := time.NewTicker(proofPullCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bbt.pullProofsForContendedNonces()
		}
	}
}

// pullProofsForContendedNonces requests all proofs at each unresolved contended nonce, once per
// round with backoff; a child does not settle a contended header, so states outlive tip advances
func (bbt *baseBlockTrack) pullProofsForContendedNonces() {
	if !bbt.enableEpochsHandler.IsFlagEnabled(common.SupernovaFlag) {
		return
	}

	currentRound := bbt.roundHandler.Index()

	bbt.mutProofPull.Lock()
	defer bbt.mutProofPull.Unlock()

	if currentRound <= bbt.lastProofPullRound {
		return
	}
	bbt.lastProofPullRound = currentRound

	for key := range bbt.getContendedUnsettledKeys() {
		_, exists := bbt.proofPullStates[key]
		if !exists {
			bbt.proofPullStates[key] = &proofPullState{nextPullRound: currentRound, backoffRounds: 1}
		}
	}

	for key, state := range bbt.proofPullStates {
		if bbt.notarizationPassedNonce(key.shardID, key.nonce) {
			delete(bbt.proofPullStates, key)
			continue
		}

		if currentRound < state.nextPullRound {
			continue
		}

		log.Debug("pulling proofs for contended nonce",
			"shardID", key.shardID,
			"nonce", key.nonce,
			"currentRound", currentRound,
		)
		bbt.requestHandler.RequestEquivalentProofByNonce(key.shardID, key.nonce)

		state.nextPullRound = currentRound + state.backoffRounds
		state.backoffRounds = min(state.backoffRounds*2, maxProofPullBackoffRounds)
	}
}

// notarizationPassedNonce is the pull terminal condition: some header at or past the nonce was
// notarized, so the arbitration this pull was feeding has concluded
func (bbt *baseBlockTrack) notarizationPassedNonce(shardID uint32, nonce uint64) bool {
	if shardID == bbt.shardCoordinator.SelfId() {
		return bbt.selfNotarizer.GetLastNotarizedHeaderNonce(core.MetachainShardId) >= nonce
	}

	return bbt.crossNotarizer.GetLastNotarizedHeaderNonce(shardID) >= nonce
}

// getContendedUnsettledKeys returns a pull key for every tracked contended header not yet settled,
// at any tracked nonce: a later child must not hide a contended ancestor from discovery
func (bbt *baseBlockTrack) getContendedUnsettledKeys() map[proofPullKey]struct{} {
	candidates, orphans := bbt.collectContendedCandidates()

	for _, orphan := range orphans {
		parent, err := bbt.headersPool.GetHeaderByHash(orphan.header.GetPrevHash())
		if err != nil || check.IfNil(parent) {
			continue
		}
		if common.IsContendedHeader(orphan.header, parent) {
			candidates = append(candidates, orphan)
		}
	}

	keys := make(map[proofPullKey]struct{})
	for _, candidate := range candidates {
		key := proofPullKey{shardID: candidate.shardID, nonce: candidate.header.GetNonce()}
		if _, exists := keys[key]; exists {
			continue
		}
		if bbt.IsSettledCrossHeader(candidate.header, candidate.hash) {
			continue
		}
		keys[key] = struct{}{}
	}

	return keys
}

// collectContendedCandidates walks the tracked headers under the read lock; headers whose parent
// is not tracked come back separately for pool resolution outside the lock
func (bbt *baseBlockTrack) collectContendedCandidates() ([]contendedHeaderCandidate, []contendedHeaderCandidate) {
	bbt.mutHeaders.RLock()
	defer bbt.mutHeaders.RUnlock()

	candidates := make([]contendedHeaderCandidate, 0)
	orphans := make([]contendedHeaderCandidate, 0)
	for shardID, headersForShard := range bbt.headers {
		for nonce, headersInfo := range headersForShard {
			if nonce == 0 {
				continue
			}

			for _, headerInfo := range headersInfo {
				if headerInfo == nil || check.IfNil(headerInfo.Header) {
					continue
				}

				candidate := contendedHeaderCandidate{shardID: shardID, header: headerInfo.Header, hash: headerInfo.Hash}
				parent := findTrackedParent(headersForShard, headerInfo.Header)
				if check.IfNil(parent) {
					orphans = append(orphans, candidate)
					continue
				}
				if common.IsContendedHeader(headerInfo.Header, parent) {
					candidates = append(candidates, candidate)
				}
			}
		}
	}

	return candidates, orphans
}

func findTrackedParent(headersForShard map[uint64][]*HeaderInfo, header data.HeaderHandler) data.HeaderHandler {
	for _, headerInfo := range headersForShard[header.GetNonce()-1] {
		if headerInfo != nil && string(headerInfo.Hash) == string(header.GetPrevHash()) {
			return headerInfo.Header
		}
	}

	return nil
}
