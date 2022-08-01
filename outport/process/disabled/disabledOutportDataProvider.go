package disabled

import (
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go/outport/process"
)

type disabledOutportDataProvider struct{}

// NewDisabledOutportDataProvider will create a new instance of disabledOutportDataProvider
func NewDisabledOutportDataProvider() *disabledOutportDataProvider {
	return &disabledOutportDataProvider{}
}

// PrepareOutportSaveBlockData wil do nothing
func (d *disabledOutportDataProvider) PrepareOutportSaveBlockData(_ process.ArgPrepareOutportSaveBlockData) (*outportcore.ArgsSaveBlockData, error) {
	return &outportcore.ArgsSaveBlockData{}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledOutportDataProvider) IsInterfaceNil() bool {
	return d == nil
}
