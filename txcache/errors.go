package txcache

import "errors"

var errNilTxGasHandler = errors.New("nil tx gas handler")
var errNilSelectionSession = errors.New("nil selection session")
var errItemAlreadyInCache = errors.New("item already in cache")
var errEmptyBunchOfTransactions = errors.New("empty bunch of transactions")
