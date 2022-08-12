package factory

import (
	"fmt"

	chainData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/blockInfoProviders"
)

// CreateAccountsAdapterAPIOnFinal creates a new instance of AccountsAdapterAPI that tracks the final blocks state
func CreateAccountsAdapterAPIOnFinal(args state.ArgsAccountsDB, chainHandler chainData.ChainHandler) (state.AccountsAdapterAPI, error) {
	provider, err := blockInfoProviders.NewFinalBlockInfo(chainHandler)
	if err != nil {
		return nil, fmt.Errorf("%w in CreateAccountsAdapterAPIOnFinal", err)
	}

	accountsAdapterApi, err := createAccountsDB(args, provider)
	if err != nil {
		return nil, fmt.Errorf("%w in CreateAccountsAdapterAPIOnFinal", err)
	}

	return accountsAdapterApi, nil
}

// CreateAccountsAdapterAPIOnCurrent creates a new instance of AccountsAdapterAPI that tracks the current blocks state
func CreateAccountsAdapterAPIOnCurrent(args state.ArgsAccountsDB, chainHandler chainData.ChainHandler) (state.AccountsAdapterAPI, error) {
	provider, err := blockInfoProviders.NewCurrentBlockInfo(chainHandler)
	if err != nil {
		return nil, fmt.Errorf("%w in CreateAccountsAdapterAPIOnCurrent", err)
	}

	accountsAdapterApi, err := createAccountsDB(args, provider)
	if err != nil {
		return nil, fmt.Errorf("%w in CreateAccountsAdapterAPIOnCurrent", err)
	}

	return accountsAdapterApi, nil
}

func createAccountsDB(args state.ArgsAccountsDB, provider state.BlockInfoProvider) (state.AccountsAdapterAPI, error) {
	accounts, err := state.NewAccountsDB(args)
	if err != nil {
		return nil, fmt.Errorf("%w in CreateAccountsAdapterAPI/createAccountsDB", err)
	}

	return state.NewAccountsDBApi(accounts, provider)
}

// CreateAccountsAdapterAPIOnHistorical creates a new instance of AccountsAdapterAPI that tracks historical state
func CreateAccountsAdapterAPIOnHistorical(args state.ArgsAccountsDB) (state.AccountsAdapterAPI, error) {
	accounts, err := state.NewAccountsDB(args)
	if err != nil {
		return nil, fmt.Errorf("%w in CreateAccountsAdapterAPIOnHistorical", err)
	}

	accountsAdapterApi, err := state.NewAccountsDBApiWithHistory(accounts)
	if err != nil {
		return nil, fmt.Errorf("%w in CreateAccountsAdapterAPIOnHistorical", err)
	}

	return accountsAdapterApi, nil
}
