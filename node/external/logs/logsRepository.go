package logs

import (
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
)

type logsRepository struct {
	storer     storage.Storer
	marshaller marshal.Marshalizer
}

func newLogsRepository(storageService dataRetriever.StorageService, marshaller marshal.Marshalizer) *logsRepository {
	storer, err := storageService.GetStorer(dataRetriever.TxLogsUnit)
	if err != nil {
		return nil
	}

	return &logsRepository{
		storer:     storer,
		marshaller: marshaller,
	}
}

// getLog loads event & logs on a best-effort basis.
// It searches for logs both in epoch N and epoch N-1, in order to overcome epoch-changing edge-cases.
func (repository *logsRepository) getLog(logKey []byte, epoch uint32) (*transaction.Log, error) {
	txLog, err := repository.doGetLog(logKey, epoch)
	if err != nil && epoch > 0 {
		txLog, errPreviousEpoch := repository.doGetLog(logKey, epoch-1)
		if errPreviousEpoch == nil {
			return txLog, nil
		}

		// We ignore errPreviousEpoch.
	}

	return txLog, err
}

func (repository *logsRepository) doGetLog(logKey []byte, epoch uint32) (*transaction.Log, error) {
	bytes, err := repository.storer.GetFromEpoch(logKey, epoch)
	if err != nil {
		return nil, fmt.Errorf("%w: %v, epoch = %d, key = %s", errCannotLoadLogs, err, epoch, hex.EncodeToString(logKey))
	}

	txLog := &transaction.Log{}
	err = repository.marshaller.Unmarshal(txLog, bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v, epoch = %d, key = %s", errCannotUnmarshalLog, err, epoch, hex.EncodeToString(logKey))
	}

	return txLog, nil
}

// getLogs loads event & logs on a best-effort basis.
// It searches for logs both in epoch N and epoch N-1, in order to overcome epoch-changing edge-cases.
func (repository *logsRepository) getLogs(logsKeys [][]byte, epoch uint32) (map[string]*transaction.Log, error) {
	logsMap, err := repository.doGetLogs(logsKeys, epoch)
	if err != nil {
		return nil, err
	}

	if epoch > 0 {
		logsMapPreviousEpoch, err := repository.doGetLogs(logsKeys, epoch-1)
		if err != nil {
			return nil, err
		}

		for key, value := range logsMapPreviousEpoch {
			logsMap[key] = value
		}
	}

	return logsMap, nil
}

// Since a lookup table of logs is useful in the "logsFacade" to JOIN "txs" with "logs" ON "tx.hash" == "log.key",
// we'll return such a structure here (the returned logs don't have to be in any particular order).
func (repository *logsRepository) doGetLogs(logsKeys [][]byte, epoch uint32) (map[string]*transaction.Log, error) {
	keyValuePairs, err := repository.storer.GetBulkFromEpoch(logsKeys, epoch)
	if err != nil {
		return nil, fmt.Errorf("%w: %v, epoch = %d", errCannotLoadLogs, err, epoch)
	}

	results := make(map[string]*transaction.Log)

	for _, pair := range keyValuePairs {
		txLog := &transaction.Log{}
		err = repository.marshaller.Unmarshal(txLog, pair.Value)
		if err != nil {
			return nil, fmt.Errorf("%w: %v, epoch = %d, key = %s", errCannotUnmarshalLog, err, epoch, hex.EncodeToString(pair.Key))
		}

		results[string(pair.Key)] = txLog
	}

	return results, nil
}
