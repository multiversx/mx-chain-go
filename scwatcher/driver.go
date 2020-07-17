package scwatcher

import "github.com/ElrondNetwork/elrond-go/data"

var _ Driver = (*ScWatcherDriver)(nil)

type ScWatcherDriver struct {
}

func NewScWatcherDriver() *ScWatcherDriver {
	return &ScWatcherDriver{}
}

// DigestBlock digests a block
func (driver *ScWatcherDriver) DigestBlock(body data.BodyHandler, header data.HeaderHandler, transactions TransactionsToDigest) {
}
