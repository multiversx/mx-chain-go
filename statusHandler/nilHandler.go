package statusHandler

// NilStatusHandler will be used when an AppStatusHandler is required, but another one isn't necessary or available
type NilStatusHandler struct {
}

// NewNillStatusHandler will return an instance of the struct
func NewNillStatusHandler() *NilStatusHandler {
	return new(NilStatusHandler)
}

// Increment method - won't do anything
func (psh *NilStatusHandler) Increment(key string) {
	return
}

// Decrement method - won't do anything
func (psh *NilStatusHandler) Decrement(key string) {
	return
}

// SetInt64Value method - won't do anything
func (psh *NilStatusHandler) SetInt64Value(key string, value int64) {
	return
}

// SetUInt64Value method - won't do anything
func (psh *NilStatusHandler) SetUInt64Value(key string, value uint64) {
	return
}

// Close method - won't do anything
func (psh *NilStatusHandler) Close() {
}
