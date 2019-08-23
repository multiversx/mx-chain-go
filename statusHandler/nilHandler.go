package statusHandler

// NilStatusHandler will be used when an AppStatusHandler is required, but another one isn't necessary or available
type NilStatusHandler struct {
}

// NewNilStatusHandler will return an instance of the struct
func NewNilStatusHandler() *NilStatusHandler {
	return new(NilStatusHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (nsh *NilStatusHandler) IsInterfaceNil() bool {
	if nsh == nil {
		return true
	}
	return false
}

// Increment method - won't do anything
func (nsh *NilStatusHandler) Increment(key string) {
}

// Decrement method - won't do anything
func (nsh *NilStatusHandler) Decrement(key string) {
}

// SetInt64Value method - won't do anything
func (nsh *NilStatusHandler) SetInt64Value(key string, value int64) {
}

// SetUInt64Value method - won't do anything
func (nsh *NilStatusHandler) SetUInt64Value(key string, value uint64) {
}

// SetStringValue method - won't do anything
func (nsh *NilStatusHandler) SetStringValue(key string, value string) {
}

// Close method - won't do anything
func (nsh *NilStatusHandler) Close() {
}
