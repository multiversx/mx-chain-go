package incomingEventsProc

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
)

type eventProcExecutedDepositOperation struct {
	depositEventProc                  IncomingEventHandler
	eventProcConfirmExecutedOperation IncomingEventHandler
}

func NewEventProcExecutedDepositOperation(
	depositEventProc IncomingEventHandler,
	eventProcConfirmExecutedOperation IncomingEventHandler,
) (*eventProcExecutedDepositOperation, error) {

	return &eventProcExecutedDepositOperation{
		depositEventProc:                  depositEventProc,
		eventProcConfirmExecutedOperation: eventProcConfirmExecutedOperation,
	}, nil
}

// ProcessEvent will process events related to confirmed outgoing bridge operations to main chain
func (ep *eventProcExecutedDepositOperation) ProcessEvent(event data.EventHandler) (*dto.EventResult, error) {
	topics := event.GetTopics()
	if len(topics) == 0 {
		return nil, fmt.Errorf("%w for event id: %s", dto.ErrInvalidNumTopicsIncomingEvent, dto.EventIDExecutedOutGoingBridgeOp)
	}

	switch string(topics[0]) {
	case dto.TopicIDDepositIncomingTransfer:
		return ep.depositEventProc.ProcessEvent(event)
	case dto.TopicIDConfirmedOutGoingOperation:
		return ep.eventProcConfirmExecutedOperation.ProcessEvent(event)
	default:
		return nil, dto.ErrInvalidIncomingTopicIdentifier
	}
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ep *eventProcExecutedDepositOperation) IsInterfaceNil() bool {
	return ep == nil
}
