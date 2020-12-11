package disabled

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

// CurrentNetworkEpochProviderHandler -
type CurrentNetworkEpochProviderHandler struct {
}

// EpochStartAction does nothing
func (cneph *CurrentNetworkEpochProviderHandler) EpochStartAction(_ data.HeaderHandler) {
}

// EpochStartPrepare does nothing
func (cneph *CurrentNetworkEpochProviderHandler) EpochStartPrepare(_ data.HeaderHandler, _ data.BodyHandler) {
}

// NotifyOrder return core.CurrentNetworkEpochProvider value
func (cneph *CurrentNetworkEpochProviderHandler) NotifyOrder() uint32 {
	return core.CurrentNetworkEpochProvider
}

// EpochIsActiveInNetwork returns true
func (cneph *CurrentNetworkEpochProviderHandler) EpochIsActiveInNetwork(_ uint32) bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (cneph *CurrentNetworkEpochProviderHandler) IsInterfaceNil() bool {
	return cneph == nil
}
