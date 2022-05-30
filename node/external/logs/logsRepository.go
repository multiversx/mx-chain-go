package logs

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type logsRepository struct {
	storer      storage.Storer
	marshalizer marshal.Marshalizer
}

func newLogsRepository(storageService dataRetriever.StorageService, marshalizer marshal.Marshalizer) *logsRepository {
	storer := storageService.GetStorer(dataRetriever.TxLogsUnit)

	return &logsRepository{
		storer:      storer,
		marshalizer: marshalizer,
	}
}

func (repository *logsRepository) getLog(logKey []byte, epoch uint32) (*transaction.Log, error) {
	bytes, err := repository.storer.GetFromEpoch(logKey, epoch)
	if err != nil {
		return nil, err
	}

	txLog := &transaction.Log{}
	err = repository.marshalizer.Unmarshal(txLog, bytes)
	if err != nil {
		return nil, err
	}

	return txLog, nil
}

// Since a lookup table of logs is useful in the "logsFacade" to JOIN "txs" with "logs" ON "tx.hash" == "log.key",
// we'll return such a structure here (the returned logs don't have to be in any particular order).
func (repository *logsRepository) getLogs(logsKeys [][]byte, epoch uint32) (map[string]*transaction.Log, error) {
	keyValuePairs, err := repository.storer.GetBulkFromEpoch(logsKeys, epoch)
	if err != nil {
		return nil, err
	}

	results := make(map[string]*transaction.Log)

	for _, pair := range keyValuePairs {
		txLog := &transaction.Log{}
		err = repository.marshalizer.Unmarshal(txLog, pair.Value)
		if err != nil {
			return nil, err
		}

		results[string(pair.Key)] = txLog
	}

	return results, nil
}
