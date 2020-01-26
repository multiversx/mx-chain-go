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
func (mjer *MetaJournalEntryRound) Revert() (AccountHandler, error) {
	mjer.account.Round = mjer.oldRound

	return mjer.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mjer *MetaJournalEntryRound) IsInterfaceNil() bool {
	return mjer == nil
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
func (mjetc *MetaJournalEntryTxCount) Revert() (AccountHandler, error) {
	mjetc.account.TxCount = mjetc.oldTxCount

	return mjetc.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mjetc *MetaJournalEntryTxCount) IsInterfaceNil() bool {
	return mjetc == nil
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
func (mjembd *MetaJournalEntryMiniBlocksData) Revert() (AccountHandler, error) {
	mjembd.account.MiniBlocks = mjembd.oldMiniBlocksData

	return mjembd.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mjembd *MetaJournalEntryMiniBlocksData) IsInterfaceNil() bool {
	return mjembd == nil
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
func (mjesrh *MetaJournalEntryShardRootHash) Revert() (AccountHandler, error) {
	mjesrh.account.ShardRootHash = mjesrh.oldShardRootHash

	return mjesrh.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mjesrh *MetaJournalEntryShardRootHash) IsInterfaceNil() bool {
	return mjesrh == nil
}
