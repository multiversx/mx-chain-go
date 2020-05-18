package alarms

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("alarms")

// RaiseTransactionsAlarm raises an alarm
func RaiseTransactionsAlarm() {
	log.Error("RaiseTransactionsAlarm")
	logger.SetLogLevel("*:DEBUG,api:INFO,txcache:TRACE,missingTransactions:TRACE,dataretriever/requesthandlers:TRACE")
}
