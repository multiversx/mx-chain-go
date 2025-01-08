package incomingHeader

import (
	"errors"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
)

const (
	minTopicsInTransferEvent  = 5
	numTransferTopics         = 3
	numExecutedBridgeOpTopics = 3
	tokensIndex               = 2
	hashOfHashesIndex         = 1
	hashOfOperationIndex      = 2
)

const (
	eventIDExecutedOutGoingBridgeOp = "execute"
	eventIDDepositIncomingTransfer  = "deposit"

	topicIDConfirmedOutGoingOperation = "executedBridgeOp"
	topicIDDepositIncomingTransfer    = "deposit"
)

type confirmedBridgeOp struct {
	hashOfHashes []byte
	hash         []byte
}

type eventData struct {
	nonce                uint64
	functionCallWithArgs []byte
	gasLimit             uint64
}

type EventResult struct {
	SCR               *scrInfo
	ConfirmedBridgeOp *confirmedBridgeOp
}

type scrInfo struct {
	scr  *smartContractResult.SmartContractResult
	hash []byte
}

type eventsResult struct {
	scrs               []*scrInfo
	confirmedBridgeOps []*confirmedBridgeOp
}

type incomingEventsProcessor struct {
	handlers map[string]IncomingEventHandler
}

func (iep *incomingEventsProcessor) RegisterProcessor(event string, proc IncomingEventHandler) error {
	if check.IfNil(proc) {
		return errors.New("dsada")
	}

	iep.handlers[event] = proc
	return nil
}

// TODO refactor this to work with processors that assign tasks based on event id
func (iep *incomingEventsProcessor) processIncomingEvents(events []data.EventHandler) (*eventsResult, error) {
	scrs := make([]*scrInfo, 0, len(events))
	confirmedBridgeOps := make([]*confirmedBridgeOp, 0, len(events))

	for idx, event := range events {
		handler, found := iep.handlers[string(event.GetIdentifier())]
		if !found {
			return nil, errInvalidIncomingEventIdentifier
		}

		res, err := handler.ProcessEvent(event)
		if err != nil {
			return nil, fmt.Errorf("%w, event idx = %d", err, idx)
		}

		if res.SCR != nil {
			scrs = append(scrs, res.SCR)
		}
		if res.ConfirmedBridgeOp != nil {
			confirmedBridgeOps = append(confirmedBridgeOps, res.ConfirmedBridgeOp)
		}
	}

	return &eventsResult{
		scrs:               scrs,
		confirmedBridgeOps: confirmedBridgeOps,
	}, nil
}
