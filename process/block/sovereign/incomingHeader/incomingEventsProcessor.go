package incomingHeader

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type eventsResult struct {
	scrs               []*SCRInfo
	confirmedBridgeOps []*ConfirmedBridgeOp
}

type incomingEventsProcessor struct {
	mut      sync.RWMutex
	handlers map[string]IncomingEventHandler
}

func (iep *incomingEventsProcessor) registerProcessor(event string, proc IncomingEventHandler) error {
	if check.IfNil(proc) {
		return errNilIncomingEventHandler
	}

	iep.mut.Lock()
	iep.handlers[event] = proc
	iep.mut.Unlock()
	return nil
}

func (iep *incomingEventsProcessor) processIncomingEvents(events []data.EventHandler) (*eventsResult, error) {
	scrs := make([]*SCRInfo, 0, len(events))
	confirmedBridgeOps := make([]*ConfirmedBridgeOp, 0, len(events))

	for idx, event := range events {
		iep.mut.RLock()
		handler, found := iep.handlers[string(event.GetIdentifier())]
		iep.mut.RUnlock()
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
