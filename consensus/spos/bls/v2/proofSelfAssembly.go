package v2

import (
	"bytes"
	"math/rand"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"

	commonConsensus "github.com/multiversx/mx-chain-go/common/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

// trySelfAssembleProof aggregates the snapshotted signature shares into a proof for the previous
// round's block; it never creates a second proof: any existing proof at the nonce aborts it
func trySelfAssembleProof(sr *spos.Subround, store signatureEvidenceHandler, ev *roundSignatureEvidence) {
	currentRound := sr.RoundHandler().Index()
	if ev.nextAssemblyRound.Load() > currentRound {
		return // already attempted this round
	}
	if !ev.assemblyRunning.CompareAndSwap(false, true) {
		return
	}
	defer ev.assemblyRunning.Store(false)
	ev.nextAssemblyRound.Store(currentRound + 1)

	if abortOnExistingProof(sr, store, ev) {
		return
	}
	if !shouldNodeSendProofForGroup(sr, ev.consensusGroup) {
		return
	}

	bitmap, agg, err := aggregateAndVerifyEvidence(sr, ev)
	if err != nil {
		if ev.getCount() < ev.threshold {
			// inflated signal was adversarial; the decision tiers demote to the wait tier
			store.DropRetained(ev.nonce)
			log.Debug("trySelfAssembleProof: evidence demoted below threshold after stripping invalid shares",
				"nonce", ev.nonce, "count", ev.getCount(), "threshold", ev.threshold)
			return
		}

		bitmap, agg, err = aggregateAndVerifyEvidence(sr, ev)
		if err != nil {
			log.Warn("trySelfAssembleProof: aggregation failed after stripping invalid shares",
				"nonce", ev.nonce, "error", err.Error())
			return
		}
	}

	proof := &block.HeaderProof{
		PubKeysBitmap:       bitmap,
		AggregatedSignature: agg,
		HeaderHash:          ev.headerHash,
		HeaderEpoch:         ev.epoch,
		HeaderNonce:         ev.nonce,
		HeaderShardId:       ev.shardID,
		HeaderRound:         ev.headerRound,
		IsStartOfEpoch:      ev.isStartOfEpoch,
	}

	// atomic add: never overwrites a proof that arrived at the nonce during aggregation
	added, existing := sr.EquivalentProofsPool().AddProofIfNoneAtNonce(proof)
	if !added {
		if !check.IfNil(existing) && !bytes.Equal(existing.GetHeaderHash(), ev.headerHash) {
			log.Warn("trySelfAssembleProof: competing proof arrived during assembly, aborting broadcast",
				"nonce", ev.nonce, "ownHash", ev.headerHash, "existingHash", existing.GetHeaderHash())
		}
		store.DropRetained(ev.nonce) // nonce settled either way
		return
	}
	store.DropRetained(ev.nonce)

	sender := equivalentProofSenderForGroup(sr, ev.consensusGroup)
	err = sr.BroadcastMessenger().BroadcastEquivalentProof(proof, []byte(sender))
	if err != nil {
		log.Warn("trySelfAssembleProof: failed to broadcast proof", "nonce", ev.nonce, "error", err.Error())
		return
	}

	log.Debug("trySelfAssembleProof: self-assembled proof has been sent",
		"nonce", ev.nonce, "headerHash", ev.headerHash)
}

func abortOnExistingProof(sr *spos.Subround, store signatureEvidenceHandler, ev *roundSignatureEvidence) bool {
	existing, err := sr.EquivalentProofsPool().GetProofByNonce(ev.nonce, ev.shardID)
	if err != nil || check.IfNil(existing) {
		return false
	}

	if !bytes.Equal(existing.GetHeaderHash(), ev.headerHash) {
		log.Warn("trySelfAssembleProof: competing proof already exists at nonce, aborting assembly",
			"nonce", ev.nonce, "ownHash", ev.headerHash, "existingHash", existing.GetHeaderHash())
	}
	store.DropRetained(ev.nonce) // nonce settled either way

	return true
}

// aggregateAndVerifyEvidence aggregates the snapshotted shares and verifies the result; on
// failure it strips the invalid shares so a retry or demotion can follow
func aggregateAndVerifyEvidence(sr *spos.Subround, ev *roundSignatureEvidence) ([]byte, []byte, error) {
	bitmap, shares := ev.getAggregationData()

	agg, err := sr.SigningHandler().AggregateSigsWithKeys(ev.consensusGroup, bitmap, shares, ev.epoch)
	if err == nil {
		err = sr.SigningHandler().VerifyAggregatedSigWithKeys(ev.consensusGroup, bitmap, ev.headerHash, agg, ev.epoch)
	}
	if err == nil {
		return bitmap, agg, nil
	}

	stripInvalidShares(sr, ev, bitmap, shares)

	return nil, nil, err
}

// stripInvalidShares verifies each share in parallel and drops the invalid ones from the evidence
func stripInvalidShares(sr *spos.Subround, ev *roundSignatureEvidence, bitmap []byte, shares [][]byte) {
	invalidIndices := make([]int, 0)
	var mut sync.Mutex

	sem := make(chan struct{}, maxParallelVerifications)
	var wg sync.WaitGroup

	numVerified := 0
	for i := range ev.consensusGroup {
		if !isIndexSetInBitmap(i, bitmap) || shares[i] == nil {
			continue
		}
		numVerified++

		wg.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			err := sr.SigningHandler().VerifySigShareWithKey([]byte(ev.consensusGroup[idx]), shares[idx], ev.headerHash, ev.epoch)
			if err == nil {
				return
			}

			log.Debug("stripInvalidShares: invalid signature share in evidence",
				"nonce", ev.nonce, "index", idx, "error", err.Error())
			mut.Lock()
			invalidIndices = append(invalidIndices, idx)
			mut.Unlock()
		}(i)
	}
	wg.Wait()

	// quorum evidence always holds at least one honest, valid share, so all shares failing
	// points to an infrastructure error, not adversarial shares: keep the evidence intact
	if numVerified > 0 && len(invalidIndices) == numVerified {
		log.Warn("stripInvalidShares: all shares failed verification, keeping evidence for retry",
			"nonce", ev.nonce, "numShares", numVerified)
		return
	}

	ev.dropShares(invalidIndices)
}

func isIndexSetInBitmap(index int, bitmap []byte) bool {
	if index/8 >= len(bitmap) {
		return false
	}

	return bitmap[index/8]&(1<<(uint16(index)%8)) != 0
}

// shouldNodeSendProofForGroup mirrors the shouldSendProof eligibility over the snapshot group
func shouldNodeSendProofForGroup(sr *spos.Subround, group []string) bool {
	for _, pk := range group {
		if pk == sr.SelfPubKey() && commonConsensus.ShouldConsiderSelfKeyInConsensus(sr.NodeRedundancyHandler()) {
			return true
		}
		if sr.IsKeyManagedBySelf([]byte(pk)) {
			return true
		}
	}

	return false
}

// equivalentProofSenderForGroup mirrors getEquivalentProofSender over the snapshot group
func equivalentProofSenderForGroup(sr *spos.Subround, group []string) string {
	managedKeys := make([]string, 0)
	for _, pk := range group {
		if pk == sr.SelfPubKey() {
			return pk // single key mode
		}
		if sr.IsKeyManagedBySelf([]byte(pk)) {
			managedKeys = append(managedKeys, pk)
		}
	}

	if len(managedKeys) == 0 {
		return sr.SelfPubKey() // fallback, should never happen
	}

	return managedKeys[rand.Intn(len(managedKeys))]
}
