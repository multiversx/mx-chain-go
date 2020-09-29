package epochproviders

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
)

type currentNetworkEpochProvider struct {
	currentEpoch        uint32
	mutCurrentEpoch     sync.RWMutex
	numActivePersisters int
}

// NewCurrentNetworkEpochProvider will return a new instance of currentNetworkEpochProvider
func NewCurrentNetworkEpochProvider(numActivePersisters int) *currentNetworkEpochProvider {
	return &currentNetworkEpochProvider{
		currentEpoch:        uint32(0),
		numActivePersisters: numActivePersisters,
	}
}

// SetCurrentEpoch will update the component's current epoch
func (cnep *currentNetworkEpochProvider) SetCurrentEpoch(epoch uint32) {
	// TODO: analyze where to call this from. For now, it is only called from epoch start bootstrapper so the value will
	// be accurate only when the node starts
	cnep.mutCurrentEpoch.Lock()
	cnep.currentEpoch = epoch
	cnep.mutCurrentEpoch.Unlock()
}

// EpochIsActiveInNetwork returns true if the persister for the given epoch is active in the network
func (cnep *currentNetworkEpochProvider) EpochIsActiveInNetwork(epoch uint32) bool {
	cnep.mutCurrentEpoch.RLock()
	defer cnep.mutCurrentEpoch.RUnlock()

	lower := core.MaxInt(int(cnep.currentEpoch)-cnep.numActivePersisters+1, 0)
	upper := cnep.currentEpoch

	return epoch >= uint32(lower) && epoch <= upper
}

// CurrentEpoch returns the current epoch in the network
func (cnep *currentNetworkEpochProvider) CurrentEpoch() uint32 {
	cnep.mutCurrentEpoch.RLock()
	defer cnep.mutCurrentEpoch.RUnlock()

	return cnep.currentEpoch
}

// IsInterfaceNil returns true if there is no value under the interface
func (cnep *currentNetworkEpochProvider) IsInterfaceNil() bool {
	return cnep == nil
}
