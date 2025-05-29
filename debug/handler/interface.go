package handler

import (
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

type InterceptorTxDebugHandler interface {
	Process(data process.InterceptedData, msg p2p.MessageP2P)
	PrintReceivedTxsBroadcastAndCleanRecords()
}
