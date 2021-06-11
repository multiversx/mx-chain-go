package transactionLog

import (
	"encoding/hex"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type printTxLogProcessor struct {
	storeLogsInCache bool
	logs             map[string]*transaction.Log
	mut              sync.RWMutex
	storer           storage.Storer
	marshalizer      marshal.Marshalizer
}

// NewPrintTxLogProcessor -
func NewPrintTxLogProcessor() *printTxLogProcessor {
	return &printTxLogProcessor{}
}

// GetLog -
func (tlp *printTxLogProcessor) GetLog(_ []byte) (data.LogHandler, error) {
	return nil, nil
}

// GetLogFromCache -
func (tlp *printTxLogProcessor) GetLogFromCache(_ []byte) (data.LogHandler, bool) {
	return nil, false
}

// EnableLogToBeSavedInCache -
func (tlp *printTxLogProcessor) EnableLogToBeSavedInCache() {
}

// Clean -
func (tlp *printTxLogProcessor) Clean() {
	return
}

// SaveLog -
func (tlp *printTxLogProcessor) SaveLog(txHash []byte, _ data.TransactionHandler, logEntries []*vmcommon.LogEntry) error {
	if len(logEntries) == 0 {
		return nil
	}

	log.Info("printTxLogProcessor.SaveLog", "transaction hash", hex.EncodeToString(txHash))
	for _, entry := range logEntries {
		log.Info("entry",
			"identifier", string(entry.Identifier),
			"address", entry.Address,
			"topics", prepareTopics(entry.Topics))
	}
	return nil
}

func prepareTopics(topics [][]byte) string {
	all := ""
	for _, topic := range topics {
		all += string(topic) + " "
	}

	return all
}

// IsInterfaceNil returns true if there is no value under the interface
func (tlp *printTxLogProcessor) IsInterfaceNil() bool {
	return tlp == nil
}
