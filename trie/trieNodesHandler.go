package trie

type trieNodesHandler struct {
	hashesOrder   []string
	existingNodes map[string]node
	missingHashes map[string]struct{}
}

func newTrieNodesHandler() *trieNodesHandler {
	return &trieNodesHandler{
		hashesOrder:   make([]string, 0),
		existingNodes: make(map[string]node),
		missingHashes: make(map[string]struct{}),
	}
}

func (handler *trieNodesHandler) addInitialRootHash(rootHash string) {
	handler.missingHashes[rootHash] = struct{}{}
	handler.hashesOrder = append(handler.hashesOrder, rootHash)
}

func (handler *trieNodesHandler) processMissingHashWasFound(n node, hash string) {
	handler.existingNodes[hash] = n
	delete(handler.missingHashes, hash)
}

func (handler *trieNodesHandler) jobDone() bool {
	return len(handler.hashesOrder) == 0
}

func (handler *trieNodesHandler) noMissingHashes() bool {
	return len(handler.missingHashes) == 0
}

func (handler *trieNodesHandler) hashIsMissing(hash string) bool {
	_, isMissing := handler.missingHashes[hash]
	return isMissing
}

func (handler *trieNodesHandler) getExistingNode(hash string) (node, bool) {
	element, exists := handler.existingNodes[hash]
	return element, exists
}

func (handler *trieNodesHandler) replaceParentWithChildren(index int, parentHash string, children []node, missingChildrenHashes [][]byte) {
	delete(handler.existingNodes, parentHash)

	allChildrenHashes := make([]string, 0, len(children)+len(missingChildrenHashes))
	for _, c := range children {
		hash := string(c.getHash())
		allChildrenHashes = append(allChildrenHashes, hash)
		handler.existingNodes[hash] = c
	}
	for _, m := range missingChildrenHashes {
		hash := string(m)
		allChildrenHashes = append(allChildrenHashes, string(m))
		handler.missingHashes[hash] = struct{}{}
	}

	handler.hashesOrder = replaceHashesAtPosition(index, handler.hashesOrder, allChildrenHashes)
}

func replaceHashesAtPosition(index int, initial []string, newData []string) []string {
	if index >= len(initial) || index < 0 {
		return initial
	}

	after := initial[index+1:]
	if len(newData) == 0 {
		copy(initial[index:], after)
		initial = initial[:len(initial)-1]
		return initial
	}

	//add some space
	initial = append(initial, newData[:len(newData)-1]...)
	copy(initial[index+len(newData):], after)
	copy(initial[index:], newData)
	return initial
}
