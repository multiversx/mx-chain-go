package v2

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
)

// maxAllowedSizeInBytes defines how many bytes are allowed as payload in a message
const maxAllowedSizeInBytes = uint32(core.MegabyteSize * 95 / 100)

// subroundBlock defines the data needed by the subround Block
type subroundBlock struct {
	*spos.Subround

	processingThresholdPercentage int
	worker                        spos.WorkerHandler
}

// NewSubroundBlock creates a subroundBlock object
func NewSubroundBlock(
	baseSubround *spos.Subround,
	processingThresholdPercentage int,
	worker spos.WorkerHandler,
) (*subroundBlock, error) {
	err := checkNewSubroundBlockParams(baseSubround)
	if err != nil {
		return nil, err
	}

	if check.IfNil(worker) {
		return nil, spos.ErrNilWorker
	}

	srBlock := subroundBlock{
		Subround:                      baseSubround,
		processingThresholdPercentage: processingThresholdPercentage,
		worker:                        worker,
	}

	srBlock.Job = srBlock.doBlockJob
	srBlock.Check = srBlock.doBlockConsensusCheck
	srBlock.Extend = srBlock.worker.Extend

	return &srBlock, nil
}

func checkNewSubroundBlockParams(
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

// doBlockJob method does the job of the subround Block
func (sr *subroundBlock) doBlockJob(ctx context.Context) bool {
	isSelfLeader := sr.IsSelfLeader()
	if !isSelfLeader { // is NOT self leader in this round?
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

	// This must be done after createBlock, in order to have the proper epoch set
	wasSetProofOk := sr.addProofOnHeader(header)
	if !wasSetProofOk {
		return false
	}

	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		log.Debug("doBlockJob.GetLeader", "error", errGetLeader)
		return false
	}

	sentWithSuccess := sr.sendBlock(header, body, leader)
	if !sentWithSuccess {
		return false
	}

	err = sr.SetJobDone(leader, sr.Current(), true)
	if err != nil {
		log.Debug("doBlockJob.SetSelfJobDone", "error", err.Error())
		return false
	}

	// placeholder for subroundBlock.doBlockJob script

	sr.ConsensusCoreHandler.ScheduledProcessor().StartScheduledProcessing(header, body, sr.GetRoundTimeStamp())

	return true
}

func printLogMessage(ctx context.Context, baseMessage string, err error) {
	if common.IsContextDone(ctx) {
		log.Debug(baseMessage + " context is closing")
		return
	}

	log.Debug(baseMessage, "error", err.Error())
}

func (sr *subroundBlock) sendBlock(header data.HeaderHandler, body data.BodyHandler, leader string) bool {
	marshalizedBody, err := sr.Marshalizer().Marshal(body)
	if err != nil {
		log.Debug("sendBlock.Marshal: body", "error", err.Error())
		return false
	}

	marshalizedHeader, err := sr.Marshalizer().Marshal(header)
	if err != nil {
		log.Debug("sendBlock.Marshal: header", "error", err.Error())
		return false
	}

	if sr.couldBeSentTogether(marshalizedBody, marshalizedHeader) {
		return sr.sendHeaderAndBlockBody(header, body, marshalizedBody, marshalizedHeader)
	}

	if !sr.sendBlockBody(body, marshalizedBody) || !sr.sendBlockHeader(header, marshalizedHeader) {
		return false
	}

	return true
}

func (sr *subroundBlock) couldBeSentTogether(marshalizedBody []byte, marshalizedHeader []byte) bool {
	// TODO[cleanup cns finality]: remove this method
	if sr.EnableEpochsHandler().IsFlagEnabled(common.EquivalentMessagesFlag) {
		return false
	}

	bodyAndHeaderSize := uint32(len(marshalizedBody) + len(marshalizedHeader))
	log.Debug("couldBeSentTogether",
		"body size", len(marshalizedBody),
		"header size", len(marshalizedHeader),
		"body and header size", bodyAndHeaderSize,
		"max allowed size in bytes", maxAllowedSizeInBytes)
	return bodyAndHeaderSize <= maxAllowedSizeInBytes
}

func (sr *subroundBlock) createBlock(header data.HeaderHandler) (data.HeaderHandler, data.BodyHandler, error) {
	startTime := sr.GetRoundTimeStamp()
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
) bool {
	headerHash := sr.Hasher().Compute(string(marshalizedHeader))

	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		log.Debug("sendBlockBodyAndHeader.GetLeader", "error", errGetLeader)
		return false
	}

	cnsMsg := consensus.NewConsensusMessage(
		headerHash,
		nil,
		marshalizedBody,
		marshalizedHeader,
		[]byte(leader),
		nil,
		int(bls.MtBlockBodyAndHeader),
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
		return false
	}

	log.Debug("step 1: block body and header have been sent",
		"nonce", headerHandler.GetNonce(),
		"hash", headerHash)

	sr.SetData(headerHash)
	sr.SetBody(bodyHandler)
	sr.SetHeader(headerHandler)

	return true
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
		int(bls.MtBlockBody),
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

	sr.SetBody(bodyHandler)

	return true
}

