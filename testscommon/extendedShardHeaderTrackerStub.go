package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// ExtendedShardHeaderTrackerStub -
type ExtendedShardHeaderTrackerStub struct {
	BlockTrackerStub
	ComputeLongestExtendedShardChainFromLastNotarizedCalled func() ([]data.HeaderHandler, [][]byte, error)
}

// ComputeLongestExtendedShardChainFromLastNotarized -
func (eshts *ExtendedShardHeaderTrackerStub) ComputeLongestExtendedShardChainFromLastNotarized() ([]data.HeaderHandler, [][]byte, error) {
	if eshts.ComputeLongestExtendedShardChainFromLastNotarizedCalled != nil {
		return eshts.ComputeLongestMetaChainFromLastNotarizedCalled()
	}
	return nil, nil, nil
}
