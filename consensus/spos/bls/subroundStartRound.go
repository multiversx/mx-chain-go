package bls

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/data"
)

// subroundStartRound defines the data needed by the subround StartRound
type subroundStartRound struct {
	*spos.Subround
	processingThresholdPercentage int
	executeStoredMessages         func()
	resetConsensusMessages        func()

	indexer indexer.Indexer
}

// NewSubroundStartRound creates a subroundStartRound object
func NewSubroundStartRound(
	baseSubround *spos.Subround,
	extend func(subroundId int),
	processingThresholdPercentage int,
	executeStoredMessages func(),
	resetConsensusMessages func(),
) (*subroundStartRound, error) {
	err := checkNewSubroundStartRoundParams(
		baseSubround,
	)
	if err != nil {
		return nil, err
	}

	srStartRound := subroundStartRound{
		Subround:                      baseSubround,
		processingThresholdPercentage: processingThresholdPercentage,
		executeStoredMessages:         executeStoredMessages,
		resetConsensusMessages:        resetConsensusMessages,
		indexer:                       indexer.NewNilIndexer(),
	}
	srStartRound.Job = srStartRound.doStartRoundJob
	srStartRound.Check = srStartRound.doStartRoundConsensusCheck
	srStartRound.Extend = extend
	baseSubround.EpochStartRegistrationHandler().RegisterHandler(&srStartRound)

	return &srStartRound, nil
}

func checkNewSubroundStartRoundParams(
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

// SetIndexer method set indexer
func (sr *subroundStartRound) SetIndexer(indexer indexer.Indexer) {
	sr.indexer = indexer
}

// doStartRoundJob method does the job of the subround StartRound
func (sr *subroundStartRound) doStartRoundJob() bool {
	sr.ResetConsensusState()
	sr.RoundIndex = sr.Rounder().Index()
	sr.RoundTimeStamp = sr.Rounder().TimeStamp()
	topic := spos.GetConsensusTopicID(sr.ShardCoordinator())
	sr.GetAntiFloodHandler().ResetForTopic(topic)
	sr.resetConsensusMessages()
	return true
}

// doStartRoundConsensusCheck method checks if the consensus is achieved in the subround StartRound
func (sr *subroundStartRound) doStartRoundConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.IsSubroundFinished(sr.Current()) {
		return true
	}

	if sr.initCurrentRound() {
		return true
	}

	return false
}

func (sr *subroundStartRound) initCurrentRound() bool {
	nodeState := sr.BootStrapper().GetNodeState()
	if nodeState != core.NsSynchronized { // if node is not synchronized yet, it has to continue the bootstrapping mechanism
		return false
	}

	if sr.NodeRedundancyHandler().IsRedundancyNode() {
		//TODO: Add code for redundancy mechanism
	}

	sr.AppStatusHandler().SetStringValue(core.MetricConsensusRoundState, "")

	err := sr.generateNextConsensusGroup(sr.Rounder().Index())
	if err != nil {
		log.Debug("initCurrentRound.generateNextConsensusGroup",
			"round index", sr.Rounder().Index(),
			"error", err.Error())

		sr.RoundCanceled = true

		return false
	}

	leader, err := sr.GetLeader()
	if err != nil {
		log.Debug("initCurrentRound.GetLeader", "error", err.Error())

		sr.RoundCanceled = true

		return false
	}

	msg := ""
	if leader == sr.SelfPubKey() {
		sr.AppStatusHandler().Increment(core.MetricCountLeader)
		sr.AppStatusHandler().SetStringValue(core.MetricConsensusRoundState, "proposed")
		sr.AppStatusHandler().SetStringValue(core.MetricConsensusState, "proposer")
		msg = " (my turn)"
	}

	log.Debug("step 0: preparing the round",
		"leader", core.GetTrimmedPk(hex.EncodeToString([]byte(leader))),
		"messsage", msg)

	pubKeys := sr.ConsensusGroup()

	if sr.NodeRedundancyHandler().IsRedundancyNode() {
		//TODO: Add code for redundancy mechanism
	}

	sr.indexRoundIfNeeded(pubKeys)

	selfIndex, err := sr.SelfConsensusGroupIndex()
	if err != nil {
		log.Debug("not in consensus group")
		sr.AppStatusHandler().SetStringValue(core.MetricConsensusState, "not in consensus group")
	} else {
		if leader != sr.SelfPubKey() {
			sr.AppStatusHandler().Increment(core.MetricCountConsensus)
		}
		sr.AppStatusHandler().SetStringValue(core.MetricConsensusState, "participant")
	}

	err = sr.MultiSigner().Reset(pubKeys, uint16(selfIndex))
	if err != nil {
		log.Debug("initCurrentRound.Reset", "error", err.Error())

		sr.RoundCanceled = true

		return false
	}

	startTime := sr.RoundTimeStamp
	maxTime := sr.Rounder().TimeDuration() * time.Duration(sr.processingThresholdPercentage) / 100
	if sr.Rounder().RemainingTime(startTime, maxTime) < 0 {
		log.Debug("canceled round, time is out",
			"round", sr.SyncTimer().FormattedCurrentTime(), sr.Rounder().Index(),
			"subround", sr.Name())

		sr.RoundCanceled = true

		return false
	}

	sr.SetStatus(sr.Current(), spos.SsFinished)

	// execute stored messages which were received in this new round but before this initialisation
	go sr.executeStoredMessages()

	return true
}

