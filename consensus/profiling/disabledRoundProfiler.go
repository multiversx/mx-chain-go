package profiling

import "time"

type disabledRoundProfiler struct{}

// NewDisabledRoundProfiler returns a no-op round profiler
func NewDisabledRoundProfiler() *disabledRoundProfiler {
	return &disabledRoundProfiler{}
}

// OnRoundStart does nothing
func (d *disabledRoundProfiler) OnRoundStart(_ int64, _ time.Time) {}

// Close does nothing
func (d *disabledRoundProfiler) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledRoundProfiler) IsInterfaceNil() bool {
	return d == nil
}
