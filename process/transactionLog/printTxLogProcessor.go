package transactionLog

import (
	"encoding/hex"
	"strings"

	"github.com/multiversx/mx-chain-core-go/data"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

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
func (tlp *printTxLogProcessor) GetLogFromCache(_ []byte) (*data.LogData, bool) {
	return &data.LogData{}, false
}

// EnableLogToBeSavedInCache -
func (tlp *printTxLogProcessor) EnableLogToBeSavedInCache() {
}

// Clean -
func (tlp *printTxLogProcessor) Clean() {
}

// GetAllCurrentLogs -
func (tlp *printTxLogProcessor) GetAllCurrentLogs() []*data.LogData {
	return []*data.LogData{}
}

// SaveLog -
func (tlp *printTxLogProcessor) SaveLog(txHash []byte, _ data.TransactionHandler, logEntries []*vmcommon.LogEntry) error {
	if len(logEntries) == 0 {
		return nil
	}

	log.Debug("printTxLogProcessor.SaveLog", "transaction hash", hex.EncodeToString(txHash))
	for _, entry := range logEntries {
		log.Debug("printTxLogProcessor.entry",
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
