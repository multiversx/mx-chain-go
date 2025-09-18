package dblookupext

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/storage"
)

type executionResultsProcessor struct {
	marshaller marshal.Marshalizer
	storer     storage.Storer
}

func newExecutionResultsProcessor(storer storage.Storer, marshaller marshal.Marshalizer) *executionResultsProcessor {
	return &executionResultsProcessor{
		marshaller: marshaller,
		storer:     storer,
	}
}

func (erp *executionResultsProcessor) saveExecutionResultsFromHeader(header data.HeaderHandler) error {
	for _, executionResult := range header.GetExecutionResultsHandlers() {
		if check.IfNil(executionResult) {
			continue
		}

		executionResultBytes, err := erp.marshaller.Marshal(executionResult)
		if err != nil {
			return err
		}

		err = erp.storer.PutInEpoch(executionResult.GetHeaderHash(), executionResultBytes, executionResult.GetHeaderEpoch())
		if err != nil {
			return err
		}
	}

	return nil
}
