package txcache

// evictionJournal keeps a short journal about the eviction process
// This is useful for debugging and reasoning about the eviction
type evictionJournal struct {
	evictionPerformed   bool
	passOneNumTxs       uint32
	passOneNumSenders   uint32
	passTwoNumTxs       uint32
	passTwoNumSenders   uint32
	passThreeNumTxs     uint32
	passThreeNumSenders uint32
	passThreeNumSteps   uint32
}

func (journal *evictionJournal) display() {
	log.Trace("Eviction journal:")
	log.Trace("1st pass:", "txs", journal.passOneNumTxs, "senders", journal.passOneNumSenders)
	log.Trace("2nd pass:", "txs", journal.passTwoNumTxs, "senders", journal.passTwoNumSenders)
	log.Trace("3rd pass:", "steps", journal.passThreeNumSteps, "txs", journal.passThreeNumTxs, "senders", journal.passThreeNumSenders)
}
