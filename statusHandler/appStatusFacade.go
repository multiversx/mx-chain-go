package statusHandler

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// AppStatusFacade will be used for handling multiple monitoring tools at once
type AppStatusFacade struct {
	handlers []core.AppStatusHandler
}

// NewAppStatusFacadeWithHandlers will receive the handlers which should receive monitored data
func NewAppStatusFacadeWithHandlers(aphs ...core.AppStatusHandler) (*AppStatusFacade, error) {
	if aphs == nil {
		return nil, ErrNilHandlersSlice
	}
	return &AppStatusFacade{
		handlers: aphs,
	}, nil
}

// Increment method - will increment the value for a key for every handler
func (asf *AppStatusFacade) Increment(key string) {
	go func() {
		for _, ash := range asf.handlers {
			ash.Increment(key)
		}
	}()
}

// Decrement method - will decrement the value for a key for every handler
func (asf *AppStatusFacade) Decrement(key string) {
	go func() {
		for _, ash := range asf.handlers {
			ash.Decrement(key)
		}
	}()
}

// SetInt64Value method - will update the value for a key for every handler
func (asf *AppStatusFacade) SetInt64Value(key string, value int64) {
	go func() {
		for _, ash := range asf.handlers {
			ash.SetInt64Value(key, value)
		}
	}()
}

// SetUInt64Value method - will update the value for a key for every handler
func (asf *AppStatusFacade) SetUInt64Value(key string, value uint64) {
	go func() {
		for _, ash := range asf.handlers {
			ash.SetUInt64Value(key, value)
		}
	}()
}

// Close method will close all the handlers
func (asf *AppStatusFacade) Close() {
	go func() {
		for _, ash := range asf.handlers {
			ash.Close()
		}
	}()
}
