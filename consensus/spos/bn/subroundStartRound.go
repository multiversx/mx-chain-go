package bn

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

// doStartRoundJob method is the function which actually does the job of the StartRound subround
// (it is used as the handler function of the doSubroundJob pointer variable function in Subround struct,
// from spos package)
func (wrk *Worker) doStartRoundJob() bool {
	wrk.BlockBody = nil
	wrk.Header = nil
	wrk.SPoS.Data = nil
	wrk.SPoS.ResetRoundStatus()
	wrk.SPoS.ResetRoundState()
	wrk.cleanReceivedMessages()

	err := wrk.generateNextConsensusGroup()

	if err != nil {
		log.Error(err.Error())

		wrk.SPoS.Chr.SetSelfSubround(-1)

		return false
	}

	leader, err := wrk.SPoS.GetLeader()

	if err != nil {
		log.Error(err.Error())
		leader = "Unknown"
	}

	msg := ""
	if leader == wrk.SPoS.SelfPubKey() {
		msg = " (MY TURN)"
	}

	log.Info(fmt.Sprintf("%sStep 0: Preparing for this round with leader %s%s\n",
		wrk.SPoS.Chr.GetFormattedTime(), hex.EncodeToString([]byte(leader)), msg))

	pubKeys := wrk.SPoS.ConsensusGroup()

	selfIndex, err := wrk.SPoS.IndexSelfConsensusGroup()

	if err != nil {
		log.Info(fmt.Sprintf("%sCanceled round %d in subround %s, NOT IN THE CONSENSUS GROUP\n",
			wrk.SPoS.Chr.GetFormattedTime(), wrk.SPoS.Chr.Round().Index(), getSubroundName(SrStartRound)))

		wrk.SPoS.Chr.SetSelfSubround(-1)

		return false
	}

	err = wrk.multiSigner.Reset(pubKeys, uint16(selfIndex))

	if err != nil {
		log.Error(err.Error())

		wrk.SPoS.Chr.SetSelfSubround(-1)

		return false
	}

	wrk.SPoS.SetStatus(SrStartRound, spos.SsFinished)

	return true
}

func (wrk *Worker) generateNextConsensusGroup() error {
	// TODO: replace random source with last block signature
	headerHash := wrk.BlockChain.CurrentBlockHeaderHash
	if wrk.BlockChain.CurrentBlockHeaderHash == nil {
		headerHash = wrk.BlockChain.GenesisHeaderHash
	}

	randomSource := fmt.Sprintf("%d-%s", wrk.SPoS.Chr.Round().Index(), toB64(headerHash))

	log.Info(fmt.Sprintf("random source used to determine the next consensus group is: %s\n", randomSource))

	nextConsensusGroup, err := wrk.SPoS.GetNextConsensusGroup(randomSource, wrk.validatorGroupSelector)

	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("consensus group for round %d is formed by next validators:\n",
		wrk.SPoS.Chr.Round().Index()))

	for i := 0; i < len(nextConsensusGroup); i++ {
		log.Info(fmt.Sprintf("%s", hex.EncodeToString([]byte(nextConsensusGroup[i]))))
	}

	log.Info(fmt.Sprintf("\n"))

	wrk.SPoS.SetConsensusGroup(nextConsensusGroup)

	return nil
}

// checkStartRoundConsensus method checks if the consensus is achieved in the start subround.
func (wrk *Worker) checkStartRoundConsensus() bool {
	return !wrk.SPoS.Chr.IsCancelled()
}

// extendStartRound method just call the doStartRoundJob method to be sure that the init will be done
func (wrk *Worker) extendStartRound() {
	wrk.SPoS.SetStatus(SrStartRound, spos.SsExtended)

	log.Info(fmt.Sprintf("%sStep 0: Extended the (START_ROUND) subround\n", wrk.SPoS.Chr.GetFormattedTime()))

	wrk.doStartRoundJob()
}
