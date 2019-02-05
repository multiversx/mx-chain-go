package state_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

//------- JournalEntryCreation

func TestJournalEntryCreation_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.RemoveAccountCalled = func(addressContainer state.AddressContainer) error {
		wasCalled = true

		return nil
	}

	adr := mock.NewAddressMock()
	jec := state.NewJournalEntryCreation(adr)
	err := jec.Revert(acntAdapter)

	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, adr, jec.DirtiedAddress())
}

func TestJournalEntryCreation_RevertNilAddressContainerShouldErr(t *testing.T) {
	t.Parallel()

	jec := state.NewJournalEntryCreation(nil)
	err := jec.Revert(mock.NewAccountsAdapterMock())
	assert.NotNil(t, err)
}

func TestJournalEntryCreation_RevertNilAccountAdapterShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	jec := state.NewJournalEntryCreation(adr)

	err := jec.Revert(nil)
	assert.NotNil(t, err)
}

func TestJournalEntryCreation_RevertAccountAdapterErrorShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	jec := state.NewJournalEntryCreation(adr)

	wasCalled := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.RemoveAccountCalled = func(addressContainer state.AddressContainer) error {
		wasCalled = true

		return errors.New("failure")
	}

	err := jec.Revert(acntAdapter)

	assert.NotNil(t, err)
	assert.True(t, wasCalled)
}

//------- JournalEntryNonce

func TestJournalEntryNonce_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalled = true

		return nil
	}

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	acnt.Nonce = 445

	jec := state.NewJournalEntryNonce(acnt, 1)
	err := jec.Revert(acntAdapter)

	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, adr, jec.DirtiedAddress())
	assert.Equal(t, uint64(1), acnt.Nonce)
}

func TestJournalEntryNonce_RevertNilAccountShouldErr(t *testing.T) {
	t.Parallel()

	jen := state.NewJournalEntryNonce(nil, 1)
	err := jen.Revert(mock.NewAccountsAdapterMock())
	assert.NotNil(t, err)
}

func TestJournalEntryNonce_RevertNilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	jen := state.NewJournalEntryNonce(acnt, 1)

	err := jen.Revert(nil)
	assert.NotNil(t, err)
}

func TestJournalEntryNonce_RevertNilAddressShouldErr(t *testing.T) {
	t.Parallel()

	acnt := mock.NewJournalizedAccountWrapMock(nil)
	jen := state.NewJournalEntryNonce(acnt, 1)
	err := jen.Revert(mock.NewAccountsAdapterMock())
	assert.NotNil(t, err)
}

func TestJournalEntryNonce_RevertAccountAdapterErrorShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)

	acnt = mock.NewJournalizedAccountWrapMock(adr)
	wasCalled := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalled = true

		return nil
	}

	acnt.Nonce = 445

	jen := state.NewJournalEntryNonce(acnt, 1)
	err := jen.Revert(acntAdapter)

	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, adr, jen.DirtiedAddress())
	assert.Equal(t, uint64(1), acnt.Nonce)
}

//------- JournalEntryBalance

func TestJournalEntryBalance_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalled = true

		return nil
	}

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	acnt.Balance = big.NewInt(445)

	jec := state.NewJournalEntryBalance(acnt, big.NewInt(2))
	err := jec.Revert(acntAdapter)

	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, adr, jec.DirtiedAddress())
	assert.Equal(t, big.NewInt(2), acnt.Balance)
}

func TestJournalEntryBalance_RevertNilAccountShouldErr(t *testing.T) {
	t.Parallel()

	jeb := state.NewJournalEntryBalance(nil, big.NewInt(2))
	err := jeb.Revert(mock.NewAccountsAdapterMock())
	assert.NotNil(t, err)
}

func TestJournalEntryBalance_RevertNilAccountAdapterShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	jeb := state.NewJournalEntryBalance(acnt, big.NewInt(2))

	err := jeb.Revert(nil)
	assert.NotNil(t, err)
}

func TestJournalEntryBalance_RevertNilAddressShouldErr(t *testing.T) {
	t.Parallel()

	acnt := mock.NewJournalizedAccountWrapMock(nil)
	jen := state.NewJournalEntryBalance(acnt, big.NewInt(2))
	err := jen.Revert(mock.NewAccountsAdapterMock())
	assert.NotNil(t, err)
}

func TestJournalEntryBalance_RevertAccountAdapterErrorShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)

	acnt = mock.NewJournalizedAccountWrapMock(adr)
	wasCalled := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalled = true

		return nil
	}

	acnt.Balance = big.NewInt(445)

	jeb := state.NewJournalEntryBalance(acnt, big.NewInt(2))
	err := jeb.Revert(acntAdapter)

	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, adr, jeb.DirtiedAddress())
	assert.Equal(t, big.NewInt(2), acnt.Balance)
}

//------- JournalEntryCodeHash

