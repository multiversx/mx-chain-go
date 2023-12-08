package bls

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

// maxAllowedSizeInBytes defines how many bytes are allowed as payload in a message
const maxAllowedSizeInBytes = uint32(core.MegabyteSize * 95 / 100)

// subroundBlock defines the data needed by the subround Block
type subroundBlock struct {
	*spos.Subround

	processingThresholdPercentage int
}

// NewSubroundBlock creates a subroundBlock object
func NewSubroundBlock(
	baseSubround *spos.Subround,
	extend func(subroundId int),
	processingThresholdPercentage int,
) (*subroundBlock, error) {
	err := checkNewSubroundBlockParams(baseSubround)
	if err != nil {
		return nil, err
	}

	srBlock := subroundBlock{
		Subround:                      baseSubround,
		processingThresholdPercentage: processingThresholdPercentage,
	}

	srBlock.Job = srBlock.doBlockJob
	srBlock.Check = srBlock.doBlockConsensusCheck
	srBlock.Extend = extend

	return &srBlock, nil
}

func checkNewSubroundBlockParams(
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

// doBlockJob method does the job of the subround Block
func (sr *subroundBlock) doBlockJob(ctx context.Context) bool {
	if !sr.IsSelfLeaderInCurrentRound() && !sr.IsMultiKeyLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if sr.RoundHandler().Index() <= sr.getRoundInLastCommittedBlock() {
		return false
	}

	if sr.IsLeaderJobDone(sr.Current()) {
		return false
	}

	if sr.IsSubroundFinished(sr.Current()) {
		return false
	}

	metricStatTime := time.Now()
	defer sr.computeSubroundProcessingMetric(metricStatTime, common.MetricCreatedProposedBlock)

	header, err := sr.createHeader()
	if err != nil {
		printLogMessage(ctx, "doBlockJob.createHeader", err)
		return false
	}

	header, body, err := sr.createBlock(header)
	if err != nil {
		printLogMessage(ctx, "doBlockJob.createBlock", err)
		return false
	}

	sentWithSuccess, signatureShare := sr.sendBlock(header, body)
	if !sentWithSuccess {
		return false
	}

	if sr.EnableEpochsHandler().IsConsensusPropagationChangesFlagEnabled() {
		return sr.processBlock(ctx, sr.RoundHandler().Index(), []byte(sr.SelfPubKey()), signatureShare)
	}

	// TODO[cleanup cns finality]: remove these lines once the above epoch will be active
	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		log.Debug("doBlockJob.GetLeader", "error", errGetLeader)
		return false
	}

	err = sr.SetJobDone(leader, sr.Current(), true)
	if err != nil {
		log.Debug("doBlockJob.SetSelfJobDone", "error", err.Error())
		return false
	}

	// placeholder for subroundBlock.doBlockJob script

	sr.ConsensusCoreHandler.ScheduledProcessor().StartScheduledProcessing(header, body, sr.RoundTimeStamp)

	return true
}

func printLogMessage(ctx context.Context, baseMessage string, err error) {
	if common.IsContextDone(ctx) {
		log.Debug(baseMessage + " context is closing")
		return
	}

	log.Debug(baseMessage, "error", err.Error())
}

func (sr *subroundBlock) sendBlock(header data.HeaderHandler, body data.BodyHandler) (bool, []byte) {
	marshalizedBody, err := sr.Marshalizer().Marshal(body)
	if err != nil {
		log.Debug("sendBlock.Marshal: body", "error", err.Error())
		return false, nil
	}

	marshalizedHeader, err := sr.Marshalizer().Marshal(header)
	if err != nil {
		log.Debug("sendBlock.Marshal: header", "error", err.Error())
		return false, nil
	}

	var signatureShare []byte
	if sr.EnableEpochsHandler().IsConsensusPropagationChangesFlagEnabled() {
		selfIndex, err := sr.SelfConsensusGroupIndex()
		if err != nil {
			log.Debug("sendBlock.SelfConsensusGroupIndex: not in consensus group")
			return false, nil
		}

		headerHash := sr.Hasher().Compute(string(marshalizedHeader))
		signatureShare, err = sr.SigningHandler().CreateSignatureShareForPublicKey(
			headerHash,
			uint16(selfIndex),
			header.GetEpoch(),
			[]byte(sr.SelfPubKey()),
		)
		if err != nil {
			log.Debug("sendBlock.CreateSignatureShareForPublicKey", "error", err.Error())
			return false, nil
		}
	}

	if sr.couldBeSentTogether(marshalizedBody, marshalizedHeader) {
		return sr.sendHeaderAndBlockBody(header, body, marshalizedBody, marshalizedHeader, signatureShare)
	}

	if !sr.sendBlockBody(body, marshalizedBody) || !sr.sendBlockHeader(header, marshalizedHeader, signatureShare) {
		return false, nil
	}

	return true, signatureShare
}

