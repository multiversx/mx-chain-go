package txcache

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type printedTransaction struct {
	Hash       string `json:"hash"`
	PPU        uint64 `json:"ppu"`
	Nonce      uint64 `json:"nonce"`
	Sender     string `json:"sender"`
	GasPrice   uint64 `json:"gasPrice"`
	GasLimit   uint64 `json:"gasLimit"`
	Receiver   string `json:"receiver"`
	DataLength int    `json:"dataLength"`
}

// Diagnose checks the state of the cache for inconsistencies and displays a summary, senders and transactions.
func (cache *TxCache) Diagnose(_ bool) {
	cache.diagnoseCounters()
	cache.diagnoseTransactions()
}

func (cache *TxCache) diagnoseCounters() {
	if log.GetLevel() > logger.LogDebug {
		return
	}

	sizeInBytes := cache.NumBytes()
	numTxsEstimate := int(cache.CountTx())
	numTxsInChunks := cache.txByHash.backingMap.Count()
	txsKeys := cache.txByHash.backingMap.Keys()
	numSendersEstimate := int(cache.CountSenders())
	numSendersInChunks := cache.txListBySender.backingMap.Count()
	sendersKeys := cache.txListBySender.backingMap.Keys()

	fine := numSendersEstimate == numSendersInChunks
	fine = fine && (int(numSendersEstimate) == len(sendersKeys))
	fine = fine && (numTxsEstimate == numTxsInChunks && numTxsEstimate == len(txsKeys))

	log.Debug("diagnoseCounters",
		"fine", fine,
		"numTxsEstimate", numTxsEstimate,
		"numTxsInChunks", numTxsInChunks,
		"len(txsKeys)", len(txsKeys),
		"sizeInBytes", sizeInBytes,
		"numBytesThreshold", cache.config.NumBytesThreshold,
		"numSendersEstimate", numSendersEstimate,
		"numSendersInChunks", numSendersInChunks,
		"len(sendersKeys)", len(sendersKeys),
	)
}

func (cache *TxCache) diagnoseTransactions() {
	if logDiagnoseTransactions.GetLevel() > logger.LogTrace {
		return
	}

	transactions := cache.getAllTransactions()
	if len(transactions) == 0 {
		return
	}

	numToDisplay := core.MinInt(diagnosisMaxTransactionsToDisplay, len(transactions))
	logDiagnoseTransactions.Trace("diagnoseTransactions", "numTransactions", len(transactions), "numToDisplay", numToDisplay)
	logDiagnoseTransactions.Trace(marshalTransactionsToNewlineDelimitedJSON(transactions[:numToDisplay], "diagnoseTransactions"))
}

// marshalTransactionsToNewlineDelimitedJSON converts a list of transactions to a newline-delimited JSON string.
// Note: each line is indexed, to improve readability. The index is easily removable if separate analysis is needed.
func marshalTransactionsToNewlineDelimitedJSON(transactions []*WrappedTransaction, linePrefix string) string {
	builder := strings.Builder{}
	builder.WriteString("\n")

	for i, wrappedTx := range transactions {
		printedTx := convertWrappedTransactionToPrintedTransaction(wrappedTx)
		printedTxJSON, _ := json.Marshal(printedTx)

		builder.WriteString(fmt.Sprintf("%s#%d: ", linePrefix, i))
		builder.WriteString(string(printedTxJSON))
		builder.WriteString("\n")
	}

	builder.WriteString("\n")
	return builder.String()
}

func convertWrappedTransactionToPrintedTransaction(wrappedTx *WrappedTransaction) *printedTransaction {
	transaction := wrappedTx.Tx

	return &printedTransaction{
		Hash:       hex.EncodeToString(wrappedTx.TxHash),
		Nonce:      transaction.GetNonce(),
		Receiver:   hex.EncodeToString(transaction.GetRcvAddr()),
		Sender:     hex.EncodeToString(transaction.GetSndAddr()),
		GasPrice:   transaction.GetGasPrice(),
		GasLimit:   transaction.GetGasLimit(),
		DataLength: len(transaction.GetData()),
		PPU:        wrappedTx.PricePerUnit,
	}
}

func displaySelectionOutcome(contextualLogger logger.Logger, linePrefix string, transactions []*WrappedTransaction) {
	if contextualLogger.GetLevel() > logger.LogTrace {
		return
	}

	if len(transactions) > 0 {
		contextualLogger.Trace("displaySelectionOutcome - transactions (as newline-separated JSON):")
		contextualLogger.Trace(marshalTransactionsToNewlineDelimitedJSON(transactions, linePrefix))
	} else {
		contextualLogger.Trace("displaySelectionOutcome - transactions: none")
	}
}
