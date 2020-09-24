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
func (cnrp *currentNetworkEpochProvider) SetCurrentEpoch(epoch uint32) {
	// TODO: analyze where to call this from. For now, the only solution seems to be by requesting
	// the last epoch start metablock from network. Analyze if a mechanism similar to epoch bootstrapper
	// is needed, so a consensus has to be made over the received meta block. This may be slow.
	cnrp.mutCurrentEpoch.Lock()
	cnrp.currentEpoch = epoch
	cnrp.mutCurrentEpoch.Unlock()
}

// EpochIsActiveInNetwork returns true if the persister for the given epoch is active in the network
func (cnrp *currentNetworkEpochProvider) EpochIsActiveInNetwork(epoch uint32) bool {
	cnrp.mutCurrentEpoch.RLock()
	defer cnrp.mutCurrentEpoch.RUnlock()

	lower := core.MaxInt(int(cnrp.currentEpoch)-cnrp.numActivePersisters+1, 0)
	upper := cnrp.currentEpoch

	return epoch >= uint32(lower) && epoch <= upper
}

// CurrentEpoch returns the
func (cnrp *currentNetworkEpochProvider) CurrentEpoch() uint32 {
	cnrp.mutCurrentEpoch.RLock()
	defer cnrp.mutCurrentEpoch.RUnlock()

	return cnrp.currentEpoch
}

// IsInterfaceNil returns true if there is no value under the interface
func (cnrp *currentNetworkEpochProvider) IsInterfaceNil() bool {
	return cnrp == nil
}
