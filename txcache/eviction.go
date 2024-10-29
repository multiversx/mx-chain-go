package txcache

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// evictionJournal keeps a short journal about the eviction process
// This is useful for debugging and reasoning about the eviction
type evictionJournal struct {
	numTxs uint32
}

// doEviction does cache eviction
// We do not allow more evictions to start concurrently
func (cache *TxCache) doEviction() *evictionJournal {
	if cache.isEvictionInProgress.IsSet() {
		return nil
	}

	if !cache.isCapacityExceeded() {
		return nil
	}

	cache.evictionMutex.Lock()
	defer cache.evictionMutex.Unlock()

	_ = cache.isEvictionInProgress.SetReturningPrevious()
	defer cache.isEvictionInProgress.Reset()

	if !cache.isCapacityExceeded() {
		return nil
	}

	logRemove.Debug("doEviction(): before eviction",
		"num bytes", cache.NumBytes(),
		"num txs", cache.CountTx(),
		"num senders", cache.CountSenders(),
	)

	stopWatch := core.NewStopWatch()
	stopWatch.Start("eviction")

	// TODO: reimplement.
	evictionJournal := evictionJournal{}

	stopWatch.Stop("eviction")

	logRemove.Debug(
		"doEviction(): after eviction",
		"num bytes", cache.NumBytes(),
		"num now", cache.CountTx(),
		"num senders", cache.CountSenders(),
		"duration", stopWatch.GetMeasurement("eviction"),
		"evicted txs", evictionJournal.numTxs,
	)

	return &evictionJournal
}

func (cache *TxCache) isCapacityExceeded() bool {
	return cache.areThereTooManyBytes() || cache.areThereTooManySenders() || cache.areThereTooManyTxs()
}

func (cache *TxCache) areThereTooManyBytes() bool {
	numBytes := cache.NumBytes()
	tooManyBytes := numBytes > int(cache.config.NumBytesThreshold)
	return tooManyBytes
}

func (cache *TxCache) areThereTooManySenders() bool {
	numSenders := cache.CountSenders()
	tooManySenders := numSenders > uint64(cache.config.CountThreshold)
	return tooManySenders
}

func (cache *TxCache) areThereTooManyTxs() bool {
	numTxs := cache.CountTx()
	tooManyTxs := numTxs > uint64(cache.config.CountThreshold)
	return tooManyTxs
}