func (sr *subroundBlock) couldBeSentTogether(marshalizedBody []byte, marshalizedHeader []byte) bool {
	bodyAndHeaderSize := uint32(len(marshalizedBody) + len(marshalizedHeader))
	log.Debug("couldBeSentTogether",
		"body size", len(marshalizedBody),
		"header size", len(marshalizedHeader),
		"body and header size", bodyAndHeaderSize,
		"max allowed size in bytes", maxAllowedSizeInBytes)
	return bodyAndHeaderSize <= maxAllowedSizeInBytes
}

func (sr *subroundBlock) createBlock(header data.HeaderHandler) (data.HeaderHandler, data.BodyHandler, error) {
	startTime := sr.RoundTimeStamp
	maxTime := time.Duration(sr.EndTime())
	haveTimeInCurrentSubround := func() bool {
		return sr.RoundHandler().RemainingTime(startTime, maxTime) > 0
	}

	finalHeader, blockBody, err := sr.BlockProcessor().CreateBlock(
		header,
		haveTimeInCurrentSubround,
	)
	if err != nil {
		return nil, nil, err
	}

	return finalHeader, blockBody, nil
}

// sendHeaderAndBlockBody method sends the proposed header and block body in the subround Block
func (sr *subroundBlock) sendHeaderAndBlockBody(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
	marshalizedBody []byte,
	marshalizedHeader []byte,
	signature []byte,
) (bool, []byte) {
	headerHash := sr.Hasher().Compute(string(marshalizedHeader))

	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		log.Debug("sendBlockBodyAndHeader.GetLeader", "error", errGetLeader)
		return false, nil
	}

	cnsMsg := consensus.NewConsensusMessage(
		headerHash,
		signature,
		marshalizedBody,
		marshalizedHeader,
		[]byte(leader),
		nil,
		int(MtBlockBodyAndHeader),
		sr.RoundHandler().Index(),
		sr.ChainID(),
		nil,
		nil,
		nil,
		sr.GetAssociatedPid([]byte(leader)),
		nil,
	)

	err := sr.BroadcastMessenger().BroadcastConsensusMessage(cnsMsg)
	if err != nil {
		log.Debug("sendHeaderAndBlockBody.BroadcastConsensusMessage", "error", err.Error())
		return false, nil
	}

	log.Debug("step 1: block body and header have been sent",
		"nonce", headerHandler.GetNonce(),
		"hash", headerHash)

	sr.Data = headerHash
	sr.Body = bodyHandler
	sr.Header = headerHandler

	return true, signature
}

// sendBlockBody method sends the proposed block body in the subround Block
func (sr *subroundBlock) sendBlockBody(
	bodyHandler data.BodyHandler,
	marshalizedBody []byte,
) bool {
	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		log.Debug("sendBlockBody.GetLeader", "error", errGetLeader)
		return false
	}

	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		marshalizedBody,
		nil,
		[]byte(leader),
		nil,
		int(MtBlockBody),
		sr.RoundHandler().Index(),
		sr.ChainID(),
		nil,
		nil,
		nil,
		sr.GetAssociatedPid([]byte(leader)),
		nil,
	)

	err := sr.BroadcastMessenger().BroadcastConsensusMessage(cnsMsg)
	if err != nil {
		log.Debug("sendBlockBody.BroadcastConsensusMessage", "error", err.Error())
		return false
	}

	log.Debug("step 1: block body has been sent")

	sr.Body = bodyHandler

	return true
}

// sendBlockHeader method sends the proposed block header in the subround Block
func (sr *subroundBlock) sendBlockHeader(
	headerHandler data.HeaderHandler,
	marshalizedHeader []byte,
	signature []byte,
) bool {
	headerHash := sr.Hasher().Compute(string(marshalizedHeader))

	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		log.Debug("sendBlockBody.GetLeader", "error", errGetLeader)
		return false
	}

	cnsMsg := consensus.NewConsensusMessage(
		headerHash,
		signature,
		nil,
		marshalizedHeader,
		[]byte(leader),
		nil,
		int(MtBlockHeader),
		sr.RoundHandler().Index(),
		sr.ChainID(),
		nil,
		nil,
		nil,
		sr.GetAssociatedPid([]byte(leader)),
		nil,
	)

	err := sr.BroadcastMessenger().BroadcastConsensusMessage(cnsMsg)
	if err != nil {
		log.Debug("sendBlockHeader.BroadcastConsensusMessage", "error", err.Error())
		return false
	}

	log.Debug("step 1: block header has been sent",
		"nonce", headerHandler.GetNonce(),
		"hash", headerHash)

	sr.Data = headerHash
	sr.Header = headerHandler

	return true
}

