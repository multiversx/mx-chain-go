package commonSubround

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

var log = logger.DefaultLogger()

// SubroundStartRound defines the data needed by the subround StartRound
type SubroundStartRound struct {
	*spos.Subround
	processingThresholdPercentage int
	getSubroundName               func(subroundId int) string
	executeStoredMessages         func()
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
		baseSubround,
		processingThresholdPercentage,
		getSubroundName,
		executeStoredMessages,
	}
	srStartRound.Job = srStartRound.doStartRoundJob
	srStartRound.Check = srStartRound.doStartRoundConsensusCheck
	srStartRound.Extend = extend

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

	err := sr.generateNextConsensusGroup(sr.Rounder().Index())
	if err != nil {
		log.Error(err.Error())

		sr.RoundCanceled = true

		return false
	}

	leader, err := sr.GetLeader()
	if err != nil {
		log.Info(err.Error())

		sr.RoundCanceled = true

		return false
	}

	msg := ""
	if leader == sr.SelfPubKey() {
		msg = " (my turn)"
	}

	log.Info(fmt.Sprintf("%sStep 0: preparing for this round with leader %s%s\n",
		sr.SyncTimer().FormattedCurrentTime(), core.GetTrimmedPk(hex.EncodeToString([]byte(leader))), msg))

	pubKeys := sr.ConsensusGroup()

	selfIndex, err := sr.SelfConsensusGroupIndex()
	if err != nil {
		log.Info(fmt.Sprintf("%scanceled round %d in subround %s, not in the consensus group\n",
			sr.SyncTimer().FormattedCurrentTime(), sr.Rounder().Index(), sr.getSubroundName(sr.Current())))

		sr.RoundCanceled = true

		return false
	}

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

	// execute stored messages which were received in this new round but before this initialisation
	go sr.executeStoredMessages()

	return true
}

func (sr *SubroundStartRound) generateNextConsensusGroup(roundIndex int32) error {
	currentHeader := sr.Blockchain().GetCurrentBlockHeader()
	if currentHeader == nil {
		currentHeader = sr.Blockchain().GetGenesisHeader()
		if currentHeader == nil {
			return spos.ErrNilHeader
		}
	}

	randomSource := fmt.Sprintf("%d-%s", roundIndex, toB64(currentHeader.GetRandSeed()))

	log.Info(fmt.Sprintf("random source used to determine the next consensus group is: %s\n", randomSource))

	nextConsensusGroup, err := sr.GetNextConsensusGroup(randomSource, sr.ValidatorGroupSelector())
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

	return nil
}
