package commonSubround

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"

	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

var log = logger.GetOrCreate("consensus/spos/commonsubround")

// SubroundStartRound defines the data needed by the subround StartRound
type SubroundStartRound struct {
	*spos.Subround
	processingThresholdPercentage int
	getSubroundName               func(subroundId int) string
	executeStoredMessages         func()

	appStatusHandler core.AppStatusHandler
	indexer          indexer.Indexer
}

// NewSubroundStartRound creates a SubroundStartRound object
func NewSubroundStartRound(
	baseSubround *spos.Subround,
	extend func(subroundId int),
	processingThresholdPercentage int,
	getSubroundName func(subroundId int) string,
	executeStoredMessages func(),
) (*SubroundStartRound, error) {
	err := checkNewSubroundStartRoundParams(
		baseSubround,
	)
	if err != nil {
		return nil, err
	}

	srStartRound := SubroundStartRound{
		Subround:                      baseSubround,
		processingThresholdPercentage: processingThresholdPercentage,
		getSubroundName:               getSubroundName,
		executeStoredMessages:         executeStoredMessages,
		appStatusHandler:              statusHandler.NewNilStatusHandler(),
		indexer:                       indexer.NewNilIndexer(),
	}
	srStartRound.Job = srStartRound.doStartRoundJob
	srStartRound.Check = srStartRound.doStartRoundConsensusCheck
	srStartRound.Extend = extend
	baseSubround.EpochStartSubscriber().RegisterHandler(&srStartRound)

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

// SetAppStatusHandler method set appStatusHandler
func (sr *SubroundStartRound) SetAppStatusHandler(ash core.AppStatusHandler) error {
	if ash == nil || ash.IsInterfaceNil() {
		return spos.ErrNilAppStatusHandler
	}

	sr.appStatusHandler = ash
	return nil
}

// SetIndexer method set indexer
func (sr *SubroundStartRound) SetIndexer(indexer indexer.Indexer) {
	sr.indexer = indexer
}

// doStartRoundJob method does the job of the subround StartRound
func (sr *SubroundStartRound) doStartRoundJob() bool {
	sr.ResetConsensusState()
	sr.RoundIndex = sr.Rounder().Index()
	sr.RoundTimeStamp = sr.Rounder().TimeStamp()
	return true
}

// doStartRoundConsensusCheck method checks if the consensus is achieved in the subround StartRound
func (sr *SubroundStartRound) doStartRoundConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.Status(sr.Current()) == spos.SsFinished {
		return true
	}

	if sr.initCurrentRound() {
		return true
	}

	return false
}

func (sr *SubroundStartRound) initCurrentRound() bool {
	if sr.BootStrapper().ShouldSync() { // if node is not synchronized yet, it has to continue the bootstrapping mechanism
		return false
	}
	sr.appStatusHandler.SetStringValue(core.MetricConsensusRoundState, "")

	err := sr.generateNextConsensusGroup(sr.Rounder().Index())
	if err != nil {
		log.Debug("generateNextConsensusGroup", "error", err.Error())

		sr.RoundCanceled = true

		return false
	}

	leader, err := sr.GetLeader()
	if err != nil {
		log.Debug("GetLeader", "error", err.Error())

		sr.RoundCanceled = true

		return false
	}

	msg := ""
	if leader == sr.SelfPubKey() {
		sr.appStatusHandler.Increment(core.MetricCountLeader)
		sr.appStatusHandler.SetStringValue(core.MetricConsensusRoundState, "proposed")
		sr.appStatusHandler.SetStringValue(core.MetricConsensusState, "proposer")
		msg = " (my turn)"
	}

	log.Debug("step 0: preparing the round",
		"time [s]", sr.SyncTimer().FormattedCurrentTime(),
		"leader", core.GetTrimmedPk(hex.EncodeToString([]byte(leader))),
		"messsage", msg)

	pubKeys := sr.ConsensusGroup()

	sr.indexRoundIfNeeded(pubKeys)

	selfIndex, err := sr.SelfConsensusGroupIndex()
	if err != nil {
		log.Debug("canceled round, not in the consensus group",
			"time [s]", sr.SyncTimer().FormattedCurrentTime(),
			"round", sr.Rounder().Index(),
			"subround", sr.getSubroundName(sr.Current()))

		sr.RoundCanceled = true

		sr.appStatusHandler.SetStringValue(core.MetricConsensusState, "not in consensus group")

		return false
	}

	sr.appStatusHandler.Increment(core.MetricCountConsensus)
	sr.appStatusHandler.SetStringValue(core.MetricConsensusState, "participant")

	err = sr.MultiSigner().Reset(pubKeys, uint16(selfIndex))
	if err != nil {
		log.Debug("Reset", "error", err.Error())

		sr.RoundCanceled = true

		return false
	}

	startTime := time.Time{}
	startTime = sr.RoundTimeStamp
	maxTime := sr.Rounder().TimeDuration() * time.Duration(sr.processingThresholdPercentage) / 100
	if sr.Rounder().RemainingTime(startTime, maxTime) < 0 {
		log.Debug("canceled round, time is out",
			"time [s]", sr.SyncTimer().FormattedCurrentTime(),
			"round", sr.SyncTimer().FormattedCurrentTime(), sr.Rounder().Index(),
			"subround", sr.getSubroundName(sr.Current()))

		sr.RoundCanceled = true

		return false
	}

	sr.SetStatus(sr.Current(), spos.SsFinished)

	// execute stored messages which were received in this new round but before this initialisation
	go sr.executeStoredMessages()

	return true
}

