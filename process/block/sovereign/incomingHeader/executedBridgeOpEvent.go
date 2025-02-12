package incomingHeader

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
)

type executedBridgeOpEventProc struct {
	depositEventProc IncomingEventHandler
}

// ProcessEvent will process events related to confirmed outgoing bridge operations to main chain
func (eep *executedBridgeOpEventProc) ProcessEvent(event data.EventHandler) (*EventResult, error) {
	topics := event.GetTopics()
	if len(topics) == 0 {
		return nil, fmt.Errorf("%w for event id: %s", errInvalidNumTopicsIncomingEvent, eventIDExecutedOutGoingBridgeOp)
	}

	var confirmedOp *ConfirmedBridgeOp
	var err error
	switch string(topics[0]) {
	case topicIDDepositIncomingTransfer:
		return eep.depositEventProc.ProcessEvent(event)
	case topicIDConfirmedOutGoingOperation:
		confirmedOp, err = getConfirmedBridgeOperation(topics)
	default:
		return nil, errInvalidIncomingTopicIdentifier
	}

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

	return &ConfirmedBridgeOp{
		HashOfHashes: topics[hashOfHashesIndex],
		Hash:         topics[hashOfOperationIndex],
	}, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (eep *executedBridgeOpEventProc) IsInterfaceNil() bool {
	return eep == nil
}
