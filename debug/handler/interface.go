package handler

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

// InterceptorTxDebugHandler defines what an interceptor debug handler should be able to do
type InterceptorTxDebugHandler interface {
	Process(data process.InterceptedData, msg p2p.MessageP2P, fromConnectedPeer core.PeerID)
	PrintReceivedTxsBroadcastAndCleanRecords()
}
