package bls

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
)

// worker defines the data needed by spos to communicate between nodes which are in the validators group
type worker struct {
}

// NewConsensusService creates a new worker object
func NewConsensusService() (*worker, error) {
	wrk := worker{}

	return &wrk, nil
}

//InitReceivedMessages initializes the MessagesType map for all messages for the current ConsensusService
func (wrk *worker) InitReceivedMessages() map[consensus.MessageType][]*consensus.Message {
	receivedMessages := make(map[consensus.MessageType][]*consensus.Message)
	receivedMessages[MtBlockBodyAndHeader] = make([]*consensus.Message, 0)
	receivedMessages[MtBlockBody] = make([]*consensus.Message, 0)
	receivedMessages[MtBlockHeader] = make([]*consensus.Message, 0)
	receivedMessages[MtSignature] = make([]*consensus.Message, 0)
	receivedMessages[MtBlockHeaderFinalInfo] = make([]*consensus.Message, 0)

	return receivedMessages
}

//GetStringValue gets the name of the messageType
func (wrk *worker) GetStringValue(messageType consensus.MessageType) string {
	return getStringValue(messageType)
}

//GetSubroundName gets the subround name for the subround id provided
func (wrk *worker) GetSubroundName(subroundId int) string {
	return getSubroundName(subroundId)
}

//IsMessageWithBlockBodyAndHeader returns if the current messageType is about block body and header
func (wrk *worker) IsMessageWithBlockBodyAndHeader(msgType consensus.MessageType) bool {
	return msgType == MtBlockBodyAndHeader
}

//IsMessageWithBlockHeader returns if the current messageType is about block header
func (wrk *worker) IsMessageWithBlockHeader(msgType consensus.MessageType) bool {
	return msgType == MtBlockHeader
}

//IsMessageWithSignature returns if the current messageType is about signature
func (wrk *worker) IsMessageWithSignature(msgType consensus.MessageType) bool {
	return msgType == MtSignature
}

//IsSubroundSignature returns if the current subround is about signature
func (wrk *worker) IsSubroundSignature(subroundId int) bool {
	return subroundId == SrSignature
}

//IsSubroundStartRound returns if the current subround is about start round
func (wrk *worker) IsSubroundStartRound(subroundId int) bool {
	return subroundId == SrStartRound
}

//GetMessageRange provides the MessageType range used in checks by the consensus
func (wrk *worker) GetMessageRange() []consensus.MessageType {
	var v []consensus.MessageType

	for i := MtBlockBodyAndHeader; i <= MtBlockHeaderFinalInfo; i++ {
		v = append(v, i)
	}

	return v
}

//CanProceed returns if the current messageType can proceed further if previous subrounds finished
func (wrk *worker) CanProceed(consensusState *spos.ConsensusState, msgType consensus.MessageType) bool {
	switch msgType {
	case MtBlockBodyAndHeader:
		return consensusState.Status(SrStartRound) == spos.SsFinished
	case MtBlockBody:
		return consensusState.Status(SrStartRound) == spos.SsFinished
	case MtBlockHeader:
		return consensusState.Status(SrStartRound) == spos.SsFinished
	case MtSignature:
		return consensusState.Status(SrBlock) == spos.SsFinished
	case MtBlockHeaderFinalInfo:
		return consensusState.Status(SrSignature) == spos.SsFinished
	}

	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (wrk *worker) IsInterfaceNil() bool {
	return wrk == nil
}
