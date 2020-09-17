package spos

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

type RoundConsensus struct {
	*roundConsensus
}

func NewRoundConsensusWrapper(rcns *roundConsensus) *RoundConsensus {
	return &RoundConsensus{
		roundConsensus: rcns,
	}
}

// worker

func (wrk *Worker) BlockProcessor() process.BlockProcessor {
	return wrk.blockProcessor
}

func (wrk *Worker) SetBlockProcessor(blockProcessor process.BlockProcessor) {
	wrk.blockProcessor = blockProcessor
}

func (wrk *Worker) Bootstrapper() process.Bootstrapper {
	return wrk.bootstrapper
}

func (wrk *Worker) SetBootstrapper(bootstrapper process.Bootstrapper) {
	wrk.bootstrapper = bootstrapper
}

func (wrk *Worker) BroadcastMessenger() consensus.BroadcastMessenger {
	return wrk.broadcastMessenger
}

func (wrk *Worker) SetBroadcastMessenger(broadcastMessenger consensus.BroadcastMessenger) {
	wrk.broadcastMessenger = broadcastMessenger
}

func (wrk *Worker) ConsensusState() *ConsensusState {
	return wrk.consensusState
}

func (wrk *Worker) SetConsensusState(consensusState *ConsensusState) {
	wrk.consensusState = consensusState
}

func (wrk *Worker) ForkDetector() process.ForkDetector {
	return wrk.forkDetector
}

func (wrk *Worker) SetForkDetector(forkDetector process.ForkDetector) {
	wrk.forkDetector = forkDetector
}

func (wrk *Worker) Marshalizer() marshal.Marshalizer {
	return wrk.marshalizer
}

func (wrk *Worker) SetMarshalizer(marshalizer marshal.Marshalizer) {
	wrk.marshalizer = marshalizer
}

func (wrk *Worker) Rounder() consensus.Rounder {
	return wrk.rounder
}

func (wrk *Worker) SetRounder(rounder consensus.Rounder) {
	wrk.rounder = rounder
}

func (wrk *Worker) CheckSignature(cnsData *consensus.Message) error {
	return wrk.peerSignatureHandler.VerifyPeerSignature(cnsData.PubKey, core.PeerID(cnsData.OriginatorPid), cnsData.Signature)
}

func (wrk *Worker) ExecuteMessage(cnsDtaList []*consensus.Message) {
	wrk.executeMessage(cnsDtaList)
}

func (wrk *Worker) InitReceivedMessages() {
	wrk.initReceivedMessages()
}

func (wrk *Worker) ReceivedSyncState(isNodeSynchronized bool) {
	wrk.receivedSyncState(isNodeSynchronized)
}

func (wrk *Worker) ReceivedMessages() map[consensus.MessageType][]*consensus.Message {
	wrk.mutReceivedMessages.RLock()
	defer wrk.mutReceivedMessages.RUnlock()

	return wrk.receivedMessages
}

func (wrk *Worker) SetReceivedMessages(messageType consensus.MessageType, cnsDta []*consensus.Message) {
	wrk.mutReceivedMessages.Lock()
	wrk.receivedMessages[messageType] = cnsDta
	wrk.mutReceivedMessages.Unlock()
}

func (wrk *Worker) NilReceivedMessages() {
	wrk.mutReceivedMessages.Lock()
	wrk.receivedMessages = nil
	wrk.mutReceivedMessages.Unlock()
}

func (wrk *Worker) ReceivedMessagesCalls() map[consensus.MessageType]func(*consensus.Message) bool {
	wrk.mutReceivedMessagesCalls.RLock()
	defer wrk.mutReceivedMessagesCalls.RUnlock()

	return wrk.receivedMessagesCalls
}

func (wrk *Worker) SetReceivedMessagesCalls(messageType consensus.MessageType, f func(*consensus.Message) bool) {
	wrk.mutReceivedMessagesCalls.Lock()
	wrk.receivedMessagesCalls[messageType] = f
	wrk.mutReceivedMessagesCalls.Unlock()
}

func (wrk *Worker) ExecuteMessageChannel() chan *consensus.Message {
	return wrk.executeMessageChannel
}

func (wrk *Worker) ConsensusStateChangedChannel() chan bool {
	return wrk.consensusStateChangedChannel
}

func (wrk *Worker) SetConsensusStateChangedChannel(consensusStateChangedChannel chan bool) {
	wrk.consensusStateChangedChannel = consensusStateChangedChannel
}

func (wrk *Worker) CheckSelfState(cnsDta *consensus.Message) error {
	return wrk.checkSelfState(cnsDta)
}

func (rcns *RoundConsensus) EligibleList() map[string]struct{} {
	return rcns.eligibleNodes
}

func (wrk *Worker) AppStatusHandler() core.AppStatusHandler {
	return wrk.appStatusHandler
}

func (wrk *Worker) CheckConsensusMessageValidity(cnsMsg *consensus.Message, originator core.PeerID) error {
	return wrk.checkConsensusMessageValidity(cnsMsg, originator)
}

func (wrk *Worker) CheckMessageWithFinalInfoValidity(cnsMsg *consensus.Message) error {
	return wrk.checkMessageWithFinalInfoValidity(cnsMsg)
}

func (wrk *Worker) CheckMessageWithSignatureValidity(cnsMsg *consensus.Message) error {
	return wrk.checkMessageWithSignatureValidity(cnsMsg)
}

func (wrk *Worker) CheckMessageWithBlockHeaderValidity(cnsMsg *consensus.Message) error {
	return wrk.checkMessageWithBlockHeaderValidity(cnsMsg)
}

func (wrk *Worker) CheckMessageWithBlockBodyValidity(cnsMsg *consensus.Message) error {
	return wrk.checkMessageWithBlockBodyValidity(cnsMsg)
}

func (wrk *Worker) CheckMessageWithBlockBodyAndHeaderValidity(cnsMsg *consensus.Message) error {
	return wrk.checkMessageWithBlockBodyAndHeaderValidity(cnsMsg)
}

func (wrk *Worker) CheckConsensusMessageValidityForMessageType(cnsMsg *consensus.Message) error {
	return wrk.checkConsensusMessageValidityForMessageType(cnsMsg)
}

func (wrk *Worker) IsBlockHeaderHashSizeValid(cnsMsg *consensus.Message) bool {
	return wrk.isBlockHeaderHashSizeValid(cnsMsg)
}

func (wrk *Worker) AddMessageTypeToPublicKey(pk []byte, msgType consensus.MessageType) {
	wrk.addMessageTypeToPublicKey(pk, msgType)
}

func (wrk *Worker) IsMessageTypeLimitReached(pk []byte, msgType consensus.MessageType) bool {
	return wrk.isMessageTypeLimitReached(pk, msgType)
}

func (wrk *Worker) GetNumOfMessageTypeForPublicKey(pk []byte, msgType consensus.MessageType) uint32 {
	wrk.mutPkConsensusMessages.RLock()
	defer wrk.mutPkConsensusMessages.RUnlock()

	mapMsgType, ok := wrk.mapPkConsensusMessages[string(pk)]
	if !ok {
		return uint32(0)
	}

	return mapMsgType[msgType]
}
