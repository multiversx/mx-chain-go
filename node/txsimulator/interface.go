package txsimulator

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// TransactionProcessor defines the operations needed do be done by a transaction processor
type TransactionProcessor interface {
	ProcessTransaction(transaction *transaction.Transaction) (vmcommon.ReturnCode, error)
	GetSmartContractResults() map[string]data.TransactionHandler
	SetAccountsAdapter(accounts state.AccountsAdapter)
	IsInterfaceNil() bool
}
