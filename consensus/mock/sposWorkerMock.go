package mock

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// SposWorkerMock -
type SposWorkerMock struct {
	AddReceivedMessageCallCalled func(
		messageType consensus.MessageType,
		receivedMessageCall func(cnsDta *consensus.Message) bool,
	)
	AddReceivedHeaderHandlerCalled         func(handler func(data.HeaderHandler))
	RemoveAllReceivedMessagesCallsCalled   func()
	ProcessReceivedMessageCalled           func(message p2p.MessageP2P) error
	SendConsensusMessageCalled             func(cnsDta *consensus.Message) bool
	ExtendCalled                           func(subroundId int)
	GetConsensusStateChangedChannelsCalled func() chan bool
	GetBroadcastBlockCalled                func(data.BodyHandler, data.HeaderHandler) error
	GetBroadcastHeaderCalled               func(data.HeaderHandler) error
	ExecuteStoredMessagesCalled            func()
	DisplayStatisticsCalled                func()
	ReceivedHeaderCalled                   func(headerHandler data.HeaderHandler, headerHash []byte)
	SetAppStatusHandlerCalled              func(ash core.AppStatusHandler) error
	ResetConsensusMessagesCalled           func()
}

// AddReceivedMessageCall -
func (sposWorkerMock *SposWorkerMock) AddReceivedMessageCall(messageType consensus.MessageType,
	receivedMessageCall func(cnsDta *consensus.Message) bool) {
	sposWorkerMock.AddReceivedMessageCallCalled(messageType, receivedMessageCall)
}

// AddReceivedHeaderHandler -
func (sposWorkerMock *SposWorkerMock) AddReceivedHeaderHandler(handler func(data.HeaderHandler)) {
	if sposWorkerMock.AddReceivedHeaderHandlerCalled != nil {
		sposWorkerMock.AddReceivedHeaderHandlerCalled(handler)
	}
}

// RemoveAllReceivedMessagesCalls -
func (sposWorkerMock *SposWorkerMock) RemoveAllReceivedMessagesCalls() {
	sposWorkerMock.RemoveAllReceivedMessagesCallsCalled()
}

// ProcessReceivedMessage -
func (sposWorkerMock *SposWorkerMock) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	return sposWorkerMock.ProcessReceivedMessageCalled(message)
}

// SendConsensusMessage -
func (sposWorkerMock *SposWorkerMock) SendConsensusMessage(cnsDta *consensus.Message) bool {
	return sposWorkerMock.SendConsensusMessageCalled(cnsDta)
}

// Extend -
func (sposWorkerMock *SposWorkerMock) Extend(subroundId int) {
	sposWorkerMock.ExtendCalled(subroundId)
}

// GetConsensusStateChangedChannel -
func (sposWorkerMock *SposWorkerMock) GetConsensusStateChangedChannel() chan bool {
	return sposWorkerMock.GetConsensusStateChangedChannelsCalled()
}

// BroadcastBlock -
func (sposWorkerMock *SposWorkerMock) BroadcastBlock(body data.BodyHandler, header data.HeaderHandler) error {
	return sposWorkerMock.GetBroadcastBlockCalled(body, header)
}

// ExecuteStoredMessages -
func (sposWorkerMock *SposWorkerMock) ExecuteStoredMessages() {
	sposWorkerMock.ExecuteStoredMessagesCalled()
}

// DisplayStatistics -
func (sposWorkerMock *SposWorkerMock) DisplayStatistics() {
	if sposWorkerMock.DisplayStatisticsCalled != nil {
		sposWorkerMock.DisplayStatisticsCalled()
	}
}

// ReceivedHeader -
func (sposWorkerMock *SposWorkerMock) ReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	if sposWorkerMock.ReceivedHeaderCalled != nil {
		sposWorkerMock.ReceivedHeaderCalled(headerHandler, headerHash)
	}
}

// Close -
func (sposWorkerMock *SposWorkerMock) Close() error {
	return nil
}

// StartWorking -
func (sposWorkerMock *SposWorkerMock) StartWorking() {
}

// ResetConsensusMessages -
func (sposWorkerMock *SposWorkerMock) ResetConsensusMessages() {
	if sposWorkerMock.ResetConsensusMessagesCalled != nil {
		sposWorkerMock.ResetConsensusMessagesCalled()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (sposWorkerMock *SposWorkerMock) IsInterfaceNil() bool {
	return sposWorkerMock == nil
}
