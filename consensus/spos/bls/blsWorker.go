package bls

import (
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

// peerMaxMessagesPerSec defines how many messages can be propagated by a pid in a round. The value was chosen by
// following the next premises:
//  1. a leader can propagate as maximum as 3 messages per round: proposed header block + proposed body + final info;
//  2. due to the fact that a delayed signature of the proposer (from previous round) can be received in the current round
//     adds an extra 1 to the total value, reaching value 4;
//  3. Because the leader might be selected in the next round and might have an empty data pool, it can send the newly
//     empty proposed block at the very beginning of the next round. One extra message here, yielding to a total of 5.
//  4. If we consider the forks that can appear on the system wee need to add one more to the value.
//
// Validators only send one signature message in a round, treating the edge case of a delayed message, will need at most
// 2 messages per round (which is ok as it is below the set value of 5)
const peerMaxMessagesPerSec = uint32(6)

// defaultMaxNumOfMessageTypeAccepted represents the maximum number of the same message type accepted in one round to be
// received from the same public key for the default message types
const defaultMaxNumOfMessageTypeAccepted = uint32(1)

// maxNumOfMessageTypeSignatureAccepted represents the maximum number of the signature message type accepted in one round to be
// received from the same public key
const maxNumOfMessageTypeSignatureAccepted = uint32(2)

// worker defines the data needed by spos to communicate between nodes which are in the validators group
type worker struct {
}

// NewConsensusService creates a new worker object
func NewConsensusService() (*worker, error) {
	wrk := worker{}

	return &wrk, nil
}

// InitReceivedMessages initializes the MessagesType map for all messages for the current ConsensusService
func (wrk *worker) InitReceivedMessages() map[consensus.MessageType][]*consensus.Message {
	receivedMessages := make(map[consensus.MessageType][]*consensus.Message)
	receivedMessages[MtBlockBodyAndHeader] = make([]*consensus.Message, 0)
	receivedMessages[MtBlockBody] = make([]*consensus.Message, 0)
	receivedMessages[MtBlockHeader] = make([]*consensus.Message, 0)
	receivedMessages[MtSignature] = make([]*consensus.Message, 0)
	receivedMessages[MtBlockHeaderFinalInfo] = make([]*consensus.Message, 0)
	receivedMessages[MtInvalidSigners] = make([]*consensus.Message, 0)

	return receivedMessages
}

// GetMaxMessagesInARoundPerPeer returns the maximum number of messages a peer can send per round for BLS
func (wrk *worker) GetMaxMessagesInARoundPerPeer() uint32 {
	return peerMaxMessagesPerSec
}

// GetStringValue gets the name of the messageType
func (wrk *worker) GetStringValue(messageType consensus.MessageType) string {
	return getStringValue(messageType)
}

// GetSubroundName gets the subround name for the subround id provided
func (wrk *worker) GetSubroundName(subroundId int) string {
	return getSubroundName(subroundId)
}

// IsMessageWithBlockBodyAndHeader returns if the current messageType is about block body and header
func (wrk *worker) IsMessageWithBlockBodyAndHeader(msgType consensus.MessageType) bool {
	return msgType == MtBlockBodyAndHeader
}

// IsMessageWithBlockBody returns if the current messageType is about block body
func (wrk *worker) IsMessageWithBlockBody(msgType consensus.MessageType) bool {
	return msgType == MtBlockBody
}

// IsMessageWithBlockHeader returns if the current messageType is about block header
func (wrk *worker) IsMessageWithBlockHeader(msgType consensus.MessageType) bool {
	return msgType == MtBlockHeader
}

// IsMessageWithSignature returns if the current messageType is about signature
func (wrk *worker) IsMessageWithSignature(msgType consensus.MessageType) bool {
	return msgType == MtSignature
}

// IsMessageWithFinalInfo returns if the current messageType is about header final info
func (wrk *worker) IsMessageWithFinalInfo(msgType consensus.MessageType) bool {
	return msgType == MtBlockHeaderFinalInfo
}

// IsMessageWithInvalidSigners returns if the current messageType is about invalid signers
func (wrk *worker) IsMessageWithInvalidSigners(msgType consensus.MessageType) bool {
	return msgType == MtInvalidSigners
}

// IsMessageTypeValid returns if the current messageType is valid
func (wrk *worker) IsMessageTypeValid(msgType consensus.MessageType) bool {
	isMessageTypeValid := msgType == MtBlockBodyAndHeader ||
		msgType == MtBlockBody ||
		msgType == MtBlockHeader ||
		msgType == MtSignature ||
		msgType == MtBlockHeaderFinalInfo ||
		msgType == MtInvalidSigners

	return isMessageTypeValid
}

// IsSubroundSignature returns if the current subround is about signature
func (wrk *worker) IsSubroundSignature(subroundId int) bool {
	return subroundId == SrSignature
}

// IsSubroundStartRound returns if the current subround is about start round
func (wrk *worker) IsSubroundStartRound(subroundId int) bool {
	return subroundId == SrStartRound
}

// GetMessageRange provides the MessageType range used in checks by the consensus
func (wrk *worker) GetMessageRange() []consensus.MessageType {
	var v []consensus.MessageType

	for i := MtBlockBodyAndHeader; i <= MtInvalidSigners; i++ {
		v = append(v, i)
	}

	return v
}

// CanProceed returns if the current messageType can proceed further if previous subrounds finished
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
	case MtInvalidSigners:
		return consensusState.Status(SrSignature) == spos.SsFinished
	}

	return false
}

// GetMaxNumOfMessageTypeAccepted returns the maximum number of accepted consensus message types per round, per public key
func (wrk *worker) GetMaxNumOfMessageTypeAccepted(msgType consensus.MessageType) uint32 {
	if msgType == MtSignature {
		return maxNumOfMessageTypeSignatureAccepted
	}

	return defaultMaxNumOfMessageTypeAccepted
}

// IsInterfaceNil returns true if there is no value under the interface
func (wrk *worker) IsInterfaceNil() bool {
	return wrk == nil
}