func (sr *subroundBlock) createHeader() (data.HeaderHandler, error) {
	var nonce uint64
	var prevHash []byte
	var prevRandSeed []byte

	currentHeader := sr.Blockchain().GetCurrentBlockHeader()
	if check.IfNil(currentHeader) {
		nonce = sr.Blockchain().GetGenesisHeader().GetNonce() + 1
		prevHash = sr.Blockchain().GetGenesisHeaderHash()
		prevRandSeed = sr.Blockchain().GetGenesisHeader().GetRandSeed()
	} else {
		nonce = currentHeader.GetNonce() + 1
		prevHash = sr.Blockchain().GetCurrentBlockHeaderHash()
		prevRandSeed = currentHeader.GetRandSeed()
	}

	round := uint64(sr.RoundHandler().Index())
	hdr, err := sr.BlockProcessor().CreateNewHeader(round, nonce)
	if err != nil {
		return nil, err
	}

	err = hdr.SetPrevHash(prevHash)
	if err != nil {
		return nil, err
	}

	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		return nil, errGetLeader
	}

	randSeed, err := sr.SigningHandler().CreateSignatureForPublicKey(prevRandSeed, []byte(leader))
	if err != nil {
		return nil, err
	}

	err = hdr.SetShardID(sr.ShardCoordinator().SelfId())
	if err != nil {
		return nil, err
	}

	err = hdr.SetTimeStamp(uint64(sr.RoundHandler().TimeStamp().Unix()))
	if err != nil {
		return nil, err
	}

	err = hdr.SetPrevRandSeed(prevRandSeed)
	if err != nil {
		return nil, err
	}

	err = hdr.SetRandSeed(randSeed)
	if err != nil {
		return nil, err
	}

	err = hdr.SetChainID(sr.ChainID())
	if err != nil {
		return nil, err
	}

	return hdr, nil
}

