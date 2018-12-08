package execution

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

// TransactionExecutor is the main interface for transaction execution engine
type TransactionExecutor interface {
	SChandler() func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error
	SetSChandler(func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error)

	ProcessTransaction(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error
}
