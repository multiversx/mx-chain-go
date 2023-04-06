package logs

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("node/external/logs")

type logsFacade struct {
	repository *logsRepository
	converter  *logsConverter
}

// NewLogsFacade creates a new logs facade
func NewLogsFacade(args ArgsNewLogsFacade) (*logsFacade, error) {
	err := args.check()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errCannotCreateLogsFacade, err)
	}

	repository := newLogsRepository(args.StorageService, args.Marshaller)
	converter := newLogsConverter(args.PubKeyConverter)

	return &logsFacade{
		repository: repository,
		converter:  converter,
	}, nil
}

// GetLog loads a transaction log (from storage)
func (facade *logsFacade) GetLog(logKey []byte, epoch uint32) (*transaction.ApiLogs, error) {
	txLog, err := facade.repository.getLog(logKey, epoch)
	if err != nil {
		return nil, err
	}

	apiResource := facade.converter.txLogToApiResource(logKey, txLog)

	return apiResource, nil
}

// IncludeLogsInTransactions loads transaction logs from storage and includes them in the provided transaction objects
// Note: the transaction objects MUST have the field "HashBytes" set in advance.
func (facade *logsFacade) IncludeLogsInTransactions(txs []*transaction.ApiTransactionResult, logsKeys [][]byte, epoch uint32) error {
	logsByKey, err := facade.repository.getLogs(logsKeys, epoch)
	if err != nil {
		return err
	}

	for _, tx := range txs {
		key := tx.HashBytes
		txLog, ok := logsByKey[string(key)]
		if ok {
			tx.Logs = facade.converter.txLogToApiResource(key, txLog)
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (facade *logsFacade) IsInterfaceNil() bool {
	return facade == nil
}
