package outport

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

var _ Driver = (*disabledOutportDriver)(nil)

type disabledOutportDriver struct {
}

// NewDisabledOutportDriver creates a disabled outport driver
func NewDisabledOutportDriver() *disabledOutportDriver {
	return &disabledOutportDriver{}
}

// DigestCommittedBlock does nothing
func (driver *disabledOutportDriver) DigestCommittedBlock(headerHash []byte, header data.HeaderHandler) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (driver *disabledOutportDriver) IsInterfaceNil() bool {
	return driver == nil
}
