package bls

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/p2p"
)

const timeBetweenSignaturesChecks = time.Millisecond * 5

type subroundEndRound struct {
	*spos.Subround
	processingThresholdPercentage      int
	appStatusHandler                   core.AppStatusHandler
	mutProcessingEndRound              sync.Mutex
	sentSignatureTracker               spos.SentSignaturesTracker
	worker                             spos.WorkerHandler
	mutEquivalentProofsCriticalSection sync.RWMutex
}

// NewSubroundEndRound creates a subroundEndRound object
func NewSubroundEndRound(
	baseSubround *spos.Subround,
	processingThresholdPercentage int,
	appStatusHandler core.AppStatusHandler,
	sentSignatureTracker spos.SentSignaturesTracker,
	worker spos.WorkerHandler,
) (*subroundEndRound, error) {
	err := checkNewSubroundEndRoundParams(
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

	srEndRound := subroundEndRound{
		Subround:                      baseSubround,
		processingThresholdPercentage: processingThresholdPercentage,
		appStatusHandler:              appStatusHandler,
		mutProcessingEndRound:         sync.Mutex{},
		sentSignatureTracker:          sentSignatureTracker,
		worker:                        worker,
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
	if baseSubround.ConsensusState == nil {
		return spos.ErrNilConsensusState
	}

	err := spos.ValidateConsensusCore(baseSubround.ConsensusCoreHandler)

	return err
}

// receivedBlockHeaderFinalInfo method is called when a block header final info is received
func (sr *subroundEndRound) receivedBlockHeaderFinalInfo(_ context.Context, cnsDta *consensus.Message) bool {
	sr.mutEquivalentProofsCriticalSection.Lock()
	defer sr.mutEquivalentProofsCriticalSection.Unlock()

	node := string(cnsDta.PubKey)

	if !sr.IsConsensusDataSet() {
		return false
	}
	if check.IfNil(sr.Header) {
		return false
	}

	// TODO[cleanup cns finality]: remove if statement
	isSenderAllowed := sr.IsNodeInConsensusGroup(node)
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, sr.Header.GetEpoch()) {
		isNodeLeader := sr.IsNodeLeaderInCurrentRound(node) && sr.ShouldConsiderSelfKeyInConsensus()
		isSenderAllowed = isNodeLeader || sr.IsMultiKeyLeaderInCurrentRound()
	}
	if !isSenderAllowed { // is NOT this node leader in current round?
		sr.PeerHonestyHandler().ChangeScore(
			node,
			spos.GetConsensusTopicID(sr.ShardCoordinator()),
			spos.LeaderPeerHonestyDecreaseFactor,
		)

		return false
	}

	// TODO[cleanup cns finality]: remove if
	isSelfSender := sr.IsNodeSelf(node) || sr.IsKeyManagedByCurrentNode([]byte(node))
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, sr.Header.GetEpoch()) {
		isSelfSender = sr.IsSelfLeaderInCurrentRound() || sr.IsMultiKeyLeaderInCurrentRound()
	}
	if isSelfSender {
		return false
	}

	if !sr.IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.RoundHandler().Index(), sr.Current()) {
		return false
	}

	hasProof := sr.worker.HasEquivalentMessage(cnsDta.BlockHeaderHash)
	if hasProof && sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, sr.Header.GetEpoch()) {
		return true
	}

	if !sr.isBlockHeaderFinalInfoValid(cnsDta) {
		return false
	}

	log.Debug("step 3: block header final info has been received",
		"PubKeysBitmap", cnsDta.PubKeysBitmap,
		"AggregateSignature", cnsDta.AggregateSignature,
		"LeaderSignature", cnsDta.LeaderSignature)

	sr.PeerHonestyHandler().ChangeScore(
		node,
		spos.GetConsensusTopicID(sr.ShardCoordinator()),
		spos.LeaderPeerHonestyIncreaseFactor,
	)

	return sr.doEndRoundJobByParticipant(cnsDta)
}

func (sr *subroundEndRound) isBlockHeaderFinalInfoValid(cnsDta *consensus.Message) bool {
	header := sr.Header.ShallowClone()

	// TODO[cleanup cns finality]: remove this
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, header.GetEpoch()) {
		return sr.verifySignatures(header, cnsDta)
	}

	err := sr.HeaderSigVerifier().VerifySignatureForHash(header, cnsDta.BlockHeaderHash, cnsDta.PubKeysBitmap, cnsDta.AggregateSignature)
	if err != nil {
		log.Debug("isBlockHeaderFinalInfoValid.VerifySignatureForHash", "error", err.Error())
		return false
	}

	return true
}

