package consensus

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/p2p"
)

// SposWorkerMock -
type SposWorkerMock struct {
	AddReceivedMessageCallCalled func(
		messageType consensus.MessageType,
		receivedMessageCall func(ctx context.Context, cnsDta *consensus.Message) bool,
	)
	AddReceivedHeaderHandlerCalled         func(handler func(data.HeaderHandler))
	RemoveAllReceivedHeaderHandlersCalled  func()
	AddReceivedProofHandlerCalled          func(handler func(proofHandler consensus.ProofHandler))
	RemoveAllReceivedMessagesCallsCalled   func()
	ProcessReceivedMessageCalled           func(message p2p.MessageP2P) ([]byte, error)
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
	ResetConsensusStateCalled              func()
	ReceivedProofCalled                    func(proofHandler consensus.ProofHandler)
	ResetConsensusRoundStateCalled         func()
	ResetInvalidSignersCacheCalled         func()
}

// ResetConsensusRoundState -
func (sposWorkerMock *SposWorkerMock) ResetConsensusRoundState() {
	if sposWorkerMock.ResetConsensusRoundStateCalled != nil {
		sposWorkerMock.ResetConsensusRoundStateCalled()
	}
}

// AddReceivedMessageCall -
func (sposWorkerMock *SposWorkerMock) AddReceivedMessageCall(messageType consensus.MessageType,
	receivedMessageCall func(ctx context.Context, cnsDta *consensus.Message) bool) {
	if sposWorkerMock.AddReceivedMessageCallCalled != nil {
		sposWorkerMock.AddReceivedMessageCallCalled(messageType, receivedMessageCall)
	}
}

// AddReceivedHeaderHandler -
func (sposWorkerMock *SposWorkerMock) AddReceivedHeaderHandler(handler func(data.HeaderHandler)) {
	if sposWorkerMock.AddReceivedHeaderHandlerCalled != nil {
		sposWorkerMock.AddReceivedHeaderHandlerCalled(handler)
	}
}

// RemoveAllReceivedHeaderHandlers -
func (sposWorkerMock *SposWorkerMock) RemoveAllReceivedHeaderHandlers() {
	if sposWorkerMock.RemoveAllReceivedHeaderHandlersCalled != nil {
		sposWorkerMock.RemoveAllReceivedHeaderHandlersCalled()
	}
}

func (sposWorkerMock *SposWorkerMock) AddReceivedProofHandler(handler func(proofHandler consensus.ProofHandler)) {
	if sposWorkerMock.AddReceivedProofHandlerCalled != nil {
		sposWorkerMock.AddReceivedProofHandlerCalled(handler)
	}
}

// RemoveAllReceivedMessagesCalls -
func (sposWorkerMock *SposWorkerMock) RemoveAllReceivedMessagesCalls() {
	if sposWorkerMock.RemoveAllReceivedMessagesCallsCalled != nil {
		sposWorkerMock.RemoveAllReceivedMessagesCallsCalled()
	}
}

// ProcessReceivedMessage -
func (sposWorkerMock *SposWorkerMock) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID, _ p2p.MessageHandler) ([]byte, error) {
	if sposWorkerMock.ProcessReceivedMessageCalled == nil {
		return sposWorkerMock.ProcessReceivedMessageCalled(message)
	}
	return nil, nil
}

// SendConsensusMessage -
func (sposWorkerMock *SposWorkerMock) SendConsensusMessage(cnsDta *consensus.Message) bool {
	if sposWorkerMock.SendConsensusMessageCalled != nil {
		return sposWorkerMock.SendConsensusMessageCalled(cnsDta)
	}
	return false
}

// Extend -
func (sposWorkerMock *SposWorkerMock) Extend(subroundId int) {
	if sposWorkerMock.ExtendCalled != nil {
		sposWorkerMock.ExtendCalled(subroundId)
	}
}

// GetConsensusStateChangedChannel -
func (sposWorkerMock *SposWorkerMock) GetConsensusStateChangedChannel() chan bool {
	if sposWorkerMock.GetConsensusStateChangedChannelsCalled != nil {
		return sposWorkerMock.GetConsensusStateChangedChannelsCalled()
	}

	return nil
}

// BroadcastBlock -
func (sposWorkerMock *SposWorkerMock) BroadcastBlock(body data.BodyHandler, header data.HeaderHandler) error {
	if sposWorkerMock.GetBroadcastBlockCalled != nil {
		return sposWorkerMock.GetBroadcastBlockCalled(body, header)
	}
	return nil
}

// ExecuteStoredMessages -
func (sposWorkerMock *SposWorkerMock) ExecuteStoredMessages() {
	if sposWorkerMock.ExecuteStoredMessagesCalled != nil {
		sposWorkerMock.ExecuteStoredMessagesCalled()
	}
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

// ResetConsensusState -
func (sposWorkerMock *SposWorkerMock) ResetConsensusState() {
	if sposWorkerMock.ResetConsensusStateCalled != nil {
		sposWorkerMock.ResetConsensusStateCalled()
	}
}

// ReceivedProof -
func (sposWorkerMock *SposWorkerMock) ReceivedProof(proofHandler consensus.ProofHandler) {
	if sposWorkerMock.ReceivedProofCalled != nil {
		sposWorkerMock.ReceivedProofCalled(proofHandler)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (sposWorkerMock *SposWorkerMock) IsInterfaceNil() bool {
	return sposWorkerMock == nil
}

// ResetInvalidSignersCache -
func (sposWorkerMock *SposWorkerMock) ResetInvalidSignersCache() {
	if sposWorkerMock.ResetInvalidSignersCacheCalled != nil {
		sposWorkerMock.ResetInvalidSignersCacheCalled()
	}
}
