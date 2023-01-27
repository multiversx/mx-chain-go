package state

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type accountsRepository struct {
	finalStateAccountsWrapper      AccountsAdapterAPI
	currentStateAccountsWrapper    AccountsAdapterAPI
	historicalStateAccountsWrapper AccountsAdapterAPI
}

// ArgsAccountsRepository is the DTO for the NewAccountsRepository constructor function
type ArgsAccountsRepository struct {
	FinalStateAccountsWrapper      AccountsAdapterAPI
	CurrentStateAccountsWrapper    AccountsAdapterAPI
	HistoricalStateAccountsWrapper AccountsAdapterAPI
}

// NewAccountsRepository creates a new accountsRepository instance
func NewAccountsRepository(args ArgsAccountsRepository) (*accountsRepository, error) {
	if check.IfNil(args.CurrentStateAccountsWrapper) {
		return nil, fmt.Errorf("%w for CurrentStateAccountsWrapper", ErrNilAccountsAdapter)
	}
	if check.IfNil(args.FinalStateAccountsWrapper) {
		return nil, fmt.Errorf("%w for FinalStateAccountsWrapper", ErrNilAccountsAdapter)
	}
	if check.IfNil(args.HistoricalStateAccountsWrapper) {
		return nil, fmt.Errorf("%w for HistoricalStateAccountsWrapper", ErrNilAccountsAdapter)
	}

	return &accountsRepository{
		finalStateAccountsWrapper:      args.FinalStateAccountsWrapper,
		currentStateAccountsWrapper:    args.CurrentStateAccountsWrapper,
		historicalStateAccountsWrapper: args.HistoricalStateAccountsWrapper,
	}, nil
}

// GetAccountWithBlockInfo will return the account handler with the block info providing the address and the query option
func (repository *accountsRepository) GetAccountWithBlockInfo(address []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
	accountsAdapter, err := repository.selectStateAccounts(options)
	if err != nil {
		return nil, nil, err
	}

	convertedOptions := holders.NewRootHashHolder(options.BlockRootHash, options.HintEpoch)
	return accountsAdapter.GetAccountWithBlockInfo(address, convertedOptions)
}

// GetCodeWithBlockInfo will return the code with the block info providing the code hash and the query option
func (repository *accountsRepository) GetCodeWithBlockInfo(codeHash []byte, options api.AccountQueryOptions) ([]byte, common.BlockInfo, error) {
	accountsAdapter, err := repository.selectStateAccounts(options)
	if err != nil {
		return nil, nil, err
	}

	convertedOptions := holders.NewRootHashHolder(options.BlockRootHash, options.HintEpoch)
	return accountsAdapter.GetCodeWithBlockInfo(codeHash, convertedOptions)
}

func (repository *accountsRepository) selectStateAccounts(options api.AccountQueryOptions) (AccountsAdapterAPI, error) {
	if len(options.BlockRootHash) > 0 {
		return repository.historicalStateAccountsWrapper, nil
	}
	if options.OnFinalBlock {
		return repository.finalStateAccountsWrapper, nil
	}
	if options.OnStartOfEpoch.HasValue {
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
	errHistorical := repository.historicalStateAccountsWrapper.Close()
	errFinal := repository.finalStateAccountsWrapper.Close()
	errCurrent := repository.currentStateAccountsWrapper.Close()

	if errHistorical != nil {
		return errHistorical
	}
	if errFinal != nil {
		return errFinal
	}

	return errCurrent
}

// IsInterfaceNil returns true if there is no value under the interface
func (repository *accountsRepository) IsInterfaceNil() bool {
	return repository == nil
}
