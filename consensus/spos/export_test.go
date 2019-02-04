package spos

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

func (sposWorker *SPOSConsensusWorker) CheckSignaturesValidity(bitmap []byte) error {
	return sposWorker.checkSignaturesValidity(bitmap)
}

func (sposWorker *SPOSConsensusWorker) GenCommitmentHash() ([]byte, error) {
	return sposWorker.genCommitmentHash()
}

func (sposWorker *SPOSConsensusWorker) GenBitmap(subround chronology.SubroundId) []byte {
	return sposWorker.genBitmap(subround)
}

func (sposWorker *SPOSConsensusWorker) CheckCommitmentsValidity(bitmap []byte) error {
	return sposWorker.checkCommitmentsValidity(bitmap)
}

func (sposWorker *SPOSConsensusWorker) ShouldDropConsensusMessage(cnsDta *ConsensusData) bool {
	return sposWorker.shouldDropConsensusMessage(cnsDta)
}

func (sposWorker *SPOSConsensusWorker) CheckSignature(cnsData *ConsensusData) error {
	return sposWorker.checkSignature(cnsData)
}

func (sposWorker *SPOSConsensusWorker) ProcessReceivedBlock(cnsDta *ConsensusData) bool {
	return sposWorker.processReceivedBlock(cnsDta)
}

func (sposWorker *SPOSConsensusWorker) HaveTime() time.Duration {
	return sposWorker.haveTime()
}

func (sposWorker *SPOSConsensusWorker) InitReceivedMessages() {
	sposWorker.initReceivedMessages()
}

func (sposWorker *SPOSConsensusWorker) CleanReceivedMessages() {
	sposWorker.cleanReceivedMessages()
}

func (sposWorker *SPOSConsensusWorker) ExecuteMessage(cnsDtaList []*ConsensusData) {
	sposWorker.executeMessage(cnsDtaList)
}
