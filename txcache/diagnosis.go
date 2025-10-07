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

type printedTrackedBlock struct {
	Hash         string `json:"hash"`
	PreviousHash string `json:"previousHash"`
	RootHash     string `json:"rootHash"`
	Nonce        uint64 `json:"nonce"`
}

// Diagnose checks the state of the cache for inconsistencies and displays a summary, senders and transactions.
func (cache *TxCache) Diagnose(_ bool) {
	cache.diagnoseTransactions()
}

func (cache *TxCache) GetDimensionOfTrackedBlocks() uint64 {
	return cache.tracker.getDimensionOfTrackedBlocks()
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

// marshalTrackedBlockToNewlineDelimitedJSON converts a list of tracked blocks to a newline-delimited JSON string.
// Note: each line is indexed, to improve readability. The index is easily removable if separate analysis is needed.
func marshalTrackedBlockToNewlineDelimitedJSON(trackedBlocks []*trackedBlock, linePrefix string) string {
	builder := strings.Builder{}
	builder.WriteString("\n")

	for i, block := range trackedBlocks {
		printedBlock := convertTrackedBlockToPrintedBlock(block)
		printedBlockJSON, _ := json.Marshal(printedBlock)

		builder.WriteString(fmt.Sprintf("%s#%d: ", linePrefix, i))
		builder.WriteString(string(printedBlockJSON))
		builder.WriteString("\n")
	}

	builder.WriteString("\n")
	return builder.String()
}

func convertTrackedBlockToPrintedBlock(block *trackedBlock) *printedTrackedBlock {
	return &printedTrackedBlock{
		Hash:         hex.EncodeToString(block.hash),
		PreviousHash: hex.EncodeToString(block.prevHash),
		RootHash:     hex.EncodeToString(block.rootHash),
		Nonce:        block.nonce,
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

func displayTrackedBlocks(contextualLogger logger.Logger, linePrefix string, trackedBlocks []*trackedBlock) {
	if contextualLogger.GetLevel() > logger.LogTrace {
		return
	}

	if len(trackedBlocks) > 0 {
		contextualLogger.Trace("displayTrackedBlocks - trackedBlocks (as newline-separated JSON):")
		contextualLogger.Trace(marshalTrackedBlockToNewlineDelimitedJSON(trackedBlocks, linePrefix))
	} else {
		contextualLogger.Trace("displayTrackedBlocks - trackedBlocks: none")
	}
}
