package metrics

import "github.com/ElrondNetwork/elrond-go-core/core"

type disabledConnectionsWatcher struct{}

// NewDisabledConnectionsWatcher returns a disabled ConnectionWatcher implementation
func NewDisabledConnectionsWatcher() *disabledConnectionsWatcher {
	return &disabledConnectionsWatcher{}
}

// NewKnownConnection does nothing
func (dcw *disabledConnectionsWatcher) NewKnownConnection(_ core.PeerID, _ string) {}

// Close does nothing and returns nil
func (dcw *disabledConnectionsWatcher) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dcw *disabledConnectionsWatcher) IsInterfaceNil() bool {
	return dcw == nil
}
