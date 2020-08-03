package outport

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ Driver = (*DisabledOutportDriver)(nil)

type DisabledOutportDriver struct {
}

func NewDisabledOutportDriver() *DisabledOutportDriver {
	return &DisabledOutportDriver{}
}

// DigestBlock does nothing
func (driver *DisabledOutportDriver) DigestBlock(header data.HeaderHandler, body data.BodyHandler, txCoordinator process.TransactionCoordinator) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (driver *DisabledOutportDriver) IsInterfaceNil() bool {
	return driver == nil
}
