package fees

import (
	"bytes"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("fees/common")

// GetAccumulatedFeesInEpoch returns the accumulated fees in epoch from the header
func GetAccumulatedFeesInEpoch(
	epochStartMetaBlock data.MetaHeaderHandler,
	headersPool dataRetriever.HeadersPool,
	marshaller marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) *big.Int {
	if checkNilParams(epochStartMetaBlock, headersPool, marshaller, storageService) {
		return big.NewInt(0)
	}

	if !epochStartMetaBlock.IsHeaderV3() {
		return epochStartMetaBlock.GetAccumulatedFeesInEpoch()
	}

	execResultWithFees := getExecutionResultWithFees(epochStartMetaBlock, headersPool, marshaller, storageService)
	return execResultWithFees.GetAccumulatedFeesInEpoch()
}

// GetDeveloperFeesInEpoch returns the developer fees in epoch from the header
func GetDeveloperFeesInEpoch(
	epochStartMetaBlock data.MetaHeaderHandler,
	headersPool dataRetriever.HeadersPool,
	marshaller marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) *big.Int {
	if checkNilParams(epochStartMetaBlock, headersPool, marshaller, storageService) {
		return big.NewInt(0)
	}

	if !epochStartMetaBlock.IsHeaderV3() {
		return epochStartMetaBlock.GetDevFeesInEpoch()
	}

	execResultWithFees := getExecutionResultWithFees(epochStartMetaBlock, headersPool, marshaller, storageService)
	return execResultWithFees.GetDevFeesInEpoch()
}

func checkNilParams(
	epochStartMetaBlock data.MetaHeaderHandler,
	headersPool dataRetriever.HeadersPool,
	marshaller marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) bool {
	if check.IfNil(epochStartMetaBlock) {
		return true
	}
	if check.IfNil(headersPool) {
		return true
	}
	if check.IfNil(marshaller) {
		return true
	}
	if check.IfNil(storageService) {
		return true
	}
	return false
}

func getExecutionResultWithFees(
	epochStartMetaBlock data.MetaHeaderHandler,
	headersPool dataRetriever.HeadersPool,
	marshaller marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) data.BaseMetaExecutionResultHandler {
	execResultWithNoFees := &block.BaseMetaExecutionResult{
		AccumulatedFeesInEpoch: big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
	}

	// For epoch start block, search for the epoch change proposed block
	for _, execResult := range epochStartMetaBlock.GetExecutionResultsHandlers() {
		executedMetaHeaderHash := execResult.GetHeaderHash()
		executedMetaHeader, err := process.GetMetaHeader(
			executedMetaHeaderHash,
			headersPool,
			marshaller,
			storageService,
		)
		if err != nil {
			log.Error("getExecutionResultWithFees: cannot find executed header", "error", err)
			// saving a metric should not be a blocking error, we will return empty fee fields
			return execResultWithNoFees
		}

		if !executedMetaHeader.IsEpochChangeProposed() {
			continue
		}

		// search for the execution result of previous hash of proposed epoch change header
		prevHash := executedMetaHeader.GetPrevHash()
		for _, execResultOnProposedEpochChangeHeader := range executedMetaHeader.GetExecutionResultsHandlers() {
			if bytes.Equal(execResultOnProposedEpochChangeHeader.GetHeaderHash(), prevHash) {
				baseMetaExecResult, ok := execResultOnProposedEpochChangeHeader.(data.BaseMetaExecutionResultHandler)
				if !ok {
					log.Warn("getExecutionResultWithFees: cannot cast execResultOnProposedEpochChangeHeader to BaseMetaExecutionResult")
					continue
				}

				return baseMetaExecResult
			}
		}
	}

	log.Error("getExecutionResultWithFees: cannot find prev header of epoch change proposed")
	// saving a metric should not be a blocking error, we will return empty fee fields
	return execResultWithNoFees
}
