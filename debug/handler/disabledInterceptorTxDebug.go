package handler

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

type disabledInterceptorTxDebug struct{}

// NewDisabledInterceptorTxDebug will create a new instance of *disabledInterceptorTxDebug
func NewDisabledInterceptorTxDebug() *disabledInterceptorTxDebug {
	return new(disabledInterceptorTxDebug)
}

// Process -
func (dit *disabledInterceptorTxDebug) Process(_ process.InterceptedData, _ p2p.MessageP2P, _ core.PeerID) {

}

// PrintReceivedTxsBroadcastAndCleanRecords -
func (dit *disabledInterceptorTxDebug) PrintReceivedTxsBroadcastAndCleanRecords() {}
