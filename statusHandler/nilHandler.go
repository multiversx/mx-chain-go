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
	return nsh == nil
}

// AddUint64 method - won't do anything
func (nsh *NilStatusHandler) AddUint64(_ string, _ uint64) {
}

// Increment method - won't do anything
func (nsh *NilStatusHandler) Increment(_ string) {
}

// Decrement method - won't do anything
func (nsh *NilStatusHandler) Decrement(_ string) {
}

// SetInt64Value method - won't do anything
func (nsh *NilStatusHandler) SetInt64Value(_ string, _ int64) {
}

// SetUInt64Value method - won't do anything
func (nsh *NilStatusHandler) SetUInt64Value(_ string, _ uint64) {
}

// SetStringValue method - won't do anything
func (nsh *NilStatusHandler) SetStringValue(_ string, _ string) {
}

// Close method - won't do anything
func (nsh *NilStatusHandler) Close() {
}
