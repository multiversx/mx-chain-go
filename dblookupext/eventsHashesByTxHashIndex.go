//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. resultsHashesByTxHash.proto

package dblookupext

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common/logging"
	"github.com/multiversx/mx-chain-go/storage"
)

type eventsHashesByTxHash struct {
	marshalizer marshal.Marshalizer
	storer      storage.Storer
}

func newEventsHashesByTxHash(storer storage.Storer, marshalizer marshal.Marshalizer) *eventsHashesByTxHash {
	return &eventsHashesByTxHash{
		marshalizer: marshalizer,
		storer:      storer,
	}
}

func (eht *eventsHashesByTxHash) saveResultsHashes(epoch uint32, scResults, receipts map[string]data.TransactionHandler) error {
	resultsHashes := eht.groupTransactionOutcomesByTransactionHash(epoch, scResults, receipts)
	results := eht.mergeRecordsFromStorageIfExists(resultsHashes)

	for txHash, resultHashes := range results {
		resultHashesBytes, err := eht.marshalizer.Marshal(resultHashes)
		if err != nil {
			continue
		}

		err = eht.storer.Put([]byte(txHash), resultHashesBytes)
		if err != nil {
			logging.LogErrAsWarnExceptAsDebugIfClosingError(log, err,
				"saveResultsHashes() cannot save resultHashesByte",
				"err", err.Error())
			continue
		}
	}

	return nil
}

func (eht *eventsHashesByTxHash) groupTransactionOutcomesByTransactionHash(
	epoch uint32,
	scResults,
	receipts map[string]data.TransactionHandler,
) map[string]*ResultsHashesByTxHash {
	resultsHashesMap := eht.groupSmartContractResults(epoch, scResults)

	for receiptHash, receiptHandler := range receipts {
		rec, ok := receiptHandler.(*receipt.Receipt)
		if !ok {
			log.Error("groupTransactionOutcomesByTransactionHash() cannot cast TransactionHandler to Receipt")
			continue
		}

		originalTxHash := string(rec.TxHash)
		resultsHashesMap[originalTxHash] = &ResultsHashesByTxHash{
			ReceiptsHash: []byte(receiptHash),
		}
	}

	return resultsHashesMap
}

func (eht *eventsHashesByTxHash) groupSmartContractResults(
	epoch uint32,
	scrResults map[string]data.TransactionHandler,
) map[string]*ResultsHashesByTxHash {
	resultsHashesMap := make(map[string]*ResultsHashesByTxHash)
	for scrHash, scrHandler := range scrResults {
		scrResult, ok := scrHandler.(*smartContractResult.SmartContractResult)
		if !ok {
			log.Error("groupSmartContractResults() cannot cast TransactionHandler to SmartContractResult")
			continue
		}

		originalTxHash := string(scrResult.OriginalTxHash)

		if _, hashExists := resultsHashesMap[originalTxHash]; !hashExists {
			resultsHashesMap[originalTxHash] = &ResultsHashesByTxHash{
				ScResultsHashesAndEpoch: []*ScResultsHashesAndEpoch{
					{
						Epoch: epoch,
					},
				},
			}
		}
		scResultsHashesAndEpoch := resultsHashesMap[originalTxHash].ScResultsHashesAndEpoch[0]
		scResultsHashesAndEpoch.ScResultsHashes = append(scResultsHashesAndEpoch.ScResultsHashes, []byte(scrHash))
	}

	return resultsHashesMap
}

func (eht *eventsHashesByTxHash) mergeRecordsFromStorageIfExists(
	records map[string]*ResultsHashesByTxHash,
) map[string]*ResultsHashesByTxHash {
	for originalTxHash, results := range records {
		rawBytes, err := eht.storer.Get([]byte(originalTxHash))
		if err != nil {
			continue
		}

		record := &ResultsHashesByTxHash{}
		err = eht.marshalizer.Unmarshal(record, rawBytes)
		if err != nil {
			continue
		}

		results.ScResultsHashesAndEpoch = append(record.ScResultsHashesAndEpoch, results.ScResultsHashesAndEpoch...)
	}

	return records
}

func (eht *eventsHashesByTxHash) getEventsHashesByTxHash(txHash []byte, epoch uint32) (*ResultsHashesByTxHash, error) {
	rawBytes, err := eht.storer.GetFromEpoch(txHash, epoch)
	if err != nil {
		if storage.IsNotFoundInStorageErr(err) {
			err = fmt.Errorf("%w: %v", storage.ErrKeyNotFound, err)
		}

		return nil, err
	}

	record := &ResultsHashesByTxHash{}
	err = eht.marshalizer.Unmarshal(record, rawBytes)
	if err != nil {
		return nil, err
	}

	return record, nil
}
