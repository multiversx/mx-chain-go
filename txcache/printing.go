package txcache

import (
	"encoding/hex"
	"encoding/json"
	"strings"
)

type printedTransaction struct {
	Hash     string `json:"hash"`
	Nonce    uint64 `json:"nonce"`
	GasPrice uint64 `json:"gasPrice"`
	GasLimit uint64 `json:"gasLimit"`
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
}

type printedSender struct {
	Address      string `json:"address"`
	Score        int    `json:"score"`
	Nonce        uint64 `json:"nonce"`
	IsNonceKnown bool   `json:"isNonceKnown"`
}

func marshalSendersToNewlineDelimitedJson(senders []*txListForSender) string {
	builder := strings.Builder{}
	builder.WriteString("\n")

	for _, txListForSender := range senders {
		printedSender := convertTxListForSenderToPrintedSender(txListForSender)
		printedSenderJson, _ := json.Marshal(printedSender)
		builder.WriteString(string(printedSenderJson))
	}

	builder.WriteString("\n")
	return builder.String()
}

func marshalTransactionsToNewlineDelimitedJson(transactions []*WrappedTransaction) string {
	builder := strings.Builder{}
	builder.WriteString("\n")

	for _, wrappedTx := range transactions {
		printedTx := convertWrappedTransactionToPrintedTransaction(wrappedTx)
		printedTxJson, _ := json.Marshal(printedTx)
		builder.WriteString(string(printedTxJson))
	}

	builder.WriteString("\n")
	return builder.String()
}

func convertWrappedTransactionToPrintedTransaction(wrappedTx *WrappedTransaction) *printedTransaction {
	transaction := wrappedTx.Tx

	return &printedTransaction{
		Hash:     hex.EncodeToString(wrappedTx.TxHash),
		Nonce:    transaction.GetNonce(),
		Receiver: hex.EncodeToString(transaction.GetRcvAddr()),
		Sender:   hex.EncodeToString(transaction.GetSndAddr()),
		GasPrice: transaction.GetGasPrice(),
		GasLimit: transaction.GetGasLimit(),
	}
}

func convertTxListForSenderToPrintedSender(txListForSender *txListForSender) *printedSender {
	return &printedSender{
		Address:      hex.EncodeToString([]byte(txListForSender.sender)),
		Score:        txListForSender.getScore(),
		Nonce:        txListForSender.accountNonce.Get(),
		IsNonceKnown: txListForSender.accountNonceKnown.IsSet(),
	}
}
