package statusHandler

import "github.com/ElrondNetwork/elrond-go/core"

// AppStatusFacade will be used for handling multiple monitoring tools at once
type AppStatusFacade struct {
	handlers []core.AppStatusHandler
}

// NewAppStatusFacadeWithHandlers will receive the handlers which should receive monitored data
func NewAppStatusFacadeWithHandlers(aphs ...core.AppStatusHandler) AppStatusFacade {
	var appStatusFacade AppStatusFacade
	appStatusFacade.handlers = aphs
	return appStatusFacade
}

// Increment method - will increment the value for a key for every handler
func (asf *AppStatusFacade) Increment(key string) {
	for _, ash := range asf.handlers {
		ash.Increment(key)
	}
}

// Decrement method - will decrement the value for a key for every handler
func (asf *AppStatusFacade) Decrement(key string) {
	for _, ash := range asf.handlers {
		ash.Decrement(key)
	}
}

// SetInt64Value method - will update the value for a key for every handler
func (asf *AppStatusFacade) SetInt64Value(key string, value int64) {
	for _, ash := range asf.handlers {
		ash.SetInt64Value(key, value)
	}
}

// SetUInt64Value method - will update the value for a key for every handler
func (asf *AppStatusFacade) SetUInt64Value(key string, value uint64) {
	for _, ash := range asf.handlers {
		ash.SetUInt64Value(key, value)
	}
}

// GetValue method - will fetch the value for a key from the first handler which has it - TESTING ONLY
func (asf *AppStatusFacade) GetValue(key string) float64 {
	for _, ash := range asf.handlers {
		return ash.GetValue(key)
	}
	return float64(0)
}

// Close method will close all the handlers
func (asf *AppStatusFacade) Close() {
	for _, ash := range asf.handlers {
		ash.Close()
	}
}
