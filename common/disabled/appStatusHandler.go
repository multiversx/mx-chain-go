package disabled

type appStatusHandler struct {
}

// NewAppStatusHandler creates a new instance of disabled app status handler
func NewAppStatusHandler() *appStatusHandler {
	return &appStatusHandler{}
}

// AddUint64 does nothing
func (ash *appStatusHandler) AddUint64(key string, value uint64) {
}

// Increment does nothing
func (ash *appStatusHandler) Increment(key string) {
}

// Decrement does nothing
func (ash *appStatusHandler) Decrement(key string) {
}

// SetInt64Value does nothing
func (ash *appStatusHandler) SetInt64Value(key string, value int64) {
}

// SetUInt64Value does nothing
func (ash *appStatusHandler) SetUInt64Value(key string, value uint64) {
}

// SetStringValue does nothing
func (ash *appStatusHandler) SetStringValue(key string, value string) {
}

// Close does nothing
func (ash *appStatusHandler) Close() {
}

// IsInterfaceNil returns nil if there is no value under the interface
func (ash *appStatusHandler) IsInterfaceNil() bool {
	return ash == nil
}
