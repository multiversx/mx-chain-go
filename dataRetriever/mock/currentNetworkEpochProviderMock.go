package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
)

type CurrentNetworkEpochProviderStub struct {
	currentEpoch        uint32
	mutCurrentEpoch     sync.RWMutex
	numActivePersisters int
}

// NewCurrentNetworkEpochProvider will return a new instance of CurrentNetworkEpochProviderStub
func NewCurrentNetworkEpochProvider(numActivePersisters int) *CurrentNetworkEpochProviderStub {
	return &CurrentNetworkEpochProviderStub{
		currentEpoch:        uint32(0),
		numActivePersisters: numActivePersisters,
	}
}

// SetCurrentEpoch will update the component's current epoch
func (cnep *CurrentNetworkEpochProviderStub) SetCurrentEpoch(epoch uint32) {
	cnep.mutCurrentEpoch.Lock()
	cnep.currentEpoch = epoch
	cnep.mutCurrentEpoch.Unlock()
}

// EpochIsActiveInNetwork returns true if the persister for the given epoch is active in the network
func (cnep *CurrentNetworkEpochProviderStub) EpochIsActiveInNetwork(epoch uint32) bool {
	cnep.mutCurrentEpoch.RLock()
	defer cnep.mutCurrentEpoch.RUnlock()

	lower := core.MaxInt(int(cnep.currentEpoch)-cnep.numActivePersisters+1, 0)
	upper := cnep.currentEpoch

	return epoch >= uint32(lower) && epoch <= upper
}

// CurrentEpoch returns the current network epoch
func (cnep *CurrentNetworkEpochProviderStub) CurrentEpoch() uint32 {
	cnep.mutCurrentEpoch.RLock()
	defer cnep.mutCurrentEpoch.RUnlock()

	return cnep.currentEpoch
}

// IsInterfaceNil returns true if there is no value under the interface
func (cnep *CurrentNetworkEpochProviderStub) IsInterfaceNil() bool {
	return cnep == nil
}
