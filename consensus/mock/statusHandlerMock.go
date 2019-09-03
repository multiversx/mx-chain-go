package mock

// AppStatusHandlerMock is an empty implementation of AppStatusHandler in order to be used in constructors
type AppStatusHandlerMock struct {
}

func (ashs *AppStatusHandlerMock) IsInterfaceNil() bool {
	if ashs == nil {
		return true
	}
	return false
}

// Increment method won't do anything
func (ashs *AppStatusHandlerMock) Increment(key string) {
}

// Decrement method won't do anything
func (ashs *AppStatusHandlerMock) Decrement(key string) {
}

// SetInt64Value method won't do anything
func (ashs *AppStatusHandlerMock) SetInt64Value(key string, value int64) {
}

// SetUInt64Value method won't do anything
func (ashs *AppStatusHandlerMock) SetUInt64Value(key string, value uint64) {
}

// SetStringValue method won't do anything
func (ashs *AppStatusHandlerMock) SetStringValue(key string, value string) {
}

// Close won't do anything
func (ashs *AppStatusHandlerMock) Close() {
}
