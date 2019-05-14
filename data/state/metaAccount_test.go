package state_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/stretchr/testify/assert"
)

func TestMetaAccount_MarshalUnmarshal_ShouldWork(t *testing.T) {
	t.Parallel()

	addr := &mock.AddressMock{}
	addrTr := &mock.AccountTrackerStub{}
	acnt, _ := state.NewMetaAccount(addr, addrTr)

	marshalizer := mock.MarshalizerMock{}
	buff, _ := marshalizer.Marshal(&acnt)

	acntRecovered, _ := state.NewMetaAccount(addr, addrTr)
	_ = marshalizer.Unmarshal(acntRecovered, buff)

	assert.Equal(t, acnt, acntRecovered)
}

func TestMetaAccount_NewAccountNilAddress(t *testing.T) {
	t.Parallel()

	acc, err := state.NewMetaAccount(nil, &mock.AccountTrackerStub{})

	assert.Nil(t, acc)
	assert.Equal(t, err, state.ErrNilAddressContainer)
}

func TestMetaAccount_NewMetaAccountNilAaccountTracker(t *testing.T) {
	t.Parallel()

	acc, err := state.NewMetaAccount(&mock.AddressMock{}, nil)

	assert.Nil(t, acc)
	assert.Equal(t, err, state.ErrNilAccountTracker)
}

func TestMetaAccount_NewMetaAccountOk(t *testing.T) {
	t.Parallel()

	acc, err := state.NewMetaAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})

	assert.NotNil(t, acc)
	assert.Nil(t, err)
}

func TestMetaAccount_AddressContainer(t *testing.T) {
	t.Parallel()

	addr := &mock.AddressMock{}
	acc, err := state.NewMetaAccount(addr, &mock.AccountTrackerStub{})

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, addr, acc.AddressContainer())
}

func TestMetaAccount_GetCode(t *testing.T) {
	t.Parallel()

	acc, err := state.NewMetaAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})

	code := []byte("code")
	acc.SetCode(code)

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, code, acc.GetCode())
}

func TestMetaAccount_GetCodeHash(t *testing.T) {
	t.Parallel()

	acc, err := state.NewMetaAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})

	code := []byte("code")
	acc.CodeHash = code

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, code, acc.GetCodeHash())
}

func TestMetaAccount_SetCodeHash(t *testing.T) {
	t.Parallel()

	acc, err := state.NewMetaAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})

	code := []byte("code")
	acc.SetCodeHash(code)

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, code, acc.GetCodeHash())
}

func TestMetaAccount_GetRootHash(t *testing.T) {
	t.Parallel()

	acc, err := state.NewMetaAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})

	root := []byte("root")
	acc.RootHash = root

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, root, acc.GetRootHash())
}

func TestMetaAccount_SetRootHash(t *testing.T) {
	t.Parallel()

	acc, err := state.NewMetaAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})

	root := []byte("root")
	acc.SetRootHash(root)

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, root, acc.GetRootHash())
}

func TestMetaAccount_DataTrie(t *testing.T) {
	t.Parallel()

	acc, err := state.NewMetaAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})
	trie := &mock.TrieMock{}

	acc.SetDataTrie(trie)

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, trie, acc.DataTrie())
}

func TestMetaAccount_SetRoundWithJournal(t *testing.T) {
	t.Parallel()

	journalizeCalled := 0
	saveAccountCalled := 0
	tracker := &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeCalled++
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
	}

	acc, err := state.NewMetaAccount(&mock.AddressMock{}, tracker)

	round := uint64(0)
	err = acc.SetRoundWithJournal(round)

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, round, acc.Round)
	assert.Equal(t, 1, journalizeCalled)
	assert.Equal(t, 1, saveAccountCalled)
}

func TestMetaAccount_SetTxCountWithJournal(t *testing.T) {
	t.Parallel()

	journalizeCalled := 0
	saveAccountCalled := 0
	tracker := &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeCalled++
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
	}

	acc, err := state.NewMetaAccount(&mock.AddressMock{}, tracker)

	txCount := big.NewInt(15)
	err = acc.SetTxCountWithJournal(txCount)

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, txCount, acc.TxCount)
	assert.Equal(t, 1, journalizeCalled)
	assert.Equal(t, 1, saveAccountCalled)
}

func TestMetaAccount_SetCodeHashWithJournal(t *testing.T) {
	t.Parallel()

	journalizeCalled := 0
	saveAccountCalled := 0
	tracker := &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeCalled++
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
	}

	acc, err := state.NewMetaAccount(&mock.AddressMock{}, tracker)

	codeHash := []byte("codehash")
	err = acc.SetCodeHashWithJournal(codeHash)

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, codeHash, acc.CodeHash)
	assert.Equal(t, 1, journalizeCalled)
	assert.Equal(t, 1, saveAccountCalled)
}

func TestMetaAccount_SetRootHashWithJournal(t *testing.T) {
	t.Parallel()

	journalizeCalled := 0
	saveAccountCalled := 0
	tracker := &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeCalled++
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
	}

	acc, err := state.NewMetaAccount(&mock.AddressMock{}, tracker)

	rootHash := []byte("roothash")
	err = acc.SetRootHashWithJournal(rootHash)

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, rootHash, acc.RootHash)
	assert.Equal(t, 1, journalizeCalled)
	assert.Equal(t, 1, saveAccountCalled)
}

func TestMetaAccount_SetMiniBlocksDataWithJournal(t *testing.T) {
	t.Parallel()

	journalizeCalled := 0
	saveAccountCalled := 0
	tracker := &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeCalled++
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
	}

	acc, err := state.NewMetaAccount(&mock.AddressMock{}, tracker)

	mbs := make([]*state.MiniBlockData, 2)
	err = acc.SetMiniBlocksDataWithJournal(mbs)

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, mbs, acc.MiniBlocks)
	assert.Equal(t, 1, journalizeCalled)
	assert.Equal(t, 1, saveAccountCalled)
}

func TestMetaAccount_SetShardRootHashWithJournal(t *testing.T) {
	t.Parallel()

	journalizeCalled := 0
	saveAccountCalled := 0
	tracker := &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeCalled++
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
	}

	acc, err := state.NewMetaAccount(&mock.AddressMock{}, tracker)

	shardRootHash := []byte("shardroothash")
	err = acc.SetShardRootHashWithJournal(shardRootHash)

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, shardRootHash, acc.ShardRootHash)
	assert.Equal(t, 1, journalizeCalled)
	assert.Equal(t, 1, saveAccountCalled)
}
