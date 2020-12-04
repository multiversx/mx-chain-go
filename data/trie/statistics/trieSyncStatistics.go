package statistics

import "sync"

type trieSyncStatistics struct {
	sync.RWMutex
	numReceived int
	numMissing  int
	missingMap  map[string]int
}

// NewTrieSyncStatistics returns a structure able to collect sync statistics from a trie and store them
func NewTrieSyncStatistics() *trieSyncStatistics {
	return &trieSyncStatistics{
		missingMap: make(map[string]int),
	}
}

// Reset will reset the contained values to 0
func (tss *trieSyncStatistics) Reset() {
	tss.Lock()
	tss.numReceived = 0
	tss.numMissing = 0
	tss.Unlock()
}

// AddNumReceived will add the provided value to the existing numReceived
func (tss *trieSyncStatistics) AddNumReceived(value int) {
	tss.Lock()
	tss.numReceived += value
	tss.Unlock()
}

// SetNumMissing will write the provided value on the existing numMissing
func (tss *trieSyncStatistics) SetNumMissing(rootHash []byte, value int) {
	tss.Lock()
	defer tss.Unlock()

	existing, found := tss.missingMap[string(rootHash)]
	if value == 0 {
		if !found {
			return
		}

		delete(tss.missingMap, string(rootHash))
		tss.numMissing -= existing
		return
	}

	tss.numMissing += value - existing
	tss.missingMap[string(rootHash)] = value
}

// NumReceived returns the received nodes
func (tss *trieSyncStatistics) NumReceived() int {
	tss.RLock()
	defer tss.RUnlock()

	return tss.numReceived
}

// NumMissing returns the missing nodes
func (tss *trieSyncStatistics) NumMissing() int {
	tss.RLock()
	defer tss.RUnlock()

	return tss.numMissing
}

// IsInterfaceNil returns true if there is no value under the interface
func (tss *trieSyncStatistics) IsInterfaceNil() bool {
	return tss == nil
}
