package workItems

import "github.com/ElrondNetwork/elrond-go/data/state"

type itemAccounts struct {
	indexer  saveAccountsIndexer
	accounts []state.UserAccountHandler
}

// NewItemAccounts will create a new instance of itemAccounts
func NewItemAccounts(
	indexer saveAccountsIndexer,
	accounts []state.UserAccountHandler,
) WorkItemHandler {
	return &itemAccounts{
		indexer:  indexer,
		accounts: accounts,
	}
}

// Save will save information about an account
func (wiv *itemAccounts) Save() error {
	err := wiv.indexer.SaveAccounts(wiv.accounts)
	if err != nil {
		log.Warn("itemAccounts.Save",
			"could not index account",
			"error", err.Error())
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (wiv *itemAccounts) IsInterfaceNil() bool {
	return wiv == nil
}
