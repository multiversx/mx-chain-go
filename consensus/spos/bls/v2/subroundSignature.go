package v2

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	atomicCore "github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"

	commonConsensus "github.com/multiversx/mx-chain-go/common/consensus"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
)

const timeSpentBetweenChecks = time.Millisecond

type subroundSignature struct {
	*spos.Subround
	appStatusHandler     core.AppStatusHandler
	sentSignatureTracker spos.SentSignaturesTracker
	signatureThrottler   core.Throttler
}

// NewSubroundSignature creates a subroundSignature object
func NewSubroundSignature(
	baseSubround *spos.Subround,
	appStatusHandler core.AppStatusHandler,
	sentSignatureTracker spos.SentSignaturesTracker,
	worker spos.WorkerHandler,
	signatureThrottler core.Throttler,
) (*subroundSignature, error) {
	err := checkNewSubroundSignatureParams(
		baseSubround,
	)
	if err != nil {
		return nil, err
	}
	if check.IfNil(appStatusHandler) {
		return nil, spos.ErrNilAppStatusHandler
	}
	if check.IfNil(sentSignatureTracker) {
		return nil, ErrNilSentSignatureTracker
	}
	if check.IfNil(worker) {
		return nil, spos.ErrNilWorker
	}
	if check.IfNil(signatureThrottler) {
		return nil, spos.ErrNilThrottler
	}

	srSignature := subroundSignature{
		Subround:             baseSubround,
		appStatusHandler:     appStatusHandler,
		sentSignatureTracker: sentSignatureTracker,
		signatureThrottler:   signatureThrottler,
	}
	srSignature.Job = srSignature.doSignatureJob
	srSignature.Check = srSignature.doSignatureConsensusCheck
	srSignature.Extend = worker.Extend

	return &srSignature, nil
}

func checkNewSubroundSignatureParams(
	baseSubround *spos.Subround,
) error {
	if baseSubround == nil {
		return spos.ErrNilSubround
	}
	if check.IfNil(baseSubround.ConsensusStateHandler) {
		return spos.ErrNilConsensusState
	}

	err := spos.ValidateConsensusCore(baseSubround.ConsensusCoreHandler)

	return err
}

// doSignatureJob method does the job of the subround Signature
func (sr *subroundSignature) doSignatureJob(ctx context.Context) bool {
	if !sr.CanDoSubroundJob(sr.Current()) {
		return false
	}
	if check.IfNil(sr.GetHeader()) {
		log.Error("doSignatureJob", "error", spos.ErrNilHeader)
		return false
	}

	proofAlreadyReceived := sr.EquivalentProofsPool().HasProof(sr.ShardCoordinator().SelfId(), sr.GetData())
	if proofAlreadyReceived {
		sr.SetStatus(sr.Current(), spos.SsFinished)
		log.Debug("step 2: subround has been finished, proof already received",
			"subround", sr.Name())

		return true
	}

	if sr.HasProofForCompetingBlock() {
		log.Debug("step 2: subround cannot proceed, proof for competing block exists")
		return false
	}

	// Wait once for the entire node if competing block detected
	nonce := sr.GetHeader().GetNonce()
	currentHash := sr.GetData()
	shouldAbort := sr.waitIfCompetingBlockForNode(ctx, nonce, currentHash)
	if shouldAbort {
		return false
	}

	isSelfSingleKeyInConsensusGroup := sr.IsNodeInConsensusGroup(sr.SelfPubKey()) && commonConsensus.ShouldConsiderSelfKeyInConsensus(sr.NodeRedundancyHandler())
	if isSelfSingleKeyInConsensusGroup {
		if !sr.doSignatureJobForSingleKey(ctx) {
			return false
		}
	}

	if !sr.doSignatureJobForManagedKeys(ctx) {
		return false
	}

	sr.SetStatus(sr.Current(), spos.SsFinished)
	log.Debug("step 2: subround has been finished",
		"subround", sr.Name())

	return true
}

func (sr *subroundSignature) createAndSendSignatureMessage(signatureShare []byte, pkBytes []byte) bool {
	cnsMsg := consensus.NewConsensusMessage(
		sr.GetData(),
		signatureShare,
		nil,
		nil,
		pkBytes,
		nil,
		int(bls.MtSignature),
		sr.RoundHandler().Index(),
		sr.ChainID(),
		nil,
		nil,
		nil,
		sr.GetAssociatedPid(pkBytes),
		nil,
	)

	err := sr.BroadcastMessenger().BroadcastConsensusMessage(cnsMsg)
	if err != nil {
		log.Debug("createAndSendSignatureMessage.BroadcastConsensusMessage",
			"error", err.Error(), "pk", pkBytes)
		return false
	}

	//log.Debug("step 2: signature has been sent", "pk", pkBytes)

	return true
}

