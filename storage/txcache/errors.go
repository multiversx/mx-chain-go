package txcache

import "fmt"

var errTxNotFound = fmt.Errorf("tx not found in cache")
var errMapsSyncInconsistency = fmt.Errorf("maps sync inconsistency between 'txByHash' and 'txListBySender'")
var errTxDuplicated = fmt.Errorf("duplicated tx")