func TestJournalEntryCodeHash_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalled = true

		return nil
	}

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	acnt.CodeHash = []byte("aaaa")

	jec := state.NewJournalEntryCodeHash(acnt, make([]byte, mock.HasherMock{}.Size()))
	err := jec.Revert(acntAdapter)

	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, adr, jec.DirtiedAddress())
	assert.Equal(t, make([]byte, mock.HasherMock{}.Size()), acnt.CodeHash)
}

func TestJournalEntryCodeHash_RevertNilAccountShouldErr(t *testing.T) {
	t.Parallel()

	jech := state.NewJournalEntryCodeHash(nil, []byte("aaa"))
	err := jech.Revert(mock.NewAccountsAdapterMock())
	assert.NotNil(t, err)
}

func TestJournalEntryCodeHash_RevertNilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	jech := state.NewJournalEntryCodeHash(acnt, []byte("bbb"))
	err := jech.Revert(nil)
	assert.NotNil(t, err)
}

func TestJournalEntryCodeHash_RevertNilAddressShouldErr(t *testing.T) {
	t.Parallel()

	acnt := mock.NewJournalizedAccountWrapMock(nil)
	jech := state.NewJournalEntryCodeHash(acnt, []byte("aaa"))
	err := jech.Revert(mock.NewAccountsAdapterMock())
	assert.NotNil(t, err)
}

func TestJournalEntryCodeHash_RevertNilCodeHashShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalled = true

		return nil
	}

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	jech := state.NewJournalEntryCodeHash(acnt, nil)
	err := jech.Revert(acntAdapter)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestJournalEntryCodeHash_RevertAccountsAdapterErrorShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	wasCalled := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalled = true

		return nil
	}

	acnt.CodeHash = []byte("aaaa")

	jech := state.NewJournalEntryCodeHash(acnt, make([]byte, mock.HasherMock{}.Size()))
	err := jech.Revert(acntAdapter)

	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, adr, jech.DirtiedAddress())
	assert.Equal(t, make([]byte, mock.HasherMock{}.Size()), acnt.CodeHash)
}

//------- JournalEntryCode

func TestJournalEntryCode_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.RemoveCodeCalled = func(codeHash []byte) error {
		wasCalled = true

		return nil
	}

	jec := state.NewJournalEntryCode(make([]byte, mock.HasherMock{}.Size()))
	err := jec.Revert(acntAdapter)

	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, nil, jec.DirtiedAddress())
}

func TestJournalEntryCode_RevertNilCodeHashShouldErr(t *testing.T) {
	t.Parallel()

	jec := state.NewJournalEntryCode(nil)
	err := jec.Revert(mock.NewAccountsAdapterMock())
	assert.NotNil(t, err)
}

func TestJournalEntryCodeRevertNilAccountAdapterShouldErr(t *testing.T) {
	t.Parallel()

	jec := state.NewJournalEntryCode([]byte("a"))
	err := jec.Revert(nil)
	assert.NotNil(t, err)
}

func TestJournalEntryCode_RevertAccountAdapterErrorShouldErr(t *testing.T) {
	t.Parallel()

	jec := state.NewJournalEntryCode(make([]byte, mock.HasherMock{}.Size()))

	wasCalled := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.RemoveCodeCalled = func(codeHash []byte) error {
		wasCalled = true

		return errors.New("failure")
	}

	err := jec.Revert(acntAdapter)

	assert.NotNil(t, err)
	assert.True(t, wasCalled)
}

//------- JournalEntryRootHash

func TestJournalEntryRootHash_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalledSave := false
	wasCalledRetrieved := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalledSave = true

		return nil
	}

	acntAdapter.RetrieveDataTrieCalled = func(acountWrapper state.AccountWrapper) error {
		wasCalledRetrieved = true

		return nil
	}

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	acnt.RootHash = []byte("aaaa")

	jerh := state.NewJournalEntryRootHash(acnt, make([]byte, mock.HasherMock{}.Size()))
	err := jerh.Revert(acntAdapter)

	assert.Nil(t, err)
	assert.True(t, wasCalledSave)
	assert.True(t, wasCalledRetrieved)
	assert.Equal(t, adr, jerh.DirtiedAddress())
	assert.Equal(t, make([]byte, mock.HasherMock{}.Size()), acnt.RootHash)
}

func TestJournalEntryRootHash_RevertNilAccountShouldErr(t *testing.T) {
	t.Parallel()

	jerh := state.NewJournalEntryRootHash(nil, []byte("aaa"))
	err := jerh.Revert(mock.NewAccountsAdapterMock())
	assert.NotNil(t, err)
}

func TestJournalEntryRootHash_RevertNilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	jerh := state.NewJournalEntryRootHash(acnt, []byte("bbb"))
	err := jerh.Revert(nil)
	assert.NotNil(t, err)
}