// receivedBlockBodyAndHeader method is called when a block body and a block header is received
func (sr *subroundBlock) receivedBlockBodyAndHeader(ctx context.Context, cnsDta *consensus.Message) bool {
	sw := core.NewStopWatch()
	sw.Start("receivedBlockBodyAndHeader")

	defer func() {
		sw.Stop("receivedBlockBodyAndHeader")
		log.Debug("time measurements of receivedBlockBodyAndHeader", sw.GetMeasurements()...)
	}()

	node := string(cnsDta.PubKey)

	if sr.IsConsensusDataSet() {
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

	if sr.IsBlockBodyAlreadyReceived() {
		return false
	}

	if sr.IsHeaderAlreadyReceived() {
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.RoundHandler().Index(), sr.Current()) {
		return false
	}

	if sr.EnableEpochsHandler().IsConsensusPropagationChangesFlagEnabled() {
		err := sr.SigningHandler().VerifySingleSignature(cnsDta.PubKey, cnsDta.BlockHeaderHash, cnsDta.SignatureShare)
		if err != nil {
			log.Debug("VerifySingleSignature: confirmed that node provided invalid signature",
				"pubKey", cnsDta.PubKey,
				"blockHeaderHash", cnsDta.BlockHeaderHash,
				"error", err.Error(),
			)
			return false
		}
	}

	sr.Data = cnsDta.BlockHeaderHash
	sr.Body = sr.BlockProcessor().DecodeBlockBody(cnsDta.Body)
	sr.Header = sr.BlockProcessor().DecodeBlockHeader(cnsDta.Header)

	isInvalidData := check.IfNil(sr.Body) || sr.isInvalidHeaderOrData()
	if isInvalidData {
		return false
	}

	log.Debug("step 1: block body and header have been received",
		"nonce", sr.Header.GetNonce(),
		"hash", cnsDta.BlockHeaderHash)

	sw.Start("processReceivedBlock")
	blockProcessedWithSuccess := sr.processReceivedBlock(ctx, cnsDta)
	sw.Stop("processReceivedBlock")

	sr.PeerHonestyHandler().ChangeScore(
		node,
		spos.GetConsensusTopicID(sr.ShardCoordinator()),
		spos.LeaderPeerHonestyIncreaseFactor,
	)

	return blockProcessedWithSuccess
}

func (sr *subroundBlock) saveLeaderSignature(nodeKey []byte, signature []byte) bool {
	if len(signature) == 0 {
		return true
	}

	node := string(nodeKey)
	pkForLogs := core.GetTrimmedPk(hex.EncodeToString(nodeKey))

	index, err := sr.ConsensusGroupIndex(node)
	if err != nil {
		log.Debug("saveLeaderSignature.ConsensusGroupIndex",
			"node", pkForLogs,
			"error", err.Error())
		return false
	}

	err = sr.SigningHandler().StoreSignatureShare(uint16(index), signature)
	if err != nil {
		log.Debug("saveLeaderSignature.StoreSignatureShare",
			"node", pkForLogs,
			"index", index,
			"error", err.Error())
		return false
	}

	return true
}

func (sr *subroundBlock) isInvalidHeaderOrData() bool {
	return sr.Data == nil || check.IfNil(sr.Header) || sr.Header.CheckFieldsForNil() != nil
}

// receivedBlockBody method is called when a block body is received through the block body channel
func (sr *subroundBlock) receivedBlockBody(ctx context.Context, cnsDta *consensus.Message) bool {
	node := string(cnsDta.PubKey)

	if !sr.IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		sr.PeerHonestyHandler().ChangeScore(
			node,
			spos.GetConsensusTopicID(sr.ShardCoordinator()),
			spos.LeaderPeerHonestyDecreaseFactor,
		)

		return false
	}

	if sr.IsBlockBodyAlreadyReceived() {
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.RoundHandler().Index(), sr.Current()) {
		return false
	}

	sr.Body = sr.BlockProcessor().DecodeBlockBody(cnsDta.Body)

	if check.IfNil(sr.Body) {
		return false
	}

	log.Debug("step 1: block body has been received")

	blockProcessedWithSuccess := sr.processReceivedBlock(ctx, cnsDta)

	sr.PeerHonestyHandler().ChangeScore(
		node,
		spos.GetConsensusTopicID(sr.ShardCoordinator()),
		spos.LeaderPeerHonestyIncreaseFactor,
	)

	return blockProcessedWithSuccess
}

