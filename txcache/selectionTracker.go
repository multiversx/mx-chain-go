package txcache

type selectionTracker struct {
}

// Will receive TxCache pointer (interface) in constructor.
// In TxCache, we'll instantiate this: newSelectionTracker(this).

func onReceivedBlock(block any) {
	// Maybe block header and block body.
}
