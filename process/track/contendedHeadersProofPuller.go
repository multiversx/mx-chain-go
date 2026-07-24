package track

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
)

// proofPullCheckInterval is shorter than any round duration; requests still go out at most once
// per round, gated by the round index
const proofPullCheckInterval = 200 * time.Millisecond

const maxProofPullBackoffRounds = 8

type proofPullState struct {
	nonce         uint64
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

// pullProofsForContendedTips actively discovers competing proofs by requesting all proofs at the
// nonce of each contended-unsettled tracked tip, at most once per round with per-nonce backoff
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

	contendedTips := bbt.getContendedUnsettledTips()

	for shardID := range bbt.proofPullPerShard {
		_, stillContended := contendedTips[shardID]
		if !stillContended {
			delete(bbt.proofPullPerShard, shardID)
		}
	}

	for shardID, tip := range contendedTips {
		state := bbt.proofPullPerShard[shardID]
		if state == nil || state.nonce != tip.GetNonce() {
			state = &proofPullState{nonce: tip.GetNonce(), nextPullRound: currentRound, backoffRounds: 1}
			bbt.proofPullPerShard[shardID] = state
		}

		if currentRound < state.nextPullRound {
			continue
		}

		log.Debug("pulling proofs for contended chain tip",
			"shardID", shardID,
			"nonce", tip.GetNonce(),
			"tipRound", tip.GetRound(),
			"currentRound", currentRound,
		)
		bbt.requestHandler.RequestEquivalentProofByNonce(shardID, tip.GetNonce())

		state.nextPullRound = currentRound + state.backoffRounds
		state.backoffRounds = min(state.backoffRounds*2, maxProofPullBackoffRounds)
	}
}

// getContendedUnsettledTips returns, per shard, the tracked tip that skipped at least one round
// after its parent (a competing proof could exist in the skipped rounds) and has no proofed child
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
