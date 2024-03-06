package factory

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/state/trackableDataTrie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// sovereignAccountCreator has method to create a new account
type sovereignAccountCreator struct {
	hasher              hashing.Hasher
	marshaller          marshal.Marshalizer
	esdtAsBalance       state.ESDTAsBalanceHandler
	enableEpochsHandler common.EnableEpochsHandler
}

// NewSovereignAccountCreator creates a new instance of AccountCreator
func NewSovereignAccountCreator(args ArgsAccountCreator) (state.AccountFactory, error) {
	if check.IfNil(args.Hasher) {
		return nil, errors.ErrNilHasher
	}
	if check.IfNil(args.Marshaller) {
		return nil, errors.ErrNilMarshalizer
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, errors.ErrNilEnableEpochsHandler
	}

	esdtAsBalance, err := accounts.NewESDTAsBalance("WEGLD-bd4d79", args.Marshaller)
	if err != nil {
		return nil, err
	}

	return &sovereignAccountCreator{
		hasher:              args.Hasher,
		marshaller:          args.Marshaller,
		enableEpochsHandler: args.EnableEpochsHandler,
		esdtAsBalance:       esdtAsBalance,
	}, nil
}

// CreateAccount calls the new Account creator and returns the result
func (s *sovereignAccountCreator) CreateAccount(address []byte) (vmcommon.AccountHandler, error) {
	tdt, err := trackableDataTrie.NewTrackableDataTrie(address, s.hasher, s.marshaller, s.enableEpochsHandler)
	if err != nil {
		return nil, err
	}

	dataTrieLeafParser, err := parsers.NewDataTrieLeafParser(address, s.marshaller, s.enableEpochsHandler)
	if err != nil {
		return nil, err
	}

	return accounts.NewSovereignAccount(address, tdt, dataTrieLeafParser, s.esdtAsBalance)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *sovereignAccountCreator) IsInterfaceNil() bool {
	return s == nil
}
