package disabled

import "time"

type disabledSenderHandler struct {
}

// NewDisabledSenderHandler returns a new instance of disabledSenderHandler
func NewDisabledSenderHandler() *disabledSenderHandler {
	return &disabledSenderHandler{}
}

// ExecutionReadyChannel returns a new chan
func (sender *disabledSenderHandler) ExecutionReadyChannel() <-chan time.Time {
	return make(chan time.Time)
}

// Execute does nothing
func (sender *disabledSenderHandler) Execute() {
}

// Close does nothing
func (sender *disabledSenderHandler) Close() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (sender *disabledSenderHandler) IsInterfaceNil() bool {
	return sender == nil
}
