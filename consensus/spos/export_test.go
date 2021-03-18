package spos

import (
	"fmt"

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

func (wrk *Worker) RoundHandler() consensus.RoundHandler {
	return wrk.roundHandler
}

func (wrk *Worker) SetNodeRedundancyHandler(nodeRedundancyHandler consensus.NodeRedundancyHandler) {
	wrk.nodeRedundancyHandler = nodeRedundancyHandler
}

func (wrk *Worker) SetRoundHandler(roundHandler consensus.RoundHandler) {
	wrk.roundHandler = roundHandler
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

// consensusMessageValidator

func (cmv *consensusMessageValidator) CheckConsensusMessageValidity(cnsMsg *consensus.Message, originator core.PeerID) error {
	return cmv.checkConsensusMessageValidity(cnsMsg, originator)
}

func (cmv *consensusMessageValidator) CheckMessageWithFinalInfoValidity(cnsMsg *consensus.Message) error {
	return cmv.checkMessageWithFinalInfoValidity(cnsMsg)
}

func (cmv *consensusMessageValidator) CheckMessageWithSignatureValidity(cnsMsg *consensus.Message) error {
	return cmv.checkMessageWithSignatureValidity(cnsMsg)
}

func (cmv *consensusMessageValidator) CheckMessageWithBlockHeaderValidity(cnsMsg *consensus.Message) error {
	return cmv.checkMessageWithBlockHeaderValidity(cnsMsg)
}

func (cmv *consensusMessageValidator) CheckMessageWithBlockBodyValidity(cnsMsg *consensus.Message) error {
	return cmv.checkMessageWithBlockBodyValidity(cnsMsg)
}

func (cmv *consensusMessageValidator) CheckMessageWithBlockBodyAndHeaderValidity(cnsMsg *consensus.Message) error {
	return cmv.checkMessageWithBlockBodyAndHeaderValidity(cnsMsg)
}

func (cmv *consensusMessageValidator) CheckConsensusMessageValidityForMessageType(cnsMsg *consensus.Message) error {
	return cmv.checkConsensusMessageValidityForMessageType(cnsMsg)
}

func (cmv *consensusMessageValidator) IsBlockHeaderHashSizeValid(cnsMsg *consensus.Message) bool {
	return cmv.isBlockHeaderHashSizeValid(cnsMsg)
}

func (cmv *consensusMessageValidator) AddMessageTypeToPublicKey(pk []byte, round int64, msgType consensus.MessageType) {
	cmv.addMessageTypeToPublicKey(pk, round, msgType)
}

func (cmv *consensusMessageValidator) IsMessageTypeLimitReached(pk []byte, round int64, msgType consensus.MessageType) bool {
	return cmv.isMessageTypeLimitReached(pk, round, msgType)
}

func (cmv *consensusMessageValidator) GetNumOfMessageTypeForPublicKey(pk []byte, round int64, msgType consensus.MessageType) uint32 {
	cmv.mutPkConsensusMessages.RLock()
	defer cmv.mutPkConsensusMessages.RUnlock()

	key := fmt.Sprintf("%s_%d", string(pk), round)

	mapMsgType, ok := cmv.mapPkConsensusMessages[key]
	if !ok {
		return uint32(0)
	}

	return mapMsgType[msgType]
}

func (cmv *consensusMessageValidator) ResetConsensusMessages() {
	cmv.resetConsensusMessages()
}
