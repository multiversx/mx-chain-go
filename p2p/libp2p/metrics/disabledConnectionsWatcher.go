package metrics

import "github.com/ElrondNetwork/elrond-go-core/core"

type disabledConnectionsWatcher struct{}

// NewDisabledConnectionsWatcher returns a disabled ConnectionWatcher implementation
func NewDisabledConnectionsWatcher() *disabledConnectionsWatcher {
	return &disabledConnectionsWatcher{}
}

// Connected does nothing
func (dcw *disabledConnectionsWatcher) Connected(_ core.PeerID, _ string) {}

// Disconnected does nothing
func (dcw *disabledConnectionsWatcher) Disconnected(_ core.PeerID) {}

// Close does nothing and returns nil
func (dcw *disabledConnectionsWatcher) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dcw *disabledConnectionsWatcher) IsInterfaceNil() bool {
	return dcw == nil
}
