package mock

import (
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/outport"
)

// StatusComponentsStub -
type StatusComponentsStub struct {
	Outport              outport.OutportHandler
	SoftwareVersionCheck statistics.SoftwareVersionChecker
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

// IsInterfaceNil -
func (scs *StatusComponentsStub) IsInterfaceNil() bool {
	return scs == nil
}
