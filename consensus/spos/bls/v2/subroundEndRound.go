package v2

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/bits"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/display"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
)

const timeBetweenSignaturesChecks = time.Millisecond * 5

type subroundEndRound struct {
	*spos.Subround
	processingThresholdPercentage int
	appStatusHandler              core.AppStatusHandler
	mutProcessingEndRound         sync.Mutex
	sentSignatureTracker          spos.SentSignaturesTracker
	worker                        spos.WorkerHandler
	signatureThrottler            core.Throttler
}

// NewSubroundEndRound creates a subroundEndRound object
func NewSubroundEndRound(
	baseSubround *spos.Subround,
	processingThresholdPercentage int,
	appStatusHandler core.AppStatusHandler,
	sentSignatureTracker spos.SentSignaturesTracker,
	worker spos.WorkerHandler,
	signatureThrottler core.Throttler,
) (*subroundEndRound, error) {
	err := checkNewSubroundEndRoundParams(baseSubround)
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

	srEndRound := subroundEndRound{
		Subround:                      baseSubround,
		processingThresholdPercentage: processingThresholdPercentage,
		appStatusHandler:              appStatusHandler,
		mutProcessingEndRound:         sync.Mutex{},
		sentSignatureTracker:          sentSignatureTracker,
		worker:                        worker,
		signatureThrottler:            signatureThrottler,
	}
	srEndRound.Job = srEndRound.doEndRoundJob
	srEndRound.Check = srEndRound.doEndRoundConsensusCheck
	srEndRound.Extend = worker.Extend

	return &srEndRound, nil
}

func checkNewSubroundEndRoundParams(
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

func (sr *subroundEndRound) isProofForCurrentConsensus(proof consensus.ProofHandler) bool {
	return bytes.Equal(sr.GetData(), proof.GetHeaderHash())
}

// receivedProof method is called when a block header final info is received
func (sr *subroundEndRound) receivedProof(proof consensus.ProofHandler) {
	sr.mutProcessingEndRound.Lock()
	defer sr.mutProcessingEndRound.Unlock()

	if sr.IsJobDone(sr.SelfPubKey(), sr.Current()) {
		return
	}
	if !sr.IsConsensusDataSet() {
		return
	}
	if check.IfNil(sr.GetHeader()) {
		return
	}
	if !sr.isProofForCurrentConsensus(proof) {
		return
	}

	// no need to re-verify the proof since it was already verified when it was added to the proofs pool
	log.Debug("step 3: block header final info has been received",
		"PubKeysBitmap", proof.GetPubKeysBitmap(),
		"AggregateSignature", proof.GetAggregatedSignature(),
		"HederHash", proof.GetHeaderHash())

	sr.doEndRoundJobByNode()
}

// receivedInvalidSignersInfo method is called when a message with invalid signers has been received
func (sr *subroundEndRound) receivedInvalidSignersInfo(_ context.Context, cnsDta *consensus.Message) bool {
	messageSender := string(cnsDta.PubKey)

	if !sr.IsConsensusDataSet() {
		return false
	}
	if check.IfNil(sr.GetHeader()) {
		return false
	}

	isSelfSender := sr.IsNodeSelf(messageSender) || sr.IsKeyManagedBySelf([]byte(messageSender))
	if isSelfSender {
		return false
	}

	if !sr.IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.RoundHandler().Index(), sr.Current()) {
		return false
	}

	if len(cnsDta.InvalidSigners) == 0 {
		return false
	}

	err := sr.verifyInvalidSigners(cnsDta.InvalidSigners)
	if err != nil {
		log.Trace("receivedInvalidSignersInfo.verifyInvalidSigners", "error", err.Error())
		return false
	}

	log.Debug("step 3: invalid signers info has been evaluated")

	sr.PeerHonestyHandler().ChangeScore(
		messageSender,
		spos.GetConsensusTopicID(sr.ShardCoordinator()),
		spos.LeaderPeerHonestyIncreaseFactor,
	)

	return true
}