func (sr *subroundSignature) completeSignatureSubRound(pk string) bool {
	err := sr.SetJobDone(pk, sr.Current(), true)
	if err != nil {
		log.Debug("doSignatureJob.SetSelfJobDone",
			"subround", sr.Name(),
			"error", err.Error(),
			"pk", []byte(pk),
		)
		return false
	}

	return true
}

// doSignatureConsensusCheck method checks if the consensus in the subround Signature is achieved
func (sr *subroundSignature) doSignatureConsensusCheck() bool {
	if sr.GetRoundCanceled() {
		return false
	}

	if sr.IsSubroundFinished(sr.Current()) {
		return true
	}

	if check.IfNil(sr.GetHeader()) {
		return false
	}

	isSelfInConsensusGroup := sr.IsSelfInConsensusGroup()
	if !isSelfInConsensusGroup {
		log.Debug("step 2: subround has been finished",
			"subround", sr.Name())
		sr.SetStatus(sr.Current(), spos.SsFinished)

		return true
	}

	if sr.IsSelfJobDone(sr.Current()) {
		log.Debug("step 2: subround has been finished",
			"subround", sr.Name())
		sr.SetStatus(sr.Current(), spos.SsFinished)
		sr.appStatusHandler.SetStringValue(common.MetricConsensusRoundState, "signed")

		return true
	}

	return false
}

func (sr *subroundSignature) doSignatureJobForManagedKeys(ctx context.Context) bool {
	numMultiKeysSignaturesSent := int32(0)
	sentSigForAllKeys := atomicCore.Flag{}
	sentSigForAllKeys.SetValue(true)

	wg := sync.WaitGroup{}

	for idx, pk := range sr.ConsensusGroup() {
		pkBytes := []byte(pk)
		if !sr.IsKeyManagedBySelf(pkBytes) {
			continue
		}

		if sr.IsJobDone(pk, sr.Current()) {
			continue
		}

		err := sr.checkGoRoutinesThrottler(ctx)
		if err != nil {
			return false
		}
		sr.signatureThrottler.StartProcessing()
		wg.Add(1)

		go func(ctx context.Context, idx int, pk string) {
			defer sr.signatureThrottler.EndProcessing()

			signatureSent := sr.sendSignatureForManagedKey(ctx, idx, pk)
			if signatureSent {
				atomic.AddInt32(&numMultiKeysSignaturesSent, 1)
			} else {
				sentSigForAllKeys.SetValue(false)
			}
			wg.Done()
		}(ctx, idx, pk)
	}

	wg.Wait()

	if numMultiKeysSignaturesSent > 0 {
		log.Debug("step 2: multi keys signatures have been sent", "num", numMultiKeysSignaturesSent)
	}

	return sentSigForAllKeys.IsSet()
}

func (sr *subroundSignature) sendSignatureForManagedKey(_ context.Context, idx int, pk string) bool {
	pkBytes := []byte(pk)
	nonce := sr.GetHeader().GetNonce()
	currentHash := sr.GetData()

	signatureShare, err := sr.SigningHandler().CreateSignatureShareForPublicKey(
		currentHash,
		uint16(idx),
		sr.GetHeader().GetEpoch(),
		pkBytes,
	)
	if err != nil {
		log.Debug("sendSignatureForManagedKey.CreateSignatureShareForPublicKey", "error", err.Error())
		return false
	}

	// Record the signed nonce before broadcast so competing block detection works
	// even if the broadcast itself fails
	sr.sentSignatureTracker.RecordSignedNonce(pkBytes, nonce, currentHash)

	// with the equivalent messages feature on, signatures from all managed keys must be broadcast, as the aggregation is done by any participant
	ok := sr.createAndSendSignatureMessage(signatureShare, pkBytes)
	if !ok {
		return false
	}
	sr.sentSignatureTracker.SignatureSent(pkBytes)

	return sr.completeSignatureSubRound(pk)
}

func (sr *subroundSignature) checkGoRoutinesThrottler(ctx context.Context) error {
	for {
		if sr.signatureThrottler.CanProcess() {
			break
		}
		select {
		case <-time.After(timeSpentBetweenChecks):
			continue
		case <-ctx.Done():
			return fmt.Errorf("%w while checking the throttler", spos.ErrTimeIsOut)
		}
	}
	return nil
}

