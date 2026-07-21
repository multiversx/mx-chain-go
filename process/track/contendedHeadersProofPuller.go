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

type headCandidate struct {
	shardID  uint32
	head     data.HeaderHandler
	headHash []byte
	parent   data.HeaderHandler
}

func (bbt *baseBlockTrack) pullProofsForContendedHeadsLoop(ctx context.Context) {
	ticker := time.NewTicker(proofPullCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bbt.pullProofsForContendedHeads()
		}
	}
}

// pullProofsForContendedHeads actively discovers competing proofs by requesting all proofs at the
// nonce of each contended-unsettled tracked head, at most once per round with per-nonce backoff
func (bbt *baseBlockTrack) pullProofsForContendedHeads() {
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

	contendedHeads := bbt.getContendedUnsettledHeads()

	for shardID := range bbt.proofPullPerShard {
		_, stillContended := contendedHeads[shardID]
		if !stillContended {
			delete(bbt.proofPullPerShard, shardID)
		}
	}

	for shardID, head := range contendedHeads {
		state := bbt.proofPullPerShard[shardID]
		if state == nil || state.nonce != head.GetNonce() {
			state = &proofPullState{nonce: head.GetNonce(), nextPullRound: currentRound, backoffRounds: 1}
			bbt.proofPullPerShard[shardID] = state
		}

		if currentRound < state.nextPullRound {
			continue
		}

		log.Debug("pulling proofs for contended head",
			"shardID", shardID,
			"nonce", head.GetNonce(),
			"headRound", head.GetRound(),
			"currentRound", currentRound,
		)
		bbt.requestHandler.RequestEquivalentProofByNonce(shardID, head.GetNonce())

		state.nextPullRound = currentRound + state.backoffRounds
		state.backoffRounds = min(state.backoffRounds*2, maxProofPullBackoffRounds)
	}
}

// getContendedUnsettledHeads returns, per shard, the tracked head that skipped at least one round
// after its parent and is not settled yet
func (bbt *baseBlockTrack) getContendedUnsettledHeads() map[uint32]data.HeaderHandler {
	candidates := bbt.getTrackedHeadCandidates()

	contendedHeads := make(map[uint32]data.HeaderHandler)
	for _, candidate := range candidates {
		parent := candidate.parent
		if check.IfNil(parent) {
			var err error
			parent, err = bbt.headersPool.GetHeaderByHash(candidate.head.GetPrevHash())
			if err != nil {
				continue
			}
		}

		if !common.IsContendedHeader(candidate.head, parent) {
			continue
		}

		if bbt.IsSettledCrossHeader(candidate.head, candidate.headHash) {
			continue
		}

		contendedHeads[candidate.shardID] = candidate.head
	}

	return contendedHeads
}

// getTrackedHeadCandidates returns, per shard, the highest-nonce tracked header (lowest round on
// ties) together with its tracked parent, if any
func (bbt *baseBlockTrack) getTrackedHeadCandidates() []headCandidate {
	bbt.mutHeaders.RLock()
	defer bbt.mutHeaders.RUnlock()

	candidates := make([]headCandidate, 0, len(bbt.headers))
	for shardID, headersForShard := range bbt.headers {
		headInfo := getHighestNonceLowestRoundHeader(headersForShard)
		if headInfo == nil || headInfo.Header.GetNonce() == 0 {
			continue
		}

		head := headInfo.Header
		var parent data.HeaderHandler
		for _, headerInfo := range headersForShard[head.GetNonce()-1] {
			if string(headerInfo.Hash) == string(head.GetPrevHash()) {
				parent = headerInfo.Header
				break
			}
		}

		candidates = append(candidates, headCandidate{shardID: shardID, head: head, headHash: headInfo.Hash, parent: parent})
	}

	return candidates
}

func getHighestNonceLowestRoundHeader(headersForShard map[uint64][]*HeaderInfo) *HeaderInfo {
	var head *HeaderInfo
	highestNonce := uint64(0)
	for nonce, headersInfo := range headersForShard {
		if nonce < highestNonce {
			continue
		}

		for _, headerInfo := range headersInfo {
			if check.IfNil(headerInfo.Header) {
				continue
			}

			keepCurrentHead := nonce == highestNonce && head != nil && headerInfo.Header.GetRound() >= head.Header.GetRound()
			if keepCurrentHead {
				continue
			}

			head = headerInfo
			highestNonce = nonce
		}
	}

	return head
}
