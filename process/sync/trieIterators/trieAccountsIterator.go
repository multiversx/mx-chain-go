package trieIterators

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/state/parsers"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("trieIterators")

// TrieAccountIteratorHandler represents a type that maps a handler for the trie's accounts iterator
type TrieAccountIteratorHandler func(account state.UserAccountHandler) error

type trieAccountsIterator struct {
	marshaller marshal.Marshalizer
	accounts   state.AccountsAdapter
}

// ArgsTrieAccountsIterator holds the arguments needed to create a new trie Accounts iterator
type ArgsTrieAccountsIterator struct {
	Marshaller marshal.Marshalizer
	Accounts   state.AccountsAdapter
}

// NewTrieAccountsIterator returns a new instance of trieAccountsIterator
func NewTrieAccountsIterator(args ArgsTrieAccountsIterator) (*trieAccountsIterator, error) {
	if check.IfNil(args.Marshaller) {
		return nil, errNilMarshaller
	}
	if check.IfNil(args.Accounts) {
		return nil, errNilAccountsAdapter
	}

	return &trieAccountsIterator{
		marshaller: args.Marshaller,
		accounts:   args.Accounts,
	}, nil
}

// Process will iterate over the entire trie and iterate over the Accounts while calling the received handlers
func (t *trieAccountsIterator) Process(handlers ...TrieAccountIteratorHandler) error {
	if len(handlers) == 0 {
		return nil
	}

	rootHash, err := t.accounts.RootHash()
	if err != nil {
		return err
	}

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err = t.accounts.GetAllLeaves(iteratorChannels, context.Background(), rootHash, parsers.NewMainTrieLeafParser())
	if err != nil {
		return err
	}

	return t.iterateOverHandlers(iteratorChannels, handlers)
}

func (t *trieAccountsIterator) iterateOverHandlers(iteratorChannels *common.TrieIteratorChannels, handlers []TrieAccountIteratorHandler) error {
	log.Debug("starting the trie's accounts iteration with calling the handlers")
	for leaf := range iteratorChannels.LeavesChan {
		userAddress, isAccount := t.getAddress(leaf)
		if !isAccount {
			continue
		}

		acc, err := t.accounts.GetExistingAccount(userAddress)
		if err != nil {
			return err
		}

		userAccount, ok := acc.(state.UserAccountHandler)
		if !ok {
			continue
		}

		for _, handler := range handlers {
			err = handler(userAccount)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *trieAccountsIterator) getAddress(kv core.KeyValueHolder) ([]byte, bool) {
	userAccount := &accounts.UserAccountData{}
	errUnmarshal := t.marshaller.Unmarshal(userAccount, kv.Value())
	if errUnmarshal != nil {
		// probably a code node
		return nil, false
	}
	if len(userAccount.RootHash) == 0 {
		return nil, false
	}

	return kv.Key(), true
}
