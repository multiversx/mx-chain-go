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

// ProcessEvent handles incoming events related to confirmed outgoing bridge operations to the main chain.
// Each confirmed event is identified by a unique identifier (e.g., confirming that a deposit or
// changeValidatorSet operation has been executed).
//
// The event's topics ([][]byte) are expected to be:
// 1. topic[0] = executedBridgeOp – Indicates the execution of a bridge operation.
// 2. topic[1] = hashOfHashes – A hash representing all outgoing operations.
// 3. topic[2] = hashOfOperation – The hash of a specific operation included in `HashOfHashes`.
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
