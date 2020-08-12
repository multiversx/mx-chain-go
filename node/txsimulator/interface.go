package txsimulator

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type TransactionProcessor interface {
	ProcessTransaction(transaction *transaction.Transaction) (vmcommon.ReturnCode, error)
	SetAccountsAdapter(accounts state.AccountsAdapter)
	IsInterfaceNil() bool
}
