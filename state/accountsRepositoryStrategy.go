package state

import (
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type accountsRepositoryStrategy struct {
	getBlockInfo    func() (*accountBlockInfo, error)
	accountsAdapter AccountsAdapter
	trieController  *accountsDBApiTrieController
}

func newAccountsRepositoryStrategy(
	getBlockInfo func() (*accountBlockInfo, error),
	accountsAdapter AccountsAdapter,
) *accountsRepositoryStrategy {
	return &accountsRepositoryStrategy{
		getBlockInfo:    getBlockInfo,
		accountsAdapter: accountsAdapter,
		trieController:  newAccountsDBApiTrieController(accountsAdapter),
	}
}

func (strategy *accountsRepositoryStrategy) getAccount(address []byte) (vmcommon.AccountHandler, AccountBlockInfo, error) {
	blockInfo, err := strategy.getBlockInfo()
	if err != nil {
		return nil, nil, err
	}

	err = strategy.trieController.recreateTrieIfNecessary(blockInfo.rootHash)
	if err != nil {
		return nil, nil, err
	}

	account, err := strategy.accountsAdapter.GetExistingAccount(address)
	if err != nil {
		return nil, nil, err
	}

	return account, blockInfo, nil
}

func (strategy *accountsRepositoryStrategy) getCode(codeHash []byte) ([]byte, AccountBlockInfo) {
	blockInfo, err := strategy.getBlockInfo()
	if err != nil {
		return nil, nil
	}

	err = strategy.trieController.recreateTrieIfNecessary(blockInfo.rootHash)
	if err != nil {
		return nil, nil
	}

	code := strategy.accountsAdapter.GetCode(codeHash)
	return code, blockInfo
}

func (strategy *accountsRepositoryStrategy) close() error {
	return strategy.accountsAdapter.Close()
}
