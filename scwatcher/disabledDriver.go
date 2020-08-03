package scwatcher

import "github.com/ElrondNetwork/elrond-go/data"

var _ Driver = (*DisabledScWatcherDriver)(nil)

type DisabledScWatcherDriver struct {
}

func NewDisabledScWatcherDriver() *DisabledScWatcherDriver {
	return &DisabledScWatcherDriver{}
}

// DigestBlock does nothing
func (driver *DisabledScWatcherDriver) DigestBlock(body data.BodyHandler, header data.HeaderHandler, transactions TransactionsToDigest) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (driver *DisabledScWatcherDriver) IsInterfaceNil() bool {
	return driver == nil
}
