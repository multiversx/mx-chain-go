package trie

func (ts *trieSyncer) trieNodeIntercepted(hash []byte, val interface{}) {
	ts.mutOperation.Lock()
	defer ts.mutOperation.Unlock()

	log.Trace("trie node intercepted", "hash", hash)

	n, ok := ts.nodesForTrie[string(hash)]
	if !ok || n.received {
		return
	}

	interceptedNode, ok := val.(*InterceptedTrieNode)
	if !ok {
		return
	}

	ts.nodesForTrie[string(hash)] = trieNodeInfo{
		trieNode: interceptedNode.node,
		received: true,
	}
}