func (sr *subroundEndRound) verifySignatures(header data.HeaderHandler, cnsDta *consensus.Message) bool {
	err := header.SetPubKeysBitmap(cnsDta.PubKeysBitmap)
	if err != nil {
		log.Debug("verifySignatures.SetPubKeysBitmap", "error", err.Error())
		return false
	}

	err = header.SetSignature(cnsDta.AggregateSignature)
	if err != nil {
		log.Debug("verifySignatures.SetSignature", "error", err.Error())
		return false
	}

	err = header.SetLeaderSignature(cnsDta.LeaderSignature)
	if err != nil {
		log.Debug("verifySignatures.SetLeaderSignature", "error", err.Error())
		return false
	}

	err = sr.HeaderSigVerifier().VerifyLeaderSignature(header)
	if err != nil {
		log.Debug("verifySignatures.VerifyLeaderSignature", "error", err.Error())
		return false
	}
	err = sr.HeaderSigVerifier().VerifySignature(header)
	if err != nil {
		log.Debug("verifySignatures.VerifySignature", "error", err.Error())
		return false
	}

	return true
}

// receivedInvalidSignersInfo method is called when a message with invalid signers has been received
func (sr *subroundEndRound) receivedInvalidSignersInfo(_ context.Context, cnsDta *consensus.Message) bool {
	messageSender := string(cnsDta.PubKey)

	if !sr.IsConsensusDataSet() {
		return false
	}
	if check.IfNil(sr.Header) {
		return false
	}

	// TODO[cleanup cns finality]: remove if statement
	isSenderAllowed := sr.IsNodeInConsensusGroup(messageSender)
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, sr.Header.GetEpoch()) {
		isSelfLeader := sr.IsNodeLeaderInCurrentRound(messageSender) && sr.ShouldConsiderSelfKeyInConsensus()
		isSenderAllowed = isSelfLeader || sr.IsMultiKeyLeaderInCurrentRound()
	}
	if !isSenderAllowed { // is NOT this node leader in current round?
		sr.PeerHonestyHandler().ChangeScore(
			messageSender,
			spos.GetConsensusTopicID(sr.ShardCoordinator()),
			spos.LeaderPeerHonestyDecreaseFactor,
		)

		return false
	}

	// TODO[cleanup cns finality]: update this check
	isSelfSender := messageSender == sr.SelfPubKey() || sr.IsKeyManagedByCurrentNode([]byte(messageSender))
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, sr.Header.GetEpoch()) {
		isSelfSender = sr.IsSelfLeaderInCurrentRound() || sr.IsMultiKeyLeaderInCurrentRound()
	}
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

	err = sr.SigningHandler().VerifySingleSignature(cnsMsg.PubKey, cnsMsg.BlockHeaderHash, cnsMsg.AggregateSignature)
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

func (sr *subroundEndRound) receivedHeader(headerHandler data.HeaderHandler) {
	isFlagEnabledForHeader := sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, headerHandler.GetEpoch())
	isLeader := sr.IsSelfLeaderInCurrentRound() || sr.IsMultiKeyLeaderInCurrentRound()
	isLeaderBeforeActivation := isLeader && !isFlagEnabledForHeader
	if sr.ConsensusGroup() == nil || isLeaderBeforeActivation {
		return
	}

	receivedHeaderHash, err := core.CalculateHash(sr.Marshalizer(), sr.Hasher(), headerHandler)
	if err != nil {
		return
	}
	isHeaderAlreadyReceived := bytes.Equal(receivedHeaderHash, sr.Data)
	if isFlagEnabledForHeader && isHeaderAlreadyReceived {
		return
	}

	sr.AddReceivedHeader(headerHandler)

	sr.doEndRoundJobByParticipant(nil)
}

// doEndRoundJob method does the job of the subround EndRound
func (sr *subroundEndRound) doEndRoundJob(_ context.Context) bool {
	if check.IfNil(sr.Header) {
		return false
	}

	// TODO[cleanup cns finality]: remove this code block
	isFlagEnabled := sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, sr.Header.GetEpoch())
	if !sr.IsSelfLeaderInCurrentRound() && !sr.IsMultiKeyLeaderInCurrentRound() && !isFlagEnabled {
		if sr.IsNodeInConsensusGroup(sr.SelfPubKey()) || sr.IsMultiKeyInConsensusGroup() {
			err := sr.prepareBroadcastBlockDataForValidator()
			if err != nil {
				log.Warn("validator in consensus group preparing for delayed broadcast",
					"error", err.Error())
			}
		}

		return sr.doEndRoundJobByParticipant(nil)
	}

	if !sr.IsNodeInConsensusGroup(sr.SelfPubKey()) && !sr.IsMultiKeyInConsensusGroup() {
		return sr.doEndRoundJobByParticipant(nil)
	}

	return sr.doEndRoundJobByLeader()
}

