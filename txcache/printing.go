package txcache

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type printedTransaction struct {
	Hash       string `json:"hash"`
	Nonce      uint64 `json:"nonce"`
	GasPrice   uint64 `json:"gasPrice"`
	GasLimit   uint64 `json:"gasLimit"`
	Sender     string `json:"sender"`
	Receiver   string `json:"receiver"`
	DataLength int    `json:"dataLength"`
	Fee        string `json:"fee"`
}

type printedSender struct {
	Address       string `json:"address"`
	Nonce         uint64 `json:"nonce"`
	IsNonceKnown  bool   `json:"isNonceKnown"`
	HasInitialGap bool   `json:"hasInitialGap"`
	NumTxs        uint64 `json:"numTxs"`
}

// marshalSendersToNewlineDelimitedJson converts a list of senders to a newline-delimited JSON string.
// Note: each line is indexed, to improve readability. The index is easily removable for if separate analysis is needed.
func marshalSendersToNewlineDelimitedJson(senders []*txListForSender) string {
	builder := strings.Builder{}
	builder.WriteString("\n")

	for i, txListForSender := range senders {
		printedSender := convertTxListForSenderToPrintedSender(txListForSender)
		printedSenderJson, _ := json.Marshal(printedSender)

		builder.WriteString(fmt.Sprintf("#%d: ", i))
		builder.WriteString(string(printedSenderJson))
		builder.WriteString("\n")
	}

	builder.WriteString("\n")
	return builder.String()
}

// marshalTransactionsToNewlineDelimitedJson converts a list of transactions to a newline-delimited JSON string.
// Note: each line is indexed, to improve readability. The index is easily removable for if separate analysis is needed.
func marshalTransactionsToNewlineDelimitedJson(transactions []*WrappedTransaction) string {
	builder := strings.Builder{}
	builder.WriteString("\n")

	for i, wrappedTx := range transactions {
		printedTx := convertWrappedTransactionToPrintedTransaction(wrappedTx)
		printedTxJson, _ := json.Marshal(printedTx)

		builder.WriteString(fmt.Sprintf("#%d: ", i))
		builder.WriteString(string(printedTxJson))
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
		Fee:        wrappedTx.TxFee.String(),
	}
}

func convertTxListForSenderToPrintedSender(txListForSender *txListForSender) *printedSender {
	return &printedSender{
		Address:       hex.EncodeToString([]byte(txListForSender.sender)),
		Nonce:         txListForSender.accountNonce.Get(),
		IsNonceKnown:  txListForSender.accountNonceKnown.IsSet(),
		HasInitialGap: txListForSender.hasInitialGapWithLock(),
		NumTxs:        txListForSender.countTxWithLock(),
	}
}
