//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. resultsHashesByTxHash.proto

package dblookupext

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
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
			log.Warn("saveResultsHashes() cannot save resultHashesByte",
				"error", err.Error())
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
	resultsHashesMap := make(map[string]*ResultsHashesByTxHash, 0)
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
		return nil, err
	}

	record := &ResultsHashesByTxHash{}
	err = eht.marshalizer.Unmarshal(record, rawBytes)
	if err != nil {
		return nil, err
	}

	return record, nil
}