// TODO[cleanup cns finality]: rename this method, as this will be done by each participant
func (sr *subroundEndRound) doEndRoundJobByLeader() bool {
	if check.IfNil(sr.Header) {
		return false
	}

	sender, err := sr.getSender()
	if err != nil {
		return false
	}

	if !sr.waitForSignalSync() {
		return false
	}

	sr.mutEquivalentProofsCriticalSection.Lock()
	defer sr.mutEquivalentProofsCriticalSection.Unlock()

	if !sr.shouldSendFinalInfo() {
		return false
	}

	if !sr.sendFinalInfo(sender) {
		return false
	}

	// broadcast header
	// TODO[Sorin next PR]: decide if we send this with the delayed broadcast
	err = sr.BroadcastMessenger().BroadcastHeader(sr.Header, sender)
	if err != nil {
		log.Warn("broadcastHeader.BroadcastHeader", "error", err.Error())
	}

	startTime := time.Now()
	err = sr.BlockProcessor().CommitBlock(sr.Header, sr.Body)
	elapsedTime := time.Since(startTime)
	if elapsedTime >= common.CommitMaxTime {
		log.Warn("doEndRoundJobByLeader.CommitBlock", "elapsed time", elapsedTime)
	} else {
		log.Debug("elapsed time to commit block",
			"time [s]", elapsedTime,
		)
	}
	if err != nil {
		log.Debug("doEndRoundJobByLeader.CommitBlock", "error", err)
		return false
	}

	sr.SetStatus(sr.Current(), spos.SsFinished)

	sr.worker.DisplayStatistics()

	log.Debug("step 3: Body and Header have been committed and header has been broadcast")

	err = sr.broadcastBlockDataLeader(sender)
	if err != nil {
		log.Debug("doEndRoundJobByLeader.broadcastBlockDataLeader", "error", err.Error())
	}

	msg := fmt.Sprintf("Added proposed block with nonce  %d  in blockchain", sr.Header.GetNonce())
	log.Debug(display.Headline(msg, sr.SyncTimer().FormattedCurrentTime(), "+"))

	sr.updateMetricsForLeader()

	return true
}

func (sr *subroundEndRound) sendFinalInfo(sender []byte) bool {
	bitmap := sr.GenerateBitmap(SrSignature)
	err := sr.checkSignaturesValidity(bitmap)
	if err != nil {
		log.Debug("sendFinalInfo.checkSignaturesValidity", "error", err.Error())
		return false
	}

	// Aggregate sig and add it to the block
	bitmap, sig, err := sr.aggregateSigsAndHandleInvalidSigners(bitmap)
	if err != nil {
		log.Debug("sendFinalInfo.aggregateSigsAndHandleInvalidSigners", "error", err.Error())
		return false
	}

	// TODO[cleanup cns finality]: remove this code block
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, sr.Header.GetEpoch()) {
		err = sr.Header.SetPubKeysBitmap(bitmap)
		if err != nil {
			log.Debug("sendFinalInfo.SetPubKeysBitmap", "error", err.Error())
			return false
		}

		err = sr.Header.SetSignature(sig)
		if err != nil {
			log.Debug("sendFinalInfo.SetSignature", "error", err.Error())
			return false
		}

		// Header is complete so the leader can sign it
		leaderSignature, err := sr.signBlockHeader(sender)
		if err != nil {
			log.Error(err.Error())
			return false
		}

		err = sr.Header.SetLeaderSignature(leaderSignature)
		if err != nil {
			log.Debug("sendFinalInfo.SetLeaderSignature", "error", err.Error())
			return false
		}
	} else {
		proof := data.HeaderProof{
			AggregatedSignature: sig,
			PubKeysBitmap:       bitmap,
		}

		sr.worker.SetValidEquivalentProof(sr.GetData(), proof)

		sr.Blockchain().SetCurrentHeaderProof(proof)
	}

	ok := sr.ScheduledProcessor().IsProcessedOKWithTimeout()
	// placeholder for subroundEndRound.doEndRoundJobByLeader script
	if !ok {
		return false
	}

	roundHandler := sr.RoundHandler()
	if roundHandler.RemainingTime(roundHandler.TimeStamp(), roundHandler.TimeDuration()) < 0 {
		log.Debug("sendFinalInfo: time is out -> cancel broadcasting final info and header",
			"round time stamp", roundHandler.TimeStamp(),
			"current time", time.Now())
		return false
	}

	// broadcast header and final info section
	leaderSigToBroadcast := sr.Header.GetLeaderSignature()
	if sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, sr.Header.GetEpoch()) {
		leaderSigToBroadcast = nil
	}

	return sr.createAndBroadcastHeaderFinalInfoForKey(sig, bitmap, leaderSigToBroadcast, sender)
}

