package disabled

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

// epochProvider will be used for regular validator/observer nodes
type epochProvider struct {
}

// NewEpochProvider returns a new disabled epoch provider instance
func NewEpochProvider() *epochProvider {
	return &epochProvider{}
}

// EpochIsActiveInNetwork returns true
func (ep *epochProvider) EpochIsActiveInNetwork(_ uint32) bool {
	return true
}

// EpochStartAction does nothing
func (ep *epochProvider) EpochStartAction(_ data.HeaderHandler) {
}

// EpochStartPrepare does nothing
func (ep *epochProvider) EpochStartPrepare(_ data.HeaderHandler, _ data.BodyHandler) {
}

// NotifyOrder will return the core.CurrentNetworkEpochProvider value
func (ep *epochProvider) NotifyOrder() uint32 {
	return core.CurrentNetworkEpochProvider
}

// IsInterfaceNil returns true if there is no value under the interface
func (ep *epochProvider) IsInterfaceNil() bool {
	return ep == nil
}
