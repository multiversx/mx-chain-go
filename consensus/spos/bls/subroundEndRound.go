package bls

import (
	"bytes"
	"context"
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
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/p2p"
)

type aggregatedSigsResult struct {
	bitmap              []byte
	aggregatedSig       []byte
	extraAggregatedSigs map[string][]byte
}

type subroundEndRound struct {
	*spos.Subround
	processingThresholdPercentage int
	displayStatistics             func()
	mutProcessingEndRound         sync.Mutex
	getMessageToVerifySigFunc     func() []byte

	extraSignersHolder SubRoundEndExtraSignersHolder
}

// NewSubroundEndRound creates a subroundEndRound object
func NewSubroundEndRound(
	baseSubround *spos.Subround,
	extend func(subroundId int),
	processingThresholdPercentage int,
	displayStatistics func(),
	extraSignersHolder SubRoundEndExtraSignersHolder,
) (*subroundEndRound, error) {
	err := checkNewSubroundEndRoundParams(
		baseSubround,
	)
	if err != nil {
		return nil, err
	}
	if check.IfNil(extraSignersHolder) {
		return nil, errors.ErrNilEndRoundExtraSignersHolder
	}

	srEndRound := subroundEndRound{
		baseSubround,
		processingThresholdPercentage,
		displayStatistics,
		sync.Mutex{},
		nil,
		extraSignersHolder,
	}
	srEndRound.Job = srEndRound.doEndRoundJob
	srEndRound.Check = srEndRound.doEndRoundConsensusCheck
	srEndRound.Extend = extend
	srEndRound.getMessageToVerifySigFunc = srEndRound.getMessageToVerifySig

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
	node := string(cnsDta.PubKey)

	if !sr.IsConsensusDataSet() {
		return false
	}

	if !sr.IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		sr.PeerHonestyHandler().ChangeScore(
			node,
			spos.GetConsensusTopicID(sr.ShardCoordinator()),
			spos.LeaderPeerHonestyDecreaseFactor,
		)

		return false
	}

	if sr.IsSelfLeaderInCurrentRound() || sr.IsMultiKeyLeaderInCurrentRound() {
		return false
	}

	if !sr.IsConsensusDataEqual(cnsDta.HeaderHash) {
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.RoundHandler().Index(), sr.Current()) {
		return false
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
	if check.IfNil(sr.Header) {
		return false
	}

	header := sr.Header.ShallowClone()
	err := header.SetPubKeysBitmap(cnsDta.PubKeysBitmap)
	if err != nil {
		log.Debug("isBlockHeaderFinalInfoValid.SetPubKeysBitmap", "error", err.Error())
		return false
	}

	err = header.SetSignature(cnsDta.AggregateSignature)
	if err != nil {
		log.Debug("isBlockHeaderFinalInfoValid.SetSignature", "error", err.Error())
		return false
	}

	err = header.SetLeaderSignature(cnsDta.LeaderSignature)
	if err != nil {
		log.Debug("isBlockHeaderFinalInfoValid.SetLeaderSignature", "error", err.Error())
		return false
	}

	err = sr.extraSignersHolder.HaveConsensusHeaderWithFullInfo(header, cnsDta)
	if err != nil {
		log.Debug("isBlockHeaderFinalInfoValid.extraSignersHolder.haveConsensusHeaderWithFullInfo", "error", err.Error())
		return false
	}

	err = sr.HeaderSigVerifier().VerifyLeaderSignature(header)
	if err != nil {
		log.Debug("isBlockHeaderFinalInfoValid.VerifyLeaderSignature", "error", err.Error())
		return false
	}

	err = sr.HeaderSigVerifier().VerifySignature(header)
	if err != nil {
		log.Debug("isBlockHeaderFinalInfoValid.VerifySignature", "error", err.Error())
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

	if !sr.IsNodeLeaderInCurrentRound(messageSender) { // is NOT this node leader in current round?
		sr.PeerHonestyHandler().ChangeScore(
			messageSender,
			spos.GetConsensusTopicID(sr.ShardCoordinator()),
			spos.LeaderPeerHonestyDecreaseFactor,
		)

		return false
	}

	if sr.IsSelfLeaderInCurrentRound() {
		return false
	}

	if !sr.IsConsensusDataEqual(cnsDta.HeaderHash) {
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

	headerHash := sr.getHeaderHashToVerifySig(cnsMsg)
	err = sr.SigningHandler().VerifySingleSignature(cnsMsg.PubKey, headerHash, cnsMsg.SignatureShare)
	if err != nil {
		log.Debug("verifyInvalidSigner: confirmed that node provided invalid signature",
			"pubKey", cnsMsg.PubKey,
			"headerHash", headerHash,
			"error", err.Error(),
		)
		sr.applyBlacklistOnNode(msg.Peer())
	}

	return nil
}

func (sr *subroundEndRound) getHeaderHashToVerifySig(cnsMsg *consensus.Message) []byte {
	if sr.EnableEpochHandler().IsConsensusModelV2Enabled() {
		return cnsMsg.ProcessedHeaderHash
	}

	return cnsMsg.HeaderHash
}

func (sr *subroundEndRound) applyBlacklistOnNode(peer core.PeerID) {
	sr.PeerBlacklistHandler().BlacklistPeer(peer, common.InvalidSigningBlacklistDuration)
}

func (sr *subroundEndRound) receivedHeader(headerHandler data.HeaderHandler) {
	if sr.ConsensusGroup() == nil || sr.IsSelfLeaderInCurrentRound() || sr.IsMultiKeyLeaderInCurrentRound() {
		return
	}

	sr.AddReceivedHeader(headerHandler)

	sr.doEndRoundJobByParticipant(nil)
}

// doEndRoundJob method does the job of the subround EndRound
func (sr *subroundEndRound) doEndRoundJob(_ context.Context) bool {
	if !sr.IsSelfLeaderInCurrentRound() && !sr.IsMultiKeyLeaderInCurrentRound() {
		if sr.IsNodeInConsensusGroup(sr.SelfPubKey()) || sr.IsMultiKeyInConsensusGroup() {
			err := sr.prepareBroadcastBlockDataForValidator()
			if err != nil {
				log.Warn("validator in consensus group preparing for delayed broadcast",
					"error", err.Error())
			}
		}

		return sr.doEndRoundJobByParticipant(nil)
	}

	return sr.doEndRoundJobByLeader()
}

func (sr *subroundEndRound) doEndRoundJobByLeader() bool {
	bitmap := sr.generateBitmap()
	err := sr.checkSignaturesValidity(bitmap)
	if err != nil {
		log.Debug("doEndRoundJobByLeader.checkSignaturesValidity", "error", err.Error())
		return false
	}

	if check.IfNil(sr.Header) {
		log.Error("doEndRoundJobByLeader.CheckNilHeader", "error", spos.ErrNilHeader)
		return false
	}

	// Aggregate sig and add it to the block
	aggSigsRes, err := sr.aggregateSigsAndHandleInvalidSigners(bitmap)
	if err != nil {
		log.Debug("doEndRoundJobByLeader.aggregateSigsAndHandleInvalidSigners", "error", err.Error())
		return false
	}

	bitmap = aggSigsRes.bitmap
	err = sr.Header.SetPubKeysBitmap(bitmap)
	if err != nil {
		log.Debug("doEndRoundJobByLeader.SetPubKeysBitmap", "error", err.Error())
		return false
	}

	err = sr.Header.SetSignature(aggSigsRes.aggregatedSig)
	if err != nil {
		log.Debug("doEndRoundJobByLeader.SetSignature", "error", err.Error())
		return false
	}

	err = sr.extraSignersHolder.SetAggregatedSignatureInHeader(sr.Header, aggSigsRes.extraAggregatedSigs)
	if err != nil {
		return false
	}

	// Header is complete so the leader can sign it
	leaderPubKey, leaderSignature, err := sr.signBlockHeader()
	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.Header.SetLeaderSignature(leaderSignature)
	if err != nil {
		log.Debug("doEndRoundJobByLeader.SetLeaderSignature", "error", err.Error())
		return false
	}

	err = sr.extraSignersHolder.SignAndSetLeaderSignature(sr.Header, leaderPubKey)
	if err != nil {
		log.Debug("doEndRoundJobByLeader.extraSignatureAggregator.SignAndSetLeaderSignature", "error", err.Error())
		return false
	}

	ok := sr.ScheduledProcessor().IsProcessedOKWithTimeout()
	// placeholder for subroundEndRound.doEndRoundJobByLeader script
	if !ok {
		return false
	}

	roundHandler := sr.RoundHandler()
	if roundHandler.RemainingTime(roundHandler.TimeStamp(), roundHandler.TimeDuration()) < 0 {
		log.Debug("doEndRoundJob: time is out -> cancel broadcasting final info and header",
			"round time stamp", roundHandler.TimeStamp(),
			"current time", time.Now())
		return false
	}

	// broadcast header and final info section

	sr.createAndBroadcastHeaderFinalInfo()

	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		log.Debug("doEndRoundJobByLeader.GetLeader", "error", errGetLeader)
		return false
	}

	// broadcast header
	err = sr.BroadcastMessenger().BroadcastHeader(sr.Header, []byte(leader))
	if err != nil {
		log.Debug("doEndRoundJobByLeader.BroadcastHeader", "error", err.Error())
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

	sr.displayStatistics()

	log.Debug("step 3: Body and Header have been committed and header has been broadcast")

	err = sr.broadcastBlockDataLeader()
	if err != nil {
		log.Debug("doEndRoundJobByLeader.broadcastBlockDataLeader", "error", err.Error())
	}

	msg := fmt.Sprintf("Added proposed block with nonce  %d  in blockchain", sr.Header.GetNonce())
	log.Debug(display.Headline(msg, sr.SyncTimer().FormattedCurrentTime(), "+"))

	sr.updateMetricsForLeader()

	return true
}

func (sr *subroundEndRound) aggregateSigsAndHandleInvalidSigners(bitmap []byte) (*aggregatedSigsResult, error) {
	sig, err := sr.SigningHandler().AggregateSigs(bitmap, sr.Header.GetEpoch())
	if err != nil {
		log.Debug("doEndRoundJobByLeader.AggregateSigs", "error", err.Error())

		return sr.handleInvalidSignersOnAggSigFail()
	}

	// placeholder for AggregateSignatures
	extraSigs, err := sr.extraSignersHolder.AggregateSignatures(bitmap, sr.Header.GetEpoch())
	if err != nil {
		log.Debug("doEndRoundJobByLeader.extraAggregatedSig.AggregateSignatures", "error", err.Error())

		// TODO :MariusC. Here we should add behavior to handle invalid sigs on outgoing operations
		return nil, err
	}

	err = sr.SigningHandler().SetAggregatedSig(sig)
	if err != nil {
		log.Debug("doEndRoundJobByLeader.SetAggregatedSig", "error", err.Error())
		return nil, err
	}

	err = sr.SigningHandler().Verify(sr.getMessageToVerifySigFunc(), bitmap, sr.Header.GetEpoch())
	if err != nil {
		log.Debug("doEndRoundJobByLeader.Verify", "error", err.Error())

		return sr.handleInvalidSignersOnAggSigFail()
	}

	err = sr.extraSignersHolder.VerifyAggregatedSignatures(sr.Header, bitmap)
	if err != nil {
		log.Debug("doEndRoundJobByLeader.extraSignersHolder.verifyAggregatedSignatures", "error", err.Error())
		// TODO: MariusC. Here we should add behavior to handle invalid sigs on outgoing operations
		return nil, err
	}

	return &aggregatedSigsResult{
		bitmap:              bitmap,
		aggregatedSig:       sig,
		extraAggregatedSigs: extraSigs,
	}, nil
}

func (sr *subroundEndRound) verifyNodesOnAggSigFail() ([]string, error) {
	invalidPubKeys := make([]string, 0)
	pubKeys := sr.ConsensusGroup()
	msg := sr.getMessageToVerifySigFunc()

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
		err = sr.SigningHandler().VerifySignatureShare(uint16(i), sigShare, msg, sr.Header.GetEpoch())
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

func (sr *subroundEndRound) handleInvalidSignersOnAggSigFail() (*aggregatedSigsResult, error) {
	invalidPubKeys, err := sr.verifyNodesOnAggSigFail()
	if err != nil {
		log.Debug("doEndRoundJobByLeader.verifyNodesOnAggSigFail", "error", err.Error())
		return nil, err
	}

	invalidSigners, err := sr.getFullMessagesForInvalidSigners(invalidPubKeys)
	if err != nil {
		log.Debug("doEndRoundJobByLeader.getFullMessagesForInvalidSigners", "error", err.Error())
		return nil, err
	}

	if len(invalidSigners) > 0 {
		sr.createAndBroadcastInvalidSigners(invalidSigners)
	}

	bitmap, sig, err := sr.computeAggSigOnValidNodes()
	if err != nil {
		log.Debug("doEndRoundJobByLeader.computeAggSigOnValidNodes", "error", err.Error())
		return nil, err
	}

	return &aggregatedSigsResult{
		bitmap:              bitmap,
		aggregatedSig:       sig,
		extraAggregatedSigs: nil,
	}, nil
}

func (sr *subroundEndRound) computeAggSigOnValidNodes() ([]byte, []byte, error) {
	threshold := sr.Threshold(sr.Current())
	numValidSigShares := sr.ComputeSize(SrSignature)

	if check.IfNil(sr.Header) {
		return nil, nil, spos.ErrNilHeader
	}

	if numValidSigShares < threshold {
		return nil, nil, fmt.Errorf("%w: number of valid sig shares lower than threshold, numSigShares: %d, threshold: %d",
			spos.ErrInvalidNumSigShares, numValidSigShares, threshold)
	}

	bitmap := sr.generateBitmap()
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

func (sr *subroundEndRound) generateBitmap() []byte {
	if sr.EnableEpochHandler().IsConsensusModelV2Enabled() {
		processedHeaderHash := sr.getMessageToVerifySigFunc()
		return sr.GenerateBitmapForHash(SrSignature, processedHeaderHash)
	}

	return sr.GenerateBitmap(SrSignature)
}

func (sr *subroundEndRound) createAndBroadcastHeaderFinalInfo() {
	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		log.Debug("createAndBroadcastHeaderFinalInfo.GetLeader", "error", errGetLeader)
		return
	}

	cnsMsg := consensus.NewConsensusMessage(
		sr.GetData(),
		nil,
		nil,
		nil,
		[]byte(leader),
		nil,
		int(MtBlockHeaderFinalInfo),
		sr.RoundHandler().Index(),
		sr.ChainID(),
		sr.Header.GetPubKeysBitmap(),
		sr.Header.GetSignature(),
		sr.Header.GetLeaderSignature(),
		sr.GetAssociatedPid([]byte(leader)),
		nil,
		sr.getProcessedHeaderHash(),
	)

	// placeholder for AddAggregatedSignature
	err := sr.extraSignersHolder.AddLeaderAndAggregatedSignatures(sr.Header, cnsMsg)
	if err != nil {
		log.Debug("doEndRoundJob.extraSignatureAggregator.AddLeaderAndAggregatedSignatures", "error", err.Error())
		return
	}

	err = sr.BroadcastMessenger().BroadcastConsensusMessage(cnsMsg)
	if err != nil {
		log.Debug("doEndRoundJob.BroadcastConsensusMessage", "error", err.Error())
		return
	}

	log.Debug("step 3: block header final info has been sent",
		"PubKeysBitmap", sr.Header.GetPubKeysBitmap(),
		"AggregateSignature", sr.Header.GetSignature(),
		"LeaderSignature", sr.Header.GetLeaderSignature())
}

func (sr *subroundEndRound) createAndBroadcastInvalidSigners(invalidSigners []byte) {
	cnsMsg := consensus.NewConsensusMessage(
		sr.GetData(),
		nil,
		nil,
		nil,
		[]byte(sr.SelfPubKey()),
		nil,
		int(MtInvalidSigners),
		sr.RoundHandler().Index(),
		sr.ChainID(),
		nil,
		nil,
		nil,
		sr.CurrentPid(),
		invalidSigners,
		sr.getProcessedHeaderHash(),
	)

	err := sr.BroadcastMessenger().BroadcastConsensusMessage(cnsMsg)
	if err != nil {
		log.Debug("doEndRoundJob.BroadcastConsensusMessage", "error", err.Error())
		return
	}

	log.Debug("step 3: invalid signers info has been sent")
}

func (sr *subroundEndRound) getProcessedHeaderHash() []byte {
	if sr.EnableEpochHandler().IsConsensusModelV2Enabled() {
		return sr.getMessageToVerifySigFunc()
	}

	return nil
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

	sr.SetStatus(sr.Current(), spos.SsFinished)

	if sr.IsNodeInConsensusGroup(sr.SelfPubKey()) || sr.IsMultiKeyInConsensusGroup() {
		err = sr.setHeaderForValidator(header)
		if err != nil {
			log.Warn("doEndRoundJobByParticipant", "error", err.Error())
		}
	}

	sr.displayStatistics()

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

	err = sr.extraSignersHolder.HaveConsensusHeaderWithFullInfo(header, cnsDta)
	if err != nil {
		return false, nil
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
		receivedHeader := receivedHeaders[index].ShallowClone()
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

func (sr *subroundEndRound) signBlockHeader() ([]byte, []byte, error) {
	headerClone := sr.Header.ShallowClone()
	err := headerClone.SetLeaderSignature(nil)
	if err != nil {
		return nil, nil, err
	}

	marshalizedHdr, err := sr.Marshalizer().Marshal(headerClone)
	if err != nil {
		return nil, nil, err
	}

	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		return nil, nil, errGetLeader
	}

	leaderPubKey := []byte(leader)
	leaderSignature, err := sr.SigningHandler().CreateSignatureForPublicKey(marshalizedHdr, leaderPubKey)
	if err != nil {
		return nil, nil, err
	}

	return leaderPubKey, leaderSignature, nil
}

func (sr *subroundEndRound) updateMetricsForLeader() {
	sr.AppStatusHandler().Increment(common.MetricCountAcceptedBlocks)
	sr.AppStatusHandler().SetStringValue(common.MetricConsensusRoundState,
		fmt.Sprintf("valid block produced in %f sec", time.Since(sr.RoundHandler().TimeStamp()).Seconds()))
}

func (sr *subroundEndRound) broadcastBlockDataLeader() error {
	miniBlocks, transactions, err := sr.BlockProcessor().MarshalizedDataToBroadcast(sr.Header, sr.Body)
	if err != nil {
		return err
	}

	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		log.Debug("broadcastBlockDataLeader.GetLeader", "error", errGetLeader)
		return errGetLeader
	}

	return sr.BroadcastMessenger().BroadcastBlockDataLeader(sr.Header, miniBlocks, transactions, []byte(leader))
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

	if sr.IsSubroundFinished(sr.Current()) {
		return true
	}

	return false
}

func (sr *subroundEndRound) checkSignaturesValidity(bitmap []byte) error {
	nbBitsBitmap := len(bitmap) * 8
	consensusGroup := sr.ConsensusGroup()
	consensusGroupSize := len(consensusGroup)
	size := consensusGroupSize

	if consensusGroupSize > nbBitsBitmap {
		size = nbBitsBitmap
	}

	for i := 0; i < size; i++ {
		indexRequired := (bitmap[i/8] & (1 << uint16(i%8))) > 0
		if !indexRequired {
			continue
		}

		pubKey := consensusGroup[i]
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

// IsInterfaceNil returns true if there is no value under the interface
func (sr *subroundEndRound) IsInterfaceNil() bool {
	return sr == nil
}

func (sr *subroundEndRound) getMessageToVerifySig() []byte {
	return sr.GetData()
}
