package mock

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/state"
)

// TxExecutionProcessorStub -
type TxExecutionProcessorStub struct {
	ExecuteTransactionCalled func(nonce uint64, sndAddr []byte, rcvAddress []byte, value *big.Int, data []byte) error
	AccountExistsCalled      func(address []byte) bool
	GetNonceCalled           func(senderBytes []byte) (uint64, error)
	AddBalanceCalled         func(senderBytes []byte, value *big.Int) error
	AddNonceCalled           func(senderBytes []byte, nonce uint64) error
}

// ExecuteTransaction -
func (teps *TxExecutionProcessorStub) ExecuteTransaction(nonce uint64, sndAddr []byte, rcvAddress []byte, value *big.Int, data []byte) error {
	if teps.ExecuteTransactionCalled != nil {
		return teps.ExecuteTransactionCalled(nonce, sndAddr, rcvAddress, value, data)
	}

	return nil
}

// GetAccount -
func (teps *TxExecutionProcessorStub) GetAccount(address []byte) (state.UserAccountHandler, bool) {
	if teps.AccountExistsCalled != nil {
		return nil, teps.AccountExistsCalled(address)
	}

	return nil, false
}

// GetNonce -
func (teps *TxExecutionProcessorStub) GetNonce(senderBytes []byte) (uint64, error) {
	if teps.GetNonceCalled != nil {
		return teps.GetNonceCalled(senderBytes)
	}

	return 0, nil
}

// AddBalance -
func (teps *TxExecutionProcessorStub) AddBalance(senderBytes []byte, value *big.Int) error {
	if teps.AddBalanceCalled != nil {
		return teps.AddBalanceCalled(senderBytes, value)
	}

	return nil
}

// AddNonce -
func (teps *TxExecutionProcessorStub) AddNonce(senderBytes []byte, nonce uint64) error {
	if teps.AddNonceCalled != nil {
		return teps.AddNonceCalled(senderBytes, nonce)
	}

	return nil
}

// GetExecutedTransactions -
func (teps *TxExecutionProcessorStub) GetExecutedTransactions() []data.TransactionHandler {
	return nil
}

// IsInterfaceNil -
func (teps *TxExecutionProcessorStub) IsInterfaceNil() bool {
	return teps == nil
}
