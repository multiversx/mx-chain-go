package mainFactoryMocks

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/process"
)

// StatusComponentsStub -
type StatusComponentsStub struct {
	Outport                  outport.OutportHandler
	SoftwareVersionCheck     statistics.SoftwareVersionChecker
	AppStatusHandler         core.AppStatusHandler
	ManagedPeersMonitorField common.ManagedPeersMonitor
}

// Create -
func (scs *StatusComponentsStub) Create() error {
	return nil
}

// Close -
func (scs *StatusComponentsStub) Close() error {
	return nil
}

// CheckSubcomponents -
func (scs *StatusComponentsStub) CheckSubcomponents() error {
	return nil
}

// OutportHandler -
func (scs *StatusComponentsStub) OutportHandler() outport.OutportHandler {
	return scs.Outport
}

// SoftwareVersionChecker -
func (scs *StatusComponentsStub) SoftwareVersionChecker() statistics.SoftwareVersionChecker {
	return scs.SoftwareVersionCheck
}

// ManagedPeersMonitor -
func (scs *StatusComponentsStub) ManagedPeersMonitor() common.ManagedPeersMonitor {
	return scs.ManagedPeersMonitorField
}

func (scs *StatusComponentsStub) SetForkDetector(forkDetector process.ForkDetector) error {
	return nil
}

func (scs *StatusComponentsStub) StartPolling() error {
	return nil
}

// String -
func (scs *StatusComponentsStub) String() string {
	return "StatusComponentsMock"
}

// IsInterfaceNil -
func (scs *StatusComponentsStub) IsInterfaceNil() bool {
	return scs == nil
}
