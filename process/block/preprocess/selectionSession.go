package preprocess

import (
	"errors"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
)

type selectionSession struct {
	transactionsProcessor process.TransactionProcessor
	accountsProvider      *state.AccountsEphemeralProvider
}

// ArgsSelectionSession holds the arguments for creating a new selection session.
type ArgsSelectionSession struct {
	AccountsAdapter       state.AccountsAdapter
	TransactionsProcessor process.TransactionProcessor
}

// NewSelectionSession creates a new selection session.
func NewSelectionSession(args ArgsSelectionSession) (*selectionSession, error) {
	if check.IfNil(args.TransactionsProcessor) {
		return nil, process.ErrNilTxProcessor
	}

	// Provider is not concurrency-safe, but it's never accessed concurrently.
	accountsProvider, err := state.NewAccountsEphemeralProvider(args.AccountsAdapter)
	if err != nil {
		return nil, err
	}

	return &selectionSession{
		transactionsProcessor: args.TransactionsProcessor,
		accountsProvider:      accountsProvider,
	}, nil
}

// GetAccountNonceAndBalance returns the nonce of the account, the balance of the account, and whether it's currently existing on-chain.
// Will be called by the transactions pool, during transactions selection.
func (session *selectionSession) GetAccountNonceAndBalance(address []byte) (uint64, *big.Int, bool, error) {
	return session.accountsProvider.GetAccountNonceAndBalance(address)
}

// IsIncorrectlyGuarded checks if a transaction is incorrectly guarded (not executable).
// Will be called by mempool during transaction selection.
// See: MX-16157, MX-16772.
func (session *selectionSession) IsIncorrectlyGuarded(tx data.TransactionHandler) bool {
	address := tx.GetSndAddr()
	account, err := session.accountsProvider.GetUserAccount(address)
	if err != nil || check.IfNil(account) {
		// Unexpected failure. In case of any error (or missing account), we assume the transaction is "correctly guarded".
		return false
	}

	txTyped, ok := tx.(*transaction.Transaction)
	if !ok {
		return false
	}

	err = session.transactionsProcessor.VerifyGuardian(txTyped, account)
	return errors.Is(err, process.ErrTransactionNotExecutable)
}

// GetRootHash returns the current root hash
func (session *selectionSession) GetRootHash() ([]byte, error) {
	return session.accountsProvider.GetRootHash()
}

// IsInterfaceNil returns true if there is no value under the interface
func (session *selectionSession) IsInterfaceNil() bool {
	return session == nil
}
