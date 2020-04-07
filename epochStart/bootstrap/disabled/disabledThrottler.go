package disabled

type disabledThrottler struct {
}

// NewThrottler returns a new instance of a disabledThrottler
func NewThrottler() *disabledThrottler {
	return &disabledThrottler{}
}

// CanProcess will return true always
func (d *disabledThrottler) CanProcess() bool {
	return true
}

// StartProcessing won't do anything
func (d *disabledThrottler) StartProcessing() {
}

// EndProcessing won't do anything
func (d *disabledThrottler) EndProcessing() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledThrottler) IsInterfaceNil() bool {
	return d == nil
}
