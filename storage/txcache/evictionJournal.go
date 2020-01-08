package txcache

// evictionJournal keeps a short journal about the eviction process
// This is useful for debugging and reasoning about the eviction
type evictionJournal struct {
	evictionPerformed   bool
	passZeroNumTxs      uint32
	passZeroNumSenders  uint32
	passZeroNumSteps    uint32
	passOneNumTxs       uint32
	passOneNumSenders   uint32
	passTwoNumTxs       uint32
	passTwoNumSenders   uint32
	passThreeNumTxs     uint32
	passThreeNumSenders uint32
	passThreeNumSteps   uint32
}

func (journal *evictionJournal) display() {
	// log.Info("Eviction journal:")
	// log.Info("Pass 0:", "txs", journal.passOneNumTxs, "senders", journal.passOneNumSenders)
	// log.Info("Pass 1:", "txs", journal.passOneNumTxs, "senders", journal.passOneNumSenders)
	// log.Info("Pass 2:", "txs", journal.passTwoNumTxs, "senders", journal.passTwoNumSenders)
	// log.Info("Pass 3:", "steps", journal.passThreeNumSteps, "txs", journal.passThreeNumTxs, "senders", journal.passThreeNumSenders)
}
