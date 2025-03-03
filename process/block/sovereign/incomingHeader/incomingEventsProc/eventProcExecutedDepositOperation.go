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

// ProcessEvent handles events related to executed outgoing deposit token bridge operations to the main chain.
// Each event is identified by dto.EventIDExecutedOutGoingBridgeOp.
//
// An outgoing deposit operation can have two possible outcomes:
//   - successful – The operation is confirmed.
//   - failed – The tokens are returned to the original caller.
//     ⚠ Note: The current method of returning tokens for failed cases will be revised in future versions.
//
// Expected event topics ([][]byte):
// - topic[0] = dto.TopicIDDepositIncomingTransfer → Indicates a failed operation.
//   - The event is treated as an **incoming deposit event**, and the tokens are returned.
//   - The remaining topic fields follow the format defined in `eventProcDepositTokens.go`.
//
// - topic[1] = dto.TopicIDConfirmedOutGoingOperation → Indicates a successful operation.
//   - The operation is confirmed.
//   - The remaining topic fields follow the format defined in `eventProcConfirmExecutedOperation.go`.
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
