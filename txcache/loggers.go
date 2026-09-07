package txcache

import logger "github.com/multiversx/mx-chain-logger-go"

var log = logger.GetOrCreate("txcache/main")
var logAdd = logger.GetOrCreate("txcache/add")
var logRemove = logger.GetOrCreate("txcache/remove")
var logSelect = logger.GetOrCreate("txcache/select")
var logDiagnoseTransactions = logger.GetOrCreate("txcache/diagnose/transactions")
