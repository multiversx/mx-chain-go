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

type trackedTip struct {
	shardID uint32
	tip     data.HeaderHandler
	tipHash []byte
	parent  data.HeaderHandler
}

func (bbt *baseBlockTrack) pullProofsForContendedTipsLoop(ctx context.Context) {
	ticker := time.NewTicker(proofPullCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bbt.pullProofsForContendedTips()
		}
	}
}

// pullProofsForContendedTips requests all proofs at each unresolved contended nonce, once per
// round with backoff; a child does not settle a contended header, so states outlive tip advances
func (bbt *baseBlockTrack) pullProofsForContendedTips() {
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

	for shardID, tip := range bbt.getContendedUnsettledTips() {
		key := proofPullKey{shardID: shardID, nonce: tip.GetNonce()}
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

// getContendedUnsettledTips returns, per shard, the tracked tip that skipped at least one round
// after its parent and is not settled yet
func (bbt *baseBlockTrack) getContendedUnsettledTips() map[uint32]data.HeaderHandler {
	tips := bbt.getTrackedTips()

	contendedTips := make(map[uint32]data.HeaderHandler)
	for _, tracked := range tips {
		parent := tracked.parent
		if check.IfNil(parent) {
			var err error
			parent, err = bbt.headersPool.GetHeaderByHash(tracked.tip.GetPrevHash())
			if err != nil {
				continue
			}
		}

		if !common.IsContendedHeader(tracked.tip, parent) {
			continue
		}

		if bbt.IsSettledCrossHeader(tracked.tip, tracked.tipHash) {
			continue
		}

		contendedTips[tracked.shardID] = tracked.tip
	}

	return contendedTips
}

// getTrackedTips returns, per shard, the highest-nonce tracked header (lowest round on ties)
// together with its tracked parent, if any
func (bbt *baseBlockTrack) getTrackedTips() []trackedTip {
	bbt.mutHeaders.RLock()
	defer bbt.mutHeaders.RUnlock()

	tips := make([]trackedTip, 0, len(bbt.headers))
	for shardID, headersForShard := range bbt.headers {
		tipInfo := getHighestNonceLowestRoundHeader(headersForShard)
		if tipInfo == nil || tipInfo.Header.GetNonce() == 0 {
			continue
		}

		tip := tipInfo.Header
		var parent data.HeaderHandler
		for _, headerInfo := range headersForShard[tip.GetNonce()-1] {
			if string(headerInfo.Hash) == string(tip.GetPrevHash()) {
				parent = headerInfo.Header
				break
			}
		}

		tips = append(tips, trackedTip{shardID: shardID, tip: tip, tipHash: tipInfo.Hash, parent: parent})
	}

	return tips
}

func getHighestNonceLowestRoundHeader(headersForShard map[uint64][]*HeaderInfo) *HeaderInfo {
	var tip *HeaderInfo
	highestNonce := uint64(0)
	for nonce, headersInfo := range headersForShard {
		if nonce < highestNonce {
			continue
		}

		for _, headerInfo := range headersInfo {
			if check.IfNil(headerInfo.Header) {
				continue
			}

			keepCurrentTip := nonce == highestNonce && tip != nil && headerInfo.Header.GetRound() >= tip.Header.GetRound()
			if keepCurrentTip {
				continue
			}

			tip = headerInfo
			highestNonce = nonce
		}
	}

	return tip
}
