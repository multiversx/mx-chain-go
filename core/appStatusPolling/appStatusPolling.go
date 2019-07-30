package appStatusPolling

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
)

// AppStatusPolling will update an AppStatusHandler by polling components at a predefined interval
type AppStatusPolling struct {
	pollingDuration     time.Duration
	mutRegisteredFunc   sync.RWMutex
	registeredFunctions []func(appStatusHandler core.AppStatusHandler)
	appStatusHandler    core.AppStatusHandler
}

// NewAppStatusPolling will return an instance of AppStatusPolling
func NewAppStatusPolling(appStatusHandler core.AppStatusHandler, pollingDuration time.Duration) (*AppStatusPolling, error) {
	if appStatusHandler == nil {
		return nil, ErrNilAppStatusHandler
	}
	if pollingDuration < 0 {
		return nil, ErrPollingDurationNegative
	}
	return &AppStatusPolling{
		pollingDuration:  pollingDuration,
		appStatusHandler: appStatusHandler,
	}, nil
}

// RegisterPollingFunc will register a new handler function
func (asp *AppStatusPolling) RegisterPollingFunc(handler func(appStatusHandler core.AppStatusHandler)) error {
	if handler == nil {
		return ErrNilHandlerFunc
	}
	asp.mutRegisteredFunc.Lock()
	asp.registeredFunctions = append(asp.registeredFunctions, handler)
	asp.mutRegisteredFunc.Unlock()
	return nil
}

// Poll will notify the AppStatusHandler at a given time
func (asp *AppStatusPolling) Poll() {
	go func() {
		for {
			time.Sleep(asp.pollingDuration)

			asp.mutRegisteredFunc.Lock()
			for _, handler := range asp.registeredFunctions {
				handler(asp.appStatusHandler)
			}
			asp.mutRegisteredFunc.Unlock()
		}
	}()
}
