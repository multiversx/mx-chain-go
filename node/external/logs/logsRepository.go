package logs

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/disabled"
)

type logsRepository struct {
	storer     storage.Storer
	marshaller marshal.Marshalizer
}

func newLogsRepository(storageService dataRetriever.StorageService, marshaller marshal.Marshalizer) *logsRepository {
	storer := storageService.GetStorer(dataRetriever.TxLogsUnit)
	if check.IfNil(storer) {
		log.Debug("could not find TxLogsUnit, using a disabled storer instead...")
		storer = disabled.NewStorer()
	}

	return &logsRepository{
		storer:     storer,
		marshaller: marshaller,
	}
}

func (repository *logsRepository) getLog(logKey []byte, epoch uint32) (*transaction.Log, error) {
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

// Since a lookup table of logs is useful in the "logsFacade" to JOIN "txs" with "logs" ON "tx.hash" == "log.key",
// we'll return such a structure here (the returned logs don't have to be in any particular order).
func (repository *logsRepository) getLogs(logsKeys [][]byte, epoch uint32) (map[string]*transaction.Log, error) {
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
