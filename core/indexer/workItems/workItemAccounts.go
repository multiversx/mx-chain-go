package workItems

import "github.com/ElrondNetwork/elrond-go/data/state"

type itemAccount struct {
	indexer saveAccountIndexer
	account state.UserAccountHandler
}

// NewItemAccount will create a new instance of itemAccount
func NewItemAccount(
	indexer saveAccountIndexer,
	account state.UserAccountHandler,
) WorkItemHandler {
	return &itemAccount{
		indexer: indexer,
		account: account,
	}
}

// Save will save information about an account
func (wiv *itemAccount) Save() error {
	err := wiv.indexer.SaveAccount(wiv.account)
	if err != nil {
		log.Warn("itemAccount.Save",
			"could not index account",
			"error", err.Error())
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (wiv *itemAccount) IsInterfaceNil() bool {
	return wiv == nil
}
