package processor

import (
	"bytes"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

// ArgsSovereignHeaderInterceptorProcessor is a struct placeholder used to create a new sovereign extended header interceptor processor
type ArgsSovereignHeaderInterceptorProcessor struct {
	BlockBlackList           process.TimeCacher
	Hasher                   hashing.Hasher
	Marshaller               marshal.Marshalizer
	IncomingHeaderSubscriber process.IncomingHeaderSubscriber
	HeadersPool              dataRetriever.HeadersPool
}

type sovereignHeaderInterceptorProcessor struct {
	blackList                process.TimeCacher
	Hasher                   hashing.Hasher
	Marshaller               marshal.Marshalizer
	IncomingHeaderSubscriber process.IncomingHeaderSubscriber
	headersPool              dataRetriever.HeadersPool
}

// NewSovereignHdrInterceptorProcessor creates a new sovereign extended header interceptor processor
func NewSovereignHdrInterceptorProcessor(args *ArgsSovereignHeaderInterceptorProcessor) (*sovereignHeaderInterceptorProcessor, error) {
	if args == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(args.BlockBlackList) {
		return nil, process.ErrNilBlackListCacher
	}

	return &sovereignHeaderInterceptorProcessor{
		blackList:                args.BlockBlackList,
		Hasher:                   args.Hasher,
		Marshaller:               args.Marshaller,
		IncomingHeaderSubscriber: args.IncomingHeaderSubscriber,
		headersPool:              args.HeadersPool,
	}, nil
}

// Validate checks if the intercepted data can be processed
func (hip *sovereignHeaderInterceptorProcessor) Validate(data process.InterceptedData, _ core.PeerID) error {
	interceptedHdr, ok := data.(process.ExtendedHeaderValidatorHandler)
	if !ok {
		return fmt.Errorf("sovereignHeaderInterceptorProcessor.Validate error: %w", process.ErrWrongTypeAssertion)
	}

	hip.blackList.Sweep()
	isBlackListed := hip.blackList.Has(string(interceptedHdr.Hash()))
	if isBlackListed {
		return process.ErrHeaderIsBlackListed
	}

	extendedHdr := interceptedHdr.GetExtendedHeader()
	computedExtendedHeader, err := hip.IncomingHeaderSubscriber.CreateExtendedHeader(extendedHdr)
	if err != nil {
		return err
	}

	computedHeaderHash, err := core.CalculateHash(hip.Marshaller, hip.Hasher, computedExtendedHeader)
	if bytes.Compare(computedHeaderHash, interceptedHdr.Hash()) != 0 {
		return errors.ErrInvalidReceivedSovereignProof
	}

	return nil
}

// Save will save the received data into headers pool
func (hip *sovereignHeaderInterceptorProcessor) Save(data process.InterceptedData, _ core.PeerID, _ string) error {
	interceptedHdr, ok := data.(process.ExtendedHeaderValidatorHandler)
	if !ok {
		return fmt.Errorf("sovereignHeaderInterceptorProcessor.Save error: %w", process.ErrWrongTypeAssertion)
	}

	if _, err := hip.headersPool.GetHeaderByHash(interceptedHdr.Hash()); err == nil {
		return nil
	}

	return hip.IncomingHeaderSubscriber.AddHeader(interceptedHdr.Hash(), interceptedHdr.GetExtendedHeader())
}

// RegisterHandler does nothing
func (hip *sovereignHeaderInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (hip *sovereignHeaderInterceptorProcessor) IsInterfaceNil() bool {
	return hip == nil
}
