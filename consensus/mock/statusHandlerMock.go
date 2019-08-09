package mock

// AppStatusHandlerMock is an empty implementation of AppStatusHandler in order to be used in constructors
type AppStatusHandlerMock struct {
}

// Increment method won't do anything
func (AppStatusHandlerMock) Increment(key string) {
}

// Decrement method won't do anything
func (AppStatusHandlerMock) Decrement(key string) {
}

// SetInt64Value method won't do anything
func (AppStatusHandlerMock) SetInt64Value(key string, value int64) {
}

// SetUInt64Value method won't do anything
func (AppStatusHandlerMock) SetUInt64Value(key string, value uint64) {
}

// SetStringValue method won't do anything
func (AppStatusHandlerMock) SetStringValue(key string, value string) {
}

// Close won't do anything
func (AppStatusHandlerMock) Close() {
}
