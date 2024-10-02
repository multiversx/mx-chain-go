package trie

import (
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/storage"
)

// GetLeafHashesAndPutNodesInRamStorage will return the leaf node hashes and put the rest of nodes in a storer
func GetLeafHashesAndPutNodesInRamStorage(
	branchNodesMap map[string][]byte,
	db storage.Persister,
	hasher hashing.Hasher,
	marshaller marshal.Marshalizer,
) ([][]byte, error) {
	leafHashes := make([][]byte, 0)
	for nodeHash, branchNodeSerialized := range branchNodesMap {
		decodedNode, err := decodeNode(branchNodeSerialized, marshaller, hasher)
		if err != nil {
			return nil, err
		}

		childrenHashes := decodedNode.getChildrenHashes()
		if len(childrenHashes) == 0 {
			continue
		}

		leafHashes = append(leafHashes, getLeafHashesFromChildrenHashes(childrenHashes, branchNodesMap)...)

		err = db.Put([]byte(nodeHash), branchNodeSerialized)
		if err != nil {
			return nil, err
		}
	}

	return leafHashes, nil
}

func getLeafHashesFromChildrenHashes(childrenHashes [][]byte, nodesMap map[string][]byte) [][]byte {
	leafHashes := make([][]byte, 0)
	for _, childHash := range childrenHashes {
		_, isBranchNodeOrExtensionNode := nodesMap[string(childHash)]
		if isBranchNodeOrExtensionNode {
			continue
		}

		leafHashes = append(leafHashes, childHash)
	}

	return leafHashes
}
