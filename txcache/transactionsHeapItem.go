package txcache

type TransactionsHeapItem struct {
	senderIndex      int
	transactionIndex int
	transaction      *WrappedTransaction
}
