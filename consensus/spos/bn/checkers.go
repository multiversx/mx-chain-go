package bn

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

func (wrk *Worker) isConsensusDataNotSet() bool {
	isConsensusDataNotSet := wrk.SPoS.Data == nil

	return isConsensusDataNotSet
}

func (wrk *Worker) isConsensusDataAlreadySet() bool {
	isConsensusDataAlreadySet := wrk.SPoS.Data != nil

	return isConsensusDataAlreadySet
}

func (wrk *Worker) isSelfJobDone(currentSubround chronology.SubroundId) bool {
	isSelfJobDone, err := wrk.SPoS.GetSelfJobDone(currentSubround)

	if err != nil {
		log.Error(err.Error())
		return true
	}

	return isSelfJobDone
}

func (wrk *Worker) isJobDone(node string, currentSubround chronology.SubroundId) bool {
	isJobDone, err := wrk.SPoS.RoundConsensus.GetJobDone(node, currentSubround)

	if err != nil {
		log.Error(err.Error())
		return true
	}

	return isJobDone
}

func (wrk *Worker) isCurrentRoundFinished(currentSubround chronology.SubroundId) bool {
	isCurrentRoundFinished := wrk.SPoS.Status(currentSubround) == spos.SsFinished

	return isCurrentRoundFinished
}

func (wrk *Worker) isMessageReceivedFromItself(node string) bool {
	isMessageReceivedFromItself := node == wrk.SPoS.SelfPubKey()

	return isMessageReceivedFromItself
}

func (wrk *Worker) isMessageReceivedTooLate() bool {
	isMessageReceivedTooLate := wrk.SPoS.Chr.GetSubround() > chronology.SubroundId(SrEndRound)

	return isMessageReceivedTooLate
}

func (wrk *Worker) isMessageReceivedForOtherRound(roundIndex int32) bool {
	isMessageReceivedForOtherRound := roundIndex != wrk.SPoS.Chr.Round().Index()

	return isMessageReceivedForOtherRound
}

func (wrk *Worker) isBlockBodyAlreadyReceived() bool {
	isBlockBodyAlreadyReceived := wrk.BlockBody != nil

	return isBlockBodyAlreadyReceived
}

func (wrk *Worker) isHeaderAlreadyReceived() bool {
	isHeaderAlreadyReceived := wrk.Header != nil

	return isHeaderAlreadyReceived
}

func (wrk *Worker) canDoSubroundJob(currentSubround chronology.SubroundId) bool {
	if wrk.isConsensusDataNotSet() {
		return false
	}

	if wrk.isSelfJobDone(currentSubround) {
		return false
	}

	if wrk.isCurrentRoundFinished(currentSubround) {
		return false
	}

	return true
}

func (wrk *Worker) canReceiveMessage(node string, roundIndex int32, currentSubround chronology.SubroundId) bool {
	if wrk.isMessageReceivedFromItself(node) {
		return false
	}

	if wrk.isMessageReceivedTooLate() {
		return false
	}

	if wrk.isMessageReceivedForOtherRound(roundIndex) {
		return false
	}

	if wrk.isJobDone(node, currentSubround) {
		return false
	}

	if wrk.isCurrentRoundFinished(currentSubround) {
		return false
	}

	return true
}
