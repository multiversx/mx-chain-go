package v2

import (
	"bytes"
	"context"
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
	signatureEvidence    signatureEvidenceHandler
}

// NewSubroundSignature creates a subroundSignature object
func NewSubroundSignature(
	baseSubround *spos.Subround,
	appStatusHandler core.AppStatusHandler,
	sentSignatureTracker spos.SentSignaturesTracker,
	worker spos.WorkerHandler,
	signatureThrottler core.Throttler,
	signatureEvidence signatureEvidenceHandler,
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
	if check.IfNil(signatureEvidence) {
		return nil, ErrNilSignatureEvidence
	}

	srSignature := subroundSignature{
		Subround:             baseSubround,
		appStatusHandler:     appStatusHandler,
		sentSignatureTracker: sentSignatureTracker,
		signatureThrottler:   signatureThrottler,
		signatureEvidence:    signatureEvidence,
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

	if !sr.isTimeLeft() {
		log.Debug("step 2: timeout while handling signature job")
		return false
	}

	nonce := sr.GetHeader().GetNonce()
	currentHash := sr.GetData()

	if sr.shouldAbortOnSignatureEvidence(ctx, nonce, currentHash) {
		return false
	}

	pkBytes := sr.getPkForCompetingBlock(nonce, currentHash)
	hasCompetingBlockForPk := len(pkBytes) != 0
	if hasCompetingBlockForPk {
		shouldAbort := sr.waitIfCompetingBlock(ctx, pkBytes, nonce, currentHash)
		if shouldAbort {
			return false
		}
	}

	// Wait once for the entire node if competing block detected
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

	log.Debug("step 2: signature has been sent", "pk", pkBytes)

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

func (sr *subroundSignature) waitForSignatures(
	timeLeft time.Duration,
) bool {
	if timeLeft <= 0 {
		sr.SignaturesCtxCancel()
		return false
	}

	timer := time.NewTimer(timeLeft)
	defer timer.Stop()

	select {
	case <-sr.SignaturesDone():
		sr.SignaturesCtxCancel()
		return true
	case <-timer.C:
		sr.SignaturesCtxCancel()
		log.Debug("timeout while waiting for signatures to be created")
		return false
	}
}

func (sr *subroundSignature) isTimeLeft() bool {
	timeLeft := sr.RoundHandler().RemainingTime(sr.RoundHandler().TimeStamp(), time.Duration(sr.EndTime()))
	return timeLeft > 0
}

func (sr *subroundSignature) doSignatureJobForManagedKeys(ctx context.Context) bool {
	// wait for optimistic signatures creation to finish
	timeLeft := sr.RoundHandler().RemainingTime(sr.RoundHandler().TimeStamp(), time.Duration(sr.EndTime()))

	sigCtx, cancel := context.WithTimeout(ctx, timeLeft)
	defer cancel()

	signaturesReady := sr.waitForSignatures(timeLeft)
	if !signaturesReady {
		log.Debug("doSignatureJobForManagedKeys: timeout while waiting for signatures to be created")
		return false
	}

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

		select {
		case <-sigCtx.Done():
			log.Debug("doSignatureJobForManagedKeys: timeout while sending signatures")
			wg.Wait()
			return false
		default:
		}

		err := checkGoRoutinesThrottler(sigCtx, sr.signatureThrottler)
		if err != nil {
			log.Debug("doSignatureJobForManagedKeys.checkGoRoutinesThrottler", "err", err)
			wg.Wait()
			return false
		}
		sr.signatureThrottler.StartProcessing()
		wg.Add(1)

		go func(sigCtx context.Context, idx int, pk string) {
			defer wg.Done()
			defer sr.signatureThrottler.EndProcessing()

			signatureSent := sr.sendSignatureForManagedKey(sigCtx, idx, pk)
			if signatureSent {
				atomic.AddInt32(&numMultiKeysSignaturesSent, 1)
			} else {
				sentSigForAllKeys.SetValue(false)
			}
		}(sigCtx, idx, pk)
	}

	wg.Wait()

	if numMultiKeysSignaturesSent > 0 {
		log.Debug("step 2: multi keys signatures have been sent", "num", numMultiKeysSignaturesSent)
	}

	return sentSigForAllKeys.IsSet()
}

func (sr *subroundSignature) sendSignatureForManagedKey(ctx context.Context, idx int, pk string) bool {
	pkBytes := []byte(pk)
	nonce := sr.GetHeader().GetNonce()
	currentHash := sr.GetData()

	select {
	case <-ctx.Done():
		log.Debug("sendSignatureForManagedKey: timeout while sending signature", "idx", idx, "pk", pk)
		return false
	default:
	}

	if !sr.reserveSignatureSlot(pkBytes, nonce, currentHash) {
		return false
	}

	signatureShare, err := sr.SigningHandler().SignatureShare(uint16(idx))
	if err != nil {
		log.Debug("sendSignatureForManagedKey.SignatureShare", "error", err.Error())
		return false
	}

	// Record the signed nonce before broadcast so competing block detection works
	// even if the broadcast itself fails
	sr.sentSignatureTracker.RecordSignedNonce(pkBytes, nonce, currentHash, sr.RoundHandler().Index())

	// with the equivalent messages feature on, signatures from all managed keys must be broadcast, as the aggregation is done by any participant
	ok := sr.createAndSendSignatureMessage(signatureShare, pkBytes)
	if !ok {
		return false
	}
	sr.sentSignatureTracker.SignatureSent(pkBytes)

	return sr.completeSignatureSubRound(pk)
}

func (sr *subroundSignature) doSignatureJobForSingleKey(ctx context.Context) bool {
	pkBytes := []byte(sr.SelfPubKey())
	nonce := sr.GetHeader().GetNonce()
	currentHash := sr.GetData()

	selfIndex, err := sr.SelfConsensusGroupIndex()
	if err != nil {
		log.Debug("doSignatureJobForSingleKey.SelfConsensusGroupIndex: not in consensus group")
		return false
	}

	if !sr.reserveSignatureSlot(pkBytes, nonce, currentHash) {
		return false
	}

	signatureShare, err := sr.SigningHandler().CreateSignatureShareForPublicKey(
		ctx,
		currentHash,
		uint16(selfIndex),
		sr.GetHeader().GetEpoch(),
		pkBytes,
	)
	if err != nil {
		log.Debug("doSignatureJobForSingleKey.CreateSignatureShareForPublicKey", "error", err.Error())
		return false
	}

	if !sr.isTimeLeft() {
		log.Debug("doSignatureJobForSingleKey: timeout while handling single key signature")
		return false
	}

	// Record the signed nonce before broadcast so competing block detection works
	// even if the broadcast itself fails
	sr.sentSignatureTracker.RecordSignedNonce(pkBytes, nonce, currentHash, sr.RoundHandler().Index())

	// leader also sends his signature here
	ok := sr.createAndSendSignatureMessage(signatureShare, pkBytes)
	if !ok {
		return false
	}

	return sr.completeSignatureSubRound(sr.SelfPubKey())
}

// reserveSignatureSlot hard-refuses a second signature for a different hash in the same round
func (sr *subroundSignature) reserveSignatureSlot(pkBytes []byte, nonce uint64, currentHash []byte) bool {
	roundIndex := sr.RoundHandler().Index()
	if sr.sentSignatureTracker.ReserveSignatureInRound(pkBytes, roundIndex, currentHash) {
		return true
	}

	log.Warn("refusing second signature for a different hash in the same round",
		"pk", pkBytes,
		"nonce", nonce,
		"round", roundIndex,
		"hash", currentHash)

	return false
}

func (sr *subroundSignature) getPkForCompetingBlock(nonce uint64, currentHash []byte) []byte {
	currentRound := sr.RoundHandler().Index()

	// check self key
	selfPk := []byte(sr.SelfPubKey())
	previousHash, round, exists := sr.sentSignatureTracker.GetSignedNonceInfo(selfPk, nonce)
	if exists && !bytes.Equal(previousHash, currentHash) && round >= currentRound-1 {
		return selfPk
	}

	// check managed keys
	for _, pk := range sr.ConsensusGroup() {
		pkBytes := []byte(pk)
		if !sr.IsKeyManagedBySelf(pkBytes) {
			continue
		}

		previousHash, round, exists := sr.sentSignatureTracker.GetSignedNonceInfo(pkBytes, nonce)
		if exists && !bytes.Equal(previousHash, currentHash) && round >= currentRound-1 {
			return pkBytes
		}
	}

	return nil
}

// shouldAbortOnSignatureEvidence refuses to sign a competing block on quorum-level evidence for
// the previous round's block (triggering proof self-assembly) and waits on significant evidence
func (sr *subroundSignature) shouldAbortOnSignatureEvidence(ctx context.Context, nonce uint64, currentHash []byte) bool {
	ev, ok := sr.signatureEvidence.GetPreviousRoundEvidence(nonce, sr.RoundHandler().Index())
	if !ok {
		ev, ok = sr.signatureEvidence.GetRetainedQuorumEvidence(nonce)
	}
	if !ok || bytes.Equal(ev.headerHash, currentHash) {
		return false
	}

	count := ev.getCount()
	if count >= ev.threshold {
		go trySelfAssembleProof(sr.Subround, sr.signatureEvidence, ev)
		log.Debug("refusing to sign competing block: previous round reached quorum",
			"nonce", nonce,
			"previousHash", ev.headerHash,
			"observedShares", count)
		return true
	}

	if float64(count) >= significantEvidenceFraction*float64(ev.threshold) {
		return sr.waitForPreviousBlockProof(ctx, ev.headerHash, nonce)
	}

	// few shares observed: the previous block lacked support, proceed
	return false
}

// waitIfCompetingBlock waits if this node already signed a different block for the same nonce
// in the current or previous round. The delay is measured from round start.
// Returns true if signing should be aborted.
func (sr *subroundSignature) waitIfCompetingBlock(ctx context.Context, pkBytes []byte, nonce uint64, currentHash []byte) bool {
	previousHash, round, exists := sr.sentSignatureTracker.GetSignedNonceInfo(pkBytes, nonce)
	if !exists {
		return false
	}

	if bytes.Equal(previousHash, currentHash) {
		return false
	}

	currentRound := sr.RoundHandler().Index()
	if round < currentRound-1 {
		log.Debug("waitIfCompetingBlock: signed hash is from an older round, signing immediately",
			"nonce", nonce,
			"signedRound", round,
			"currentRound", currentRound)
		return false
	}

	return sr.waitForPreviousBlockProof(ctx, previousHash, nonce)
}

// waitForPreviousBlockProof delays signing (delay measured from round start, capped to fit the
// signature subround) to give the previous block's proof time to arrive; returns true to abort
func (sr *subroundSignature) waitForPreviousBlockProof(ctx context.Context, previousHash []byte, nonce uint64) bool {
	roundStart := sr.GetRoundTimeStamp()
	targetTime := time.Duration(float64(sr.RoundHandler().TimeDuration()) * competingBlockSignDelay)
	delay := sr.RoundHandler().RemainingTime(roundStart, targetTime)
	if delay <= 0 {
		log.Debug("waitForPreviousBlockProof: already past competing block delay deadline, proceeding to sign")
		return false
	}

	sigEndDuration := time.Duration(sr.EndTime())
	remaining := sr.RoundHandler().RemainingTime(roundStart, sigEndDuration)
	safetyMargin := 10 * time.Millisecond
	maxDelay := remaining - safetyMargin
	if maxDelay <= 0 {
		log.Debug("waitForPreviousBlockProof: no time remaining before signature send deadline, proceeding to sign")
		return false
	}
	if delay > maxDelay {
		delay = maxDelay
	}

	log.Debug("waitForPreviousBlockProof: competing block detected, delaying before signing",
		"nonce", nonce,
		"previousHash", previousHash,
		"delay", delay)

	shardID := sr.ShardCoordinator().SelfId()
	deadline := time.After(delay)
	ticker := time.NewTicker(timeBetweenSignaturesChecks)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug("waitForPreviousBlockProof: context cancelled, aborting")
			return true
		case <-ticker.C:
			if sr.EquivalentProofsPool().HasProof(shardID, previousHash) {
				log.Debug("waitForPreviousBlockProof: proof arrived for previous block, aborting signing",
					"nonce", nonce,
					"previousHash", previousHash)
				return true
			}
			if sr.HasProofForCompetingBlock() {
				log.Debug("waitForPreviousBlockProof: competing block proof detected, aborting signing")
				return true
			}
		case <-deadline:
			log.Debug("waitForPreviousBlockProof: delay expired with no proof for previous block, proceeding to sign",
				"nonce", nonce)
			return false
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (sr *subroundSignature) IsInterfaceNil() bool {
	return sr == nil
}
