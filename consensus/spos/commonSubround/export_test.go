package commonSubround

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

// subroundStartRound

func (sr *SubroundStartRound) DoStartRoundJob() bool {
	return sr.doStartRoundJob()
}

func (sr *SubroundStartRound) DoStartRoundConsensusCheck() bool {
	return sr.doStartRoundConsensusCheck()
}

func (sr *SubroundStartRound) GenerateNextConsensusGroup(roundIndex int32) error {
	return sr.generateNextConsensusGroup(roundIndex)
}

func (sr *SubroundStartRound) InitCurrentRound() bool {
	return sr.initCurrentRound()
}

// subroundBlock

func (sr *SubroundBlock) BlockChain() data.ChainHandler {
	return sr.Blockchain()
}

func (sr *SubroundBlock) DoBlockJob() bool {
	return sr.doBlockJob()
}

func (sr *SubroundBlock) ProcessReceivedBlock(cnsDta *consensus.Message) bool {
	return sr.processReceivedBlock(cnsDta)
}

func (sr *SubroundBlock) DoBlockConsensusCheck() bool {
	return sr.doBlockConsensusCheck()
}

func (sr *SubroundBlock) IsBlockReceived(threshold int) bool {
	return sr.isBlockReceived(threshold)
}

func (sr *SubroundBlock) CreateHeader() (data.HeaderHandler, error) {
	return sr.createHeader()
}
