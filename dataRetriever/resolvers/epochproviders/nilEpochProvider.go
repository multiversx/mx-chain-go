package epochproviders

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

// NilEpochProvider will be used for regular validator/observer nodes
type NilEpochProvider struct {
}

// EpochIsActiveInNetwork returns true
func (nep *NilEpochProvider) EpochIsActiveInNetwork(_ uint32) bool {
	return true
}

// EpochStartAction does nothing
func (nep *NilEpochProvider) EpochStartAction(_ data.HeaderHandler) {
}

// EpochStartPrepare does nothing
func (nep *NilEpochProvider) EpochStartPrepare(_ data.HeaderHandler, _ data.BodyHandler) {
}

// NotifyOrder will return the core.CurrentNetworkEpochProvider value
func (nep *NilEpochProvider) NotifyOrder() uint32 {
	return core.CurrentNetworkEpochProvider
}

// IsInterfaceNil returns true if there is no value under the interface
func (nep *NilEpochProvider) IsInterfaceNil() bool {
	return nep == nil
}