// sendBlockHeader method sends the proposed block header in the subround Block
func (sr *subroundBlock) sendBlockHeader(
	headerHandler data.HeaderHandler,
	marshalledHeader []byte,
) bool {
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, headerHandler.GetEpoch()) {
		return sr.sendBlockHeaderBeforeEquivalentProofs(headerHandler, marshalledHeader)
	}

	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		log.Debug("sendBlockHeader.GetLeader", "error", errGetLeader)
		return false
	}

	err := sr.BroadcastMessenger().BroadcastHeader(headerHandler, []byte(leader))
	if err != nil {
		log.Warn("sendBlockHeader.BroadcastHeader", "error", err.Error())
		return false
	}

	headerHash := sr.Hasher().Compute(string(marshalledHeader))

	log.Debug("step 1: block header has been sent",
		"nonce", headerHandler.GetNonce(),
		"hash", headerHash)

	sr.SetData(headerHash)
	sr.SetHeader(headerHandler)

	return true
}

// TODO[cleanup cns finality]: remove this method
func (sr *subroundBlock) sendBlockHeaderBeforeEquivalentProofs(
	headerHandler data.HeaderHandler,
	marshalledHeader []byte,
) bool {
	headerHash := sr.Hasher().Compute(string(marshalledHeader))

	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		log.Debug("sendBlockHeaderBeforeEquivalentProofs.GetLeader", "error", errGetLeader)
		return false
	}

	cnsMsg := consensus.NewConsensusMessage(
		headerHash,
		nil,
		nil,
		marshalledHeader,
		[]byte(leader),
		nil,
		int(bls.MtBlockHeader),
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
		log.Debug("sendBlockHeaderBeforeEquivalentProofs.BroadcastConsensusMessage", "error", err.Error())
		return false
	}

	log.Debug("step 1: block header has been sent",
		"nonce", headerHandler.GetNonce(),
		"hash", headerHash)

	sr.SetData(headerHash)
	sr.SetHeader(headerHandler)

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

func (sr *subroundBlock) addProofOnHeader(header data.HeaderHandler) bool {
	// TODO[cleanup cns finality]: remove this
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, header.GetEpoch()) {
		return true
	}

	prevBlockProof, err := sr.EquivalentProofsPool().GetProof(sr.ShardCoordinator().SelfId(), sr.GetData())
	if err != nil {
		return false
	}

	if !isProofEmpty(prevBlockProof) {
		header.SetPreviousProof(prevBlockProof)
		return true
	}

	// this may happen in 2 cases:
	// 1. on the very first block, after equivalent messages flag activation
	// in this case, we set the previous proof as signature and bitmap from the previous header
	// 2. current node is leader in the first block after sync
	// in this case, we won't set the proof, return false and wait for the next round to receive a proof
	currentHeader := sr.Blockchain().GetCurrentBlockHeader()
	if check.IfNil(currentHeader) {
		log.Debug("addProofOnHeader.GetCurrentBlockHeader, nil current header")
		return false
	}

	isFlagEnabledForCurrentHeader := sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, currentHeader.GetEpoch())
	if !isFlagEnabledForCurrentHeader {
		proof := &block.HeaderProof{
			PubKeysBitmap:       currentHeader.GetSignature(),
			AggregatedSignature: currentHeader.GetPubKeysBitmap(),
			HeaderHash:          sr.Blockchain().GetCurrentBlockHeaderHash(),
			HeaderEpoch:         currentHeader.GetEpoch(),
			HeaderNonce:         currentHeader.GetNonce(),
			HeaderShardId:       currentHeader.GetShardID(),
		}
		header.SetPreviousProof(proof)
		return true
	}

	log.Debug("leader after sync, no proof for current header, will wait one round")
	return false
}

