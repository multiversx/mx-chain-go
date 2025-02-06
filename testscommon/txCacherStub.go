package testscommon

import (
	"time"

	"github.com/multiversx/mx-chain-go/storage/txcache"
)

type TxCacherStub struct {
	CacherStub
	SelectTransactionsCalled func(session txcache.SelectionSession, gasRequested uint64, maxNum int, selectionLoopMaximumDuration time.Duration) ([]*txcache.WrappedTransaction, uint64)
}

// SelectTransactions -
func (tcs *TxCacherStub) SelectTransactions(session txcache.SelectionSession, gasRequested uint64, maxNum int, selectionLoopMaximumDuration time.Duration) ([]*txcache.WrappedTransaction, uint64) {
	if tcs.SelectTransactionsCalled != nil {
		return tcs.SelectTransactionsCalled(session, gasRequested, maxNum, selectionLoopMaximumDuration)
	}

	return make([]*txcache.WrappedTransaction, 0), 0
}
