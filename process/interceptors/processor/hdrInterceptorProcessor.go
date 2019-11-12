package processor

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// HdrInterceptorProcessor is the processor used when intercepting headers
// (shard headers, meta headers) structs which satisfy HeaderHandler interface.
type HdrInterceptorProcessor struct {
	headers       storage.Cacher
	headersNonces dataRetriever.Uint64SyncMapCacher
	hdrValidator  process.HeaderValidator
	blackList     process.BlackListHandler
}

// NewHdrInterceptorProcessor creates a new TxInterceptorProcessor instance
func NewHdrInterceptorProcessor(argument *ArgHdrInterceptorProcessor) (*HdrInterceptorProcessor, error) {
	if argument == nil {
		return nil, process.ErrNilArguments
	}
	if check.IfNil(argument.Headers) {
		return nil, process.ErrNilCacher
	}
	if check.IfNil(argument.HeadersNonces) {
		return nil, process.ErrNilUint64SyncMapCacher
	}
	if check.IfNil(argument.HdrValidator) {
		return nil, process.ErrNilHdrValidator
	}
	if check.IfNil(argument.BlackList) {
		return nil, process.ErrNilBlackListHandler
	}

	return &HdrInterceptorProcessor{
		headers:       argument.Headers,
		headersNonces: argument.HeadersNonces,
		hdrValidator:  argument.HdrValidator,
		blackList:     argument.BlackList,
	}, nil
}

// Validate checks if the intercepted data can be processed
func (hip *HdrInterceptorProcessor) Validate(data process.InterceptedData) error {
	interceptedHdr, ok := data.(process.HdrValidatorHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	isBlackListed := hip.blackList.Has(string(interceptedHdr.Hash()))
	if isBlackListed {
		return process.ErrHeaderIsBlackListed
	}

	return hip.hdrValidator.HeaderValidForProcessing(interceptedHdr)
}

// Save will save the received data into the headers cacher as hash<->[plain header structure]
// and in headersNonces as nonce<->hash
func (hip *HdrInterceptorProcessor) Save(data process.InterceptedData) error {
	interceptedHdr, ok := data.(process.HdrValidatorHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	hip.headers.HasOrAdd(interceptedHdr.Hash(), interceptedHdr.HeaderHandler())

	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(interceptedHdr.HeaderHandler().GetShardID(), interceptedHdr.Hash())
	hip.headersNonces.Merge(interceptedHdr.HeaderHandler().GetNonce(), syncMap)

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hip *HdrInterceptorProcessor) IsInterfaceNil() bool {
	if hip == nil {
		return true
	}
	return false
}
