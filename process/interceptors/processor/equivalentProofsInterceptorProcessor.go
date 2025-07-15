package processor

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
)

// equivalentProofsInterceptorProcessor is the processor used when intercepting equivalent proofs
type equivalentProofsInterceptorProcessor struct {
}

// NewEquivalentProofsInterceptorProcessor creates a new equivalentProofsInterceptorProcessor
func NewEquivalentProofsInterceptorProcessor() *equivalentProofsInterceptorProcessor {
	return &equivalentProofsInterceptorProcessor{}
}

// Validate checks if the intercepted data can be processed
// returns nil as proper validity checks are done at intercepted data level
func (epip *equivalentProofsInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save returns nil
// proof is added after validity checks, at intercepted data level
func (epip *equivalentProofsInterceptorProcessor) Save(_ process.InterceptedData, _ core.PeerID, _ string) error {
	return nil
}

// RegisterHandler registers a callback function to be notified of incoming equivalent proofs
func (epip *equivalentProofsInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("equivalentProofsInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (epip *equivalentProofsInterceptorProcessor) IsInterfaceNil() bool {
	return epip == nil
}
