package txcache

import "fmt"

var errTxNotFound = fmt.Errorf("tx not found in cache")
var errTxDuplicated = fmt.Errorf("duplicated tx")
var errInvalidCacheConfig = fmt.Errorf("invalid cache config")
