package incomingEventsProc

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
)

// IncomingEventHandler defines the behaviour of an incoming cross chain event processor handler
type IncomingEventHandler interface {
	ProcessEvent(event data.EventHandler) (*dto.EventResult, error)
	IsInterfaceNil() bool
}
