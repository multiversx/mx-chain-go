package processor

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type TrieNodeInterceptorProcessor struct {
	interceptedNodes storage.Cacher
}

func NewTrieNodesInterceptorProcessor(interceptedNodes storage.Cacher) (*TrieNodeInterceptorProcessor, error) {
	if check.IfNil(interceptedNodes) {
		return nil, process.ErrNilCacher
	}

	return &TrieNodeInterceptorProcessor{}, nil
}

func (tnip *TrieNodeInterceptorProcessor) Validate(data process.InterceptedData) error {
	return nil
}

func (tnip *TrieNodeInterceptorProcessor) Save(data process.InterceptedData) error {
	nodeData, ok := data.(*trie.InterceptedTrieNode)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	tnip.interceptedNodes.Put(nodeData.Hash(), nodeData)
	return nil
}

func (tnip *TrieNodeInterceptorProcessor) IsInterfaceNil() bool {
	if tnip == nil {
		return true
	}
	return false
}
