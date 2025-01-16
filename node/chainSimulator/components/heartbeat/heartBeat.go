package heartbeat

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/heartbeat"
)

type heartBeatComponents struct {
	monitor factory.HeartbeatV2Monitor
}

// NewSyncedHeartbeatComponents will create a new instance of heartbeat components
func NewSyncedHeartbeatComponents(monitor factory.HeartbeatV2Monitor) (factory.HeartbeatV2ComponentsHandler, error) {
	if check.IfNil(monitor) {
		return nil, heartbeat.ErrNilHeartbeatMonitor
	}

	return &heartBeatComponents{
		monitor: monitor,
	}, nil
}

// Create will do nothing
func (h *heartBeatComponents) Create() error {
	return nil
}

// Close will do nothing
func (h *heartBeatComponents) Close() error {
	return nil
}

// CheckSubcomponents will do nothing
func (h *heartBeatComponents) CheckSubcomponents() error {
	return nil
}

// String will return a string
func (h *heartBeatComponents) String() string {
	return "heartBeat"
}

// Monitor will return the monitor
func (h *heartBeatComponents) Monitor() factory.HeartbeatV2Monitor {
	return h.monitor
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *heartBeatComponents) IsInterfaceNil() bool {
	return h == nil
}