func (sr *subroundEndRound) verifyInvalidSigners(invalidSigners []byte) error {
	messages, err := sr.MessageSigningHandler().Deserialize(invalidSigners)
	if err != nil {
		return err
	}

	for _, msg := range messages {
		err = sr.verifyInvalidSigner(msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sr *subroundEndRound) verifyInvalidSigner(msg p2p.MessageP2P) error {
	err := sr.MessageSigningHandler().Verify(msg)
	if err != nil {
		return err
	}

	cnsMsg := &consensus.Message{}
	err = sr.Marshalizer().Unmarshal(cnsMsg, msg.Data())
	if err != nil {
		return err
	}

	err = sr.SigningHandler().VerifySingleSignature(cnsMsg.PubKey, cnsMsg.BlockHeaderHash, cnsMsg.SignatureShare)
	if err != nil {
		log.Debug("verifyInvalidSigner: confirmed that node provided invalid signature",
			"pubKey", cnsMsg.PubKey,
			"blockHeaderHash", cnsMsg.BlockHeaderHash,
			"error", err.Error(),
		)
		sr.applyBlacklistOnNode(msg.Peer())
	}

	return nil
}

func (sr *subroundEndRound) applyBlacklistOnNode(peer core.PeerID) {
	sr.PeerBlacklistHandler().BlacklistPeer(peer, common.InvalidSigningBlacklistDuration)
}

// doEndRoundJob method does the job of the subround EndRound
func (sr *subroundEndRound) doEndRoundJob(_ context.Context) bool {
	if check.IfNil(sr.GetHeader()) {
		return false
	}

	sr.mutProcessingEndRound.Lock()
	defer sr.mutProcessingEndRound.Unlock()

	return sr.doEndRoundJobByNode()
}

func (sr *subroundEndRound) commitBlock() error {
	startTime := time.Now()
	err := sr.BlockProcessor().CommitBlock(sr.GetHeader(), sr.GetBody())
	elapsedTime := time.Since(startTime)
	if elapsedTime >= common.CommitMaxTime {
		log.Warn("doEndRoundJobByNode.CommitBlock", "elapsed time", elapsedTime)
	} else {
		log.Debug("elapsed time to commit block", "time [s]", elapsedTime)
	}
	if err != nil {
		log.Debug("doEndRoundJobByNode.CommitBlock", "error", err)
		return err
	}

	return nil
}

func (sr *subroundEndRound) doEndRoundJobByNode() bool {
	if sr.shouldSendProof() {
		if !sr.waitForSignalSync() {
			return false
		}
	}

	proof, ok := sr.sendProof()

	if !ok {
		return false
	}

	err := sr.commitBlock()
	if err != nil {
		return false
	}

	// if proof not nil, it was created and broadcasted so it has to be added to the pool
	if proof != nil {
		err = sr.EquivalentProofsPool().AddProof(proof)
		if err != nil {
			log.Debug("doEndRoundJobByNode.AddProof", "error", err)
			return false
		}
	}

	sr.SetStatus(sr.Current(), spos.SsFinished)

	sr.worker.DisplayStatistics()

	log.Debug("step 3: Body and Header have been committed")

	msg := fmt.Sprintf("Added proposed block with nonce  %d  in blockchain", sr.GetHeader().GetNonce())
	log.Debug(display.Headline(msg, sr.SyncTimer().FormattedCurrentTime(), "+"))

	sr.updateMetricsForLeader()

	return true
}

func (sr *subroundEndRound) sendProof() (data.HeaderProofHandler, bool) {
	if !sr.shouldSendProof() {
		return nil, true
	}

	bitmap := sr.GenerateBitmap(bls.SrSignature)
	err := sr.checkSignaturesValidity(bitmap)
	if err != nil {
		log.Debug("sendProof.checkSignaturesValidity", "error", err.Error())
		return nil, false
	}

	// Aggregate signatures, handle invalid signers and send final info if needed
	bitmap, sig, err := sr.aggregateSigsAndHandleInvalidSigners(bitmap)
	if err != nil {
		log.Debug("sendProof.aggregateSigsAndHandleInvalidSigners", "error", err.Error())
		return nil, false
	}

	ok := sr.ScheduledProcessor().IsProcessedOKWithTimeout()
	// placeholder for subroundEndRound.doEndRoundJobByLeader script
	if !ok {
		return nil, false
	}

	roundHandler := sr.RoundHandler()
	if roundHandler.RemainingTime(roundHandler.TimeStamp(), roundHandler.TimeDuration()) < 0 {
		log.Debug("sendProof: time is out -> cancel broadcasting final info and header",
			"round time stamp", roundHandler.TimeStamp(),
			"current time", time.Now())
		return nil, false
	}

	// broadcast header proof
	proof, err := sr.createAndBroadcastProof(sig, bitmap)
	return proof, err == nil
}

func (sr *subroundEndRound) shouldSendProof() bool {
	if sr.EquivalentProofsPool().HasProof(sr.ShardCoordinator().SelfId(), sr.GetData()) {
		log.Debug("shouldSendProof: equivalent message already processed")
		return false
	}

	return true
}

func (sr *subroundEndRound) aggregateSigsAndHandleInvalidSigners(bitmap []byte) ([]byte, []byte, error) {
	sig, err := sr.SigningHandler().AggregateSigs(bitmap, sr.GetHeader().GetEpoch())
	if err != nil {
		log.Debug("doEndRoundJobByNode.AggregateSigs", "error", err.Error())

		return sr.handleInvalidSignersOnAggSigFail()
	}

	err = sr.SigningHandler().SetAggregatedSig(sig)
	if err != nil {
		log.Debug("doEndRoundJobByNode.SetAggregatedSig", "error", err.Error())
		return nil, nil, err
	}

	// the header (hash) verified here is with leader signature on it
	err = sr.SigningHandler().Verify(sr.GetData(), bitmap, sr.GetHeader().GetEpoch())
	if err != nil {
		log.Debug("doEndRoundJobByNode.Verify", "error", err.Error())

		return sr.handleInvalidSignersOnAggSigFail()
	}

	return bitmap, sig, nil
}

func (sr *subroundEndRound) checkGoRoutinesThrottler(ctx context.Context) error {
	for {
		if sr.signatureThrottler.CanProcess() {
			break
		}

		select {
		case <-time.After(time.Millisecond):
			continue
		case <-ctx.Done():
			return spos.ErrTimeIsOut
		}
	}
	return nil
}

// verifySignature implements parallel signature verification
func (sr *subroundEndRound) verifySignature(i int, pk string, sigShare []byte) error {
	err := sr.SigningHandler().VerifySignatureShare(uint16(i), sigShare, sr.GetData(), sr.GetHeader().GetEpoch())
	if err != nil {
		log.Trace("VerifySignatureShare returned an error: ", err)
		errSetJob := sr.SetJobDone(pk, bls.SrSignature, false)
		if errSetJob != nil {
			return errSetJob
		}

		decreaseFactor := -spos.ValidatorPeerHonestyIncreaseFactor + spos.ValidatorPeerHonestyDecreaseFactor

		sr.PeerHonestyHandler().ChangeScore(
			pk,
			spos.GetConsensusTopicID(sr.ShardCoordinator()),
			decreaseFactor,
		)
		return err
	}

	log.Trace("verifyNodesOnAggSigVerificationFail: verifying signature share", "public key", pk)

	return nil
}

func (sr *subroundEndRound) verifyNodesOnAggSigFail(ctx context.Context) ([]string, error) {
	wg := &sync.WaitGroup{}
	mutex := &sync.Mutex{}
	invalidPubKeys := make([]string, 0)
	pubKeys := sr.ConsensusGroup()

	if check.IfNil(sr.GetHeader()) {
		return nil, spos.ErrNilHeader
	}

	for i, pk := range pubKeys {
		isJobDone, err := sr.JobDone(pk, bls.SrSignature)
		if err != nil || !isJobDone {
			continue
		}

		sigShare, err := sr.SigningHandler().SignatureShare(uint16(i))
		if err != nil {
			return nil, err
		}

		err = sr.checkGoRoutinesThrottler(ctx)
		if err != nil {
			return nil, err
		}

		sr.signatureThrottler.StartProcessing()

		wg.Add(1)

		go func(i int, pk string, wg *sync.WaitGroup, sigShare []byte) {
			defer func() {
				sr.signatureThrottler.EndProcessing()
				wg.Done()
			}()
			errSigVerification := sr.verifySignature(i, pk, sigShare)
			if errSigVerification != nil {
				mutex.Lock()
				invalidPubKeys = append(invalidPubKeys, pk)
				mutex.Unlock()
			}
		}(i, pk, wg, sigShare)
	}
	wg.Wait()

	return invalidPubKeys, nil
}

func (sr *subroundEndRound) getFullMessagesForInvalidSigners(invalidPubKeys []string) ([]byte, error) {
	p2pMessages := make([]p2p.MessageP2P, 0)

	for _, pk := range invalidPubKeys {
		p2pMsg, ok := sr.GetMessageWithSignature(pk)
		if !ok {
			log.Trace("message not found in state for invalid signer", "pubkey", pk)
			continue
		}

		p2pMessages = append(p2pMessages, p2pMsg)
	}

	invalidSigners, err := sr.MessageSigningHandler().Serialize(p2pMessages)
	if err != nil {
		return nil, err
	}

	return invalidSigners, nil
}

func (sr *subroundEndRound) handleInvalidSignersOnAggSigFail() ([]byte, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sr.RoundHandler().TimeDuration())
	invalidPubKeys, err := sr.verifyNodesOnAggSigFail(ctx)
	cancel()
	if err != nil {
		log.Debug("doEndRoundJobByNode.verifyNodesOnAggSigFail", "error", err.Error())
		return nil, nil, err
	}

	_, err = sr.getFullMessagesForInvalidSigners(invalidPubKeys)
	if err != nil {
		log.Debug("doEndRoundJobByNode.getFullMessagesForInvalidSigners", "error", err.Error())
		return nil, nil, err
	}

	// TODO: handle invalid signers broadcast without flooding the network
	// if len(invalidSigners) > 0 {
	// 	sr.createAndBroadcastInvalidSigners(invalidSigners)
	// }

	bitmap, sig, err := sr.computeAggSigOnValidNodes()
	if err != nil {
		log.Debug("doEndRoundJobByNode.computeAggSigOnValidNodes", "error", err.Error())
		return nil, nil, err
	}

	return bitmap, sig, nil
}

func (sr *subroundEndRound) computeAggSigOnValidNodes() ([]byte, []byte, error) {
	threshold := sr.Threshold(bls.SrSignature)
	numValidSigShares := sr.ComputeSize(bls.SrSignature)

	if check.IfNil(sr.GetHeader()) {
		return nil, nil, spos.ErrNilHeader
	}

	if numValidSigShares < threshold {
		return nil, nil, fmt.Errorf("%w: number of valid sig shares lower than threshold, numSigShares: %d, threshold: %d",
			spos.ErrInvalidNumSigShares, numValidSigShares, threshold)
	}

	bitmap := sr.GenerateBitmap(bls.SrSignature)
	err := sr.checkSignaturesValidity(bitmap)
	if err != nil {
		return nil, nil, err
	}

	sig, err := sr.SigningHandler().AggregateSigs(bitmap, sr.GetHeader().GetEpoch())
	if err != nil {
		return nil, nil, err
	}

	err = sr.SigningHandler().SetAggregatedSig(sig)
	if err != nil {
		return nil, nil, err
	}

	return bitmap, sig, nil
}

func (sr *subroundEndRound) createAndBroadcastProof(signature []byte, bitmap []byte) (*block.HeaderProof, error) {
	headerProof := &block.HeaderProof{
		PubKeysBitmap:       bitmap,
		AggregatedSignature: signature,
		HeaderHash:          sr.GetData(),
		HeaderEpoch:         sr.GetHeader().GetEpoch(),
		HeaderNonce:         sr.GetHeader().GetNonce(),
		HeaderShardId:       sr.GetHeader().GetShardID(),
	}

	err := sr.BroadcastMessenger().BroadcastEquivalentProof(headerProof, []byte(sr.SelfPubKey()))
	if err != nil {
		return nil, err
	}

	log.Debug("step 3: block header proof has been sent",
		"PubKeysBitmap", bitmap,
		"AggregateSignature", signature)

	return headerProof, nil
}

func (sr *subroundEndRound) createAndBroadcastInvalidSigners(invalidSigners []byte) {
	if !sr.ShouldConsiderSelfKeyInConsensus() {
		return
	}

	sender, err := sr.GetLeader()
	if err != nil {
		log.Debug("createAndBroadcastInvalidSigners.getSender", "error", err)
		return
	}

	cnsMsg := consensus.NewConsensusMessage(
		sr.GetData(),
		nil,
		nil,
		nil,
		[]byte(sender),
		nil,
		int(bls.MtInvalidSigners),
		sr.RoundHandler().Index(),
		sr.ChainID(),
		nil,
		nil,
		nil,
		sr.GetAssociatedPid([]byte(sender)),
		invalidSigners,
	)

	err = sr.BroadcastMessenger().BroadcastConsensusMessage(cnsMsg)
	if err != nil {
		log.Debug("doEndRoundJob.BroadcastConsensusMessage", "error", err.Error())
		return
	}

	log.Debug("step 3: invalid signers info has been sent")
}

func (sr *subroundEndRound) updateMetricsForLeader() {
	// TODO: decide if we keep these metrics the same way
	sr.appStatusHandler.Increment(common.MetricCountAcceptedBlocks)
	sr.appStatusHandler.SetStringValue(common.MetricConsensusRoundState,
		fmt.Sprintf("valid block produced in %f sec", time.Since(sr.RoundHandler().TimeStamp()).Seconds()))
}

// doEndRoundConsensusCheck method checks if the consensus is achieved
func (sr *subroundEndRound) doEndRoundConsensusCheck() bool {
	if sr.GetRoundCanceled() {
		return false
	}

	return sr.IsSubroundFinished(sr.Current())
}

// IsBitmapInvalid checks if the bitmap is valid
// TODO: remove duplicated code and use the header sig verifier instead
func (sr *subroundEndRound) IsBitmapInvalid(bitmap []byte, consensusPubKeys []string) error {
	consensusSize := len(consensusPubKeys)

	expectedBitmapSize := consensusSize / 8
	if consensusSize%8 != 0 {
		expectedBitmapSize++
	}
	if len(bitmap) != expectedBitmapSize {
		log.Debug("wrong size bitmap",
			"expected number of bytes", expectedBitmapSize,
			"actual", len(bitmap))
		return ErrWrongSizeBitmap
	}

	numOfOnesInBitmap := 0
	for index := range bitmap {
		numOfOnesInBitmap += bits.OnesCount8(bitmap[index])
	}

	minNumRequiredSignatures := core.GetPBFTThreshold(consensusSize)
	if sr.FallbackHeaderValidator().ShouldApplyFallbackValidation(sr.GetHeader()) {
		minNumRequiredSignatures = core.GetPBFTFallbackThreshold(consensusSize)
		log.Warn("HeaderSigVerifier.verifyConsensusSize: fallback validation has been applied",
			"minimum number of signatures required", minNumRequiredSignatures,
			"actual number of signatures in bitmap", numOfOnesInBitmap,
		)
	}

	if numOfOnesInBitmap >= minNumRequiredSignatures {
		return nil
	}

	log.Debug("not enough signatures",
		"minimum expected", minNumRequiredSignatures,
		"actual", numOfOnesInBitmap)

	return ErrNotEnoughSignatures
}

func (sr *subroundEndRound) checkSignaturesValidity(bitmap []byte) error {
	consensusGroup := sr.ConsensusGroup()
	err := sr.IsBitmapInvalid(bitmap, consensusGroup)
	if err != nil {
		return err
	}

	signers := headerCheck.ComputeSignersPublicKeys(consensusGroup, bitmap)
	for _, pubKey := range signers {
		isSigJobDone, err := sr.JobDone(pubKey, bls.SrSignature)
		if err != nil {
			return err
		}

		if !isSigJobDone {
			return spos.ErrNilSignature
		}
	}

	return nil
}

func (sr *subroundEndRound) isOutOfTime() bool {
	startTime := sr.GetRoundTimeStamp()
	maxTime := sr.RoundHandler().TimeDuration() * time.Duration(sr.processingThresholdPercentage) / 100
	if sr.RoundHandler().RemainingTime(startTime, maxTime) < 0 {
		log.Debug("canceled round, time is out",
			"round", sr.SyncTimer().FormattedCurrentTime(), sr.RoundHandler().Index(),
			"subround", sr.Name())

		sr.SetRoundCanceled(true)
		return true
	}

	return false
}

func (sr *subroundEndRound) getMinConsensusGroupIndexOfManagedKeys() int {
	minIdx := sr.ConsensusGroupSize()

	for idx, validator := range sr.ConsensusGroup() {
		if !sr.IsKeyManagedBySelf([]byte(validator)) {
			continue
		}

		if idx < minIdx {
			minIdx = idx
		}
	}

	return minIdx
}

func (sr *subroundEndRound) waitForSignalSync() bool {
	if sr.IsSubroundFinished(sr.Current()) {
		return true
	}

	if sr.checkReceivedSignatures() {
		return true
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sr.waitSignatures(ctx)
	timerBetweenStatusChecks := time.NewTimer(timeBetweenSignaturesChecks)

	remainingSRTime := sr.remainingTime()
	timeout := time.NewTimer(remainingSRTime)
	for {
		select {
		case <-timerBetweenStatusChecks.C:
			if sr.IsSubroundFinished(sr.Current()) {
				log.Trace("subround already finished", "subround", sr.Name())
				return true
			}

			if sr.checkReceivedSignatures() {
				return true
			}
			timerBetweenStatusChecks.Reset(timeBetweenSignaturesChecks)
		case <-timeout.C:
			log.Debug("timeout while waiting for signatures or final info", "subround", sr.Name())
			return false
		}
	}
}

func (sr *subroundEndRound) waitSignatures(ctx context.Context) {
	remainingTime := sr.remainingTime()
	if sr.IsSubroundFinished(sr.Current()) {
		return
	}
	sr.SetWaitingAllSignaturesTimeOut(true)

	select {
	case <-time.After(remainingTime):
	case <-ctx.Done():
	}
	sr.ConsensusChannel() <- true
}

// maximum time to wait for signatures
func (sr *subroundEndRound) remainingTime() time.Duration {
	startTime := sr.RoundHandler().TimeStamp()
	maxTime := time.Duration(float64(sr.StartTime()) + float64(sr.EndTime()-sr.StartTime())*waitingAllSigsMaxTimeThreshold)
	remainingTime := sr.RoundHandler().RemainingTime(startTime, maxTime)

	return remainingTime
}

// receivedSignature method is called when a signature is received through the signature channel.
// If the signature is valid, then the jobDone map corresponding to the node which sent it,
// is set on true for the subround Signature
func (sr *subroundEndRound) receivedSignature(_ context.Context, cnsDta *consensus.Message) bool {
	node := string(cnsDta.PubKey)
	pkForLogs := core.GetTrimmedPk(hex.EncodeToString(cnsDta.PubKey))

	if !sr.IsConsensusDataSet() {
		return false
	}

	if !sr.IsNodeInConsensusGroup(node) {
		sr.PeerHonestyHandler().ChangeScore(
			node,
			spos.GetConsensusTopicID(sr.ShardCoordinator()),
			spos.ValidatorPeerHonestyDecreaseFactor,
		)

		return false
	}

	if !sr.IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.RoundHandler().Index(), sr.Current()) {
		return false
	}

	index, err := sr.ConsensusGroupIndex(node)
	if err != nil {
		log.Debug("receivedSignature.ConsensusGroupIndex",
			"node", pkForLogs,
			"error", err.Error())
		return false
	}

	err = sr.SigningHandler().StoreSignatureShare(uint16(index), cnsDta.SignatureShare)
	if err != nil {
		log.Debug("receivedSignature.StoreSignatureShare",
			"node", pkForLogs,
			"index", index,
			"error", err.Error())
		return false
	}

	err = sr.SetJobDone(node, bls.SrSignature, true)
	if err != nil {
		log.Debug("receivedSignature.SetJobDone",
			"node", pkForLogs,
			"subround", sr.Name(),
			"error", err.Error())
		return false
	}

	sr.PeerHonestyHandler().ChangeScore(
		node,
		spos.GetConsensusTopicID(sr.ShardCoordinator()),
		spos.ValidatorPeerHonestyIncreaseFactor,
	)

	return true
}

func (sr *subroundEndRound) checkReceivedSignatures() bool {
	threshold := sr.Threshold(bls.SrSignature)
	if sr.FallbackHeaderValidator().ShouldApplyFallbackValidation(sr.GetHeader()) {
		threshold = sr.FallbackThreshold(bls.SrSignature)
		log.Warn("subroundEndRound.checkReceivedSignatures: fallback validation has been applied",
			"minimum number of signatures required", threshold,
			"actual number of signatures received", sr.getNumOfSignaturesCollected(),
		)
	}

	areSignaturesCollected, numSigs := sr.areSignaturesCollected(threshold)
	areAllSignaturesCollected := numSigs == sr.ConsensusGroupSize()

	isSignatureCollectionDone := areAllSignaturesCollected || (areSignaturesCollected && sr.GetWaitingAllSignaturesTimeOut())

	isSelfJobDone := sr.IsSelfJobDone(bls.SrSignature)

	shouldStopWaitingSignatures := isSelfJobDone && isSignatureCollectionDone
	if shouldStopWaitingSignatures {
		log.Debug("step 2: signatures collection done",
			"subround", sr.Name(),
			"signatures received", numSigs,
			"total signatures", len(sr.ConsensusGroup()))

		return true
	}

	return false
}

func (sr *subroundEndRound) getNumOfSignaturesCollected() int {
	n := 0

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		node := sr.ConsensusGroup()[i]

		isSignJobDone, err := sr.JobDone(node, bls.SrSignature)
		if err != nil {
			log.Debug("getNumOfSignaturesCollected.JobDone",
				"node", node,
				"subround", sr.Name(),
				"error", err.Error())
			continue
		}

		if isSignJobDone {
			n++
		}
	}

	return n
}

// areSignaturesCollected method checks if the signatures received from the nodes, belonging to the current
// jobDone group, are more than the necessary given threshold
func (sr *subroundEndRound) areSignaturesCollected(threshold int) (bool, int) {
	n := sr.getNumOfSignaturesCollected()
	return n >= threshold, n
}

// IsInterfaceNil returns true if there is no value under the interface
func (sr *subroundEndRound) IsInterfaceNil() bool {
	return sr == nil
}
