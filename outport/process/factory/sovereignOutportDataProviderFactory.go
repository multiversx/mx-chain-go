package factory

import (
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/outport/process"
	"github.com/multiversx/mx-chain-go/outport/process/disabled"
)

type sovereignOutportDataProviderFactory struct {
}

// NewSovereignOutportDataProviderFactory creates a new outport data provider factory for sovereign chain
func NewSovereignOutportDataProviderFactory() *sovereignOutportDataProviderFactory {
	return &sovereignOutportDataProviderFactory{}
}

// CreateOutportDataProvider will create a new instance of outport.DataProviderOutport
func (f *sovereignOutportDataProviderFactory) CreateOutportDataProvider(arg ArgOutportDataProviderFactory) (outport.DataProviderOutport, error) {
	if !arg.HasDrivers {
		return disabled.NewDisabledOutportDataProvider(), nil
	}

	argsOutport, err := createArgs(arg)
	if err != nil {
		return nil, err
	}

	odp, err := process.NewOutportDataProvider(*argsOutport)
	if err != nil {
		return nil, err
	}

	return process.NewSovereignOutportDataProvider(odp)
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *sovereignOutportDataProviderFactory) IsInterfaceNil() bool {
	return f == nil
}
