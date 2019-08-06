package mock

import "github.com/ElrondNetwork/elrond-go/core"

// AppStatusFacadeStub is a stub implementation of AppStatusFacade
type AppStatusFacadeStub struct {
	ListOfAphs            []core.AppStatusHandler
	IncrementHandler      func(key string)
	DecrementHandler      func(key string)
	SetUInt64ValueHandler func(key string, value uint64)
	SetInt64ValueHandler  func(key string, value int64)
	CloseHandler          func()
}

// Increment will call the handler of the stub for incrementing
func (asf *AppStatusFacadeStub) Increment(key string) {
	asf.IncrementHandler(key)
}

// Decrement will call the handler of the stub for decrementing
func (asf *AppStatusFacadeStub) Decrement(key string) {
	asf.DecrementHandler(key)
}

// SetInt64Value will call the handler of the stub for setting an int64 value
func (asf *AppStatusFacadeStub) SetInt64Value(key string, value int64) {
	asf.SetInt64ValueHandler(key, value)
}

// SetUInt64Value will call the handler of the stub for setting an uint64 value
func (asf *AppStatusFacadeStub) SetUInt64Value(key string, value uint64) {
	asf.SetUInt64ValueHandler(key, value)
}

// Close will call the handler of the stub for closing
func (asf *AppStatusFacadeStub) Close() {
	asf.CloseHandler()
}
