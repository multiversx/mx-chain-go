package handler

import (
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/process"
)

type disabledInterceptorDebugHandler struct {
}

// NewDisabledInterceptorDebugHandler returns a disabled instance of the debug handler
func NewDisabledInterceptorDebugHandler() *disabledInterceptorDebugHandler {
	return &disabledInterceptorDebugHandler{}
}

// LogRequestedData does nothing
func (didh *disabledInterceptorDebugHandler) LogRequestedData(_ string, _ [][]byte, _ int, _ int) {
}

// LogReceivedHashes does nothing
func (didh *disabledInterceptorDebugHandler) LogReceivedHashes(_ string, _ [][]byte) {
}

// LogProcessedHashes does nothing
func (didh *disabledInterceptorDebugHandler) LogProcessedHashes(_ string, _ [][]byte, _ error) {
}

// Query returns an empty slice
func (didh *disabledInterceptorDebugHandler) Query(_ string) []string {
	return make([]string, 0)
}

// LogFailedToResolveData does nothing
func (didh *disabledInterceptorDebugHandler) LogFailedToResolveData(_ string, _ []byte, _ error) {
}

// LogSucceededToResolveData does nothing
func (didh *disabledInterceptorDebugHandler) LogSucceededToResolveData(_ string, _ []byte) {
}

// Close returns nil
func (didh *disabledInterceptorDebugHandler) Close() error {
	return nil
}

// LogReceivedData -
func (didh *disabledInterceptorDebugHandler) LogReceivedData(_ process.InterceptedData, _ p2p.MessageP2P, _ core.PeerID) {
}

// EpochStartEventHandler -
func (didh *disabledInterceptorDebugHandler) EpochStartEventHandler() epochStart.ActionHandler {
	subscribeHandler := notifier.NewHandlerForEpochStart(
		func(hdr data.HeaderHandler) {},
		func(_ data.HeaderHandler) {},
		common.EpochTxBroadcastDebug,
	)

	return subscribeHandler
}

// IsInterfaceNil returns true if there is no value under the interface
func (didh *disabledInterceptorDebugHandler) IsInterfaceNil() bool {
	return didh == nil
}
