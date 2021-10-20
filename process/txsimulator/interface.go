package txsimulator

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// TransactionProcessor defines the operations needed do be done by a transaction processor
type TransactionProcessor interface {
	ProcessTransaction(transaction *transaction.Transaction) (vmcommon.ReturnCode, error)
	VerifyTransaction(transaction *transaction.Transaction) error
	IsInterfaceNil() bool
}
