package hashesCollector

import "github.com/multiversx/mx-chain-go/common"

type disabledHashesCollector struct {
}

// NewDisabledHashesCollector creates a new instance of disabledHashesCollector.
func NewDisabledHashesCollector() common.TrieHashesCollector {
	return &disabledHashesCollector{}
}

// AddDirtyHash does nothing.
func (hc *disabledHashesCollector) AddDirtyHash(_ []byte) {
}

// GetDirtyHashes returns an empty map.
func (hc *disabledHashesCollector) GetDirtyHashes() common.ModifiedHashes {
	return make(common.ModifiedHashes)
}

// AddObsoleteHashes does nothing.
func (hc *disabledHashesCollector) AddObsoleteHashes(_ []byte, _ [][]byte) {
}

// GetCollectedData returns nil data
func (hc *disabledHashesCollector) GetCollectedData() ([]byte, common.ModifiedHashes, common.ModifiedHashes) {
	return nil, nil, nil
}

// Clean does nothing.
func (hc *disabledHashesCollector) Clean() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (hc *disabledHashesCollector) IsInterfaceNil() bool {
	return hc == nil
}
