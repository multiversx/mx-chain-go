package fallback

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

// TODO: move constant to config
const SupernovaMaxRoundsWithoutCommittedStartInEpochBlock = 500

var log = logger.GetOrCreate("fallback")

type fallbackHeaderValidator struct {
	headersPool         dataRetriever.HeadersPool
	marshalizer         marshal.Marshalizer
	storageService      dataRetriever.StorageService
	enableEpochsHandler common.EnableEpochsHandler
}

// NewFallbackHeaderValidator creates a new fallbackHeaderValidator object
func NewFallbackHeaderValidator(
	headersPool dataRetriever.HeadersPool,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
	enableEpochsHandler common.EnableEpochsHandler,
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
	if check.IfNil(enableEpochsHandler) {
		return nil, errors.ErrNilEnableEpochsHandler
	}

	hv := &fallbackHeaderValidator{
		headersPool:         headersPool,
		marshalizer:         marshalizer,
		storageService:      storageService,
		enableEpochsHandler: enableEpochsHandler,
	}

	return hv, nil
}

// ShouldApplyFallbackValidationForHeaderWith returns if for the given header data fallback validation could be applied or not
func (fhv *fallbackHeaderValidator) ShouldApplyFallbackValidationForHeaderWith(shardID uint32, startOfEpochBlock bool, round uint64, prevHeaderHash []byte) bool {
	if shardID != core.MetachainShardId {
		return false
	}
	if !startOfEpochBlock {
		return false
	}

	previousHeader, err := process.GetMetaHeader(prevHeaderHash, fhv.headersPool, fhv.marshalizer, fhv.storageService)
	if err != nil {
		log.Debug("ShouldApplyFallbackValidation", "GetMetaHeader", err.Error())
		return false
	}

	isRoundTooOld := int64(round)-int64(previousHeader.GetRound()) >= fhv.getMaxRoundsWithoutCommitedStartInEpochBlock()
	return isRoundTooOld
}

func (fhv *fallbackHeaderValidator) getMaxRoundsWithoutCommitedStartInEpochBlock() int64 {
	if fhv.enableEpochsHandler.IsFlagEnabled(common.SupernovaFlag) {
		return SupernovaMaxRoundsWithoutCommittedStartInEpochBlock
	}

	return common.MaxRoundsWithoutCommittedStartInEpochBlock
}

// ShouldApplyFallbackValidation returns if for the given header could be applied fallback validation or not
func (fhv *fallbackHeaderValidator) ShouldApplyFallbackValidation(headerHandler data.HeaderHandler) bool {
	if check.IfNil(headerHandler) {
		return false
	}

	return fhv.ShouldApplyFallbackValidationForHeaderWith(headerHandler.GetShardID(), headerHandler.IsStartOfEpochBlock(), headerHandler.GetRound(), headerHandler.GetPrevHash())
}

// IsInterfaceNil returns true if there is no value under the interface
func (fhv *fallbackHeaderValidator) IsInterfaceNil() bool {
	return fhv == nil
}
