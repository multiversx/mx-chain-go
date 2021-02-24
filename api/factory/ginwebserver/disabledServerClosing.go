package ginwebserver

type disabledServerClosing struct {
}

// NewDisabledServerClosing returns a new instance of disabledServerClosing
func NewDisabledServerClosing() *disabledServerClosing {
	return &disabledServerClosing{}
}

// Start does nothing
func (d *disabledServerClosing) Start() {
}

// Close returns nil
func (d *disabledServerClosing) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledServerClosing) IsInterfaceNil() bool {
	return d == nil
}
