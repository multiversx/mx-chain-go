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
	if index >= len(initial) {
		return initial
	}

	before := initial[:index]
	after := initial[index+1:]

	result := make([]string, 0, len(before)+len(after)+len(newData))
	result = append(result, before...)
	result = append(result, newData...)
	result = append(result, after...)

	return result
}
