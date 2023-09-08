package processor

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
)

// IncomingHeaderSubscriber defines a subscriber to incoming headers
type IncomingHeaderSubscriber interface {
	AddHeader(headerHash []byte, header sovereign.IncomingHeaderHandler) error
	IsInterfaceNil() bool
}

// ArgsSovereignHeaderInterceptorProcessor is a struct placeholder used to create a new sovereign extended header interceptor processor
type ArgsSovereignHeaderInterceptorProcessor struct {
	BlockBlackList           process.TimeCacher
	Hasher                   hashing.Hasher
	Marshaller               marshal.Marshalizer
	IncomingHeaderSubscriber IncomingHeaderSubscriber
}

type sovereignHeaderInterceptorProcessor struct {
	blackList                process.TimeCacher
	Hasher                   hashing.Hasher
	Marshaller               marshal.Marshalizer
	IncomingHeaderSubscriber IncomingHeaderSubscriber
}

// NewSovereignHdrInterceptorProcessor creates a new sovereign extended header interceptor processor
func NewSovereignHdrInterceptorProcessor(argument *ArgsSovereignHeaderInterceptorProcessor) (*sovereignHeaderInterceptorProcessor, error) {
	if argument == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(argument.BlockBlackList) {
		return nil, process.ErrNilBlackListCacher
	}

	return &sovereignHeaderInterceptorProcessor{
		blackList:                argument.BlockBlackList,
		Hasher:                   argument.Hasher,
		Marshaller:               argument.Marshaller,
		IncomingHeaderSubscriber: argument.IncomingHeaderSubscriber,
	}, nil
}

// Validate checks if the intercepted data can be processed
func (hip *sovereignHeaderInterceptorProcessor) Validate(data process.InterceptedData, _ core.PeerID) error {
	_, ok := data.(process.ExtendedHeaderValidatorHandler)
	if !ok {
		return fmt.Errorf("sovereignHeaderInterceptorProcessor.Validate error: %w", process.ErrWrongTypeAssertion)
	}

	return nil
}

// Save will save the received data into headers pool
func (hip *sovereignHeaderInterceptorProcessor) Save(data process.InterceptedData, _ core.PeerID, _ string) error {
	_, ok := data.(process.ExtendedHeaderValidatorHandler)
	if !ok {
		return fmt.Errorf("sovereignHeaderInterceptorProcessor.Save error: %w", process.ErrWrongTypeAssertion)
	}

	return nil
}

// RegisterHandler does nothing
func (hip *sovereignHeaderInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (hip *sovereignHeaderInterceptorProcessor) IsInterfaceNil() bool {
	return hip == nil
}
