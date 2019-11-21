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

// SignalEndOfProcessing signals the end of processing
func (tnip *TrieNodeInterceptorProcessor) SignalEndOfProcessing(data []process.InterceptedData) {
	nodeData, ok := data[0].(*trie.InterceptedTrieNode)
	if !ok {
		log.Debug("intercepted data is not a trie node")
		return
	}

	// TODO instead of using a node to trigger the end of processing, use a dedicated channel
	//  between interceptor and sync
	nodeData.CreateEndOfProcessingTriggerNode()
	err := tnip.Save(nodeData)
	if err != nil {
		log.Debug(err.Error())
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (tnip *TrieNodeInterceptorProcessor) IsInterfaceNil() bool {
	if tnip == nil {
		return true
	}
	return false
}
