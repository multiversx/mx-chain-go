package mock

// AppStatusHandlerFake -
type AppStatusHandlerFake struct {
}

// Increment -
func (a *AppStatusHandlerFake) Increment(key string) {
}

// AddUint64 -
func (a *AppStatusHandlerFake) AddUint64(key string, val uint64) {
}

// Decrement -
func (a *AppStatusHandlerFake) Decrement(key string) {
}

// SetInt64Value -
func (a *AppStatusHandlerFake) SetInt64Value(key string, value int64) {
}

// SetUInt64Value -
func (a *AppStatusHandlerFake) SetUInt64Value(key string, value uint64) {
}

// SetStringValue -
func (a *AppStatusHandlerFake) SetStringValue(key string, value string) {
}

// Close -
func (a *AppStatusHandlerFake) Close() {
}

// IsInterfaceNil -
func (a *AppStatusHandlerFake) IsInterfaceNil() bool {
	return false
}
