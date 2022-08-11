package disabled

import "github.com/ElrondNetwork/elrond-go-core/core"

type disabledBootstrapper struct {
}

// NewDisabledBootstrapper returns a new instance of disabledBootstrapper
func NewDisabledBootstrapper() *disabledBootstrapper {
	return &disabledBootstrapper{}
}

// AddSyncStateListener won't do anything as this is a disabled component
func (d *disabledBootstrapper) AddSyncStateListener(_ func(isSyncing bool)) {
}

// GetNodeState will return a not synchronized state
func (d *disabledBootstrapper) GetNodeState() core.NodeState {
	return core.NsNotSynchronized
}

// StartSyncingBlocks won't do anything as this is a disabled component
func (d *disabledBootstrapper) StartSyncingBlocks() {
}

// Close will return a nil error as this is a disabled component
func (d *disabledBootstrapper) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledBootstrapper) IsInterfaceNil() bool {
	return d == nil
}
