package incomingHeader

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type IncomingEventHandler interface {
	ProcessEvent(event data.EventHandler) (*eventsResult, error)
	IsInterfaceNil() bool
}

type incomingEventsProcHolder struct {
	handlers map[string]IncomingEventHandler
}

func (iep *incomingEventsProcHolder) RegisterProcessor(event string, proc IncomingEventHandler) error {
	if check.IfNil(proc) {
		return errors.New("dsada")
	}

	iep.handlers[event] = proc
	return nil
}

func (iep *incomingEventsProcHolder) ProcessIncomingEvent(event data.EventHandler) (*eventsResult, error) {
	handler, found := iep.handlers[string(event.GetIdentifier())]
	if !found {
		return nil, errors.New("dsada")
	}

	return handler.ProcessEvent(event)
}

func (iep *incomingEventsProcHolder) IsInterfaceNil() bool {
	return iep == nil
}
