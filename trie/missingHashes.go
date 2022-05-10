package trie

type missingHashes struct {
	innerMap map[string]struct{}
	order    []string
}

func newMissingHashes() *missingHashes {
	return &missingHashes{
		innerMap: make(map[string]struct{}),
		order:    make([]string, 0),
	}
}

func (missing *missingHashes) add(hash string) {
	if missing.has(hash) {
		return
	}

	missing.order = append(missing.order, hash)
	missing.innerMap[hash] = struct{}{}
}

func (missing *missingHashes) has(hash string) bool {
	_, found := missing.innerMap[hash]

	return found
}

func (missing *missingHashes) remove(hash string) {
	if !missing.has(hash) {
		return
	}

	delete(missing.innerMap, hash)
	foundIdx := 0
	for idx := range missing.order {
		if missing.order[idx] == hash {
			foundIdx = idx
		}
	}
	missing.order = append(missing.order[:foundIdx], missing.order[foundIdx+1:]...)
}

func (missing *missingHashes) getHashesSliceWithCopy() []string {
	newOrder := make([]string, len(missing.order))
	copy(newOrder, missing.order)

	return newOrder
}

func (missing *missingHashes) len() int {
	return len(missing.order)
}
