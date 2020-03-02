package state_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.False(t, acc.IsInterfaceNil())
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

func TestAccount_SetAndGetNonce(t *testing.T) {
	t.Parallel()

	acc, err := state.NewAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{})
	assert.Nil(t, err)

	nonce := uint64(37)
	acc.SetNonce(nonce)

	assert.Equal(t, nonce, acc.GetNonce())
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

	trie := &mock.TrieStub{}
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
	err = acc.AddToBalance(balance)

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

func TestAccount_SetOwnerAddressWithJournal(t *testing.T) {
	t.Parallel()

	journalizeWasCalled, saveAccountWasCalled := false, false
	tracker := &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeWasCalled = true
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountWasCalled = true
			return nil
		},
	}

	acc, _ := state.NewAccount(&mock.AddressMock{}, tracker)

	owner := []byte("owner")
	err := acc.SetOwnerAddressWithJournal(owner)
	require.Nil(t, err)
	require.True(t, journalizeWasCalled)
	require.True(t, saveAccountWasCalled)
	require.Equal(t, acc.OwnerAddress, owner)
}

func TestAccount_AddAndClaimDeveloperRewardsWrongAddress(t *testing.T) {
	t.Parallel()

	tracker := &mock.AccountTrackerStub{}
	acc, _ := state.NewAccount(&mock.AddressMock{}, tracker)
	owner := []byte("owner")
	_ = acc.SetOwnerAddressWithJournal(owner)

	devRewards := big.NewInt(100)
	err := acc.AddToDeveloperReward(devRewards)
	require.Nil(t, err)

	rewards, err := acc.ClaimDeveloperRewards([]byte("wrongAddr"))
	require.Nil(t, rewards)
	require.Equal(t, state.ErrOperationNotPermitted, err)

}

func TestAccount_AddAndClaimDeveloperRewardsShouldWork(t *testing.T) {
	t.Parallel()

	tracker := &mock.AccountTrackerStub{}
	acc, _ := state.NewAccount(&mock.AddressMock{}, tracker)
	owner := []byte("owner")
	_ = acc.SetOwnerAddressWithJournal(owner)

	devRewards := big.NewInt(100)
	err := acc.AddToDeveloperReward(devRewards)
	require.Nil(t, err)

	rewards, err := acc.ClaimDeveloperRewards(owner)
	require.Nil(t, err)
	require.Equal(t, devRewards, rewards)
	require.Equal(t, big.NewInt(0), acc.DeveloperReward)
}

func TestAccount_ChangeOwnerAddress(t *testing.T) {
	t.Parallel()

	tracker := &mock.AccountTrackerStub{}

	acc, _ := state.NewAccount(mock.NewAddressMock(), tracker)
	owner := []byte("owner")
	_ = acc.SetOwnerAddressWithJournal(owner)

	newOwnerAddr := []byte("newOwner")
	err := acc.ChangeOwnerAddress([]byte("wrong"), newOwnerAddr)
	require.Equal(t, state.ErrOperationNotPermitted, err)

	err = acc.ChangeOwnerAddress(owner, newOwnerAddr)
	require.Equal(t, state.ErrInvalidAddressLength, err)

	correctNewAddr := []byte("00000000000000000000000000000000")
	err = acc.ChangeOwnerAddress(owner, correctNewAddr)
	require.Nil(t, err)
	require.Equal(t, correctNewAddr, acc.GetOwnerAddress())
}

func TestAccount_AddToBalance(t *testing.T) {
	t.Parallel()

	tracker := &mock.AccountTrackerStub{}

	acc, _ := state.NewAccount(mock.NewAddressMock(), tracker)

	value := big.NewInt(1000)
	err := acc.AddToBalance(value)
	require.Nil(t, err)

	result := acc.GetBalance()
	require.Equal(t, value, result)

	value = big.NewInt(-500)
	expected := big.NewInt(0).Add(acc.GetBalance(), value)
	err = acc.AddToBalance(value)
	require.Nil(t, err)
	require.Equal(t, expected, acc.GetBalance())

	value = big.NewInt(-1000)
	err = acc.AddToBalance(value)
	require.Equal(t, state.ErrInsufficientFunds, err)
}

func TestAccount_SubFromBalance(t *testing.T) {
	t.Parallel()

	tracker := &mock.AccountTrackerStub{}

	acc, _ := state.NewAccount(mock.NewAddressMock(), tracker)

	_ = acc.AddToBalance(big.NewInt(10000))

	value := big.NewInt(1000)
	expectedResult := big.NewInt(0).Sub(acc.Balance, value)
	err := acc.SubFromBalance(value)
	require.Nil(t, err)

	result := acc.GetBalance()
	require.Equal(t, expectedResult, result)

	value = big.NewInt(-500)
	expected := big.NewInt(0).Sub(acc.GetBalance(), value)
	err = acc.SubFromBalance(value)
	require.Nil(t, err)
	require.Equal(t, expected, acc.GetBalance())

	value = big.NewInt(10000000)
	err = acc.SubFromBalance(value)
	require.Equal(t, state.ErrInsufficientFunds, err)
}