// receivedBlockHeader method is called when a block header is received through the block header channel.
// If the block header is valid, then the validatorRoundStates map corresponding to the node which sent it,
// is set on true for the subround Block
func (sr *subroundBlock) receivedBlockHeader(ctx context.Context, cnsDta *consensus.Message) bool {
	node := string(cnsDta.PubKey)

	if sr.IsConsensusDataSet() {
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

	if sr.IsHeaderAlreadyReceived() {
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.RoundHandler().Index(), sr.Current()) {
		return false
	}

	sr.Data = cnsDta.BlockHeaderHash
	sr.Header = sr.BlockProcessor().DecodeBlockHeader(cnsDta.Header)

	if sr.isInvalidHeaderOrData() {
		return false
	}

	if sr.EnableEpochsHandler().IsConsensusPropagationChangesFlagEnabled() {
		err := sr.SigningHandler().VerifySingleSignature(cnsDta.PubKey, cnsDta.BlockHeaderHash, cnsDta.SignatureShare)
		if err != nil {
			log.Debug("VerifySingleSignature: confirmed that node provided invalid signature",
				"pubKey", cnsDta.PubKey,
				"blockHeaderHash", cnsDta.BlockHeaderHash,
				"error", err.Error(),
			)
			return false
		}
	}

	log.Debug("step 1: block header has been received",
		"nonce", sr.Header.GetNonce(),
		"hash", cnsDta.BlockHeaderHash)
	blockProcessedWithSuccess := sr.processReceivedBlock(ctx, cnsDta)

	sr.PeerHonestyHandler().ChangeScore(
		node,
		spos.GetConsensusTopicID(sr.ShardCoordinator()),
		spos.LeaderPeerHonestyIncreaseFactor,
	)

	return blockProcessedWithSuccess
}

func (sr *subroundBlock) processReceivedBlock(ctx context.Context, cnsDta *consensus.Message) bool {
	if check.IfNil(sr.Body) {
		return false
	}
	if check.IfNil(sr.Header) {
		return false
	}

	defer func() {
		sr.SetProcessingBlock(false)
	}()

	sr.SetProcessingBlock(true)

	shouldNotProcessBlock := sr.ExtendedCalled || cnsDta.RoundIndex < sr.RoundHandler().Index()
	if shouldNotProcessBlock {
		log.Debug("canceled round, extended has been called or round index has been changed",
			"round", sr.RoundHandler().Index(),
			"subround", sr.Name(),
			"cnsDta round", cnsDta.RoundIndex,
			"extended called", sr.ExtendedCalled,
		)
		return false
	}

	return sr.processBlock(ctx, cnsDta.RoundIndex, cnsDta.PubKey, cnsDta.SignatureShare)
}

func (sr *subroundBlock) processBlock(
	ctx context.Context,
	roundIndex int64,
	pubkey []byte,
	signature []byte,
) bool {
	startTime := sr.RoundTimeStamp
	maxTime := sr.RoundHandler().TimeDuration() * time.Duration(sr.processingThresholdPercentage) / 100
	remainingTimeInCurrentRound := func() time.Duration {
		return sr.RoundHandler().RemainingTime(startTime, maxTime)
	}

	metricStatTime := time.Now()
	defer sr.computeSubroundProcessingMetric(metricStatTime, common.MetricProcessedProposedBlock)

	err := sr.BlockProcessor().ProcessBlock(
		sr.Header,
		sr.Body,
		remainingTimeInCurrentRound,
	)

	if roundIndex < sr.RoundHandler().Index() {
		log.Debug("canceled round, round index has been changed",
			"round", sr.RoundHandler().Index(),
			"subround", sr.Name(),
			"cnsDta round", roundIndex,
		)
		return false
	}

	if err != nil {
		sr.printCancelRoundLogMessage(ctx, err)
		sr.RoundCanceled = true

		return false
	}

	ok := sr.saveLeaderSignature(pubkey, signature)
	if !ok {
		sr.printCancelRoundLogMessage(ctx, err)
		return false
	}

	node := string(pubkey)
	err = sr.SetJobDone(node, sr.Current(), true)
	if err != nil {
		sr.printCancelRoundLogMessage(ctx, err)
		return false
	}

	sr.ConsensusCoreHandler.ScheduledProcessor().StartScheduledProcessing(sr.Header, sr.Body, sr.RoundTimeStamp)

	return true
}

func (sr *subroundBlock) printCancelRoundLogMessage(ctx context.Context, err error) {
	if common.IsContextDone(ctx) {
		log.Debug("canceled round as the context is closing")
		return
	}

	log.Debug("canceled round",
		"round", sr.RoundHandler().Index(),
		"subround", sr.Name(),
		"error", err.Error())
}

func (sr *subroundBlock) computeSubroundProcessingMetric(startTime time.Time, metric string) {
	subRoundDuration := sr.EndTime() - sr.StartTime()
	if subRoundDuration == 0 {
		// can not do division by 0
		return
	}

	percent := uint64(time.Since(startTime)) * 100 / uint64(subRoundDuration)
	sr.AppStatusHandler().SetUInt64Value(metric, percent)
}

// doBlockConsensusCheck method checks if the consensus in the subround Block is achieved
func (sr *subroundBlock) doBlockConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.IsSubroundFinished(sr.Current()) {
		return true
	}

	threshold := sr.Threshold(sr.Current())
	if sr.isBlockReceived(threshold) {
		log.Debug("step 1: subround has been finished",
			"subround", sr.Name())
		sr.SetStatus(sr.Current(), spos.SsFinished)
		return true
	}

	return false
}

// isBlockReceived method checks if the block was received from the leader in the current round
func (sr *subroundBlock) isBlockReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		node := sr.ConsensusGroup()[i]
		isJobDone, err := sr.JobDone(node, sr.Current())
		if err != nil {
			log.Debug("isBlockReceived.JobDone",
				"node", node,
				"subround", sr.Name(),
				"error", err.Error())
			continue
		}

		if isJobDone {
			n++
		}
	}

	return n >= threshold
}

func (sr *subroundBlock) getRoundInLastCommittedBlock() int64 {
	roundInLastCommittedBlock := int64(0)
	currentHeader := sr.Blockchain().GetCurrentBlockHeader()
	if !check.IfNil(currentHeader) {
		roundInLastCommittedBlock = int64(currentHeader.GetRound())
	}

	return roundInLastCommittedBlock
}

// IsInterfaceNil returns true if there is no value under the interface
func (sr *subroundBlock) IsInterfaceNil() bool {
	return sr == nil
}
