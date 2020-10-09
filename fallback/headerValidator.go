package fallback

import (
	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
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
		log.Debug("ShouldApplyFallbackValidation.GetMetaHeader", "error", err.Error())
		return false
	}

	isRoundTooOld := int64(headerHandler.GetRound())-int64(previousHeader.GetRound()) >= core.MaxRoundsWithoutCommittedStartInEpochBlock
	return isRoundTooOld
}

// IsInterfaceNil returns true if there is no value under the interface
func (fhv *fallbackHeaderValidator) IsInterfaceNil() bool {
	return fhv == nil
}
