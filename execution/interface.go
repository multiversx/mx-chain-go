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

// Checker checks the integrity of a data structure
type Checker interface {
	Check() bool
}

// SigVerifier provides functionality to verify a signature
type SigVerifier interface {
	VerifySig() bool
}

// SignedDataValidator provides functionality and check the validity of a block header
type SignedDataValidator interface {
	SigVerifier
	Checker
}