func (sr *subroundEndRound) shouldSendFinalInfo() bool {
	// TODO[cleanup cns finality]: remove this check
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, sr.Header.GetEpoch()) {
		return true
	}

	// TODO: check if this is the best approach. Perhaps we don't want to relay only on the first received message
	if sr.worker.HasEquivalentMessage(sr.GetData()) {
		log.Debug("shouldSendFinalInfo: equivalent message already sent")
		return false
	}

	return true
}

func (sr *subroundEndRound) aggregateSigsAndHandleInvalidSigners(bitmap []byte) ([]byte, []byte, error) {
	sig, err := sr.SigningHandler().AggregateSigs(bitmap, sr.Header.GetEpoch())
	if err != nil {
		log.Debug("doEndRoundJobByLeader.AggregateSigs", "error", err.Error())

		return sr.handleInvalidSignersOnAggSigFail()
	}

	err = sr.SigningHandler().SetAggregatedSig(sig)
	if err != nil {
		log.Debug("doEndRoundJobByLeader.SetAggregatedSig", "error", err.Error())
		return nil, nil, err
	}

	err = sr.SigningHandler().Verify(sr.GetData(), bitmap, sr.Header.GetEpoch())
	if err != nil {
		log.Debug("doEndRoundJobByLeader.Verify", "error", err.Error())

		return sr.handleInvalidSignersOnAggSigFail()
	}

	return bitmap, sig, nil
}

func (sr *subroundEndRound) verifyNodesOnAggSigFail() ([]string, error) {
	invalidPubKeys := make([]string, 0)
	pubKeys := sr.ConsensusGroup()

	if check.IfNil(sr.Header) {
		return nil, spos.ErrNilHeader
	}

	for i, pk := range pubKeys {
		isJobDone, err := sr.JobDone(pk, SrSignature)
		if err != nil || !isJobDone {
			continue
		}

		sigShare, err := sr.SigningHandler().SignatureShare(uint16(i))
		if err != nil {
			return nil, err
		}

		isSuccessfull := true
		err = sr.SigningHandler().VerifySignatureShare(uint16(i), sigShare, sr.GetData(), sr.Header.GetEpoch())
		if err != nil {
			isSuccessfull = false

			err = sr.SetJobDone(pk, SrSignature, false)
			if err != nil {
				return nil, err
			}

			// use increase factor since it was added optimistically, and it proved to be wrong
			decreaseFactor := -spos.ValidatorPeerHonestyIncreaseFactor + spos.ValidatorPeerHonestyDecreaseFactor
			sr.PeerHonestyHandler().ChangeScore(
				pk,
				spos.GetConsensusTopicID(sr.ShardCoordinator()),
				decreaseFactor,
			)

			invalidPubKeys = append(invalidPubKeys, pk)
		}

		log.Trace("verifyNodesOnAggSigVerificationFail: verifying signature share", "public key", pk, "is successfull", isSuccessfull)
	}

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
	invalidPubKeys, err := sr.verifyNodesOnAggSigFail()
	if err != nil {
		log.Debug("doEndRoundJobByLeader.verifyNodesOnAggSigFail", "error", err.Error())
		return nil, nil, err
	}

	invalidSigners, err := sr.getFullMessagesForInvalidSigners(invalidPubKeys)
	if err != nil {
		log.Debug("doEndRoundJobByLeader.getFullMessagesForInvalidSigners", "error", err.Error())
		return nil, nil, err
	}

	if len(invalidSigners) > 0 {
		sr.createAndBroadcastInvalidSigners(invalidSigners)
	}

	bitmap, sig, err := sr.computeAggSigOnValidNodes()
	if err != nil {
		log.Debug("doEndRoundJobByLeader.computeAggSigOnValidNodes", "error", err.Error())
		return nil, nil, err
	}

	return bitmap, sig, nil
}

func (sr *subroundEndRound) computeAggSigOnValidNodes() ([]byte, []byte, error) {
	threshold := sr.Threshold(SrSignature)
	numValidSigShares := sr.ComputeSize(SrSignature)

	if check.IfNil(sr.Header) {
		return nil, nil, spos.ErrNilHeader
	}

	if numValidSigShares < threshold {
		return nil, nil, fmt.Errorf("%w: number of valid sig shares lower than threshold, numSigShares: %d, threshold: %d",
			spos.ErrInvalidNumSigShares, numValidSigShares, threshold)
	}

	bitmap := sr.GenerateBitmap(SrSignature)
	err := sr.checkSignaturesValidity(bitmap)
	if err != nil {
		return nil, nil, err
	}

	sig, err := sr.SigningHandler().AggregateSigs(bitmap, sr.Header.GetEpoch())
	if err != nil {
		return nil, nil, err
	}

	err = sr.SigningHandler().SetAggregatedSig(sig)
	if err != nil {
		return nil, nil, err
	}

	return bitmap, sig, nil
}

