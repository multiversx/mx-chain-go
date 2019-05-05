package bls

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
)

var log = logger.DefaultLogger()

// worker defines the data needed by spos to communicate between nodes which are in the validators group
type worker struct {
}

// NewConsensusService creates a new worker object
func NewConsensusService() (*worker, error) {
	wrk := worker{}

	return &wrk, nil
}

func (wrk *worker) InitReceivedMessages() map[consensus.MessageType][]*consensus.Message {

	receivedMessages := make(map[consensus.MessageType][]*consensus.Message)
	receivedMessages[MtBlockBody] = make([]*consensus.Message, 0)
	receivedMessages[MtBlockHeader] = make([]*consensus.Message, 0)
	receivedMessages[MtSignature] = make([]*consensus.Message, 0)

	return receivedMessages

}

func (wrk *worker) GetStringValue(messageType consensus.MessageType) string {
	return getStringValue(messageType)
}

func (wrk *worker) GetSubroundName(subroundId int) string {
	return getSubroundName(subroundId)
}
func (wrk *worker) GetMessageRange() []consensus.MessageType {
	v := []consensus.MessageType{}

	for i := MtBlockBody; i <= MtSignature; i++ {
		v = append(v, i)
	}

	return v
}

func (wrk *worker) IsFinished(consensusState *spos.ConsensusState, msgType consensus.MessageType) bool {
	finished := false
	switch msgType {
	case MtBlockBody:
		if consensusState.Status(SrStartRound) != spos.SsFinished {
			return finished
		}
	case MtBlockHeader:
		if consensusState.Status(SrStartRound) != spos.SsFinished {
			return finished
		}
	case MtSignature:
		if consensusState.Status(SrBlock) != spos.SsFinished {
			return finished
		}
	}
	return true
}