func (sr *SubroundStartRound) indexRoundIfNeeded(pubKeys []string) {
	if sr.indexer == nil || sr.indexer.IsNilIndexer() {
		return
	}

	currentHeader := sr.Blockchain().GetCurrentBlockHeader()
	if currentHeader == nil {
		currentHeader = sr.Blockchain().GetGenesisHeader()
	}

	shardId := sr.ShardCoordinator().SelfId()
	signersIndexes, err := sr.NodesCoordinator().GetValidatorsIndexes(pubKeys, currentHeader.GetEpoch())
	if err != nil {
		log.Error(err.Error())
		return
	}

	round := sr.Rounder().Index()

	roundInfo := indexer.RoundInfo{
		Index:            uint64(round),
		SignersIndexes:   signersIndexes,
		BlockWasProposed: false,
		ShardId:          shardId,
		Timestamp:        time.Duration(sr.RoundTimeStamp.Unix()),
	}

	go sr.indexer.SaveRoundInfo(roundInfo)
}

func (sr *SubroundStartRound) generateNextConsensusGroup(roundIndex int64) error {
	currentHeader := sr.Blockchain().GetCurrentBlockHeader()
	if currentHeader == nil {
		currentHeader = sr.Blockchain().GetGenesisHeader()
		if currentHeader == nil {
			return spos.ErrNilHeader
		}
	}

	randomSeed := currentHeader.GetRandSeed()

	log.Debug("random source for the next consensus group",
		"rand", randomSeed)

	shardId := sr.ShardCoordinator().SelfId()

	nextConsensusGroup, _, err := sr.GetNextConsensusGroup(
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
func (sr *SubroundStartRound) EpochStartPrepare(metaHeader data.HeaderHandler) {
	log.Trace(fmt.Sprintf("epoch %d start prepare in consensus", metaHeader.GetEpoch()))
}

// EpochStartAction is called upon a start of epoch event.
func (sr *SubroundStartRound) EpochStartAction(hdr data.HeaderHandler) {
	log.Trace(fmt.Sprintf("epoch %d start action in consensus", hdr.GetEpoch()))

	sr.changeEpoch(hdr)
}

func (sr *SubroundStartRound) changeEpoch(header data.HeaderHandler) {
	publicKeysPrevEpoch, err := sr.NodesCoordinator().GetAllEligibleValidatorsPublicKeys(header.GetEpoch() - 1)
	if err != nil {
		log.Error(fmt.Sprintf("epoch %d: %s", header.GetEpoch()-1, err.Error()))
		return
	}

	publicKeysNewEpoch, err := sr.NodesCoordinator().GetAllEligibleValidatorsPublicKeys(header.GetEpoch())
	if err != nil {
		log.Error(fmt.Sprintf("epoch %d: %s", header.GetEpoch(), err.Error()))
		return
	}

	estimatedMapSize := len(publicKeysNewEpoch) * len(publicKeysNewEpoch[0])
	shardEligible := make(map[string]struct{}, estimatedMapSize)
	// TODO: update this when inter shard shuffling is enabled
	shardId := sr.ShardCoordinator().SelfId()

	for _, pubKey := range publicKeysPrevEpoch[shardId] {
		shardEligible[string(pubKey)] = struct{}{}
	}
	for _, pubKey := range publicKeysNewEpoch[shardId] {
		shardEligible[string(pubKey)] = struct{}{}
	}

	sr.SetEligibleList(shardEligible)
}