func (sr *subroundEndRound) createAndBroadcastHeaderFinalInfoForKey(signature []byte, bitmap []byte, leaderSignature []byte, pubKey []byte) bool {
	cnsMsg := consensus.NewConsensusMessage(
		sr.GetData(),
		nil,
		nil,
		nil,
		pubKey,
		nil,
		int(MtBlockHeaderFinalInfo),
		sr.RoundHandler().Index(),
		sr.ChainID(),
		bitmap,
		signature,
		leaderSignature,
		sr.GetAssociatedPid(pubKey),
		nil,
	)

	index, err := sr.ConsensusGroupIndex(string(pubKey))
	if err != nil {
		log.Debug("createAndBroadcastHeaderFinalInfoForKey.ConsensusGroupIndex", "error", err.Error())
		return false
	}

	if !sr.EnableEpochsHandler().IsFlagEnabled(common.EquivalentMessagesFlag) {
		err = sr.BroadcastMessenger().BroadcastConsensusMessage(cnsMsg)
		if err != nil {
			log.Debug("createAndBroadcastHeaderFinalInfoForKey.BroadcastConsensusMessage", "error", err.Error())
			return false
		}

		log.Debug("step 3: block header final info has been sent",
			"PubKeysBitmap", bitmap,
			"AggregateSignature", signature,
			"LeaderSignature", leaderSignature)

		return true
	}

	sr.BroadcastMessenger().PrepareBroadcastFinalConsensusMessage(cnsMsg, index)
	log.Debug("step 3: block header final info has been sent to delayed broadcaster",
		"PubKeysBitmap", bitmap,
		"AggregateSignature", signature,
		"LeaderSignature", leaderSignature,
		"Index", index)

	return true
}

func (sr *subroundEndRound) createAndBroadcastInvalidSigners(invalidSigners []byte) {
	isSelfLeader := sr.IsSelfLeaderInCurrentRound() && sr.ShouldConsiderSelfKeyInConsensus()
	if !(isSelfLeader || sr.IsMultiKeyLeaderInCurrentRound()) {
		return
	}

	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		log.Debug("createAndBroadcastInvalidSigners.GetLeader", "error", errGetLeader)
		return
	}

	cnsMsg := consensus.NewConsensusMessage(
		sr.GetData(),
		nil,
		nil,
		nil,
		[]byte(leader),
		nil,
		int(MtInvalidSigners),
		sr.RoundHandler().Index(),
		sr.ChainID(),
		nil,
		nil,
		nil,
		sr.GetAssociatedPid([]byte(leader)),
		invalidSigners,
	)

	// TODO[Sorin next PR]: decide if we send this with the delayed broadcast
	err := sr.BroadcastMessenger().BroadcastConsensusMessage(cnsMsg)
	if err != nil {
		log.Debug("doEndRoundJob.BroadcastConsensusMessage", "error", err.Error())
		return
	}

	log.Debug("step 3: invalid signers info has been sent")
}

