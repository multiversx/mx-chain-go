package transactionEvaluator

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
)

// TransactionProcessor defines the operations needed to be done by a transaction processor
type TransactionProcessor interface {
	ProcessTransaction(transaction *transaction.Transaction) (vmcommon.ReturnCode, error)
	VerifyTransaction(transaction *transaction.Transaction) error
	GetSenderAndReceiverAccounts(transaction *transaction.Transaction) (state.UserAccountHandler, state.UserAccountHandler, error)
	IsInterfaceNil() bool
}

// DataFieldParser defines what a data field parser should be able to do
type DataFieldParser interface {
	Parse(dataField []byte, sender, receiver []byte, numOfShards uint32) *datafield.ResponseParseData
}
