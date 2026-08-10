package txcachestubs

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/txcache"
)

// TxCacheStub -
type TxCacheStub struct {
	SelectTransactionsCalled         func(session txcache.SelectionSession, options common.TxSelectionOptions, currentBlockNonce uint64) ([]*txcache.WrappedTransaction, uint64, error)
	SimulateSelectTransactionsCalled func(session txcache.SelectionSession, options common.TxSelectionOptions, currentBlockNonce uint64) ([]*txcache.WrappedTransaction, uint64, error)
	GetVirtualNonceAndRootHashCalled func(sender []byte) (uint64, []byte, error)
	SetAOTSelectionPreempterCalled   func(preempter common.AOTSelectionPreempter)
}

// SelectTransactions -
func (stub *TxCacheStub) SelectTransactions(session txcache.SelectionSession, options common.TxSelectionOptions, currentBlockNonce uint64) ([]*txcache.WrappedTransaction, uint64, error) {
	if stub.SelectTransactionsCalled != nil {
		return stub.SelectTransactionsCalled(session, options, currentBlockNonce)
	}
	return nil, 0, nil
}

// SimulateSelectTransactions -
func (stub *TxCacheStub) SimulateSelectTransactions(session txcache.SelectionSession, options common.TxSelectionOptions, currentBlockNonce uint64) ([]*txcache.WrappedTransaction, uint64, error) {
	if stub.SimulateSelectTransactionsCalled != nil {
		return stub.SimulateSelectTransactionsCalled(session, options, currentBlockNonce)
	}
	return nil, 0, nil
}

// GetVirtualNonceAndRootHash -
func (stub *TxCacheStub) GetVirtualNonceAndRootHash(sender []byte) (uint64, []byte, error) {
	if stub.GetVirtualNonceAndRootHashCalled != nil {
		return stub.GetVirtualNonceAndRootHashCalled(sender)
	}
	return 0, nil, nil
}

// SetAOTSelectionPreempter -
func (stub *TxCacheStub) SetAOTSelectionPreempter(preempter common.AOTSelectionPreempter) {
	if stub.SetAOTSelectionPreempterCalled != nil {
		stub.SetAOTSelectionPreempterCalled(preempter)
	}
}

// IsInterfaceNil -
func (stub *TxCacheStub) IsInterfaceNil() bool {
	return stub == nil
}