func (sr *subroundSignature) doSignatureJobForSingleKey(_ context.Context) bool {
	pkBytes := []byte(sr.SelfPubKey())
	nonce := sr.GetHeader().GetNonce()
	currentHash := sr.GetData()

	selfIndex, err := sr.SelfConsensusGroupIndex()
	if err != nil {
		log.Debug("doSignatureJobForSingleKey.SelfConsensusGroupIndex: not in consensus group")
		return false
	}

	signatureShare, err := sr.SigningHandler().CreateSignatureShareForPublicKey(
		currentHash,
		uint16(selfIndex),
		sr.GetHeader().GetEpoch(),
		pkBytes,
	)
	if err != nil {
		log.Debug("doSignatureJobForSingleKey.CreateSignatureShareForPublicKey", "error", err.Error())
		return false
	}

	// Record the signed nonce before broadcast so competing block detection works
	// even if the broadcast itself fails
	sr.sentSignatureTracker.RecordSignedNonce(pkBytes, nonce, currentHash)

	// leader also sends his signature here
	ok := sr.createAndSendSignatureMessage(signatureShare, pkBytes)
	if !ok {
		return false
	}

	return sr.completeSignatureSubRound(sr.SelfPubKey())
}

// waitIfCompetingBlockForNode checks if any key managed by this node previously signed a different
// hash for the given nonce. If found, waits once for the entire node instead of per-key.
func (sr *subroundSignature) waitIfCompetingBlockForNode(ctx context.Context, nonce uint64, currentHash []byte) bool {
	// Check self key first
	selfPk := []byte(sr.SelfPubKey())
	previousHash, exists := sr.sentSignatureTracker.GetSignedHash(selfPk, nonce)
	if exists && !bytes.Equal(previousHash, currentHash) {
		return sr.waitIfCompetingBlock(ctx, selfPk, nonce, currentHash)
	}

	// Check managed keys
	for _, pk := range sr.ConsensusGroup() {
		pkBytes := []byte(pk)
		if !sr.IsKeyManagedBySelf(pkBytes) {
			continue
		}
		previousHash, exists = sr.sentSignatureTracker.GetSignedHash(pkBytes, nonce)
		if exists && !bytes.Equal(previousHash, currentHash) {
			return sr.waitIfCompetingBlock(ctx, pkBytes, nonce, currentHash)
		}
	}

	return false
}

// waitIfCompetingBlock waits if this node already signed a different block for the same nonce.
// The delay is measured from round start. Returns true if signing should be aborted.
func (sr *subroundSignature) waitIfCompetingBlock(ctx context.Context, pkBytes []byte, nonce uint64, currentHash []byte) bool {
	previousHash, exists := sr.sentSignatureTracker.GetSignedHash(pkBytes, nonce)
	if !exists {
		return false
	}

	if bytes.Equal(previousHash, currentHash) {
		return false
	}

	// Delay is measured from round start, not from when this function is called
	roundStart := sr.GetRoundTimeStamp()
	targetTime := time.Duration(float64(sr.RoundHandler().TimeDuration()) * competingBlockSignDelay)
	delay := sr.RoundHandler().RemainingTime(roundStart, targetTime)
	if delay <= 0 {
		log.Debug("waitIfCompetingBlock: already past competing block delay deadline, proceeding to sign")
		return false
	}

	// Cap the delay so signing still happens within the signature subround window.
	sigEndDuration := time.Duration(sr.EndTime())
	remaining := sr.RoundHandler().RemainingTime(roundStart, sigEndDuration)
	safetyMargin := 10 * time.Millisecond
	maxDelay := remaining - safetyMargin
	if maxDelay <= 0 {
		log.Debug("waitIfCompetingBlock: no time remaining in signature subround, proceeding to sign")
		return false
	}
	if delay > maxDelay {
		delay = maxDelay
	}

	log.Debug("waitIfCompetingBlock: competing block detected, delaying before signing",
		"nonce", nonce,
		"previousHash", previousHash,
		"currentHash", currentHash,
		"delay", delay)

	shardID := sr.ShardCoordinator().SelfId()
	deadline := time.After(delay)
	ticker := time.NewTicker(timeBetweenSignaturesChecks)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug("waitIfCompetingBlock: context cancelled, aborting")
			return true
		case <-ticker.C:
			if sr.EquivalentProofsPool().HasProof(shardID, previousHash) {
				log.Debug("waitIfCompetingBlock: proof arrived for previous block, aborting signing",
					"nonce", nonce,
					"previousHash", previousHash)
				return true
			}
			if sr.HasProofForCompetingBlock() {
				log.Debug("waitIfCompetingBlock: competing block proof detected, aborting signing")
				return true
			}
		case <-deadline:
			log.Debug("waitIfCompetingBlock: delay expired with no proof for previous block, proceeding to sign",
				"nonce", nonce)
			return false
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (sr *subroundSignature) IsInterfaceNil() bool {
	return sr == nil
}
