package processor

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ process.InterceptorProcessor = (*TrieNodeInterceptorProcessor)(nil)

// TrieNodeInterceptorProcessor is the processor used when intercepting trie nodes
type TrieNodeInterceptorProcessor struct {
	interceptedNodes storage.Cacher
}

// NewTrieNodesInterceptorProcessor creates a new instance of TrieNodeInterceptorProcessor
func NewTrieNodesInterceptorProcessor(interceptedNodes storage.Cacher) (*TrieNodeInterceptorProcessor, error) {
	if check.IfNil(interceptedNodes) {
		return nil, process.ErrNilCacher
	}

	return &TrieNodeInterceptorProcessor{
		interceptedNodes: interceptedNodes,
	}, nil
}

// Validate checks if the intercepted data can be processed
func (tnip *TrieNodeInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save saves the intercepted trie node in the intercepted nodes cacher
func (tnip *TrieNodeInterceptorProcessor) Save(data process.InterceptedData, _ core.PeerID, _ string) error {
	nodeData, ok := data.(*trie.InterceptedTrieNode)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	tnip.interceptedNodes.Put(nodeData.Hash(), nodeData, nodeData.SizeInBytes())
	return nil
}

// RegisterHandler registers a callback function to be notified of incoming trie nodes
func (tnip *TrieNodeInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("trieNodeInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (tnip *TrieNodeInterceptorProcessor) IsInterfaceNil() bool {
	return tnip == nil
}
