package state

import "math/big"

//------- MetaJournalEntryRound

// MetaJournalEntryRound is used to revert a round change
type MetaJournalEntryRound struct {
	account  *MetaAccount
	oldRound uint64
}

// NewMetaJournalEntryRound outputs a new JournalEntry implementation used to revert a round change
func NewMetaJournalEntryRound(account *MetaAccount, oldRound uint64) (*MetaJournalEntryRound, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &MetaJournalEntryRound{
		account:  account,
		oldRound: oldRound,
	}, nil
}

// Revert applies undo operation
func (jen *MetaJournalEntryRound) Revert() (AccountHandler, error) {
	jen.account.Round = jen.oldRound

	return jen.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (jen *MetaJournalEntryRound) IsInterfaceNil() bool {
	if jen == nil {
		return true
	}
	return false
}

//------- MetaJournalEntryTxCount

// MetaJournalEntryTxCount is used to revert a round change
type MetaJournalEntryTxCount struct {
	account    *MetaAccount
	oldTxCount *big.Int
}

// NewMetaJournalEntryTxCount outputs a new JournalEntry implementation used to revert a TxCount change
func NewMetaJournalEntryTxCount(account *MetaAccount, oldTxCount *big.Int) (*MetaJournalEntryTxCount, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &MetaJournalEntryTxCount{
		account:    account,
		oldTxCount: oldTxCount,
	}, nil
}

// Revert applies undo operation
func (jen *MetaJournalEntryTxCount) Revert() (AccountHandler, error) {
	jen.account.TxCount = jen.oldTxCount

	return jen.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (jen *MetaJournalEntryTxCount) IsInterfaceNil() bool {
	if jen == nil {
		return true
	}
	return false
}

//------- MetaJournalEntryMiniBlocksData

// MetaJournalEntryMiniBlocksData is used to revert a round change
type MetaJournalEntryMiniBlocksData struct {
	account           *MetaAccount
	oldMiniBlocksData []*MiniBlockData
}

// NewMetaJournalEntryMiniBlocksData outputs a new JournalEntry implementation used to revert a MiniBlocksData change
func NewMetaJournalEntryMiniBlocksData(account *MetaAccount, oldMiniBlocksData []*MiniBlockData) (*MetaJournalEntryMiniBlocksData, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &MetaJournalEntryMiniBlocksData{
		account:           account,
		oldMiniBlocksData: oldMiniBlocksData,
	}, nil
}

// Revert applies undo operation
func (jen *MetaJournalEntryMiniBlocksData) Revert() (AccountHandler, error) {
	jen.account.MiniBlocks = jen.oldMiniBlocksData

	return jen.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (jen *MetaJournalEntryMiniBlocksData) IsInterfaceNil() bool {
	if jen == nil {
		return true
	}
	return false
}

//------- MetaJournalEntryShardRootHash

// MetaJournalEntryShardRootHash is used to revert a round change
type MetaJournalEntryShardRootHash struct {
	account          *MetaAccount
	oldShardRootHash []byte
}

// NewMetaJournalEntryShardRootHash outputs a new JournalEntry implementation used to revert a ShardRootHash change
func NewMetaJournalEntryShardRootHash(account *MetaAccount, oldShardRootHash []byte) (*MetaJournalEntryShardRootHash, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &MetaJournalEntryShardRootHash{
		account:          account,
		oldShardRootHash: oldShardRootHash,
	}, nil
}

// Revert applies undo operation
func (jen *MetaJournalEntryShardRootHash) Revert() (AccountHandler, error) {
	jen.account.ShardRootHash = jen.oldShardRootHash

	return jen.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (jen *MetaJournalEntryShardRootHash) IsInterfaceNil() bool {
	if jen == nil {
		return true
	}
	return false
}
