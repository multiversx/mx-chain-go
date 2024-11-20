package txcache

import "errors"

var errNilTxGasHandler = errors.New("nil tx gas handler")
var errNilAccountStateProvider = errors.New("nil account state provider")
var errItemAlreadyInCache = errors.New("item already in cache")
var errEmptyBunchOfTransactions = errors.New("empty bunch of transactions")
