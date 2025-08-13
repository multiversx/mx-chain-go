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
	accountsAdapter       state.AccountsAdapter
	transactionsProcessor process.TransactionProcessor

	// Cache of accounts, held in the scope of a single selection session.
	// Not concurrency-safe, but never accessed concurrently.
	// Entries for new (uknown) accounts are "nil" values.
	ephemeralAccountsCache map[string]state.UserAccountHandler
}

// ArgsSelectionSession holds the arguments for creating a new selection session.
type ArgsSelectionSession struct {
	AccountsAdapter       state.AccountsAdapter
	TransactionsProcessor process.TransactionProcessor
}

// NewSelectionSession creates a new selection session.
func NewSelectionSession(args ArgsSelectionSession) (*selectionSession, error) {
	if check.IfNil(args.AccountsAdapter) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.TransactionsProcessor) {
		return nil, process.ErrNilTxProcessor
	}

	return &selectionSession{
		accountsAdapter:        args.AccountsAdapter,
		transactionsProcessor:  args.TransactionsProcessor,
		ephemeralAccountsCache: make(map[string]state.UserAccountHandler),
	}, nil
}

// GetAccountNonceAndBalance returns the nonce of the account, the balance of the account, and whether it's currently existing on-chain.
// Will be called by the transactions pool, during transactions selection.
func (session *selectionSession) GetAccountNonceAndBalance(address []byte) (uint64, *big.Int, bool, error) {
	account, err := session.getCachedUserAccount(address)
	if err != nil {
		// Unexpected failure.
		return 0, nil, false, err
	}
	if check.IfNil(account) {
		// New (unknown) account.
		return 0, big.NewInt(0), false, nil
	}

	return account.GetNonce(), account.GetBalance(), true, nil
}

func (session *selectionSession) getCachedUserAccount(address []byte) (state.UserAccountHandler, error) {
	account, ok := session.ephemeralAccountsCache[string(address)]
	if ok {
		// Existing or new (unknown) account, previously-cached.
		return account, nil
	}

	account, err := session.getExistingAccountTypedAsUserAccount(address)
	if err != nil && err != state.ErrAccNotFound {
		// Unexpected failure (error different from "ErrAccNotFound").
		// Account won't be cached.
		return nil, err
	}

	// Existing account or new (unknown), we'll cache it (actual object or nil).
	session.ephemeralAccountsCache[string(address)] = account

	// Generally speaking, this isn't a good pattern: returning both nil (for unknown accounts), and a nil error.
	// However, this is a non-exported method, which should only be called with care, within this struct only.
	return account, nil
}

func (session *selectionSession) getExistingAccountTypedAsUserAccount(address []byte) (state.UserAccountHandler, error) {
	account, err := session.accountsAdapter.GetExistingAccount(address)
	if err != nil {
		return nil, err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return userAccount, nil
}

// IsIncorrectlyGuarded checks if a transaction is incorrectly guarded (not executable).
// Will be called by mempool during transaction selection.
func (session *selectionSession) IsIncorrectlyGuarded(tx data.TransactionHandler) bool {
	address := tx.GetSndAddr()
	account, err := session.getCachedUserAccount(address)
	if err != nil || check.IfNil(account) {
		// Unexpected failure.
		// In case of any error, we assume the transaction is "correctly guarded".
		// This includes the error "ErrAccNotFound", when the account is obviously not guarded: this special case is denied by interceptors beforehand.
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
	return session.accountsAdapter.RootHash()
}

// IsInterfaceNil returns true if there is no value under the interface
func (session *selectionSession) IsInterfaceNil() bool {
	return session == nil
}
