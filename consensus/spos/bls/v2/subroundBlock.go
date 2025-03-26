package v2

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

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
	mutBlockProcessing            sync.Mutex
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
	if !sr.IsSelfLeader() { // is NOT self leader in this round?
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

	// block proof verification should be done over the header that contains the leader signature
	leaderSignature, err := sr.signBlockHeader(header)
	if err != nil {
		printLogMessage(ctx, "doBlockJob.signBlockHeader", err)
		return false
	}

	err = header.SetLeaderSignature(leaderSignature)
	if err != nil {
		printLogMessage(ctx, "doBlockJob.SetLeaderSignature", err)
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

func (sr *subroundBlock) signBlockHeader(header data.HeaderHandler) ([]byte, error) {
	headerClone := header.ShallowClone()
	err := headerClone.SetLeaderSignature(nil)
	if err != nil {
		return nil, err
	}

	marshalledHdr, err := sr.Marshalizer().Marshal(headerClone)
	if err != nil {
		return nil, err
	}

	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		return nil, errGetLeader
	}

	return sr.SigningHandler().CreateSignatureForPublicKey(marshalledHdr, []byte(leader))
}

func printLogMessage(ctx context.Context, baseMessage string, err error) {
	if common.IsContextDone(ctx) {
		log.Debug(baseMessage + " context is closing")
		return
	}

	log.Debug(baseMessage, "error", err.Error())
}

func (sr *subroundBlock) sendBlock(header data.HeaderHandler, body data.BodyHandler, _ string) bool {
	marshalledBody, err := sr.Marshalizer().Marshal(body)
	if err != nil {
		log.Debug("sendBlock.Marshal: body", "error", err.Error())
		return false
	}

	marshalledHeader, err := sr.Marshalizer().Marshal(header)
	if err != nil {
		log.Debug("sendBlock.Marshal: header", "error", err.Error())
		return false
	}

	sr.logBlockSize(marshalledBody, marshalledHeader)
	if !sr.sendBlockBody(body, marshalledBody) || !sr.sendBlockHeader(header, marshalledHeader) {
		return false
	}

	return true
}

func (sr *subroundBlock) logBlockSize(marshalledBody []byte, marshalledHeader []byte) {
	bodyAndHeaderSize := uint32(len(marshalledBody) + len(marshalledHeader))
	log.Debug("logBlockSize",
		"body size", len(marshalledBody),
		"header size", len(marshalledHeader),
		"body and header size", bodyAndHeaderSize,
		"max allowed size in bytes", maxAllowedSizeInBytes)
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

func (sr *subroundBlock) getPrevHeaderAndHash() (data.HeaderHandler, []byte) {
	prevHeader := sr.Blockchain().GetCurrentBlockHeader()
	prevHeaderHash := sr.Blockchain().GetCurrentBlockHeaderHash()
	if check.IfNil(prevHeader) {
		prevHeader = sr.Blockchain().GetGenesisHeader()
		prevHeaderHash = sr.Blockchain().GetGenesisHeaderHash()
	}

	return prevHeader, prevHeaderHash
}

func (sr *subroundBlock) createHeader() (data.HeaderHandler, error) {
	prevHeader, prevHash := sr.getPrevHeaderAndHash()
	nonce := prevHeader.GetNonce() + 1
	prevRandSeed := prevHeader.GetRandSeed()

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

func (sr *subroundBlock) isHeaderForCurrentConsensus(header data.HeaderHandler) bool {
	if check.IfNil(header) {
		return false
	}
	if header.GetShardID() != sr.ShardCoordinator().SelfId() {
		return false
	}
	if header.GetRound() != uint64(sr.RoundHandler().Index()) {
		return false
	}

	prevHeader, prevHash := sr.getPrevHeaderAndHash()
	if check.IfNil(prevHeader) {
		return false
	}
	if !bytes.Equal(header.GetPrevHash(), prevHash) {
		return false
	}
	if header.GetNonce() != prevHeader.GetNonce()+1 {
		return false
	}
	prevRandSeed := prevHeader.GetRandSeed()

	return bytes.Equal(header.GetPrevRandSeed(), prevRandSeed)
}

func (sr *subroundBlock) getLeaderForHeader(headerHandler data.HeaderHandler) ([]byte, error) {
	nc := sr.NodesCoordinator()

	prevBlockEpoch := uint32(0)
	if sr.Blockchain().GetCurrentBlockHeader() != nil {
		prevBlockEpoch = sr.Blockchain().GetCurrentBlockHeader().GetEpoch()
	}
	// TODO: remove this if first block in new epoch will be validated by epoch validators
	// first block in epoch is validated by previous epoch validators
	selectionEpoch := headerHandler.GetEpoch()
	if selectionEpoch != prevBlockEpoch {
		selectionEpoch = prevBlockEpoch
	}
	leader, _, err := nc.ComputeConsensusGroup(
		headerHandler.GetPrevRandSeed(),
		headerHandler.GetRound(),
		headerHandler.GetShardID(),
		selectionEpoch,
	)
	if err != nil {
		return nil, err
	}

	return leader.PubKey(), err
}

func (sr *subroundBlock) receivedBlockHeader(headerHandler data.HeaderHandler) {
	if check.IfNil(headerHandler) {
		return
	}

	log.Debug("subroundBlock.receivedBlockHeader", "nonce", headerHandler.GetNonce(), "round", headerHandler.GetRound())
	if headerHandler.CheckFieldsForNil() != nil {
		return
	}

	isHeaderForCurrentConsensus := sr.isHeaderForCurrentConsensus(headerHandler)
	if !isHeaderForCurrentConsensus {
		log.Debug("subroundBlock.receivedBlockHeader - header is not for current consensus")
		return
	}

	isLeader := sr.IsSelfLeader()
	if sr.ConsensusGroup() == nil || isLeader {
		log.Debug("subroundBlock.receivedBlockHeader - consensus group is nil or is leader")
		return
	}

	if sr.IsConsensusDataSet() {
		log.Debug("subroundBlock.receivedBlockHeader - consensus data is set")
		return
	}

	headerLeader, err := sr.getLeaderForHeader(headerHandler)
	if err != nil {
		log.Debug("subroundBlock.receivedBlockHeader - error getting leader for header", err.Error())
		return
	}

	if !sr.IsNodeLeaderInCurrentRound(string(headerLeader)) {
		sr.PeerHonestyHandler().ChangeScore(
			string(headerLeader),
			spos.GetConsensusTopicID(sr.ShardCoordinator()),
			spos.LeaderPeerHonestyDecreaseFactor,
		)

		log.Debug("subroundBlock.receivedBlockHeader - leader is not the leader in current round")
		return
	}

	if sr.IsHeaderAlreadyReceived() {
		log.Debug("subroundBlock.receivedBlockHeader - header is already received")
		return
	}

	if !sr.CanProcessReceivedHeader(string(headerLeader)) {
		log.Debug("subroundBlock.receivedBlockHeader - can not process received header")
		return
	}

	headerHash, err := core.CalculateHash(sr.Marshalizer(), sr.Hasher(), headerHandler)
	if err != nil {
		log.Debug("subroundBlock.receivedBlockHeader", "error", err.Error())
		return
	}

	sr.SetData(headerHash)
	sr.SetHeader(headerHandler)

	log.Debug("step 1: block header has been received",
		"nonce", sr.GetHeader().GetNonce(),
		"hash", sr.GetData())

	sr.AddReceivedHeader(headerHandler)

	ctx, cancel := context.WithTimeout(context.Background(), sr.RoundHandler().TimeDuration())
	defer cancel()

	_ = sr.processReceivedBlock(ctx, int64(headerHandler.GetRound()), []byte(sr.Leader()))
	sr.PeerHonestyHandler().ChangeScore(
		sr.Leader(),
		spos.GetConsensusTopicID(sr.ShardCoordinator()),
		spos.LeaderPeerHonestyIncreaseFactor,
	)
}

// CanProcessReceivedHeader method returns true if the received header can be processed and false otherwise
func (sr *subroundBlock) CanProcessReceivedHeader(headerLeader string) bool {
	return sr.shouldProcessBlock(headerLeader)
}

func (sr *subroundBlock) shouldProcessBlock(headerLeader string) bool {
	if sr.IsNodeSelf(headerLeader) {
		return false
	}
	if sr.IsJobDone(headerLeader, sr.Current()) {
		return false
	}

	if sr.IsSubroundFinished(sr.Current()) {
		return false
	}

	return true
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

	sw := core.NewStopWatch()
	sw.Start("processReceivedBlock")

	sr.mutBlockProcessing.Lock()
	defer sr.mutBlockProcessing.Unlock()

	defer func() {
		sw.Stop("processReceivedBlock")
		log.Info("time measurements of processReceivedBlock", sw.GetMeasurements()...)

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

	// check again under critical section to avoid double execution
	if !sr.shouldProcessBlock(string(senderPK)) {
		return false
	}

	sw.Start("processBlock")
	ok := sr.processBlock(ctx, round, senderPK)
	sw.Stop("processBlock")

	return ok
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
