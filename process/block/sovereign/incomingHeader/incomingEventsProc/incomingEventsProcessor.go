package incomingEventsProc

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
)

type incomingEventsProcessor struct {
	mut      sync.RWMutex
	handlers map[string]IncomingEventHandler
}

// NewIncomingEventsProcessor creates a new incoming events processor holder
func NewIncomingEventsProcessor() *incomingEventsProcessor {
	return &incomingEventsProcessor{
		handlers: make(map[string]IncomingEventHandler),
	}
}

// RegisterProcessor registers a new incoming event handler for a specified event
func (iep *incomingEventsProcessor) RegisterProcessor(event string, proc IncomingEventHandler) error {
	if check.IfNil(proc) {
		return errNilIncomingEventHandler
	}

	iep.mut.Lock()
	iep.handlers[event] = proc
	iep.mut.Unlock()
	return nil
}

// ProcessIncomingEvents will process all incoming events and create a result of incoming scrs or confirmed bridge operations
func (iep *incomingEventsProcessor) ProcessIncomingEvents(events []data.EventHandler) (*dto.EventsResult, error) {
	scrs := make([]*dto.SCRInfo, 0, len(events))
	confirmedBridgeOps := make([]*dto.ConfirmedBridgeOp, 0, len(events))

	for idx, event := range events {
		iep.mut.RLock()
		handler, found := iep.handlers[string(event.GetIdentifier())]
		iep.mut.RUnlock()
		if !found {
			return nil, dto.ErrInvalidIncomingEventIdentifier
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

	return &dto.EventsResult{
		Scrs:               scrs,
		ConfirmedBridgeOps: confirmedBridgeOps,
	}, nil
}
