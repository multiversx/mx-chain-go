package common

// ModifiedHashes is used to memorize all old hashes and new hashes from when a trie is committed
type ModifiedHashes map[string]struct{}

// Clone is used to create a clone of the map
func (mh ModifiedHashes) Clone() ModifiedHashes {
	newMap := make(ModifiedHashes)
	for key := range mh {
		newMap[key] = struct{}{}
	}

	return newMap
}
