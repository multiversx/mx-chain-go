package fallback

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("fallback")

type fallbackHeaderValidator struct {
	headersPool    dataRetriever.HeadersPool
	marshalizer    marshal.Marshalizer
	storageService dataRetriever.StorageService
}

// NewFallbackHeaderValidator creates a new fallbackHeaderValidator object
func NewFallbackHeaderValidator(
	headersPool dataRetriever.HeadersPool,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (*fallbackHeaderValidator, error) {

	if check.IfNil(headersPool) {
		return nil, process.ErrNilHeadersDataPool
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(storageService) {
		return nil, process.ErrNilStorage
	}

	hv := &fallbackHeaderValidator{
		headersPool:    headersPool,
		marshalizer:    marshalizer,
		storageService: storageService,
	}

	return hv, nil
}

// ShouldApplyFallbackValidation returns if for the given header could be applied fallback validation or not
func (fhv *fallbackHeaderValidator) ShouldApplyFallbackValidation(headerHandler data.HeaderHandler) bool {
	if check.IfNil(headerHandler) {
		return false
	}
	if headerHandler.GetShardID() != core.MetachainShardId {
		return false
	}
	if !headerHandler.IsStartOfEpochBlock() {
		return false
	}

	previousHeader, err := process.GetMetaHeader(headerHandler.GetPrevHash(), fhv.headersPool, fhv.marshalizer, fhv.storageService)
	if err != nil {
		log.Debug("ShouldApplyFallbackValidation", "GetMetaHeader", err.Error())
		return false
	}

	isRoundTooOld := int64(headerHandler.GetRound())-int64(previousHeader.GetRound()) >= common.MaxRoundsWithoutCommittedStartInEpochBlock
	return isRoundTooOld
}

// IsInterfaceNil returns true if there is no value under the interface
func (fhv *fallbackHeaderValidator) IsInterfaceNil() bool {
	return fhv == nil
}
