package disabled

import (
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/outport/process"
)

type disabledOutportDataProvider struct{}

// NewDisabledOutportDataProvider will create a new instance of disabledOutportDataProvider
func NewDisabledOutportDataProvider() *disabledOutportDataProvider {
	return &disabledOutportDataProvider{}
}

// PrepareOutportSaveBlockData wil do nothing
func (d *disabledOutportDataProvider) PrepareOutportSaveBlockData(_ process.ArgPrepareOutportSaveBlockData) (*outportcore.OutportBlockWithHeaderAndBody, error) {
	return &outportcore.OutportBlockWithHeaderAndBody{}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledOutportDataProvider) IsInterfaceNil() bool {
	return d == nil
}
