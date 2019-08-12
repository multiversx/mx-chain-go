package statusHandler

// NilStatusHandler will be used when an AppStatusHandler is required, but another one isn't necessary or available
type NilStatusHandler struct {
}

// NewNilStatusHandler will return an instance of the struct
func NewNilStatusHandler() *NilStatusHandler {
	return new(NilStatusHandler)
}

// Increment method - won't do anything
func (psh *NilStatusHandler) Increment(key string) {
}

// Decrement method - won't do anything
func (psh *NilStatusHandler) Decrement(key string) {
}

// SetInt64Value method - won't do anything
func (psh *NilStatusHandler) SetInt64Value(key string, value int64) {
}

// SetUInt64Value method - won't do anything
func (psh *NilStatusHandler) SetUInt64Value(key string, value uint64) {
}

// SetStringValue method - won't do anything
func (psh *NilStatusHandler) SetStringValue(key string, value string) {
}

// Close method - won't do anything
func (psh *NilStatusHandler) Close() {
}
