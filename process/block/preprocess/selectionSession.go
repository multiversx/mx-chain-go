package preprocess

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
)

type selectionSession struct {
	accountsAdapter       state.AccountsAdapter
	transactionsProcessor process.TransactionProcessor
	callArgumentsParser   process.CallArgumentsParser
	esdtTransferParser    vmcommon.ESDTTransferParser
}

type argsSelectionSession struct {
	accountsAdapter       state.AccountsAdapter
	transactionsProcessor process.TransactionProcessor
	marshalizer           marshal.Marshalizer
}

func newSelectionSession(args argsSelectionSession) (*selectionSession, error) {
	if check.IfNil(args.accountsAdapter) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.transactionsProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(args.marshalizer) {
		return nil, process.ErrNilMarshalizer
	}

	argsParser := parsers.NewCallArgsParser()

	esdtTransferParser, err := parsers.NewESDTTransferParser(args.marshalizer)
	if err != nil {
		return nil, err
	}

	return &selectionSession{
		accountsAdapter:       args.accountsAdapter,
		transactionsProcessor: args.transactionsProcessor,
		callArgumentsParser:   argsParser,
		esdtTransferParser:    esdtTransferParser,
	}, nil
}

// GetAccountState returns the state of an account.
// Will be called by mempool during transaction selection.
func (session *selectionSession) GetAccountState(address []byte) (*txcache.AccountState, error) {
	account, err := session.accountsAdapter.GetExistingAccount(address)
	if err != nil {
		return nil, err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return &txcache.AccountState{
		Nonce:   userAccount.GetNonce(),
		Balance: userAccount.GetBalance(),
	}, nil
}

// IsBadlyGuarded checks if a transaction is badly guarded (not executable).
// Will be called by mempool during transaction selection.
func (session *selectionSession) IsBadlyGuarded(tx data.TransactionHandler) bool {
	address := tx.GetSndAddr()
	account, err := session.accountsAdapter.GetExistingAccount(address)
	if err != nil {
		return false
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return false
	}

	txTyped, ok := tx.(*transaction.Transaction)
	if !ok {
		return false
	}

	err = session.transactionsProcessor.VerifyGuardian(txTyped, userAccount)
	return errors.Is(err, process.ErrTransactionNotExecutable)
}

// GetTransferredValue returns the value transferred by a transaction.
func (session *selectionSession) GetTransferredValue(tx data.TransactionHandler) *big.Int {
	hasValue := tx.GetValue() != nil && tx.GetValue().Sign() != 0
	if hasValue {
		// Early exit (optimization): a transaction can either bear a regular value or be a "MultiESDTNFTTransfer".
		return tx.GetValue()
	}

	hasData := len(tx.GetData()) > 0
	if !hasData {
		// Early exit (optimization): no "MultiESDTNFTTransfer" to parse.
		return tx.GetValue()
	}

	maybeMultiTransfer := bytes.HasPrefix(tx.GetData(), []byte(core.BuiltInFunctionMultiESDTNFTTransfer))
	if !maybeMultiTransfer {
		// Early exit (optimization).
		return nil
	}

	function, args, err := session.callArgumentsParser.ParseData(string(tx.GetData()))
	if err != nil {
		return nil
	}

	if function != core.BuiltInFunctionMultiESDTNFTTransfer {
		// Early exit (optimization).
		return nil
	}

	esdtTransfers, err := session.esdtTransferParser.ParseESDTTransfers(tx.GetSndAddr(), tx.GetRcvAddr(), function, args)
	if err != nil {
		return nil
	}

	accumulatedNativeValue := big.NewInt(0)

	for _, transfer := range esdtTransfers.ESDTTransfers {
		if transfer.ESDTTokenNonce != 0 {
			continue
		}
		if string(transfer.ESDTTokenName) != vmcommon.EGLDIdentifier {
			continue
		}

		_ = accumulatedNativeValue.Add(accumulatedNativeValue, transfer.ESDTValue)
	}

	return accumulatedNativeValue
}

// IsInterfaceNil returns true if there is no value under the interface
func (session *selectionSession) IsInterfaceNil() bool {
	return session == nil
}
