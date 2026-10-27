package cutoff

import "github.com/multiversx/mx-chain-core-go/data"

type disabledBlockProcessingCutoff struct {
}

// NewDisabledBlockProcessingCutoff will return a new instance of disabledBlockProcessingCutoff
func NewDisabledBlockProcessingCutoff() *disabledBlockProcessingCutoff {
	return &disabledBlockProcessingCutoff{}
}

// HandleProcessErrorCutoff returns nil
func (d *disabledBlockProcessingCutoff) HandleProcessErrorCutoff(_ data.HeaderHandler) error {
	return nil
}

// HandlePauseCutoff does nothing
func (d *disabledBlockProcessingCutoff) HandlePauseCutoff(_ data.HeaderHandler) {
}

// HandleGracefulStopCutoff does nothing
func (d *disabledBlockProcessingCutoff) HandleGracefulStopCutoff(_ data.HeaderHandler, _ func()) {
}

// Close does nothing
func (d *disabledBlockProcessingCutoff) Close() {
}

// IsInterfaceNil returns true since this structure uses value receivers
func (d *disabledBlockProcessingCutoff) IsInterfaceNil() bool {
	return d == nil
}