func isProofEmpty(proof data.HeaderProofHandler) bool {
	return len(proof.GetAggregatedSignature()) == 0 ||
		len(proof.GetPubKeysBitmap()) == 0 ||
		len(proof.GetHeaderHash()) == 0
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

	header := sr.BlockProcessor().DecodeBlockHeader(cnsDta.Header)

	sr.SetData(cnsDta.BlockHeaderHash)
	sr.SetBody(sr.BlockProcessor().DecodeBlockBody(cnsDta.Body))
	sr.SetHeader(header)

	isInvalidData := check.IfNil(sr.GetBody()) || sr.isInvalidHeaderOrData()
	if isInvalidData {
		return false
	}

	sr.saveProofForPreviousHeaderIfNeeded()

	log.Debug("step 1: block body and header have been received",
		"nonce", sr.GetHeader().GetNonce(),
		"hash", cnsDta.BlockHeaderHash)

	sw.Start("processReceivedBlock")
	blockProcessedWithSuccess := sr.processReceivedBlock(ctx, cnsDta.RoundIndex, cnsDta.PubKey)
	sw.Stop("processReceivedBlock")

	sr.PeerHonestyHandler().ChangeScore(
		node,
		spos.GetConsensusTopicID(sr.ShardCoordinator()),
		spos.LeaderPeerHonestyIncreaseFactor,
	)

	return blockProcessedWithSuccess
}

func (sr *subroundBlock) saveProofForPreviousHeaderIfNeeded() {
	currentHeader := sr.Blockchain().GetCurrentBlockHeader()
	if check.IfNil(currentHeader) {
		log.Debug("saveProofForPreviousHeaderIfNeeded, nil current header")
		return
	}

	// TODO[cleanup cns finality]: remove this
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, currentHeader.GetEpoch()) {
		return
	}

	proof, err := sr.EquivalentProofsPool().GetProof(sr.ShardCoordinator().SelfId(), sr.GetData())
	if err != nil {
		log.Debug("saveProofForPreviousHeaderIfNeeded: do not set proof since it was not found")
		return
	}

	if !isProofEmpty(proof) {
		log.Debug("saveProofForPreviousHeaderIfNeeded: no need to set proof since it is already saved")
		return
	}

	proof = sr.GetHeader().GetPreviousProof()
	err = sr.EquivalentProofsPool().AddProof(proof)
	if err != nil {
		log.Debug("saveProofForPreviousHeaderIfNeeded: failed to add proof, %w", err)
		return
	}
}

