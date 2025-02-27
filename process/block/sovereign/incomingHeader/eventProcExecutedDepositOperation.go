package incomingHeader

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
)

type eventProcExecutedDepositOperation struct {
	depositEventProc                  IncomingEventHandler
	eventProcConfirmExecutedOperation IncomingEventHandler
}

// ProcessEvent will process events related to confirmed outgoing bridge operations to main chain
func (ep *eventProcExecutedDepositOperation) ProcessEvent(event data.EventHandler) (*EventResult, error) {
	topics := event.GetTopics()
	if len(topics) == 0 {
		return nil, fmt.Errorf("%w for event id: %s", errInvalidNumTopicsIncomingEvent, eventIDExecutedOutGoingBridgeOp)
	}

	switch string(topics[0]) {
	case topicIDDepositIncomingTransfer:
		return ep.depositEventProc.ProcessEvent(event)
	case topicIDConfirmedOutGoingOperation:
		return ep.eventProcConfirmExecutedOperation.ProcessEvent(event)
	default:
		return nil, errInvalidIncomingTopicIdentifier
	}
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ep *eventProcExecutedDepositOperation) IsInterfaceNil() bool {
	return ep == nil
}
