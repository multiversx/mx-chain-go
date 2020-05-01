package mock

// AppStatusHandlerMock -
type AppStatusHandlerMock struct {
}

// Increment -
func (a *AppStatusHandlerMock) Increment(key string) {
}

// AddUint64 -
func (a *AppStatusHandlerMock) AddUint64(key string, val uint64) {
}

// Decrement -
func (a *AppStatusHandlerMock) Decrement(key string) {
}

// SetInt64Value -
func (a *AppStatusHandlerMock) SetInt64Value(key string, value int64) {
}

// SetUInt64Value -
func (a *AppStatusHandlerMock) SetUInt64Value(key string, value uint64) {
}

// SetStringValue -
func (a *AppStatusHandlerMock) SetStringValue(key string, value string) {
}

// Close -
func (a *AppStatusHandlerMock) Close() {
}

// IsInterfaceNil -
func (a *AppStatusHandlerMock) IsInterfaceNil() bool {
	return false
}
