package txcache

import "fmt"

// ErrTxNotFound signals that the transactions was not found in the cache
var ErrTxNotFound = fmt.Errorf("tx not found in cache")

// ErrMapsSyncInconsistency signals that there's an inconsistency between the internal maps on which the cache relies
var ErrMapsSyncInconsistency = fmt.Errorf("maps sync inconsistency between 'txByHash' and 'txListBySender'")
