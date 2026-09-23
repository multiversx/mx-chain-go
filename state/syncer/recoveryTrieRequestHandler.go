package syncer

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/trie"
)

type trieNodesForEpochRequester interface {
	RequestTrieNodesForEpoch(destShardID uint32, hashes [][]byte, topic string, epoch uint32)
}

type recoveryTrieRequestHandler struct {
	trie.RequestHandler
	requester trieNodesForEpochRequester
	epoch     uint32
}

func newRecoveryTrieRequestHandler(handler trie.RequestHandler, epoch uint32) (trie.RequestHandler, error) {
	requester, ok := handler.(trieNodesForEpochRequester)
	if !ok {
		return nil, fmt.Errorf("trie request handler does not support epoch-specific requests")
	}
	return &recoveryTrieRequestHandler{
		RequestHandler: handler,
		requester:      requester,
		epoch:          epoch,
	}, nil
}

func (handler *recoveryTrieRequestHandler) RequestTrieNodes(destShardID uint32, hashes [][]byte, topic string) {
	handler.requester.RequestTrieNodesForEpoch(destShardID, hashes, topic, handler.epoch)
}
