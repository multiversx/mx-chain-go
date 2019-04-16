package bn

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

type subroundStartRound struct {
	*subround
}

// NewSubroundStartRound creates a SubroundStartRound object
func NewSubroundStartRound(
	subround *subround,
	extend func(subroundId int),
) (*subroundStartRound, error) {

	err := checkNewSubroundStartRoundParams(
		subround,
	)

	if err != nil {
		return nil, err
	}

	srStartRound := subroundStartRound{
		subround,
	}

	srStartRound.job = srStartRound.doStartRoundJob
	srStartRound.check = srStartRound.doStartRoundConsensusCheck
	srStartRound.extend = extend

	return &srStartRound, nil
}

func checkNewSubroundStartRoundParams(
	subround *subround,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	return nil
}

// doStartRoundJob method does the job of the start round subround
func (sr *subroundStartRound) doStartRoundJob() bool {
	sr.GetConsensusState().ResetConsensusState()
	sr.GetConsensusState().RoundIndex = sr.GetRounder().Index()
	sr.GetConsensusState().RoundTimeStamp = sr.GetRounder().TimeStamp()
	return true
}

// doStartRoundConsensusCheck method checks if the consensus is achieved in the start subround.
func (sr *subroundStartRound) doStartRoundConsensusCheck() bool {
	if sr.GetConsensusState().RoundCanceled {
		return false
	}

	if sr.GetConsensusState().Status(SrStartRound) == spos.SsFinished {
		return true
	}

	if sr.initCurrentRound() {
		return true
	}

	return false
}

func (sr *subroundStartRound) initCurrentRound() bool {
	if sr.GetBootStrapper().ShouldSync() { // if node is not synchronized yet, it has to continue the bootstrapping mechanism
		return false
	}

	err := sr.generateNextConsensusGroup(sr.GetRounder().Index())

	if err != nil {
		log.Error(err.Error())

		sr.GetConsensusState().RoundCanceled = true

		return false
	}

	leader, err := sr.GetConsensusState().GetLeader()

	if err != nil {
		log.Info(err.Error())

		sr.GetConsensusState().RoundCanceled = true

		return false
	}

	msg := ""
	if leader == sr.GetConsensusState().SelfPubKey() {
		msg = " (my turn)"
	}

	log.Info(fmt.Sprintf("%sStep 0: preparing for this round with leader %s%s\n",
		sr.GetSyncTimer().FormattedCurrentTime(), hex.EncodeToString([]byte(leader)), msg))

	pubKeys := sr.GetConsensusState().ConsensusGroup()

	selfIndex, err := sr.GetConsensusState().SelfConsensusGroupIndex()

	if err != nil {
		log.Info(fmt.Sprintf("%scanceled round %d in subround %s, not in the consensus group\n",
			sr.GetSyncTimer().FormattedCurrentTime(), sr.GetRounder().Index(), getSubroundName(SrStartRound)))

		sr.GetConsensusState().RoundCanceled = true

		return false
	}

	err = sr.GetMultiSigner().Reset(pubKeys, uint16(selfIndex))

	if err != nil {
		log.Error(err.Error())

		sr.GetConsensusState().RoundCanceled = true

		return false
	}

	startTime := time.Time{}
	startTime = sr.GetConsensusState().RoundTimeStamp
	maxTime := sr.GetRounder().TimeDuration() * syncThresholdPercent / 100
	if sr.GetRounder().RemainingTime(startTime, maxTime) < 0 {
		log.Info(fmt.Sprintf("%scanceled round %d in subround %s, time is out\n",
			sr.GetSyncTimer().FormattedCurrentTime(), sr.GetRounder().Index(), getSubroundName(SrStartRound)))

		sr.GetConsensusState().RoundCanceled = true

		return false
	}

	sr.GetConsensusState().SetStatus(SrStartRound, spos.SsFinished)

	return true
}

func (sr *subroundStartRound) generateNextConsensusGroup(roundIndex int32) error {
	// TODO: replace random source with last block signature
	headerHash := sr.GetChainHandler().GetCurrentBlockHeaderHash()
	if sr.GetChainHandler().GetCurrentBlockHeaderHash() == nil {
		headerHash = sr.GetChainHandler().GetGenesisHeaderHash()
	}

	randomSource := fmt.Sprintf("%d-%s", roundIndex, toB64(headerHash))

	log.Info(fmt.Sprintf("random source used to determine the next consensus group is: %s\n", randomSource))

	nextConsensusGroup, err := sr.GetConsensusState().GetNextConsensusGroup(randomSource, sr.GetValidatorGroupSelector())

	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("consensus group for round %d is formed by next validators:\n",
		roundIndex))

	for i := 0; i < len(nextConsensusGroup); i++ {
		log.Info(fmt.Sprintf("%s", hex.EncodeToString([]byte(nextConsensusGroup[i]))))
	}

	log.Info(fmt.Sprintf("\n"))

	sr.GetConsensusState().SetConsensusGroup(nextConsensusGroup)

	return nil
}
