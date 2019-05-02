package bn

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

func (wrk *worker) InitReceivedMessages() map[spos.MessageType][]*consensus.Message {

	receivedMessages := make(map[spos.MessageType][]*consensus.Message)
	receivedMessages[MtBlockBody] = make([]*consensus.Message, 0)
	receivedMessages[MtBlockHeader] = make([]*consensus.Message, 0)
	receivedMessages[MtCommitmentHash] = make([]*consensus.Message, 0)
	receivedMessages[MtBitmap] = make([]*consensus.Message, 0)
	receivedMessages[MtCommitment] = make([]*consensus.Message, 0)
	receivedMessages[MtSignature] = make([]*consensus.Message, 0)

	return receivedMessages

}

func (wrk *worker) GetStringValue(messageType spos.MessageType) string {
	return getStringValue(messageType)
}

func (wrk *worker) GetSubroundName(subroundId int) string {
	return getSubroundName(subroundId)
}
func (wrk *worker) GetMessageRange() []spos.MessageType {
	v := []spos.MessageType{}

	for i := MtBlockBody; i <= MtSignature; i++ {
		v = append(v, i)
	}

	return v
}

func (wrk *worker) IsFinished(consensusState *spos.ConsensusState, msgType spos.MessageType) bool {
	finished := true
	switch msgType {
	case MtBlockBody:
		if consensusState.Status(SrStartRound) != spos.SsFinished {
			return finished
		}
	case MtBlockHeader:
		if consensusState.Status(SrStartRound) != spos.SsFinished {
			return finished
		}
	case MtCommitmentHash:
		if consensusState.Status(SrBlock) != spos.SsFinished {
			return finished
		}
	case MtBitmap:
		if consensusState.Status(SrBlock) != spos.SsFinished {
			return finished
		}
	case MtCommitment:
		if consensusState.Status(SrBitmap) != spos.SsFinished {
			return finished
		}
	case MtSignature:
		if consensusState.Status(SrBitmap) != spos.SsFinished {
			return finished
		}
	}
	return false
}
