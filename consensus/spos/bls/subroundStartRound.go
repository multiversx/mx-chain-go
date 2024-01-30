package bls

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/outport/disabled"
)

// subroundStartRound defines the data needed by the subround StartRound
type subroundStartRound struct {
	outportMutex sync.RWMutex
	*spos.Subround
	processingThresholdPercentage int
	executeStoredMessages         func()
	resetConsensusMessages        func()

	outportHandler       outport.OutportHandler
	sentSignatureTracker spos.SentSignaturesTracker
}

// NewSubroundStartRound creates a subroundStartRound object
func NewSubroundStartRound(
	baseSubround *spos.Subround,
	extend func(subroundId int),
	processingThresholdPercentage int,
	executeStoredMessages func(),
	resetConsensusMessages func(),
	sentSignatureTracker spos.SentSignaturesTracker,
) (*subroundStartRound, error) {
	err := checkNewSubroundStartRoundParams(
		baseSubround,
	)
	if err != nil {
		return nil, err
	}
	if extend == nil {
		return nil, fmt.Errorf("%w for extend function", spos.ErrNilFunctionHandler)
	}
	if executeStoredMessages == nil {
		return nil, fmt.Errorf("%w for executeStoredMessages function", spos.ErrNilFunctionHandler)
	}
	if resetConsensusMessages == nil {
		return nil, fmt.Errorf("%w for resetConsensusMessages function", spos.ErrNilFunctionHandler)
	}
	if check.IfNil(sentSignatureTracker) {
		return nil, ErrNilSentSignatureTracker
	}

	srStartRound := subroundStartRound{
		Subround:                      baseSubround,
		processingThresholdPercentage: processingThresholdPercentage,
		executeStoredMessages:         executeStoredMessages,
		resetConsensusMessages:        resetConsensusMessages,
		outportHandler:                disabled.NewDisabledOutport(),
		sentSignatureTracker:          sentSignatureTracker,
		outportMutex:                  sync.RWMutex{},
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

// SetOutportHandler method sets outport handler
func (sr *subroundStartRound) SetOutportHandler(outportHandler outport.OutportHandler) error {
	if check.IfNil(outportHandler) {
		return outport.ErrNilDriver
	}

	sr.outportMutex.Lock()
	sr.outportHandler = outportHandler
	sr.outportMutex.Unlock()

	return nil
}

// doStartRoundJob method does the job of the subround StartRound
func (sr *subroundStartRound) doStartRoundJob(_ context.Context) bool {
	sr.ResetConsensusState()
	sr.RoundIndex = sr.RoundHandler().Index()
	sr.RoundTimeStamp = sr.RoundHandler().TimeStamp()
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
	if nodeState != common.NsSynchronized { // if node is not synchronized yet, it has to continue the bootstrapping mechanism
		return false
	}

	sr.AppStatusHandler().SetStringValue(common.MetricConsensusRoundState, "")

	err := sr.generateNextConsensusGroup(sr.RoundHandler().Index())
	if err != nil {
		log.Debug("initCurrentRound.generateNextConsensusGroup",
			"round index", sr.RoundHandler().Index(),
			"error", err.Error())

		sr.RoundCanceled = true

		return false
	}

	if sr.NodeRedundancyHandler().IsRedundancyNode() {
		sr.NodeRedundancyHandler().AdjustInactivityIfNeeded(
			sr.SelfPubKey(),
			sr.ConsensusGroup(),
			sr.RoundHandler().Index(),
		)
		// we should not return here, the multikey redundancy system relies on it
		// the NodeRedundancyHandler "thinks" it is in redundancy mode even if we use the multikey redundancy system
	}

	leader, err := sr.GetLeader()
	if err != nil {
		log.Debug("initCurrentRound.GetLeader", "error", err.Error())

		sr.RoundCanceled = true

		return false
	}

	msg := ""
	if sr.IsKeyManagedByCurrentNode([]byte(leader)) {
		msg = " (my turn in multi-key)"
	}
	if leader == sr.SelfPubKey() && sr.ShouldConsiderSelfKeyInConsensus() {
		msg = " (my turn)"
	}
	if len(msg) != 0 {
		sr.AppStatusHandler().Increment(common.MetricCountLeader)
		sr.AppStatusHandler().SetStringValue(common.MetricConsensusRoundState, "proposed")
		sr.AppStatusHandler().SetStringValue(common.MetricConsensusState, "proposer")
	}

	log.Debug("step 0: preparing the round",
		"leader", core.GetTrimmedPk(hex.EncodeToString([]byte(leader))),
		"messsage", msg)
	sr.sentSignatureTracker.StartRound()

	pubKeys := sr.ConsensusGroup()
	numMultiKeysInConsensusGroup := sr.computeNumManagedKeysInConsensusGroup(pubKeys)

	sr.indexRoundIfNeeded(pubKeys)

	isSingleKeyLeader := leader == sr.SelfPubKey() && sr.ShouldConsiderSelfKeyInConsensus()
	isLeader := isSingleKeyLeader || sr.IsKeyManagedByCurrentNode([]byte(leader))
	isSelfInConsensus := sr.IsNodeInConsensusGroup(sr.SelfPubKey()) || numMultiKeysInConsensusGroup > 0
	if !isSelfInConsensus {
		log.Debug("not in consensus group")
		sr.AppStatusHandler().SetStringValue(common.MetricConsensusState, "not in consensus group")
	} else {
		if !isLeader {
			sr.AppStatusHandler().Increment(common.MetricCountConsensus)
			sr.AppStatusHandler().SetStringValue(common.MetricConsensusState, "participant")
		}
	}

	err = sr.SigningHandler().Reset(pubKeys)
	if err != nil {
		log.Debug("initCurrentRound.Reset", "error", err.Error())

		sr.RoundCanceled = true

		return false
	}

	startTime := sr.RoundTimeStamp
	maxTime := sr.RoundHandler().TimeDuration() * time.Duration(sr.processingThresholdPercentage) / 100
	if sr.RoundHandler().RemainingTime(startTime, maxTime) < 0 {
		log.Debug("canceled round, time is out",
			"round", sr.SyncTimer().FormattedCurrentTime(), sr.RoundHandler().Index(),
			"subround", sr.Name())

		sr.RoundCanceled = true

		return false
	}

	sr.SetStatus(sr.Current(), spos.SsFinished)

	// execute stored messages which were received in this new round but before this initialisation
	go sr.executeStoredMessages()

	return true
}

func (sr *subroundStartRound) computeNumManagedKeysInConsensusGroup(pubKeys []string) int {
	numMultiKeysInConsensusGroup := 0
	for _, pk := range pubKeys {
		pkBytes := []byte(pk)
		if sr.IsKeyManagedByCurrentNode(pkBytes) {
			numMultiKeysInConsensusGroup++
			log.Trace("in consensus group with multi key",
				"pk", core.GetTrimmedPk(hex.EncodeToString(pkBytes)))
		}
		sr.IncrementRoundsWithoutReceivedMessages(pkBytes)
	}

	if numMultiKeysInConsensusGroup > 0 {
		log.Debug("in consensus group with multi keys identities", "num", numMultiKeysInConsensusGroup)
	}

	return numMultiKeysInConsensusGroup
}

func (sr *subroundStartRound) indexRoundIfNeeded(pubKeys []string) {
	sr.outportMutex.RLock()
	defer sr.outportMutex.RUnlock()

	if !sr.outportHandler.HasDrivers() {
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

	round := sr.RoundHandler().Index()

	roundInfo := &outportcore.RoundInfo{
		Round:            uint64(round),
		SignersIndexes:   signersIndexes,
		BlockWasProposed: false,
		ShardId:          shardId,
		Epoch:            epoch,
		Timestamp:        uint64(sr.RoundTimeStamp.Unix()),
	}
	roundsInfo := &outportcore.RoundsInfo{
		ShardID:    shardId,
		RoundsInfo: []*outportcore.RoundInfo{roundInfo},
	}
	sr.outportHandler.SaveRoundsInfo(roundsInfo)
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

	consensusGroupSizeForEpoch := sr.NodesCoordinator().ConsensusGroupSizeForShardAndEpoch(shardId, currentHeader.GetEpoch())
	sr.SetConsensusGroupSize(consensusGroupSizeForEpoch)

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
	return common.ConsensusOrder
}
