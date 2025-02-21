package processor

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
)

// ArgEquivalentProofsInterceptorProcessor is the argument for the interceptor processor used for equivalent proofs
type ArgEquivalentProofsInterceptorProcessor struct {
	EquivalentProofsPool EquivalentProofsPool
	Marshaller           marshal.Marshalizer
}

// equivalentProofsInterceptorProcessor is the processor used when intercepting equivalent proofs
type equivalentProofsInterceptorProcessor struct {
	equivalentProofsPool EquivalentProofsPool
	marshaller           marshal.Marshalizer
}

// NewEquivalentProofsInterceptorProcessor creates a new equivalentProofsInterceptorProcessor
func NewEquivalentProofsInterceptorProcessor(args ArgEquivalentProofsInterceptorProcessor) (*equivalentProofsInterceptorProcessor, error) {
	err := checkArgsEquivalentProofs(args)
	if err != nil {
		return nil, err
	}

	return &equivalentProofsInterceptorProcessor{
		equivalentProofsPool: args.EquivalentProofsPool,
		marshaller:           args.Marshaller,
	}, nil
}

func checkArgsEquivalentProofs(args ArgEquivalentProofsInterceptorProcessor) error {
	if check.IfNil(args.EquivalentProofsPool) {
		return process.ErrNilProofsPool
	}
	if check.IfNil(args.Marshaller) {
		return process.ErrNilMarshalizer
	}

	return nil
}

// Validate checks if the intercepted data can be processed
// returns nil as proper validity checks are done at intercepted data level
func (epip *equivalentProofsInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will save the intercepted equivalent proof inside the proofs tracker
func (epip *equivalentProofsInterceptorProcessor) Save(data process.InterceptedData, _ core.PeerID, _ string) error {
	interceptedProof, ok := data.(interceptedEquivalentProof)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	wasAdded := epip.equivalentProofsPool.AddProof(interceptedProof.GetProof())
	if !wasAdded {
		return common.ErrAlreadyExistingEquivalentProof
	}

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
