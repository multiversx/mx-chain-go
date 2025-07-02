package handler

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"time"
)

// BroadcastDebugHandler defines what an interceptor debug handler should be able to do
type BroadcastDebugHandler interface {
	Process(data process.InterceptedData, msg p2p.MessageP2P, fromConnectedPeer core.PeerID)
	PrintReceivedTxsBroadcastAndCleanRecords()
}

// NTPTime defines methods for retrieving the current time
type NTPTime interface {
	CurrentTime() time.Time
	IsInterfaceNil() bool
}
