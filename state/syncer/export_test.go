package syncer

// GetNumHandlers -
func (mtnn *missingTrieNodesNotifier) GetNumHandlers() int {
	return len(mtnn.handlers)
}
