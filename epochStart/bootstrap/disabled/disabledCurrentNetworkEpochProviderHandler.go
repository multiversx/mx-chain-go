package disabled

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

// currentNetworkEpochProviderHandler -
type currentNetworkEpochProviderHandler struct {
}

// NewCurrentNetworkEpochProviderHandler will create a new currentNetworkEpochProviderHandler instance
func NewCurrentNetworkEpochProviderHandler() *currentNetworkEpochProviderHandler {
	return &currentNetworkEpochProviderHandler{}
}

// EpochStartAction does nothing
func (cneph *currentNetworkEpochProviderHandler) EpochStartAction(_ data.HeaderHandler) {
}

// EpochStartPrepare does nothing
func (cneph *currentNetworkEpochProviderHandler) EpochStartPrepare(_ data.HeaderHandler, _ data.BodyHandler) {
}

// NotifyOrder return core.CurrentNetworkEpochProvider value
func (cneph *currentNetworkEpochProviderHandler) NotifyOrder() uint32 {
	return core.CurrentNetworkEpochProvider
}

// EpochIsActiveInNetwork returns true
func (cneph *currentNetworkEpochProviderHandler) EpochIsActiveInNetwork(_ uint32) bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (cneph *currentNetworkEpochProviderHandler) IsInterfaceNil() bool {
	return cneph == nil
}
