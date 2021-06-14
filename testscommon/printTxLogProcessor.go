package testscommon

import (
	"encoding/hex"
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
)

var log = logger.GetOrCreate("testscommon/transactionLog")

type printTxLogProcessor struct {
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
}

// SaveLog -
func (tlp *printTxLogProcessor) SaveLog(txHash []byte, _ data.TransactionHandler, logEntries []*vmcommon.LogEntry) error {
	if len(logEntries) == 0 {
		return nil
	}

	log.Info("printTxLogProcessor.SaveLog", "transaction hash", hex.EncodeToString(txHash))
	for _, entry := range logEntries {
		log.Debug("entry",
			"identifier", string(entry.Identifier),
			"address", entry.Address,
			"topics", prepareTopics(entry.Topics))
	}
	return nil
}

func prepareTopics(topics [][]byte) string {
	all := ""
	for _, topic := range topics {
		all = strings.Join([]string{all, string(topic)}, " ")
	}

	return all
}

// IsInterfaceNil returns true if there is no value under the interface
func (tlp *printTxLogProcessor) IsInterfaceNil() bool {
	return tlp == nil
}
