package mock

import "time"

type TxPoolsCleanerMock struct {
	CleanCalled         func(duration time.Duration) (bool, error)
	NumRemovedTxsCalled func() uint64
}

// Clean will check if in pools exits transactions with nonce low that transaction sender account nonce
// and if tx have low nonce will be removed from pools
func (tpc *TxPoolsCleanerMock) Clean(duration time.Duration) (bool, error) {
	return false, nil
}

// NumRemovedTxs will return the number of removed txs from pools
func (tpc *TxPoolsCleanerMock) NumRemovedTxs() uint64 {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (tpc *TxPoolsCleanerMock) IsInterfaceNil() bool {
	if tpc == nil {
		return true
	}
	return false
}
