package incomingHeader

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
)

type eventProcConfirmExecutedOperation struct {
}

// ProcessEvent will process events related to confirmed outgoing bridge operations to main chain
func (ep *eventProcConfirmExecutedOperation) ProcessEvent(event data.EventHandler) (*EventResult, error) {
	confirmedOp, err := getConfirmedBridgeOperation(event.GetTopics())
	if err != nil {
		return nil, err
	}

	return &EventResult{
		ConfirmedBridgeOp: confirmedOp,
	}, nil
}

func getConfirmedBridgeOperation(topics [][]byte) (*ConfirmedBridgeOp, error) {
	if len(topics) != numExecutedBridgeOpTopics {
		return nil, fmt.Errorf("%w for %s; num topics = %d", errInvalidNumTopicsIncomingEvent, eventIDExecutedOutGoingBridgeOp, len(topics))
	}
	if string(topics[0]) != topicIDConfirmedOutGoingOperation {
		return nil, errInvalidIncomingTopicIdentifier
	}

	return &ConfirmedBridgeOp{
		HashOfHashes: topics[hashOfHashesIndex],
		Hash:         topics[hashOfOperationIndex],
	}, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ep *eventProcConfirmExecutedOperation) IsInterfaceNil() bool {
	return ep == nil
}
