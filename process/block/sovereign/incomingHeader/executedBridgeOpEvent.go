package incomingHeader

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
)

type executedEventPoc struct {
	depositEventProc IncomingEventHandler
}

func (eep *executedEventPoc) ProcessEvent(event data.EventHandler) (*eventsResult, error) {
	topics := event.GetTopics()
	if len(topics) == 0 {
		return nil, fmt.Errorf("%w for event id: %s", errInvalidNumTopicsIncomingEvent, eventIDExecutedOutGoingBridgeOp)
	}

	var scr *scrInfo
	var confirmedOp *confirmedBridgeOp
	var err error
	switch string(topics[0]) {
	case topicIDDepositIncomingTransfer:
		resDeposit, errDeposit := eep.depositEventProc.ProcessEvent(event)
		scr, err = resDeposit.scrs[0], errDeposit
	case topicIDConfirmedOutGoingOperation:
		confirmedOp, err = getConfirmedBridgeOperation(topics)
	default:
		return nil, errInvalidIncomingTopicIdentifier
	}

	if err != nil {
		return nil, err
	}

	return &eventsResult{
		scrs:               []*scrInfo{scr},
		confirmedBridgeOps: []*confirmedBridgeOp{confirmedOp},
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
