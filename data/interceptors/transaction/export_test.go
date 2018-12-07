package transaction

func (txi *transactionInterceptor) CheckSanityTx(txNewer *TransactionNewer) bool {
	return txi.checkSanityTx(txNewer)
}
