package commonSubround

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

var log = logger.DefaultLogger()

// SubroundStartRound defines the data needed by the subround StartRound
type SubroundStartRound struct {
	*spos.Subround
	processingThresholdPercentage int
	getSubroundName               func(subroundId int) string
	executeStoredMessages         func()
	broadcastUnnotarisedBlocks    func()

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
	broadcastUnnotarisedBlocks func(),
) (*SubroundStartRound, error) {
	err := checkNewSubroundStartRoundParams(
		baseSubround,
		broadcastUnnotarisedBlocks,
	)
	if err != nil {
		return nil, err
	}

	srStartRound := SubroundStartRound{
		baseSubround,
		processingThresholdPercentage,
		getSubroundName,
		executeStoredMessages,
		broadcastUnnotarisedBlocks,
		statusHandler.NewNilStatusHandler(),
		nil,
	}
	srStartRound.Job = srStartRound.doStartRoundJob
	srStartRound.Check = srStartRound.doStartRoundConsensusCheck
	srStartRound.Extend = extend

	return &srStartRound, nil
}

func checkNewSubroundStartRoundParams(
	baseSubround *spos.Subround,
	broadcastUnnotarisedBlocks func(),
) error {
	if baseSubround == nil {
		return spos.ErrNilSubround
	}
	if baseSubround.ConsensusState == nil {
		return spos.ErrNilConsensusState
	}
	if broadcastUnnotarisedBlocks == nil {
		return spos.ErrNilBroadcastUnnotarisedBlocks
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
		log.Error(err.Error())

		sr.RoundCanceled = true

		return false
	}

	leader, err := sr.GetLeader()
	if err != nil {
		log.Error(err.Error())

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

	log.Info(fmt.Sprintf("%sStep 0: preparing for this round with leader %s%s\n",
		sr.SyncTimer().FormattedCurrentTime(), core.GetTrimmedPk(hex.EncodeToString([]byte(leader))), msg))

	pubKeys := sr.ConsensusGroup()

	sr.indexRoundIfNeeded(pubKeys)

	selfIndex, err := sr.SelfConsensusGroupIndex()
	if err != nil {
		log.Info(fmt.Sprintf("%scanceled round %d in subround %s, not in the consensus group\n",
			sr.SyncTimer().FormattedCurrentTime(), sr.Rounder().Index(), sr.getSubroundName(sr.Current())))

		sr.RoundCanceled = true

		sr.appStatusHandler.SetStringValue(core.MetricConsensusState, "not in consensus group")

		return false
	}

	sr.appStatusHandler.Increment(core.MetricCountConsensus)
	sr.appStatusHandler.SetStringValue(core.MetricConsensusState, "participant")

	err = sr.MultiSigner().Reset(pubKeys, uint16(selfIndex))
	if err != nil {
		log.Error(err.Error())

		sr.RoundCanceled = true

		return false
	}

	startTime := time.Time{}
	startTime = sr.RoundTimeStamp
	maxTime := sr.Rounder().TimeDuration() * time.Duration(sr.processingThresholdPercentage) / 100
	if sr.Rounder().RemainingTime(startTime, maxTime) < 0 {
		log.Info(fmt.Sprintf("%scanceled round %d in subround %s, time is out\n",
			sr.SyncTimer().FormattedCurrentTime(), sr.Rounder().Index(), sr.getSubroundName(sr.Current())))

		sr.RoundCanceled = true

		return false
	}

	sr.SetStatus(sr.Current(), spos.SsFinished)

	if leader == sr.SelfPubKey() {
		//TODO: Should be analyzed if call of sr.broadcastUnnotarisedBlocks() is still necessary
	}

	// execute stored messages which were received in this new round but before this initialisation
	go sr.executeStoredMessages()

	return true
}

func (sr *SubroundStartRound) indexRoundIfNeeded(pubKeys []string) {
	if sr.indexer == nil || sr.IsInterfaceNil() {
		return
	}

	validatorsPubKeys := sr.NodesCoordinator().GetAllValidatorsPublicKeys()
	shardId := sr.ShardCoordinator().SelfId()
	signersIndexes := make([]uint64, 0)

	for _, pubKey := range pubKeys {
		for index, value := range validatorsPubKeys[shardId] {
			if bytes.Equal([]byte(pubKey), value) {
				signersIndexes = append(signersIndexes, uint64(index))
			}
		}
	}

	sr.indexer.SaveRoundInfo(sr.Rounder().Index(), signersIndexes)
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

	log.Info(fmt.Sprintf("random source used to determine the next consensus group is: %s\n",
		core.ToB64(randomSeed)),
	)

	shardId := sr.ShardCoordinator().SelfId()

	nextConsensusGroup, rewardsAddresses, err := sr.GetNextConsensusGroup(
		randomSeed,
		uint64(sr.RoundIndex),
		shardId,
		sr.NodesCoordinator(),
	)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("consensus group for round %d is formed by next validators:\n",
		roundIndex))

	for i := 0; i < len(nextConsensusGroup); i++ {
		log.Info(fmt.Sprintf("%s", core.GetTrimmedPk(hex.EncodeToString([]byte(nextConsensusGroup[i])))))
	}

	log.Info(fmt.Sprintf("\n"))

	sr.SetConsensusGroup(nextConsensusGroup)

	sr.BlockProcessor().SetConsensusData(rewardsAddresses, uint64(sr.RoundIndex))

	return nil
}
