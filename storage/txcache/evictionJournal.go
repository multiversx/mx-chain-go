package txcache

// evictionJournal keeps a short journal about the eviction process
// This is useful for debugging and reasoning about the eviction
type evictionJournal struct {
	passOneNumTxs       uint32
	passOneNumSenders   uint32
	passTwoNumTxs       uint32
	passTwoNumSenders   uint32
	passThreeNumTxs     uint32
	passThreeNumSenders uint32
	passThreeNumSteps   uint32
}

func (journal *evictionJournal) display() {
	log.Info("Eviction journal:")
	log.Info("1st pass:", "txs", journal.passOneNumTxs, "senders", journal.passOneNumSenders)
	log.Info("2nd pass:", "txs", journal.passTwoNumTxs, "senders", journal.passTwoNumSenders)
	log.Info("3rd pass:", "steps", journal.passThreeNumSteps, "txs", journal.passThreeNumTxs, "senders", journal.passThreeNumSenders)
}
