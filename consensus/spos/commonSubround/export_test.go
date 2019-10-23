package commonSubround

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/data"
)

// subroundStartRound

func (sr *SubroundStartRound) DoStartRoundJob() bool {
	return sr.doStartRoundJob()
}

func (sr *SubroundStartRound) DoStartRoundConsensusCheck() bool {
	return sr.doStartRoundConsensusCheck()
}

func (sr *SubroundStartRound) GenerateNextConsensusGroup(roundIndex int64) error {
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

func (sr *SubroundBlock) CreateBody(hdr data.HeaderHandler) (data.BodyHandler, error) {
	return sr.createBody(hdr)
}

func (sr *SubroundBlock) SendBlockBody(body data.BodyHandler) bool {
	return sr.sendBlockBody(body)
}

func (sr *SubroundBlock) SendBlockHeader(header data.HeaderHandler) bool {
	return sr.sendBlockHeader(header)
}