func (sr *subroundEndRound) doEndRoundJobByParticipant(cnsDta *consensus.Message) bool {
	sr.mutProcessingEndRound.Lock()
	defer sr.mutProcessingEndRound.Unlock()

	if sr.RoundCanceled {
		return false
	}
	if !sr.IsConsensusDataSet() {
		return false
	}
	if !sr.IsSubroundFinished(sr.Previous()) {
		return false
	}
	if sr.IsSubroundFinished(sr.Current()) {
		return false
	}

	haveHeader, header := sr.haveConsensusHeaderWithFullInfo(cnsDta)
	if !haveHeader {
		return false
	}

	defer func() {
		sr.SetProcessingBlock(false)
	}()

	sr.SetProcessingBlock(true)

	shouldNotCommitBlock := sr.ExtendedCalled || int64(header.GetRound()) < sr.RoundHandler().Index()
	if shouldNotCommitBlock {
		log.Debug("canceled round, extended has been called or round index has been changed",
			"round", sr.RoundHandler().Index(),
			"subround", sr.Name(),
			"header round", header.GetRound(),
			"extended called", sr.ExtendedCalled,
		)
		return false
	}

	if sr.isOutOfTime() {
		return false
	}

	ok := sr.ScheduledProcessor().IsProcessedOKWithTimeout()
	if !ok {
		return false
	}

	startTime := time.Now()
	err := sr.BlockProcessor().CommitBlock(header, sr.Body)
	elapsedTime := time.Since(startTime)
	if elapsedTime >= common.CommitMaxTime {
		log.Warn("doEndRoundJobByParticipant.CommitBlock", "elapsed time", elapsedTime)
	} else {
		log.Debug("elapsed time to commit block",
			"time [s]", elapsedTime,
		)
	}
	if err != nil {
		log.Debug("doEndRoundJobByParticipant.CommitBlock", "error", err.Error())
		return false
	}

	isNodeInConsensus := sr.IsNodeInConsensusGroup(sr.SelfPubKey()) || sr.IsMultiKeyInConsensusGroup()
	isEquivalentMessagesFlagEnabled := sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, header.GetEpoch())
	if isNodeInConsensus && cnsDta != nil && isEquivalentMessagesFlagEnabled {
		proof := data.HeaderProof{
			AggregatedSignature: cnsDta.AggregateSignature,
			PubKeysBitmap:       cnsDta.PubKeysBitmap,
		}
		sr.Blockchain().SetCurrentHeaderProof(proof)
		sr.worker.SetValidEquivalentProof(cnsDta.BlockHeaderHash, proof)
	}

	sr.SetStatus(sr.Current(), spos.SsFinished)

	// TODO[cleanup cns finality]: remove this
	if isNodeInConsensus && !isEquivalentMessagesFlagEnabled {
		err = sr.setHeaderForValidator(header)
		if err != nil {
			log.Warn("doEndRoundJobByParticipant", "error", err.Error())
		}
	}

	sr.worker.DisplayStatistics()

	log.Debug("step 3: Body and Header have been committed")

	headerTypeMsg := "received"
	if cnsDta != nil {
		headerTypeMsg = "assembled"
	}

	msg := fmt.Sprintf("Added %s block with nonce  %d  in blockchain", headerTypeMsg, header.GetNonce())
	log.Debug(display.Headline(msg, sr.SyncTimer().FormattedCurrentTime(), "-"))
	return true
}

func (sr *subroundEndRound) haveConsensusHeaderWithFullInfo(cnsDta *consensus.Message) (bool, data.HeaderHandler) {
	if cnsDta == nil {
		return sr.isConsensusHeaderReceived()
	}

	if check.IfNil(sr.Header) {
		return false, nil
	}

	header := sr.Header.ShallowClone()
	// TODO[cleanup cns finality]: remove this
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, header.GetEpoch()) {
		err := header.SetPubKeysBitmap(cnsDta.PubKeysBitmap)
		if err != nil {
			return false, nil
		}

		err = header.SetSignature(cnsDta.AggregateSignature)
		if err != nil {
			return false, nil
		}

		err = header.SetLeaderSignature(cnsDta.LeaderSignature)
		if err != nil {
			return false, nil
		}

		return true, header
	}

	return true, header
}

func (sr *subroundEndRound) isConsensusHeaderReceived() (bool, data.HeaderHandler) {
	if check.IfNil(sr.Header) {
		return false, nil
	}

	consensusHeaderHash, err := core.CalculateHash(sr.Marshalizer(), sr.Hasher(), sr.Header)
	if err != nil {
		log.Debug("isConsensusHeaderReceived: calculate consensus header hash", "error", err.Error())
		return false, nil
	}

	receivedHeaders := sr.GetReceivedHeaders()

	var receivedHeaderHash []byte
	for index := range receivedHeaders {
		// TODO[cleanup cns finality]: remove this
		receivedHeader := receivedHeaders[index].ShallowClone()
		if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, receivedHeader.GetEpoch()) {
			err = receivedHeader.SetLeaderSignature(nil)
			if err != nil {
				log.Debug("isConsensusHeaderReceived - SetLeaderSignature", "error", err.Error())
				return false, nil
			}

			err = receivedHeader.SetPubKeysBitmap(nil)
			if err != nil {
				log.Debug("isConsensusHeaderReceived - SetPubKeysBitmap", "error", err.Error())
				return false, nil
			}

			err = receivedHeader.SetSignature(nil)
			if err != nil {
				log.Debug("isConsensusHeaderReceived - SetSignature", "error", err.Error())
				return false, nil
			}
		}

		receivedHeaderHash, err = core.CalculateHash(sr.Marshalizer(), sr.Hasher(), receivedHeader)
		if err != nil {
			log.Debug("isConsensusHeaderReceived: calculate received header hash", "error", err.Error())
			return false, nil
		}

		if bytes.Equal(receivedHeaderHash, consensusHeaderHash) {
			return true, receivedHeaders[index]
		}
	}

	return false, nil
}

