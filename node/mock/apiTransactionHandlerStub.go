package mock

import (
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

// TransactionAPIHandlerStub -
type TransactionAPIHandlerStub struct {
	GetTransactionCalled                        func(hash string, withResults bool) (*transaction.ApiTransactionResult, error)
	GetTransactionsPoolCalled                   func(fields string) (*common.TransactionsPoolAPIResponse, error)
	GetTransactionsPoolForSenderCalled          func(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error)
	GetLastPoolNonceForSenderCalled             func(sender string) (uint64, error)
	GetTransactionsPoolNonceGapsForSenderCalled func(sender string, senderAccountNonce uint64) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error)
	UnmarshalTransactionCalled                  func(txBytes []byte, txType transaction.TxType, epoch uint32) (*transaction.ApiTransactionResult, error)
	GetSelectedTransactionsCalled               func(selectionOptions common.TxSelectionOptionsAPI, blockchain coreData.ChainHandler, accountsAdapter state.AccountsAdapter) (*common.TransactionsSelectionSimulationResult, error)
	GetVirtualNonceCalled                       func(address string) (*common.VirtualNonceOfAccountResponse, error)
	UnmarshalReceiptCalled                      func(receiptBytes []byte) (*transaction.ApiReceipt, error)
	PopulateComputedFieldsCalled                func(tx *transaction.ApiTransactionResult)
	GetSCRsByTxHashCalled                       func(txHash string, scrHash string) ([]*transaction.ApiSmartContractResult, error)
}

// GetSCRsByTxHash --
func (tas *TransactionAPIHandlerStub) GetSCRsByTxHash(txHash string, scrHash string) ([]*transaction.ApiSmartContractResult, error) {
	if tas.GetSCRsByTxHashCalled != nil {
		return tas.GetSCRsByTxHashCalled(txHash, scrHash)
	}

	return nil, nil
}

// GetVirtualNonce -
func (tas *TransactionAPIHandlerStub) GetVirtualNonce(address string) (*common.VirtualNonceOfAccountResponse, error) {
	if tas.GetVirtualNonceCalled != nil {
		return tas.GetVirtualNonceCalled(address)
	}

	return nil, nil
}

// GetTransaction -
func (tas *TransactionAPIHandlerStub) GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error) {
	if tas.GetTransactionCalled != nil {
		return tas.GetTransactionCalled(hash, withResults)
	}

	return nil, nil
}

// GetTransactionsPool -
func (tas *TransactionAPIHandlerStub) GetTransactionsPool(fields string) (*common.TransactionsPoolAPIResponse, error) {
	if tas.GetTransactionsPoolCalled != nil {
		return tas.GetTransactionsPoolCalled(fields)
	}

	return nil, nil
}

// GetTransactionsPoolForSender -
func (tas *TransactionAPIHandlerStub) GetTransactionsPoolForSender(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error) {
	if tas.GetTransactionsPoolForSenderCalled != nil {
		return tas.GetTransactionsPoolForSenderCalled(sender, fields)
	}

	return nil, nil
}

// GetLastPoolNonceForSender -
func (tas *TransactionAPIHandlerStub) GetLastPoolNonceForSender(sender string) (uint64, error) {
	if tas.GetLastPoolNonceForSenderCalled != nil {
		return tas.GetLastPoolNonceForSenderCalled(sender)
	}

	return 0, nil
}

// GetTransactionsPoolNonceGapsForSender -
func (tas *TransactionAPIHandlerStub) GetTransactionsPoolNonceGapsForSender(sender string, senderAccountNonce uint64) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
	if tas.GetTransactionsPoolNonceGapsForSenderCalled != nil {
		return tas.GetTransactionsPoolNonceGapsForSenderCalled(sender, senderAccountNonce)
	}

	return nil, nil
}

// GetSelectedTransactions -
func (tas *TransactionAPIHandlerStub) GetSelectedTransactions(selectionOptions common.TxSelectionOptionsAPI, blockchain coreData.ChainHandler, accountsAdapter state.AccountsAdapter) (*common.TransactionsSelectionSimulationResult, error) {
	if tas.GetSelectedTransactionsCalled != nil {
		return tas.GetSelectedTransactionsCalled(selectionOptions, blockchain, accountsAdapter)
	}

	return nil, nil
}

// UnmarshalTransaction -
func (tas *TransactionAPIHandlerStub) UnmarshalTransaction(txBytes []byte, txType transaction.TxType, epoch uint32) (*transaction.ApiTransactionResult, error) {
	if tas.UnmarshalTransactionCalled != nil {
		return tas.UnmarshalTransactionCalled(txBytes, txType, epoch)
	}
	return nil, nil
}

// UnmarshalReceipt -
func (tas *TransactionAPIHandlerStub) UnmarshalReceipt(receiptBytes []byte) (*transaction.ApiReceipt, error) {
	if tas.UnmarshalReceiptCalled != nil {
		return tas.UnmarshalReceiptCalled(receiptBytes)
	}

	return nil, nil
}

// PopulateComputedFields -
func (tas *TransactionAPIHandlerStub) PopulateComputedFields(tx *transaction.ApiTransactionResult) {
	if tas.PopulateComputedFieldsCalled != nil {
		tas.PopulateComputedFieldsCalled(tx)
	}
}

// IsInterfaceNil -
func (tas *TransactionAPIHandlerStub) IsInterfaceNil() bool {
	return tas == nil
}
