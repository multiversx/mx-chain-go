package handler

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

type disabledBroadcastDebug struct{}

// NewDisabledBroadcastDebug will create a new instance of *disabledBroadcastDebug
func NewDisabledBroadcastDebug() *disabledBroadcastDebug {
	return new(disabledBroadcastDebug)
}

// Process -
func (db *disabledBroadcastDebug) Process(_ process.InterceptedData, _ p2p.MessageP2P, _ core.PeerID) {

}

// PrintReceivedTxsBroadcastAndCleanRecords -
func (db *disabledBroadcastDebug) PrintReceivedTxsBroadcastAndCleanRecords() {}
