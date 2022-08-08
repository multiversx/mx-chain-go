package txcache

func (cache *TxCache) initSweepable() {
	cache.sweepingListOfSenders = make([]*txListForSender, 0, estimatedNumOfSweepableSendersPerSelection)
}

func (cache *TxCache) collectSweepable(list *txListForSender) {
	if !list.sweepable.IsSet() {
		return
	}

	cache.sweepingMutex.Lock()
	cache.sweepingListOfSenders = append(cache.sweepingListOfSenders, list)
	cache.sweepingMutex.Unlock()
}

func (cache *TxCache) sweepSweepable() {
	cache.sweepingMutex.Lock()
	defer cache.sweepingMutex.Unlock()

	if len(cache.sweepingListOfSenders) == 0 {
		return
	}

	stopWatch := cache.monitorSweepingStart()
	numTxs, numSenders := cache.evictSendersAndTheirTxs(cache.sweepingListOfSenders)
	cache.initSweepable()
	cache.monitorSweepingEnd(numTxs, numSenders, stopWatch)
}