func (sr *subroundBlock) isInvalidHeaderOrData() bool {
	return sr.GetData() == nil || check.IfNil(sr.GetHeader()) || sr.GetHeader().CheckFieldsForNil() != nil
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

	sr.SetBody(sr.BlockProcessor().DecodeBlockBody(cnsDta.Body))

	if check.IfNil(sr.GetBody()) {
		return false
	}

	log.Debug("step 1: block body has been received")

	blockProcessedWithSuccess := sr.processReceivedBlock(ctx, cnsDta.RoundIndex, cnsDta.PubKey)

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
// TODO[cleanup cns finality]: remove this method
func (sr *subroundBlock) receivedBlockHeaderBeforeEquivalentProofs(ctx context.Context, cnsDta *consensus.Message) bool {
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

	header := sr.BlockProcessor().DecodeBlockHeader(cnsDta.Header)

	sr.SetData(cnsDta.BlockHeaderHash)
	sr.SetHeader(header)

	if sr.isInvalidHeaderOrData() {
		return false
	}

	sr.saveProofForPreviousHeaderIfNeeded()

	log.Debug("step 1: block header has been received",
		"nonce", sr.GetHeader().GetNonce(),
		"hash", cnsDta.BlockHeaderHash)
	blockProcessedWithSuccess := sr.processReceivedBlock(ctx, cnsDta.RoundIndex, cnsDta.PubKey)

	sr.PeerHonestyHandler().ChangeScore(
		node,
		spos.GetConsensusTopicID(sr.ShardCoordinator()),
		spos.LeaderPeerHonestyIncreaseFactor,
	)

	return blockProcessedWithSuccess
}

func (sr *subroundBlock) receivedBlockHeader(headerHandler data.HeaderHandler) {
	if check.IfNil(headerHandler) {
		return
	}

	// TODO[cleanup cns finality]: remove this check
	if !sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, headerHandler.GetEpoch()) {
		return
	}

	if headerHandler.CheckFieldsForNil() != nil {
		return
	}

	isLeader := sr.IsSelfLeader()
	if sr.ConsensusGroup() == nil || isLeader {
		return
	}

	if sr.IsConsensusDataSet() {
		return
	}

	if sr.IsHeaderAlreadyReceived() {
		return
	}

	if sr.IsSelfJobDone(sr.Current()) {
		return
	}

	if sr.IsSubroundFinished(sr.Current()) {
		return
	}

	marshalledHeader, err := sr.Marshalizer().Marshal(headerHandler)
	if err != nil {
		return
	}

	sr.SetData(sr.Hasher().Compute(string(marshalledHeader)))
	sr.SetHeader(headerHandler)

	sr.saveProofForPreviousHeaderIfNeeded()

	log.Debug("step 1: block header has been received",
		"nonce", sr.GetHeader().GetNonce(),
		"hash", sr.GetData())

	sr.PeerHonestyHandler().ChangeScore(
		sr.Leader(),
		spos.GetConsensusTopicID(sr.ShardCoordinator()),
		spos.LeaderPeerHonestyIncreaseFactor,
	)

	sr.AddReceivedHeader(headerHandler)

	ctx, cancel := context.WithTimeout(context.Background(), sr.RoundHandler().TimeDuration())
	defer cancel()

	sr.processReceivedBlock(ctx, int64(headerHandler.GetRound()), []byte(sr.Leader()))
}

func (sr *subroundBlock) processReceivedBlock(
	ctx context.Context,
	round int64,
	senderPK []byte,
) bool {
	if check.IfNil(sr.GetBody()) {
		return false
	}
	if check.IfNil(sr.GetHeader()) {
		return false
	}

	defer func() {
		sr.SetProcessingBlock(false)
	}()

	sr.SetProcessingBlock(true)

	shouldNotProcessBlock := sr.GetExtendedCalled() || round < sr.RoundHandler().Index()
	if shouldNotProcessBlock {
		log.Debug("canceled round, extended has been called or round index has been changed",
			"round", sr.RoundHandler().Index(),
			"subround", sr.Name(),
			"cnsDta round", round,
			"extended called", sr.GetExtendedCalled(),
		)
		return false
	}

	return sr.processBlock(ctx, round, senderPK)
}

func (sr *subroundBlock) processBlock(
	ctx context.Context,
	roundIndex int64,
	pubkey []byte,
) bool {
	startTime := sr.GetRoundTimeStamp()
	maxTime := sr.RoundHandler().TimeDuration() * time.Duration(sr.processingThresholdPercentage) / 100
	remainingTimeInCurrentRound := func() time.Duration {
		return sr.RoundHandler().RemainingTime(startTime, maxTime)
	}

	metricStatTime := time.Now()
	defer sr.computeSubroundProcessingMetric(metricStatTime, common.MetricProcessedProposedBlock)

	err := sr.BlockProcessor().ProcessBlock(
		sr.GetHeader(),
		sr.GetBody(),
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
		sr.SetRoundCanceled(true)

		return false
	}

	node := string(pubkey)
	err = sr.SetJobDone(node, sr.Current(), true)
	if err != nil {
		sr.printCancelRoundLogMessage(ctx, err)
		return false
	}

	sr.ConsensusCoreHandler.ScheduledProcessor().StartScheduledProcessing(sr.GetHeader(), sr.GetBody(), sr.GetRoundTimeStamp())

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
	if sr.GetRoundCanceled() {
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
