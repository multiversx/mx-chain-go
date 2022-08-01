package mock

import "github.com/ElrondNetwork/elrond-go/factory"

// HeartbeatV2ComponentsStub -
type HeartbeatV2ComponentsStub struct {
	MonitorField factory.HeartbeatV2Monitor
}

// Create -
func (hbc *HeartbeatV2ComponentsStub) Create() error {
	return nil
}

// Close -
func (hbc *HeartbeatV2ComponentsStub) Close() error {
	return nil
}

// CheckSubcomponents -
func (hbc *HeartbeatV2ComponentsStub) CheckSubcomponents() error {
	return nil
}

// String -
func (hbc *HeartbeatV2ComponentsStub) String() string {
	return ""
}

// Monitor -
func (hbc *HeartbeatV2ComponentsStub) Monitor() factory.HeartbeatV2Monitor {
	return hbc.MonitorField
}

// IsInterfaceNil -
func (hbc *HeartbeatV2ComponentsStub) IsInterfaceNil() bool {
	return hbc == nil
}
