package commonSubround

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// SubroundBlock defines the data needed by the subround Block
type SubroundBlock struct {
	*spos.Subround

	mtBlockBody                   int
	mtBlockHeader                 int
	processingThresholdPercentage int
	getSubroundName               func(subroundId int) string
}

// NewSubroundBlock creates a SubroundBlock object
func NewSubroundBlock(
	baseSubround *spos.Subround,
	extend func(subroundId int),
	mtBlockBody int,
	mtBlockHeader int,
	processingThresholdPercentage int,
	getSubroundName func(subroundId int) string,
) (*SubroundBlock, error) {
	err := checkNewSubroundBlockParams(baseSubround)
	if err != nil {
		return nil, err
	}

	srBlock := SubroundBlock{
		Subround:                      baseSubround,
		mtBlockBody:                   mtBlockBody,
		mtBlockHeader:                 mtBlockHeader,
		processingThresholdPercentage: processingThresholdPercentage,
		getSubroundName:               getSubroundName,
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
func (sr *SubroundBlock) doBlockJob() bool {
	if !sr.IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if sr.Rounder().Index() <= sr.getRoundInLastCommittedBlock() {
		return false
	}

	if sr.IsSelfJobDone(sr.Current()) {
		return false
	}

	if sr.IsCurrentSubroundFinished(sr.Current()) {
		return false
	}

	hdr, err := sr.createHeader()
	if err != nil {
		log.Debug("createHeader", "error", err.Error())
		return false
	}

	body, err := sr.createBody(hdr)
	if err != nil {
		log.Debug("createBody", "error", err.Error())
		return false
	}

	err = sr.BlockProcessor().ApplyBodyToHeader(hdr, body)
	if err != nil {
		log.Debug("ApplyBodyToHeader", "error", err.Error())
		return false
	}

	if !sr.sendBlockBody(body) ||
		!sr.sendBlockHeader(hdr) {
		return false
	}

	err = sr.SetSelfJobDone(sr.Current(), true)
	if err != nil {
		log.Debug("SetSelfJobDone", "error", err.Error())
		return false
	}

	return true
}

func (sr *SubroundBlock) createBody(header data.HeaderHandler) (data.BodyHandler, error) {
	startTime := time.Time{}
	startTime = sr.RoundTimeStamp
	maxTime := time.Duration(sr.EndTime())
	haveTimeInCurrentSubround := func() bool {
		return sr.Rounder().RemainingTime(startTime, maxTime) > 0
	}

	blockBody, err := sr.BlockProcessor().CreateBlockBody(
		header,
		haveTimeInCurrentSubround,
	)
	if err != nil {
		return nil, err
	}

	return blockBody, nil
}

// sendBlockBody method job the proposed block body in the subround Block
func (sr *SubroundBlock) sendBlockBody(blockBody data.BodyHandler) bool {
	blkStr, err := sr.Marshalizer().Marshal(blockBody)
	if err != nil {
		log.Debug("Marshal", "error", err.Error())
		return false
	}

	msg := consensus.NewConsensusMessage(
		nil,
		blkStr,
		[]byte(sr.SelfPubKey()),
		nil,
		sr.mtBlockBody,
		uint64(sr.Rounder().TimeStamp().Unix()),
		sr.Rounder().Index(),
		sr.ChainID(),
	)

	err = sr.BroadcastMessenger().BroadcastConsensusMessage(msg)
	if err != nil {
		log.Debug("BroadcastConsensusMessage", "error", err.Error())
		return false
	}

	log.Debug("step 1: block body has been sent",
		"time [s]", sr.SyncTimer().FormattedCurrentTime())

	sr.BlockBody = blockBody

	return true
}

// sendBlockHeader method job the proposed block header in the subround Block
func (sr *SubroundBlock) sendBlockHeader(hdr data.HeaderHandler) bool {
	hdrStr, err := sr.Marshalizer().Marshal(hdr)
	if err != nil {
		log.Debug("Marshal", "error", err.Error())
		return false
	}

	hdrHash := sr.Hasher().Compute(string(hdrStr))

	msg := consensus.NewConsensusMessage(
		hdrHash,
		hdrStr,
		[]byte(sr.SelfPubKey()),
		nil,
		sr.mtBlockHeader,
		uint64(sr.Rounder().TimeStamp().Unix()),
		sr.Rounder().Index(),
		sr.ChainID(),
	)

	err = sr.BroadcastMessenger().BroadcastConsensusMessage(msg)
	if err != nil {
		log.Debug("BroadcastConsensusMessage", "error", err.Error())
		return false
	}

	log.Debug("step 1: block header has been sent",
		"time [s]", sr.SyncTimer().FormattedCurrentTime(),
		"nonce", hdr.GetNonce(),
		"hash", hdrHash)

	sr.Data = hdrHash
	sr.Header = hdr

	return true
}

func (sr *SubroundBlock) createHeader() (data.HeaderHandler, error) {
	hdr := sr.BlockProcessor().CreateNewHeader()

	var prevRandSeed []byte
	if sr.Blockchain().GetCurrentBlockHeader() == nil {
		hdr.SetNonce(1)
		hdr.SetPrevHash(sr.Blockchain().GetGenesisHeaderHash())

		prevRandSeed = sr.Blockchain().GetGenesisHeader().GetRandSeed()
	} else {
		hdr.SetNonce(sr.Blockchain().GetCurrentBlockHeader().GetNonce() + 1)
		hdr.SetPrevHash(sr.Blockchain().GetCurrentBlockHeaderHash())

		prevRandSeed = sr.Blockchain().GetCurrentBlockHeader().GetRandSeed()
	}

	randSeed, err := sr.SingleSigner().Sign(sr.PrivateKey(), prevRandSeed)
	if err != nil {
		return nil, err
	}

	hdr.SetShardID(sr.ShardCoordinator().SelfId())
	hdr.SetRound(uint64(sr.Rounder().Index()))
	hdr.SetTimeStamp(uint64(sr.Rounder().TimeStamp().Unix()))
	hdr.SetPrevRandSeed(prevRandSeed)
	hdr.SetRandSeed(randSeed)
	hdr.SetChainID(sr.ChainID())

	return hdr, nil
}

// ReceivedBlockBody method is called when a block body is received through the block body channel
func (sr *SubroundBlock) ReceivedBlockBody(cnsDta *consensus.Message) bool {
	node := string(cnsDta.PubKey)

	if sr.IsBlockBodyAlreadyReceived() {
		return false
	}

	if !sr.IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), sr.Current()) {
		return false
	}

	sr.BlockBody = sr.BlockProcessor().DecodeBlockBody(cnsDta.SubRoundData)

	if sr.BlockBody == nil {
		return false
	}

	log.Debug("step 1: block body has been received",
		"time [s]", sr.SyncTimer().FormattedCurrentTime())

	blockProcessedWithSuccess := sr.processReceivedBlock(cnsDta)

	return blockProcessedWithSuccess
}

