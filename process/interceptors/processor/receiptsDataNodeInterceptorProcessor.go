package processor

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

var _ process.InterceptorProcessor = (*receiptsNodeInterceptorProcessor)(nil)

// receiptsNodeInterceptorProcessor is the processor used when intercepting receipts nodes
type receiptsNodeInterceptorProcessor struct {
	interceptedNodes storage.Cacher
}

// NewReceiptsNodesInterceptorProcessor creates a new instance of ReceiptsNodeInterceptorProcessor
func NewReceiptsNodesInterceptorProcessor(interceptedNodes storage.Cacher) (*receiptsNodeInterceptorProcessor, error) {
	if check.IfNil(interceptedNodes) {
		return nil, process.ErrNilCacher
	}

	return &receiptsNodeInterceptorProcessor{
		interceptedNodes: interceptedNodes,
	}, nil
}

// Validate checks if the intercepted data can be processed
func (p *receiptsNodeInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save saves the intercepted trie node in the intercepted nodes cacher
func (p *receiptsNodeInterceptorProcessor) Save(data process.InterceptedData, _ core.PeerID, _ string) error {
	nodeData, ok := data.(interceptedDataSizeHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	p.interceptedNodes.Put(data.Hash(), nodeData, nodeData.SizeInBytes()+len(data.Hash()))
	return nil
}

// RegisterHandler registers a callback function to be notified of incoming trie nodes
func (p *receiptsNodeInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("ReceiptsNodeInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *receiptsNodeInterceptorProcessor) IsInterfaceNil() bool {
	return p == nil
}
