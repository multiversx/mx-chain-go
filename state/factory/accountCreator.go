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

// ArgsAccountCreator holds the arguments needed to create a new account creator
type ArgsAccountCreator struct {
	Hasher              hashing.Hasher
	Marshaller          marshal.Marshalizer
	EnableEpochsHandler common.EnableEpochsHandler
}

// AccountCreator has method to create a new account
type accountCreator struct {
	hasher              hashing.Hasher
	marshaller          marshal.Marshalizer
	enableEpochsHandler common.EnableEpochsHandler
}

// NewAccountCreator creates a new instance of AccountCreator
func NewAccountCreator(args ArgsAccountCreator) (state.AccountFactory, error) {
	if check.IfNil(args.Hasher) {
		return nil, errors.ErrNilHasher
	}
	if check.IfNil(args.Marshaller) {
		return nil, errors.ErrNilMarshalizer
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, errors.ErrNilEnableEpochsHandler
	}

	return &accountCreator{
		hasher:              args.Hasher,
		marshaller:          args.Marshaller,
		enableEpochsHandler: args.EnableEpochsHandler,
	}, nil
}

// CreateAccount calls the new Account creator and returns the result
func (ac *accountCreator) CreateAccount(address []byte) (vmcommon.AccountHandler, error) {
	tdt, err := trackableDataTrie.NewTrackableDataTrie(address, ac.hasher, ac.marshaller, ac.enableEpochsHandler)
	if err != nil {
		return nil, err
	}

	dataTrieLeafParser, err := parsers.NewDataTrieLeafParser(address, ac.marshaller, ac.enableEpochsHandler)
	if err != nil {
		return nil, err
	}

	return accounts.NewUserAccount(address, tdt, dataTrieLeafParser)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ac *accountCreator) IsInterfaceNil() bool {
	return ac == nil
}
