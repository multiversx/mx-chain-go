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
	transactionsProcessor   process.TransactionProcessor
	accountsProvider        *state.AccountsEphemeralProvider
	txVersionCheckerHandler process.TxVersionCheckerHandler
}

// ArgsSelectionSession holds the arguments for creating a new selection session.
type ArgsSelectionSession struct {
	AccountsAdapter         state.AccountsAdapter
	TransactionsProcessor   process.TransactionProcessor
	TxVersionCheckerHandler process.TxVersionCheckerHandler
}

// NewSelectionSession creates a new selection session.
// TODO minimalize the interface used for AccountsAdapter
func NewSelectionSession(args ArgsSelectionSession) (*selectionSession, error) {
	if check.IfNil(args.TransactionsProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(args.TxVersionCheckerHandler) {
		return nil, process.ErrNilTransactionVersionChecker
	}

	// Provider is not concurrency-safe, but it's never accessed concurrently.
	accountsProvider, err := state.NewAccountsEphemeralProvider(args.AccountsAdapter)
	if err != nil {
		return nil, err
	}

	return &selectionSession{
		transactionsProcessor:   args.TransactionsProcessor,
		accountsProvider:        accountsProvider,
		txVersionCheckerHandler: args.TxVersionCheckerHandler,
	}, nil
}

// GetAccountNonceAndBalance returns the nonce of the account, the balance of the account, and whether it's currently existing on-chain.
// Will be called by the transactions pool, during transactions selection.
func (session *selectionSession) GetAccountNonceAndBalance(address []byte) (uint64, *big.Int, bool, error) {
	return session.accountsProvider.GetAccountNonceAndBalance(address)
}

// IsIncorrectlyGuarded checks if a transaction is incorrectly guarded (not executable).
// Will be called by mempool during transaction selection.
//
// Uses a two-phase approach to avoid upgrading lightAccountInfo to full accounts
// in the common case (non-guarded tx from non-guarded account). Only when the account
// is guarded do we need GetUserAccount (for VerifyGuardian's data trie access).
func (session *selectionSession) IsIncorrectlyGuarded(tx data.TransactionHandler) bool {
	txPtr, ok := tx.(*transaction.Transaction)
	if !ok {
		return false
	}

	isTransactionGuarded := session.txVersionCheckerHandler.IsGuardedTransaction(txPtr)

	address := tx.GetSndAddr()
	accountIsGuarded, err := session.accountsProvider.IsAccountGuarded(address)
	if err != nil {
		return false
	}

	// Non-guarded tx from non-guarded account: always OK.
	if !isTransactionGuarded && !accountIsGuarded {
		return false
	}

	// Guarded tx from non-guarded account: incorrectly guarded.
	// (matches VerifyGuardian: "guarded transaction, but account not guarded")
	// This also covers missing/new accounts (not yet in trie), which return accountIsGuarded=false.
	// A non-existent account cannot have a guardian, so filtering here is correct and avoids
	// wasting execution resources on a transaction that would always fail.
	if isTransactionGuarded && !accountIsGuarded {
		return true
	}

	// Account IS guarded — need full account for VerifyGuardian
	// (requires data trie access: GetActiveGuardian, HasPendingGuardian, etc.)
	account, err := session.accountsProvider.GetUserAccount(address)
	if err != nil || check.IfNil(account) {
		return false
	}

	err = session.transactionsProcessor.VerifyGuardian(txPtr, account)
	return errors.Is(err, process.ErrTransactionNotExecutable)
}

// IsGuarded checks if the transaction is guarded
func (session *selectionSession) IsGuarded(tx data.TransactionHandler) bool {
	txPtr, ok := tx.(*transaction.Transaction)
	if !ok {
		return false
	}

	return session.txVersionCheckerHandler.IsGuardedTransaction(txPtr)
}

// PrefetchAccounts batch-fetches multiple accounts and warms the ephemeral cache.
// Delegates to the inner AccountsEphemeralProvider, making selectionSession satisfy
// common.AccountBatchPrefetcher so virtualSessionComputer.prefetchBreadcrumbAddresses works.
func (session *selectionSession) PrefetchAccounts(addresses [][]byte) {
	session.accountsProvider.PrefetchAccounts(addresses)
}

// GetRootHash returns the current root hash
func (session *selectionSession) GetRootHash() ([]byte, error) {
	return session.accountsProvider.GetRootHash()
}

// IsInterfaceNil returns true if there is no value under the interface
func (session *selectionSession) IsInterfaceNil() bool {
	return session == nil
}
