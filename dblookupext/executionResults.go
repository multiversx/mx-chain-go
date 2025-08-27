package dblookupext

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/storage"
)

type executionResultsProcessor struct {
	marshalizer marshal.Marshalizer
	storer      storage.Storer
}

func newExecutionResultsProcessor(storer storage.Storer, marshalizer marshal.Marshalizer) *executionResultsProcessor {
	return &executionResultsProcessor{
		marshalizer: marshalizer,
		storer:      storer,
	}
}

func (erp *executionResultsProcessor) saveExecutionResultsFromHeader(header data.HeaderHandler) error {
	for _, executionResult := range header.GetExecutionResultsHandlers() {
		if executionResult.IsInterfaceNil() {
			continue
		}

		executionResultBytes, err := erp.marshalizer.Marshal(executionResult)
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