func (sr *subroundEndRound) signBlockHeader(leader []byte) ([]byte, error) {
	headerClone := sr.Header.ShallowClone()
	err := headerClone.SetLeaderSignature(nil)
	if err != nil {
		return nil, err
	}

	marshalizedHdr, err := sr.Marshalizer().Marshal(headerClone)
	if err != nil {
		return nil, err
	}

	return sr.SigningHandler().CreateSignatureForPublicKey(marshalizedHdr, leader)
}

func (sr *subroundEndRound) updateMetricsForLeader() {
	// TODO: decide if we keep these metrics the same way
	sr.appStatusHandler.Increment(common.MetricCountAcceptedBlocks)
	sr.appStatusHandler.SetStringValue(common.MetricConsensusRoundState,
		fmt.Sprintf("valid block produced in %f sec", time.Since(sr.RoundHandler().TimeStamp()).Seconds()))
}

func (sr *subroundEndRound) broadcastBlockDataLeader(sender []byte) error {
	miniBlocks, transactions, err := sr.BlockProcessor().MarshalizedDataToBroadcast(sr.Header, sr.Body)
	if err != nil {
		return err
	}

	// TODO[Sorin next PR]: decide if we send this with the delayed broadcast
	return sr.BroadcastMessenger().BroadcastBlockDataLeader(sr.Header, miniBlocks, transactions, sender)
}

func (sr *subroundEndRound) setHeaderForValidator(header data.HeaderHandler) error {
	idx, pk, miniBlocks, transactions, err := sr.getIndexPkAndDataToBroadcast()
	if err != nil {
		return err
	}

	go sr.BroadcastMessenger().PrepareBroadcastHeaderValidator(header, miniBlocks, transactions, idx, pk)

	return nil
}

func (sr *subroundEndRound) prepareBroadcastBlockDataForValidator() error {
	idx, pk, miniBlocks, transactions, err := sr.getIndexPkAndDataToBroadcast()
	if err != nil {
		return err
	}

	go sr.BroadcastMessenger().PrepareBroadcastBlockDataValidator(sr.Header, miniBlocks, transactions, idx, pk)

	return nil
}

// doEndRoundConsensusCheck method checks if the consensus is achieved
func (sr *subroundEndRound) doEndRoundConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	return sr.IsSubroundFinished(sr.Current())
}

// computeSignersPublicKeys will extract from the provided consensus group slice only the strings that matched with the bitmap
func computeSignersPublicKeys(consensusGroup []string, bitmap []byte) []string {
	nbBitsBitmap := len(bitmap) * 8
	consensusGroupSize := len(consensusGroup)
	size := consensusGroupSize
	if consensusGroupSize > nbBitsBitmap {
		size = nbBitsBitmap
	}

	result := make([]string, 0, len(consensusGroup))

	for i := 0; i < size; i++ {
		indexRequired := (bitmap[i/8] & (1 << uint16(i%8))) > 0
		if !indexRequired {
			continue
		}

		pubKey := consensusGroup[i]
		result = append(result, pubKey)
	}

	return result
}

func (sr *subroundEndRound) checkSignaturesValidity(bitmap []byte) error {
	if !sr.hasProposerSignature(bitmap) {
		return spos.ErrMissingProposerSignature
	}

	consensusGroup := sr.ConsensusGroup()
	signers := computeSignersPublicKeys(consensusGroup, bitmap)
	for _, pubKey := range signers {
		isSigJobDone, err := sr.JobDone(pubKey, SrSignature)
		if err != nil {
			return err
		}

		if !isSigJobDone {
			return spos.ErrNilSignature
		}
	}

	return nil
}

func (sr *subroundEndRound) hasProposerSignature(bitmap []byte) bool {
	// TODO[cleanup cns finality]: remove this check
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, sr.Header.GetEpoch()) {
		return true
	}

	return bitmap[0]&1 > 0
}

func (sr *subroundEndRound) isOutOfTime() bool {
	startTime := sr.RoundTimeStamp
	maxTime := sr.RoundHandler().TimeDuration() * time.Duration(sr.processingThresholdPercentage) / 100
	if sr.RoundHandler().RemainingTime(startTime, maxTime) < 0 {
		log.Debug("canceled round, time is out",
			"round", sr.SyncTimer().FormattedCurrentTime(), sr.RoundHandler().Index(),
			"subround", sr.Name())

		sr.RoundCanceled = true
		return true
	}

	return false
}

