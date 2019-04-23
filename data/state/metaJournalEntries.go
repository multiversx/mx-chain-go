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
		return nil, ErrNilAccount
	}

	return &MetaJournalEntryRound{
		account:  account,
		oldRound: oldRound,
	}, nil
}

// Revert applies undo operation
func (jen *MetaJournalEntryRound) Revert() (AccountWrapper, error) {
	jen.account.Round = jen.oldRound

	return jen.account, nil
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
		return nil, ErrNilAccount
	}

	return &MetaJournalEntryTxCount{
		account:    account,
		oldTxCount: oldTxCount,
	}, nil
}

// Revert applies undo operation
func (jen *MetaJournalEntryTxCount) Revert() (AccountWrapper, error) {
	jen.account.TxCount = jen.oldTxCount

	return jen.account, nil
}

//------- MetaJournalEntryMiniBlocksData

// MetaJournalEntryRound is used to revert a round change
type MetaJournalEntryMiniBlocksData struct {
	account           *MetaAccount
	oldMiniBlocksData []*MiniBlockData
}

// NewMetaJournalEntryRound outputs a new JournalEntry implementation used to revert a MiniBlocksData change
func NewMetaJournalEntryMiniBlocksData(account *MetaAccount, oldMiniBlocksData []*MiniBlockData) (*MetaJournalEntryMiniBlocksData, error) {
	if account == nil {
		return nil, ErrNilAccount
	}

	return &MetaJournalEntryMiniBlocksData{
		account:           account,
		oldMiniBlocksData: oldMiniBlocksData,
	}, nil
}

// Revert applies undo operation
func (jen *MetaJournalEntryMiniBlocksData) Revert() (AccountWrapper, error) {
	jen.account.MiniBlocks = jen.oldMiniBlocksData

	return jen.account, nil
}
