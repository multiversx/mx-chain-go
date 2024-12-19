package txcache

import "errors"

var errNilMempoolHost = errors.New("nil mempool host")
var errNilSelectionSession = errors.New("nil selection session")
var errItemAlreadyInCache = errors.New("item already in cache")
var errEmptyBunchOfTransactions = errors.New("empty bunch of transactions")
