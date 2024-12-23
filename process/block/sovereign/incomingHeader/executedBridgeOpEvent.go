package incomingHeader

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
)

type executedEventPoc struct {
	depositEventProc IncomingEventHandler
}

func (eep *executedEventPoc) ProcessEvent(event data.EventHandler) (*EventResult, error) {
	topics := event.GetTopics()
	if len(topics) == 0 {
		return nil, fmt.Errorf("%w for event id: %s", errInvalidNumTopicsIncomingEvent, eventIDExecutedOutGoingBridgeOp)
	}

	var resDeposit *EventResult
	var confirmedOp *confirmedBridgeOp
	var err error
	switch string(topics[0]) {
	case topicIDDepositIncomingTransfer:
		resDeposit, err = eep.depositEventProc.ProcessEvent(event)
	case topicIDConfirmedOutGoingOperation:
		confirmedOp, err = getConfirmedBridgeOperation(topics)
	default:
		return nil, errInvalidIncomingTopicIdentifier
	}

	if err != nil {
		return nil, err
	}

	return &EventResult{
		SCR:               resDeposit.SCR,
		ConfirmedBridgeOp: confirmedOp,
	}, nil
}

func getConfirmedBridgeOperation(topics [][]byte) (*confirmedBridgeOp, error) {
	if len(topics) != numExecutedBridgeOpTopics {
		return nil, fmt.Errorf("%w for %s; num topics = %d", errInvalidNumTopicsIncomingEvent, eventIDExecutedOutGoingBridgeOp, len(topics))
	}

	return &confirmedBridgeOp{
		hashOfHashes: topics[hashOfHashesIndex],
		hash:         topics[hashOfOperationIndex],
	}, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (eep *executedEventPoc) IsInterfaceNil() bool {
	return eep == nil
}
