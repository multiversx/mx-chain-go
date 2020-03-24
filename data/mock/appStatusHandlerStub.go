package mock

// AppStatusHandlerStub is a stub implementation of AppStatusHandler
type AppStatusHandlerStub struct {
	AddUint64Handler      func(key string, value uint64)
	IncrementHandler      func(key string)
	DecrementHandler      func(key string)
	SetUInt64ValueHandler func(key string, value uint64)
	SetInt64ValueHandler  func(key string, value int64)
	SetStringValueHandler func(key string, value string)
	CloseHandler          func()
}

// IsInterfaceNil -
func (ashs *AppStatusHandlerStub) IsInterfaceNil() bool {
	return ashs == nil
}

// AddUint64 will call the handler of the stub for incrementing
func (ashs *AppStatusHandlerStub) AddUint64(key string, value uint64) {
	ashs.AddUint64Handler(key, value)
}

// Increment will call the handler of the stub for incrementing
func (ashs *AppStatusHandlerStub) Increment(key string) {
	ashs.IncrementHandler(key)
}

// Decrement will call the handler of the stub for decrementing
func (ashs *AppStatusHandlerStub) Decrement(key string) {
	ashs.DecrementHandler(key)
}

// SetInt64Value will call the handler of the stub for setting an int64 value
func (ashs *AppStatusHandlerStub) SetInt64Value(key string, value int64) {
	ashs.SetInt64ValueHandler(key, value)
}

// SetUInt64Value will call the handler of the stub for setting an uint64 value
func (ashs *AppStatusHandlerStub) SetUInt64Value(key string, value uint64) {
	ashs.SetUInt64ValueHandler(key, value)
}

// SetStringValue will call the handler of the stub for setting an string value
func (ashs *AppStatusHandlerStub) SetStringValue(key string, value string) {
	ashs.SetStringValueHandler(key, value)
}

// Close will call the handler of the stub for closing
func (ashs *AppStatusHandlerStub) Close() {
	ashs.CloseHandler()
}
