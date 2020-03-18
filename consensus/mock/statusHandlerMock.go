package mock

// AppStatusHandlerMock is an empty implementation of AppStatusHandler in order to be used in constructors
type AppStatusHandlerMock struct {
}

// IsInterfaceNil -
func (ashs *AppStatusHandlerMock) IsInterfaceNil() bool {
	return ashs == nil
}

// AddUint64 -
func (ashs *AppStatusHandlerMock) AddUint64(_ string, _ uint64) {
}

// Increment method won't do anything
func (ashs *AppStatusHandlerMock) Increment(_ string) {
}

// Decrement method won't do anything
func (ashs *AppStatusHandlerMock) Decrement(_ string) {
}

// SetInt64Value method won't do anything
func (ashs *AppStatusHandlerMock) SetInt64Value(_ string, _ int64) {
}

// SetUInt64Value method won't do anything
func (ashs *AppStatusHandlerMock) SetUInt64Value(_ string, _ uint64) {
}

// SetStringValue method won't do anything
func (ashs *AppStatusHandlerMock) SetStringValue(_ string, _ string) {
}

// Close won't do anything
func (ashs *AppStatusHandlerMock) Close() {
}