func (sr *subroundStartRound) indexRoundIfNeeded(pubKeys []string) {
	if check.IfNil(sr.indexer) {
		return
	}

	currentHeader := sr.Blockchain().GetCurrentBlockHeader()
	if check.IfNil(currentHeader) {
		currentHeader = sr.Blockchain().GetGenesisHeader()
	}

	epoch := currentHeader.GetEpoch()
	shardId := sr.ShardCoordinator().SelfId()
	nodesCoordinatorShardID, err := sr.NodesCoordinator().ShardIdForEpoch(epoch)
	if err != nil {
		log.Debug("initCurrentRound.ShardIdForEpoch",
			"epoch", epoch,
			"error", err.Error())
		return
	}

	if shardId != nodesCoordinatorShardID {
		log.Debug("initCurrentRound.ShardIdForEpoch",
			"epoch", epoch,
			"shardCoordinator.ShardID", shardId,
			"nodesCoordinator.ShardID", nodesCoordinatorShardID)
		return
	}

	signersIndexes, err := sr.NodesCoordinator().GetValidatorsIndexes(pubKeys, epoch)
	if err != nil {
		log.Error(err.Error())
		return
	}

	round := sr.Rounder().Index()

	roundInfo := workItems.RoundInfo{
		Index:            uint64(round),
		SignersIndexes:   signersIndexes,
		BlockWasProposed: false,
		ShardId:          shardId,
		Timestamp:        time.Duration(sr.RoundTimeStamp.Unix()),
	}

	sr.indexer.SaveRoundsInfo([]workItems.RoundInfo{roundInfo})
}

func (sr *subroundStartRound) generateNextConsensusGroup(roundIndex int64) error {
	currentHeader := sr.Blockchain().GetCurrentBlockHeader()
	if check.IfNil(currentHeader) {
		currentHeader = sr.Blockchain().GetGenesisHeader()
		if check.IfNil(currentHeader) {
			return spos.ErrNilHeader
		}
	}

	randomSeed := currentHeader.GetRandSeed()

	log.Debug("random source for the next consensus group",
		"rand", randomSeed)

	shardId := sr.ShardCoordinator().SelfId()

	nextConsensusGroup, err := sr.GetNextConsensusGroup(
		randomSeed,
		uint64(sr.RoundIndex),
		shardId,
		sr.NodesCoordinator(),
		currentHeader.GetEpoch(),
	)
	if err != nil {
		return err
	}

	log.Trace("consensus group is formed by next validators:",
		"round", roundIndex)

	for i := 0; i < len(nextConsensusGroup); i++ {
		log.Trace(core.GetTrimmedPk(hex.EncodeToString([]byte(nextConsensusGroup[i]))))
	}

	sr.SetConsensusGroup(nextConsensusGroup)

	return nil
}

// EpochStartPrepare wis called when an epoch start event is observed, but not yet confirmed/committed.
// Some components may need to do initialisation on this event
func (sr *subroundStartRound) EpochStartPrepare(metaHdr data.HeaderHandler, _ data.BodyHandler) {
	log.Trace(fmt.Sprintf("epoch %d start prepare in consensus", metaHdr.GetEpoch()))
}

// EpochStartAction is called upon a start of epoch event.
func (sr *subroundStartRound) EpochStartAction(hdr data.HeaderHandler) {
	log.Trace(fmt.Sprintf("epoch %d start action in consensus", hdr.GetEpoch()))

	sr.changeEpoch(hdr.GetEpoch())
}

func (sr *subroundStartRound) changeEpoch(currentEpoch uint32) {
	epochNodes, err := sr.NodesCoordinator().GetConsensusWhitelistedNodes(currentEpoch)
	if err != nil {
		panic(fmt.Sprintf("consensus changing epoch failed with error %s", err.Error()))
	}

	sr.SetEligibleList(epochNodes)
}

// NotifyOrder returns the notification order for a start of epoch event
func (sr *subroundStartRound) NotifyOrder() uint32 {
	return core.ConsensusOrder
}
