package processor

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

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
func (tnip *TrieNodeInterceptorProcessor) Validate(data process.InterceptedData) error {
	return nil
}

// Save saves the intercepted trie node in the intercepted nodes cacher
func (tnip *TrieNodeInterceptorProcessor) Save(data process.InterceptedData) error {
	nodeData, ok := data.(*trie.InterceptedTrieNode)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	tnip.interceptedNodes.Put(nodeData.Hash(), nodeData)
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tnip *TrieNodeInterceptorProcessor) IsInterfaceNil() bool {
	if tnip == nil {
		return true
	}
	return false
}