func TestJournalEntryRootHash_RevertNilAddressShouldErr(t *testing.T) {
	t.Parallel()

	acnt := mock.NewJournalizedAccountWrapMock(nil)
	jerh := state.NewJournalEntryRootHash(acnt, []byte("aaa"))
	err := jerh.Revert(mock.NewAccountsAdapterMock())
	assert.NotNil(t, err)
}

func TestJournalEntryRootHash_RevertNilCodeHashShouldWork(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	jerh := state.NewJournalEntryRootHash(acnt, nil)

	wasCalledSave := false
	wasCalledRetrieved := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalledSave = true

		return nil
	}

	acntAdapter.RetrieveDataTrieCalled = func(acountWrapper state.AccountWrapper) error {
		wasCalledRetrieved = true

		return nil
	}

	err := jerh.Revert(acntAdapter)
	assert.Nil(t, err)
	assert.True(t, wasCalledRetrieved)
	assert.True(t, wasCalledSave)
}

func TestJournalEntryRootHash_RevertAccountsAdapterSaveErrorShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	wasCalledSave := false
	wasCalledRetrieved := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalledSave = true

		return errors.New("failure")
	}

	acntAdapter.RetrieveDataTrieCalled = func(acountWrapper state.AccountWrapper) error {
		wasCalledRetrieved = true

		return nil
	}

	acnt.CodeHash = []byte("aaaa")

	jerh := state.NewJournalEntryRootHash(acnt, make([]byte, mock.HasherMock{}.Size()))
	err := jerh.Revert(acntAdapter)

	assert.NotNil(t, err)
	assert.True(t, wasCalledSave)
	assert.True(t, wasCalledRetrieved)
	assert.Equal(t, adr, jerh.DirtiedAddress())
	assert.Equal(t, make([]byte, mock.HasherMock{}.Size()), acnt.RootHash)
}

func TestJournalEntryRootHash_RevertAccountsAdapterRetrieveErrorShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	wasCalledSave := false
	wasCalledRetrieved := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalledSave = true

		return nil
	}

	acntAdapter.RetrieveDataTrieCalled = func(acountWrapper state.AccountWrapper) error {
		wasCalledRetrieved = true

		return errors.New("failure")
	}

	acnt.CodeHash = []byte("aaaa")

	jerh := state.NewJournalEntryRootHash(acnt, make([]byte, mock.HasherMock{}.Size()))
	err := jerh.Revert(acntAdapter)

	assert.NotNil(t, err)
	assert.False(t, wasCalledSave)
	assert.True(t, wasCalledRetrieved)
	assert.Equal(t, adr, jerh.DirtiedAddress())
	assert.Equal(t, make([]byte, mock.HasherMock{}.Size()), acnt.RootHash)
}

//------- JournalEntryRootHash

func TestJournalEntryData_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalledClear := false

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	acnt.ClearDataCachesCalled = func() {
		wasCalledClear = true
	}

	trie := mock.NewMockTrie()

	jed := state.NewJournalEntryData(acnt, trie)
	err := jed.Revert(nil)
	assert.Nil(t, err)
	assert.True(t, wasCalledClear)
	assert.Equal(t, trie, jed.Trie())
	assert.Equal(t, nil, jed.DirtiedAddress())

}

func TestJournalEntryData_RevertNilAccountShouldWork(t *testing.T) {
	t.Parallel()

	trie := mock.NewMockTrie()

	jed := state.NewJournalEntryData(nil, trie)
	err := jed.Revert(nil)
	assert.NotNil(t, err)
}

//------- JournalEntryAppendRegistration

func TestJournalEntryAppendRegistration_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalledTrim := false

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	acnt.TrimLastRegistrationDataCalled = func() error {
		wasCalledTrim = true
		return nil
	}

	jed := state.NewJournalEntryAppendRegistration(acnt)
	err := jed.Revert(nil)
	assert.Nil(t, err)
	assert.True(t, wasCalledTrim)
	assert.Equal(t, acnt.AddressContainer(), jed.DirtiedAddress())

}

func TestJournalEntryAppendRegistration_NilAccountTrimShouldErr(t *testing.T) {
	t.Parallel()

	jed := state.NewJournalEntryAppendRegistration(nil)
	err := jed.Revert(nil)
	assert.Equal(t, state.ErrNilJurnalizingAccountWrapper, err)
}

func TestJournalEntryAppendRegistration_AppendErrTrimShouldErr(t *testing.T) {
	t.Parallel()

	wasCalledTrim := false

	adr := mock.NewAddressMock()
	acnt := mock.NewJournalizedAccountWrapMock(adr)
	acnt.TrimLastRegistrationDataCalled = func() error {
		wasCalledTrim = true
		return errors.New("failure")
	}

	jed := state.NewJournalEntryAppendRegistration(acnt)
	err := jed.Revert(nil)
	assert.NotNil(t, err)
	assert.True(t, wasCalledTrim)
	assert.Equal(t, acnt.AddressContainer(), jed.DirtiedAddress())
}
