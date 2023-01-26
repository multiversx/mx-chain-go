package disabled

type disabledHardforkHandler struct {
}

// NewDisabledHardforkHandler returns a new instance of disabledHardforkHandler
func NewDisabledHardforkHandler() *disabledHardforkHandler {
	return &disabledHardforkHandler{}
}

// ShouldTriggerHardfork returns a new chan
func (sender *disabledHardforkHandler) ShouldTriggerHardfork() <-chan struct{} {
	return make(chan struct{})
}

// Execute does nothing
func (sender *disabledHardforkHandler) Execute() {
}

// Close does nothing
func (sender *disabledHardforkHandler) Close() {
}
