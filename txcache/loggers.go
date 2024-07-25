package txcache

import logger "github.com/multiversx/mx-chain-logger-go"

var log = logger.GetOrCreate("txcache/main")
var logAdd = logger.GetOrCreate("txcache/add")
var logRemove = logger.GetOrCreate("txcache/remove")
var logSelect = logger.GetOrCreate("txcache/select")
var logDiagnoseSelection = logger.GetOrCreate("txcache/diagnose/selection")
var logDiagnoseSenders = logger.GetOrCreate("txcache/diagnose/senders")
var logDiagnoseTransactions = logger.GetOrCreate("txcache/diagnose/transactions")
