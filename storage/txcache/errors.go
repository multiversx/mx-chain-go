package txcache

import "fmt"

var errorTxNotFound = fmt.Errorf("tx not found in cache")
var errorMapsSyncInconsistency = fmt.Errorf("maps sync inconsistency between 'txByHash' and 'txListBySender'")
