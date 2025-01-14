package trie

import "github.com/multiversx/mx-chain-go/common"

type disabledGoroutinesManager struct {
	err error
}

// NewDisabledGoroutinesManager creates a new instance of disabledGoroutinesManager
func NewDisabledGoroutinesManager() *disabledGoroutinesManager {
	return &disabledGoroutinesManager{}
}

// ShouldContinueProcessing returns true if there is no error
func (d *disabledGoroutinesManager) ShouldContinueProcessing() bool {
	return d.err == nil
}

// CanStartGoRoutine returns false
func (d *disabledGoroutinesManager) CanStartGoRoutine() bool {
	return false
}

// EndGoRoutineProcessing does nothing
func (d *disabledGoroutinesManager) EndGoRoutineProcessing() {
}

// SetNewErrorChannel does nothing
func (d *disabledGoroutinesManager) SetNewErrorChannel(_ common.BufferedErrChan) error {
	return nil
}

// SetError sets the given error
func (d *disabledGoroutinesManager) SetError(err error) {
	d.err = err
}

// GetError returns the error
func (d *disabledGoroutinesManager) GetError() error {
	return d.err
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledGoroutinesManager) IsInterfaceNil() bool {
	return d == nil
}
