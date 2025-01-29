package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// ExtendedShardHeaderTrackerStub -
type ExtendedShardHeaderTrackerStub struct {
	BlockTrackerStub
	ComputeLongestExtendedShardChainFromLastNotarizedCalled func() ([]data.HeaderHandler, [][]byte, error)
	RemoveLastCrossNotarizedHeadersCalled                   func()
	RemoveLastSelfNotarizedHeadersCalled                    func()
}

// ComputeLongestExtendedShardChainFromLastNotarized -
func (eshts *ExtendedShardHeaderTrackerStub) ComputeLongestExtendedShardChainFromLastNotarized() ([]data.HeaderHandler, [][]byte, error) {
	if eshts.ComputeLongestExtendedShardChainFromLastNotarizedCalled != nil {
		return eshts.ComputeLongestMetaChainFromLastNotarizedCalled()
	}
	return nil, nil, nil
}

// RemoveLastCrossNotarizedHeaders -
func (eshts *ExtendedShardHeaderTrackerStub) RemoveLastCrossNotarizedHeaders() {
	if eshts.RemoveLastCrossNotarizedHeadersCalled != nil {
		eshts.RemoveLastCrossNotarizedHeadersCalled()
	}
}

// RemoveLastSelfNotarizedHeaders -
func (eshts *ExtendedShardHeaderTrackerStub) RemoveLastSelfNotarizedHeaders() {
	if eshts.RemoveLastSelfNotarizedHeadersCalled != nil {
		eshts.RemoveLastSelfNotarizedHeadersCalled()
	}
}

// IsGenesisLastCrossNotarizedHeader -
func (eshts *ExtendedShardHeaderTrackerStub) IsGenesisLastCrossNotarizedHeader() bool {
	return false
}
