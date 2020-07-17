package scwatcher

import "github.com/ElrondNetwork/elrond-go/data"

var _ Driver = (*DisabledScWatcherDriver)(nil)

type DisabledScWatcherDriver struct {
}

// DigestBlock does nothing
func (driver *DisabledScWatcherDriver) DigestBlock(body data.BodyHandler, header data.HeaderHandler, transactions TransactionsToDigest) {
}
