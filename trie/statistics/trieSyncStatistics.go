package statistics

import (
	"sync"
	"time"
)

type trieSyncStatistics struct {
	sync.RWMutex
	numReceived      int
	numMissing       int
	numLarge         int
	missingMap       map[string]int
	numBytesReceived uint64
	numIterations    int
	processingTime   time.Duration
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
	tss.numLarge = 0
	tss.numBytesReceived = 0
	tss.missingMap = make(map[string]int)
	tss.numIterations = 0
	tss.processingTime = time.Duration(0)
	tss.Unlock()
}

// AddNumReceived will add the provided value to the existing numReceived
func (tss *trieSyncStatistics) AddNumReceived(value int) {
	tss.Lock()
	tss.numReceived += value
	tss.Unlock()
}

// AddNumBytesReceived will add the provided value to the existing numBytesReceived
func (tss *trieSyncStatistics) AddNumBytesReceived(numBytes uint64) {
	tss.Lock()
	tss.numBytesReceived += numBytes
	tss.Unlock()
}

// AddNumLarge will add the provided value to the existing numLarge
func (tss *trieSyncStatistics) AddNumLarge(value int) {
	tss.Lock()
	tss.numLarge += value
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

// AddProcessingTime will add the processing time to the existing value
func (tss *trieSyncStatistics) AddProcessingTime(duration time.Duration) {
	tss.Lock()
	tss.processingTime += duration
	tss.Unlock()
}

// IncrementIteration will increment the iterations done on all trie syncers
func (tss *trieSyncStatistics) IncrementIteration() {
	tss.Lock()
	tss.numIterations++
	tss.Unlock()
}

// NumReceived returns the received nodes
func (tss *trieSyncStatistics) NumReceived() int {
	tss.RLock()
	defer tss.RUnlock()

	return tss.numReceived
}

// NumLarge returns the received large nodes
func (tss *trieSyncStatistics) NumLarge() int {
	tss.RLock()
	defer tss.RUnlock()

	return tss.numLarge
}

// NumMissing returns the missing nodes
func (tss *trieSyncStatistics) NumMissing() int {
	tss.RLock()
	defer tss.RUnlock()

	return tss.numMissing
}

// NumBytesReceived returns the number of bytes received
func (tss *trieSyncStatistics) NumBytesReceived() uint64 {
	tss.RLock()
	defer tss.RUnlock()

	return tss.numBytesReceived
}

// NumTries returns the number of tries that are currently syncing
func (tss *trieSyncStatistics) NumTries() int {
	tss.RLock()
	defer tss.RUnlock()

	return len(tss.missingMap)
}

// ProcessingTime will return the cumulative processing time used in tries synchronization (sum of all go routines used in trie sync processes)
func (tss *trieSyncStatistics) ProcessingTime() time.Duration {
	tss.RLock()
	defer tss.RUnlock()

	return tss.processingTime
}

// NumIterations returns the total iterations done by all go routines used in the trie sync processes
func (tss *trieSyncStatistics) NumIterations() int {
	tss.RLock()
	defer tss.RUnlock()

	return tss.numIterations
}

// IsInterfaceNil returns true if there is no value under the interface
func (tss *trieSyncStatistics) IsInterfaceNil() bool {
	return tss == nil
}
