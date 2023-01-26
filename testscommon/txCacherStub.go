package testscommon

import "github.com/multiversx/mx-chain-go/storage/txcache"

type TxCacherStub struct {
	CacherStub
	SelectTransactionsWithBandwidthCalled func(numRequested int, batchSizePerSender int, bandwidthPerSender uint64) []*txcache.WrappedTransaction
	NotifyAccountNonceCalled              func(accountKey []byte, nonce uint64)
}

// SelectTransactionsWithBandwidth -
func (tcs *TxCacherStub) SelectTransactionsWithBandwidth(numRequested int, batchSizePerSender int, bandwidthPerSender uint64) []*txcache.WrappedTransaction {
	if tcs.SelectTransactionsWithBandwidthCalled != nil {
		return tcs.SelectTransactionsWithBandwidthCalled(numRequested, batchSizePerSender, bandwidthPerSender)
	}

	return make([]*txcache.WrappedTransaction, 0)
}

// NotifyAccountNonce -
func (tcs *TxCacherStub) NotifyAccountNonce(accountKey []byte, nonce uint64) {
	if tcs.NotifyAccountNonceCalled != nil {
		tcs.NotifyAccountNonceCalled(accountKey, nonce)
	}
}
