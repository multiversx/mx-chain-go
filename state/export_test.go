package state

import (
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// LastSnapshotStarted -
const LastSnapshotStarted = lastSnapshotStarted

// NewEmptyBaseAccount -
func NewEmptyBaseAccount(address []byte, tracker DataTrieTracker) *baseAccount {
	return &baseAccount{
		address:         address,
		dataTrieTracker: tracker,
	}
}

// LoadCode -
func (adb *AccountsDB) LoadCode(accountHandler baseAccountHandler) error {
	return adb.loadCode(accountHandler)
}

// LoadDataTrie -
func (adb *AccountsDB) LoadDataTrie(accountHandler baseAccountHandler) error {
	return adb.loadDataTrie(accountHandler, adb.getMainTrie())
}

// GetAccount -
func (adb *AccountsDB) GetAccount(address []byte) (vmcommon.AccountHandler, error) {
	return adb.getAccount(address, adb.getMainTrie())
}

// GetObsoleteHashes -
func (adb *AccountsDB) GetObsoleteHashes() map[string][][]byte {
	return adb.obsoleteDataTrieHashes
}

// WaitForCompletionIfAppropriate -
func (adb *AccountsDB) WaitForCompletionIfAppropriate(stats common.SnapshotStatisticsHandler) {
	adb.waitForCompletionIfAppropriate(stats)
}

// GetCode -
func GetCode(account baseAccountHandler) []byte {
	return account.GetCodeHash()
}

// GetCodeEntry -
func GetCodeEntry(codeHash []byte, trie Updater, marshalizer marshal.Marshalizer) (*CodeEntry, error) {
	return getCodeEntry(codeHash, trie, marshalizer)
}

// RecreateTrieIfNecessary -
func (accountsDB *accountsDBApi) RecreateTrieIfNecessary() error {
	_, err := accountsDB.recreateTrieIfNecessary()

	return err
}

// DoRecreateTrieWithBlockInfo -
func (accountsDB *accountsDBApi) DoRecreateTrieWithBlockInfo(blockInfo common.BlockInfo) error {
	_, err := accountsDB.doRecreateTrieWithBlockInfo(blockInfo)

	return err
}

// SetCurrentBlockInfo -
func (accountsDB *accountsDBApi) SetCurrentBlockInfo(blockInfo common.BlockInfo) {
	accountsDB.mutRecreatedTrieBlockInfo.Lock()
	accountsDB.blockInfo = blockInfo
	accountsDB.mutRecreatedTrieBlockInfo.Unlock()
}

// EmptyErrChanReturningHadContained -
func EmptyErrChanReturningHadContained(errChan chan error) bool {
	return emptyErrChanReturningHadContained(errChan)
}
