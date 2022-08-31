package state

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/common"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type accountsRepository struct {
	finalStateAccountsWrapper   AccountsAdapterAPI
	currentStateAccountsWrapper AccountsAdapterAPI
}

// ArgsAccountsRepository is the DTO for the NewAccountsRepository constructor function
type ArgsAccountsRepository struct {
	FinalStateAccountsWrapper   AccountsAdapterAPI
	CurrentStateAccountsWrapper AccountsAdapterAPI
}

// NewAccountsRepository creates a new accountsRepository instance
func NewAccountsRepository(args ArgsAccountsRepository) (*accountsRepository, error) {
	if check.IfNil(args.CurrentStateAccountsWrapper) {
		return nil, fmt.Errorf("%w for CurrentStateAccountsWrapper", ErrNilAccountsAdapter)
	}
	if check.IfNil(args.FinalStateAccountsWrapper) {
		return nil, fmt.Errorf("%w for FinalStateAccountsWrapper", ErrNilAccountsAdapter)
	}

	return &accountsRepository{
		finalStateAccountsWrapper:   args.FinalStateAccountsWrapper,
		currentStateAccountsWrapper: args.CurrentStateAccountsWrapper,
	}, nil
}

// GetAccountWithBlockInfo will return the account handler with the block info providing the address and the query option
func (repository *accountsRepository) GetAccountWithBlockInfo(address []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
	accountsAdapter, err := repository.selectStateAccounts(options)
	if err != nil {
		return nil, nil, err
	}

	return accountsAdapter.GetAccountWithBlockInfo(address)
}

// GetCodeWithBlockInfo will return the code with the block info providing the code hash and the query option
func (repository *accountsRepository) GetCodeWithBlockInfo(codeHash []byte, options api.AccountQueryOptions) ([]byte, common.BlockInfo, error) {
	accountsAdapter, err := repository.selectStateAccounts(options)
	if err != nil {
		return nil, nil, err
	}

	return accountsAdapter.GetCodeWithBlockInfo(codeHash)
}

func (repository *accountsRepository) selectStateAccounts(options api.AccountQueryOptions) (AccountsAdapterAPI, error) {
	if options.OnFinalBlock {
		return repository.finalStateAccountsWrapper, nil
	}
	if options.OnStartOfEpoch.Value > 0 {
		// TODO implement this
		return nil, ErrFunctionalityNotImplemented
	}

	return repository.currentStateAccountsWrapper, nil
}

// GetCurrentStateAccountsWrapper gets the current state accounts wrapper
func (repository *accountsRepository) GetCurrentStateAccountsWrapper() AccountsAdapterAPI {
	return repository.currentStateAccountsWrapper
}

// Close will handle the closing of the underlying components
func (repository *accountsRepository) Close() error {
	errFinal := repository.finalStateAccountsWrapper.Close()
	errCurrent := repository.currentStateAccountsWrapper.Close()

	if errFinal != nil {
		return errFinal
	}

	return errCurrent
}

// IsInterfaceNil returns true if there is no value under the interface
func (repository *accountsRepository) IsInterfaceNil() bool {
	return repository == nil
}
