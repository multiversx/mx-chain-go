package trie

import "github.com/multiversx/mx-chain-go/storage"

// GetLeafHashesAndPutNodesInRamStorage will return the leaf node hashes and put the rest of nodes in a storer
func GetLeafHashesAndPutNodesInRamStorage(branchNodesMap map[string][]byte, db storage.Storer) ([][]byte, error) {
	leafHashes := make([][]byte, 0)
	for nodeHash, branchNodeSerialized := range branchNodesMap {
		decodedNode, err := decodeNode(branchNodeSerialized, nil, nil)
		if err != nil {
			return nil, err
		}

		hashes := decodedNode.getChildrenHashes()
		if len(hashes) == 0 {
			continue
		}

		leafHashes = append(leafHashes, hashes...)

		err = db.Put([]byte(nodeHash), branchNodeSerialized)
		if err != nil {
			return nil, err
		}
	}

	return leafHashes, nil
}