// ReceivedBlockHeader method is called when a block header is received through the block header channel.
// If the block header is valid, than the validatorRoundStates map corresponding to the node which sent it,
// is set on true for the subround Block
func (sr *SubroundBlock) ReceivedBlockHeader(cnsDta *consensus.Message) bool {
	node := string(cnsDta.PubKey)

	if sr.IsConsensusDataSet() {
		return false
	}

	if sr.IsHeaderAlreadyReceived() {
		return false
	}

	if !sr.IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), sr.Current()) {
		return false
	}

	sr.Data = cnsDta.BlockHeaderHash
	sr.Header = sr.BlockProcessor().DecodeBlockHeader(cnsDta.SubRoundData)

	if sr.Header == nil {
		return false
	}

	log.Debug("step 1: block header has been received",
		"time [s]", sr.SyncTimer().FormattedCurrentTime(),
		"nonce", sr.Header.GetNonce(),
		"hash", cnsDta.BlockHeaderHash)
	blockProcessedWithSuccess := sr.processReceivedBlock(cnsDta)

	return blockProcessedWithSuccess
}

func (sr *SubroundBlock) processReceivedBlock(cnsDta *consensus.Message) bool {
	if sr.BlockBody == nil || sr.BlockBody.IsInterfaceNil() {
		return false
	}
	if sr.Header == nil || sr.Header.IsInterfaceNil() {
		return false
	}

	defer func() {
		sr.SetProcessingBlock(false)
	}()

	sr.SetProcessingBlock(true)

	node := string(cnsDta.PubKey)

	startTime := time.Time{}
	startTime = sr.RoundTimeStamp
	maxTime := sr.Rounder().TimeDuration() * time.Duration(sr.processingThresholdPercentage) / 100
	remainingTimeInCurrentRound := func() time.Duration {
		return sr.Rounder().RemainingTime(startTime, maxTime)
	}

	err := sr.BlockProcessor().ProcessBlock(
		sr.Blockchain(),
		sr.Header,
		sr.BlockBody,
		remainingTimeInCurrentRound,
	)

	if cnsDta.RoundIndex < sr.Rounder().Index() {
		log.Debug("canceled round, meantime round index has been changed",
			"old round", cnsDta.RoundIndex,
			"subround", sr.getSubroundName(sr.Current()),
			"new round", sr.Rounder().Index())
		return false
	}

	if err != nil {
		log.Debug("canceled round",
			"round", sr.Rounder().Index(),
			"subround", sr.getSubroundName(sr.Current()),
			"error", err.Error())
		if err == process.ErrTimeIsOut {
			sr.RoundCanceled = true
		}
		return false
	}

	err = sr.SetJobDone(node, sr.Current(), true)
	if err != nil {
		log.Debug("canceled round",
			"round", sr.Rounder().Index(),
			"subround", sr.getSubroundName(sr.Current()),
			"error", err.Error())
		return false
	}

	return true
}

// doBlockConsensusCheck method checks if the consensus in the subround Block is achieved
func (sr *SubroundBlock) doBlockConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.Status(sr.Current()) == spos.SsFinished {
		return true
	}

	threshold := sr.Threshold(sr.Current())
	if sr.isBlockReceived(threshold) {
		log.Debug("step 1: subround has been finished",
			"time [s]", sr.SyncTimer().FormattedCurrentTime(),
			"subround", sr.Name())
		sr.SetStatus(sr.Current(), spos.SsFinished)
		return true
	}

	return false
}

// isBlockReceived method checks if the block was received from the leader in the current round
func (sr *SubroundBlock) isBlockReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		node := sr.ConsensusGroup()[i]
		isJobDone, err := sr.JobDone(node, sr.Current())

		if err != nil {
			log.Debug("BroadcastConsensusMessage", "error", err.Error())
			continue
		}

		if isJobDone {
			n++
		}
	}

	return n >= threshold
}

func (sr *SubroundBlock) getRoundInLastCommittedBlock() int64 {
	roundInLastCommittedBlock := int64(0)
	if sr.Blockchain().GetCurrentBlockHeader() != nil {
		roundInLastCommittedBlock = int64(sr.Blockchain().GetCurrentBlockHeader().GetRound())
	}

	return roundInLastCommittedBlock
}
