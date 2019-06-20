package state_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/stretchr/testify/assert"
)

//------- MetaJournalEntryRound

func TestNewetaJournalEntryRound_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewMetaJournalEntryRound(nil, 0)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestNewetaJournalEntryRound_ShouldWork(t *testing.T) {
	t.Parallel()
	accnt, _ := state.NewMetaAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewMetaJournalEntryRound(accnt, 0)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestNewMetaJournalEntryRound_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	round := uint64(52)
	accnt, _ := state.NewMetaAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewMetaJournalEntryRound(accnt, round)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, round, accnt.Round)
}

//------- MetaJournalEntryMiniBlocksData

func TestNewetaJournalEntryMiniBlocksData_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewMetaJournalEntryMiniBlocksData(nil, nil)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestNewetaJournalEntryMiniBlocksData_ShouldWork(t *testing.T) {
	t.Parallel()

	accnt, _ := state.NewMetaAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewMetaJournalEntryMiniBlocksData(accnt, nil)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestNewMetaJournalEntryMiniBlocksData_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	miniBlocksData := make([]*state.MiniBlockData, 2)
	accnt, _ := state.NewMetaAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewMetaJournalEntryMiniBlocksData(accnt, miniBlocksData)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, miniBlocksData, accnt.MiniBlocks)
}

//------- MetaJournalEntryTxCount

func TestNewetaJournalEntryTxCount_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewMetaJournalEntryTxCount(nil, nil)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestNewetaJournalEntryTxCount_ShouldWork(t *testing.T) {
	t.Parallel()
	accnt, _ := state.NewMetaAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewMetaJournalEntryTxCount(accnt, big.NewInt(0))

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestNewMetaJournalEntryTxCount_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	txCount := big.NewInt(34)
	accnt, _ := state.NewMetaAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewMetaJournalEntryTxCount(accnt, txCount)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, txCount, accnt.TxCount)
}

//------- MetaJournalEntryShardRootHash

func TestNewetaJournalEntryShardRootHash_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewMetaJournalEntryShardRootHash(nil, nil)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestNewMetaJournalEntryShardRootHash_ShouldWork(t *testing.T) {
	t.Parallel()
	accnt, _ := state.NewMetaAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewMetaJournalEntryShardRootHash(accnt, []byte("shardroothash"))

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestNewMetaJournalEntryShardRootHash_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	shardRootHash := []byte("shardroothash")
	accnt, _ := state.NewMetaAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewMetaJournalEntryShardRootHash(accnt, shardRootHash)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, shardRootHash, accnt.ShardRootHash)
}
