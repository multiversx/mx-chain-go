package logs

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("node/external/logs")

type logsRepository struct {
	storer          storage.Storer
	marshalizer     marshal.Marshalizer
	pubKeyConverter core.PubkeyConverter
}

func NewLogsRepository(args ArgsNewLogsRepository) (*logsRepository, error) {
	err := args.check()
	if err != nil {
		return nil, err
	}

	return &logsRepository{
		storer:          args.Storer,
		marshalizer:     args.Marshalizer,
		pubKeyConverter: args.PubKeyConverter,
	}, nil
}

func (repository *logsRepository) GetLog(txHash []byte, epoch uint32) (*transaction.ApiLogs, error) {
	bytes, err := repository.storer.GetFromEpoch(txHash, epoch)
	if err != nil {
		return nil, err
	}

	log := &transaction.Log{}
	err = repository.marshalizer.Unmarshal(log, bytes)
	if err != nil {
		return nil, err
	}

	apiLogs := repository.convertTxLogToApiLogs(txHash, log)
	return apiLogs, nil
}

func (repository *logsRepository) GetLogs(txHashes [][]byte, epoch uint32) ([]*transaction.ApiLogs, error) {
	keyValuePairs, err := repository.storer.GetBulkFromEpoch(txHashes, epoch)
	if err != nil {
		return nil, err
	}

	results := make([]*transaction.ApiLogs, 0, len(txHashes))

	for _, pair := range keyValuePairs {
		txLog := &transaction.Log{}
		err = repository.marshalizer.Unmarshal(txLog, pair.Value)
		if err != nil {
			log.Warn("GetLogs() / Unmarshal()", "hash", pair.Key, "epoch", epoch, "err", err)
			continue
		}

		apiLogs := repository.convertTxLogToApiLogs(pair.Key, txLog)
		results = append(results, apiLogs)
	}

	return results, nil
}

func (repository *logsRepository) IncludeLogsInTransactions(txs []*transaction.ApiTransactionResult, logsHashes [][]byte, epoch uint32) error {
	txsLookup := make(map[string]*transaction.ApiTransactionResult, len(txs))

	for _, tx := range txs {
		txsLookup[string(tx.HashBytes)] = tx
	}

	logs, err := repository.GetLogs(logsHashes, epoch)
	if err != nil {
		return err
	}

	for _, logEntry := range logs {
		tx, ok := txsLookup[string(logEntry.TxHashBytes)]
		if ok {
			tx.Logs = logEntry
		}
	}

	return nil
}

func (repository *logsRepository) convertTxLogToApiLogs(txHash []byte, log *transaction.Log) *transaction.ApiLogs {
	events := make([]*transaction.Events, len(log.Events))

	for i, event := range log.Events {
		events[i] = &transaction.Events{
			Address:    repository.pubKeyConverter.Encode(event.Address),
			Identifier: string(event.Identifier),
			Topics:     event.Topics,
			Data:       event.Data,
		}
	}

	return &transaction.ApiLogs{
		TxHashBytes: txHash,
		Address:     repository.pubKeyConverter.Encode(log.Address),
		Events:      events,
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (repository *logsRepository) IsInterfaceNil() bool {
	return repository == nil
}
