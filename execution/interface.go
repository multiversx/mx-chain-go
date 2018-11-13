package execution

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

// TransactionExecutor is the main interface for transaction execution engine
type TransactionExecutor interface {
	SChandler() func(accounts state.AccountsHandler, transaction *transaction.Transaction) *ExecSummary
	SetSChandler(func(accounts state.AccountsHandler, transaction *transaction.Transaction) *ExecSummary)

	ProcessTransaction(accounts state.AccountsHandler, transaction *transaction.Transaction) *ExecSummary
}
