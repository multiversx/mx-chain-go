package v1

import (
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
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
	receivedMessages[bls.MtBlockBodyAndHeader] = make([]*consensus.Message, 0)
	receivedMessages[bls.MtBlockBody] = make([]*consensus.Message, 0)
	receivedMessages[bls.MtBlockHeader] = make([]*consensus.Message, 0)
	receivedMessages[bls.MtSignature] = make([]*consensus.Message, 0)
	receivedMessages[bls.MtBlockHeaderFinalInfo] = make([]*consensus.Message, 0)
	receivedMessages[bls.MtInvalidSigners] = make([]*consensus.Message, 0)

	return receivedMessages
}

// GetMaxMessagesInARoundPerPeer returns the maximum number of messages a peer can send per round for BLS
func (wrk *worker) GetMaxMessagesInARoundPerPeer() uint32 {
	return peerMaxMessagesPerSec
}

// GetStringValue gets the name of the messageType
func (wrk *worker) GetStringValue(messageType consensus.MessageType) string {
	return bls.GetStringValue(messageType)
}

// GetSubroundName gets the subround name for the subround id provided
func (wrk *worker) GetSubroundName(subroundId int) string {
	return bls.GetSubroundName(subroundId)
}

// IsMessageWithBlockBodyAndHeader returns if the current messageType is about block body and header
func (wrk *worker) IsMessageWithBlockBodyAndHeader(msgType consensus.MessageType) bool {
	return msgType == bls.MtBlockBodyAndHeader
}

// IsMessageWithBlockBody returns if the current messageType is about block body
func (wrk *worker) IsMessageWithBlockBody(msgType consensus.MessageType) bool {
	return msgType == bls.MtBlockBody
}

// IsMessageWithBlockHeader returns if the current messageType is about block header
func (wrk *worker) IsMessageWithBlockHeader(msgType consensus.MessageType) bool {
	return msgType == bls.MtBlockHeader
}

// IsMessageWithSignature returns if the current messageType is about signature
func (wrk *worker) IsMessageWithSignature(msgType consensus.MessageType) bool {
	return msgType == bls.MtSignature
}

// IsMessageWithFinalInfo returns if the current messageType is about header final info
func (wrk *worker) IsMessageWithFinalInfo(msgType consensus.MessageType) bool {
	return msgType == bls.MtBlockHeaderFinalInfo
}

// IsMessageWithInvalidSigners returns if the current messageType is about invalid signers
func (wrk *worker) IsMessageWithInvalidSigners(msgType consensus.MessageType) bool {
	return msgType == bls.MtInvalidSigners
}

// IsMessageTypeValid returns if the current messageType is valid
func (wrk *worker) IsMessageTypeValid(msgType consensus.MessageType) bool {
	isMessageTypeValid := msgType == bls.MtBlockBodyAndHeader ||
		msgType == bls.MtBlockBody ||
		msgType == bls.MtBlockHeader ||
		msgType == bls.MtSignature ||
		msgType == bls.MtBlockHeaderFinalInfo ||
		msgType == bls.MtInvalidSigners

	return isMessageTypeValid
}

// IsSubroundSignature returns if the current subround is about signature
func (wrk *worker) IsSubroundSignature(subroundId int) bool {
	return subroundId == bls.SrSignature
}

// IsSubroundStartRound returns if the current subround is about start round
func (wrk *worker) IsSubroundStartRound(subroundId int) bool {
	return subroundId == bls.SrStartRound
}

// GetMessageRange provides the MessageType range used in checks by the consensus
func (wrk *worker) GetMessageRange() []consensus.MessageType {
	var v []consensus.MessageType

	for i := bls.MtBlockBodyAndHeader; i <= bls.MtInvalidSigners; i++ {
		v = append(v, i)
	}

	return v
}

// CanProceed returns if the current messageType can proceed further if previous subrounds finished
func (wrk *worker) CanProceed(consensusState *spos.ConsensusState, msgType consensus.MessageType) bool {
	switch msgType {
	case bls.MtBlockBodyAndHeader:
		return consensusState.Status(bls.SrStartRound) == spos.SsFinished
	case bls.MtBlockBody:
		return consensusState.Status(bls.SrStartRound) == spos.SsFinished
	case bls.MtBlockHeader:
		return consensusState.Status(bls.SrStartRound) == spos.SsFinished
	case bls.MtSignature:
		return consensusState.Status(bls.SrBlock) == spos.SsFinished
	case bls.MtBlockHeaderFinalInfo:
		return consensusState.Status(bls.SrSignature) == spos.SsFinished
	case bls.MtInvalidSigners:
		return consensusState.Status(bls.SrSignature) == spos.SsFinished
	}

	return false
}

// GetMaxNumOfMessageTypeAccepted returns the maximum number of accepted consensus message types per round, per public key
func (wrk *worker) GetMaxNumOfMessageTypeAccepted(msgType consensus.MessageType) uint32 {
	if msgType == bls.MtSignature {
		return maxNumOfMessageTypeSignatureAccepted
	}

	return defaultMaxNumOfMessageTypeAccepted
}

// IsInterfaceNil returns true if there is no value under the interface
func (wrk *worker) IsInterfaceNil() bool {
	return wrk == nil
}
