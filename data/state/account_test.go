package state_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/stretchr/testify/assert"
)

func TestAccount_MarshalUnmarshal_ShouldWork(t *testing.T) {
	t.Parallel()

	addr := &mock.AddressMock{}
	addrTr := &mock.AccountTrackerStub{}
	acnt, _ := state.NewAccount(addr, addrTr)
	acnt.Nonce = 0
	acnt.Balance = big.NewInt(56)
	acnt.CodeHash = nil
	acnt.RootHash = nil

	marshalizer := mock.MarshalizerMock{}

	buff, _ := marshalizer.Marshal(&acnt)

	acntRecovered, _ := state.NewAccount(addr, addrTr)
	_ = marshalizer.Unmarshal(acntRecovered, buff)

	assert.Equal(t, acnt, acntRecovered)
}

func TestAccount_NewAccountNilAddress(t *testing.T) {
	t.Parallel()

	acc, err := state.NewAccount(nil, &mock.AccountTrackerStub{})

	assert.Nil(t, acc)
	assert.Equal(t, err, state.ErrNilAddressContainer)
}

func TestAccount_NewAccountNilAaccountTracker(t *testing.T) {
	t.Parallel()

	acc, err := state.NewAccount(&mock.AddressMock{}, nil)

	assert.Nil(t, acc)
	assert.Equal(t, err, state.ErrNilAccountTracker)
}

func TestAccount_NewAccountOk(t *testing.T) {
	t.Parallel()

	acc, err := state.NewAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})

	assert.NotNil(t, acc)
	assert.Nil(t, err)
}

func TestAccount_AddressContainer(t *testing.T) {
	t.Parallel()

	addr := &mock.AddressMock{}
	acc, err := state.NewAccount(addr, &mock.AccountTrackerStub{})

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, addr, acc.AddressContainer())
}

func TestAccount_GetCode(t *testing.T) {
	t.Parallel()

	acc, err := state.NewAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})
	assert.Nil(t, err)

	code := []byte("code")
	acc.SetCode(code)

	assert.NotNil(t, acc)
	assert.Equal(t, code, acc.GetCode())
}

func TestAccount_GetCodeHash(t *testing.T) {
	t.Parallel()

	acc, err := state.NewAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})
	assert.Nil(t, err)

	code := []byte("code")
	acc.CodeHash = code

	assert.NotNil(t, acc)
	assert.Equal(t, code, acc.GetCodeHash())
}

func TestAccount_SetCodeHash(t *testing.T) {
	t.Parallel()

	acc, err := state.NewAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})
	assert.Nil(t, err)

	code := []byte("code")
	acc.SetCodeHash(code)

	assert.NotNil(t, acc)
	assert.Equal(t, code, acc.GetCodeHash())
}

func TestAccount_GetRootHash(t *testing.T) {
	t.Parallel()

	acc, err := state.NewAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})
	assert.Nil(t, err)

	root := []byte("root")
	acc.RootHash = root

	assert.NotNil(t, acc)
	assert.Equal(t, root, acc.GetRootHash())
}

func TestAccount_SetRootHash(t *testing.T) {
	t.Parallel()

	acc, err := state.NewAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})
	assert.Nil(t, err)

	root := []byte("root")
	acc.SetRootHash(root)

	assert.NotNil(t, acc)
	assert.Equal(t, root, acc.GetRootHash())
}

func TestAccount_DataTrie(t *testing.T) {
	t.Parallel()

	acc, err := state.NewAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})
	assert.Nil(t, err)

	trie := &mock.TrieMock{}
	acc.SetDataTrie(trie)

	assert.NotNil(t, acc)
	assert.Equal(t, trie, acc.DataTrie())
}

func TestAccount_SetNonceWithJournal(t *testing.T) {
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

	acc, err := state.NewAccount(&mock.AddressMock{}, tracker)
	assert.Nil(t, err)

	nonce := uint64(0)
	err = acc.SetNonceWithJournal(nonce)

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, nonce, acc.Nonce)
	assert.Equal(t, 1, journalizeCalled)
	assert.Equal(t, 1, saveAccountCalled)
}

func TestAccount_SetBalanceWithJournal(t *testing.T) {
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

	acc, err := state.NewAccount(&mock.AddressMock{}, tracker)
	assert.Nil(t, err)

	balance := big.NewInt(15)
	err = acc.SetBalanceWithJournal(balance)

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, balance, acc.Balance)
	assert.Equal(t, 1, journalizeCalled)
	assert.Equal(t, 1, saveAccountCalled)
}

func TestAccount_SetCodeHashWithJournal(t *testing.T) {
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

	acc, err := state.NewAccount(&mock.AddressMock{}, tracker)
	assert.Nil(t, err)

	codeHash := []byte("codehash")
	err = acc.SetCodeHashWithJournal(codeHash)

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, codeHash, acc.CodeHash)
	assert.Equal(t, 1, journalizeCalled)
	assert.Equal(t, 1, saveAccountCalled)
}

func TestAccount_SetRootHashWithJournal(t *testing.T) {
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

	acc, err := state.NewAccount(&mock.AddressMock{}, tracker)
	assert.Nil(t, err)

	rootHash := []byte("roothash")
	err = acc.SetRootHashWithJournal(rootHash)

	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, rootHash, acc.RootHash)
	assert.Equal(t, 1, journalizeCalled)
	assert.Equal(t, 1, saveAccountCalled)
}
