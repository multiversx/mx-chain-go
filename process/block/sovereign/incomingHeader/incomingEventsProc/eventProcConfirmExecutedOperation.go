package incomingEventsProc

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
)

const (
	hashOfHashesIndex         = 1
	hashOfOperationIndex      = 2
	numExecutedBridgeOpTopics = 3
)

type eventProcConfirmExecutedOperation struct {
}

func NewEventProcConfirmExecutedOperation() *eventProcConfirmExecutedOperation {
	return &eventProcConfirmExecutedOperation{}
}

// ProcessEvent will process events related to confirmed outgoing bridge operations to main chain
func (ep *eventProcConfirmExecutedOperation) ProcessEvent(event data.EventHandler) (*dto.EventResult, error) {
	confirmedOp, err := getConfirmedBridgeOperation(event.GetTopics())
	if err != nil {
		return nil, err
	}

	return &dto.EventResult{
		ConfirmedBridgeOp: confirmedOp,
	}, nil
}

func getConfirmedBridgeOperation(topics [][]byte) (*dto.ConfirmedBridgeOp, error) {
	if len(topics) != numExecutedBridgeOpTopics {
		return nil, fmt.Errorf("%w for %s; num topics = %d", dto.ErrInvalidNumTopicsIncomingEvent, dto.EventIDExecutedOutGoingBridgeOp, len(topics))
	}
	if string(topics[0]) != dto.TopicIDConfirmedOutGoingOperation {
		return nil, dto.ErrInvalidIncomingTopicIdentifier
	}

	return &dto.ConfirmedBridgeOp{
		HashOfHashes: topics[hashOfHashesIndex],
		Hash:         topics[hashOfOperationIndex],
	}, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ep *eventProcConfirmExecutedOperation) IsInterfaceNil() bool {
	return ep == nil
}