func (sr *subroundEndRound) getIndexPkAndDataToBroadcast() (int, []byte, map[uint32][]byte, map[string][][]byte, error) {
	minIdx := sr.getMinConsensusGroupIndexOfManagedKeys()

	idx, err := sr.SelfConsensusGroupIndex()
	if err == nil {
		if idx < minIdx {
			minIdx = idx
		}
	}

	if minIdx == sr.ConsensusGroupSize() {
		return -1, nil, nil, nil, err
	}

	miniBlocks, transactions, err := sr.BlockProcessor().MarshalizedDataToBroadcast(sr.Header, sr.Body)
	if err != nil {
		return -1, nil, nil, nil, err
	}

	consensusGroup := sr.ConsensusGroup()
	pk := []byte(consensusGroup[minIdx])

	return minIdx, pk, miniBlocks, transactions, nil
}

func (sr *subroundEndRound) getMinConsensusGroupIndexOfManagedKeys() int {
	minIdx := sr.ConsensusGroupSize()

	for idx, validator := range sr.ConsensusGroup() {
		if !sr.IsKeyManagedByCurrentNode([]byte(validator)) {
			continue
		}

		if idx < minIdx {
			minIdx = idx
		}
	}

	return minIdx
}

func (sr *subroundEndRound) getSender() ([]byte, error) {
	// TODO[cleanup cns finality]: remove this code block
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, sr.Header.GetEpoch()) {
		leader, errGetLeader := sr.GetLeader()
		if errGetLeader != nil {
			log.Debug("GetLeader", "error", errGetLeader)
			return nil, errGetLeader
		}

		return []byte(leader), nil
	}

	for _, pk := range sr.ConsensusGroup() {
		pkBytes := []byte(pk)
		if !sr.IsKeyManagedByCurrentNode(pkBytes) {
			continue
		}

		return pkBytes, nil
	}

	return []byte(sr.SelfPubKey()), nil
}

func (sr *subroundEndRound) waitForSignalSync() bool {
	// TODO[cleanup cns finality]: remove this
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, sr.Header.GetEpoch()) {
		return true
	}

	if sr.checkReceivedSignatures() {
		return true
	}

	go sr.waitAllSignatures()

	timerBetweenStatusChecks := time.NewTimer(timeBetweenSignaturesChecks)

	remainingSRTime := sr.remainingTime()
	timeout := time.NewTimer(remainingSRTime)
	for {
		select {
		case <-timerBetweenStatusChecks.C:
			if sr.IsSubroundFinished(sr.Current()) {
				log.Trace("subround already finished", "subround", sr.Name())
				return false
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

func (sr *subroundEndRound) waitAllSignatures() {
	remainingTime := sr.remainingTime()
	time.Sleep(remainingTime)

	if sr.IsSubroundFinished(sr.Current()) {
		return
	}

	sr.WaitingAllSignaturesTimeOut = true

	select {
	case sr.ConsensusChannel() <- true:
	default:
	}
}

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
	// TODO[cleanup cns finality]: remove this check
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, sr.Header.GetEpoch()) {
		return true
	}

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

	err = sr.SetJobDone(node, SrSignature, true)
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
	threshold := sr.Threshold(SrSignature)
	if sr.FallbackHeaderValidator().ShouldApplyFallbackValidation(sr.Header) {
		threshold = sr.FallbackThreshold(SrSignature)
		log.Warn("subroundEndRound.checkReceivedSignatures: fallback validation has been applied",
			"minimum number of signatures required", threshold,
			"actual number of signatures received", sr.getNumOfSignaturesCollected(),
		)
	}

	areSignaturesCollected, numSigs := sr.areSignaturesCollected(threshold)
	areAllSignaturesCollected := numSigs == sr.ConsensusGroupSize()

	isSignatureCollectionDone := areAllSignaturesCollected || (areSignaturesCollected && sr.WaitingAllSignaturesTimeOut)

	selfJobDone := true
	if sr.IsNodeInConsensusGroup(sr.SelfPubKey()) {
		selfJobDone = sr.IsSelfJobDone(SrSignature)
	}
	multiKeyJobDone := true
	if sr.IsMultiKeyInConsensusGroup() {
		multiKeyJobDone = sr.IsMultiKeyJobDone(SrSignature)
	}

	hasProof := sr.worker.HasEquivalentMessage(sr.Data)
	shouldStopWaitingSignatures := selfJobDone && multiKeyJobDone && isSignatureCollectionDone
	if shouldStopWaitingSignatures || hasProof {
		log.Debug("step 2: signatures collection done or proof already received",
			"subround", sr.Name(),
			"signatures received", numSigs,
			"total signatures", len(sr.ConsensusGroup()),
			"has proof", hasProof)

		return true
	}

	return false
}

func (sr *subroundEndRound) getNumOfSignaturesCollected() int {
	n := 0

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		node := sr.ConsensusGroup()[i]

		isSignJobDone, err := sr.JobDone(node, SrSignature)
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
