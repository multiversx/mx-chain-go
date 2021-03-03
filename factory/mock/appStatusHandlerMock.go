package mock

// AppStatusHandlerMock -
type AppStatusHandlerMock struct {
	CloseCalled func()
}

// Increment -
func (a *AppStatusHandlerMock) Increment(_ string) {
}

// AddUint64 -
func (a *AppStatusHandlerMock) AddUint64(_ string, _ uint64) {
}

// Decrement -
func (a *AppStatusHandlerMock) Decrement(_ string) {
}

// SetInt64Value -
func (a *AppStatusHandlerMock) SetInt64Value(_ string, _ int64) {
}

// SetUInt64Value -
func (a *AppStatusHandlerMock) SetUInt64Value(_ string, _ uint64) {
}

// SetStringValue -
func (a *AppStatusHandlerMock) SetStringValue(_ string, _ string) {
}

// Close -
func (a *AppStatusHandlerMock) Close() {
	if a.CloseCalled != nil {
		a.CloseCalled()
	}
}

// IsInterfaceNil -
func (a *AppStatusHandlerMock) IsInterfaceNil() bool {
	return false
}
