package state_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewSimpleAccountWrapInvalidValsShouldErr(t *testing.T) {
	t.Parallel()

	_, err := state.NewAccountWrap(nil, state.NewAccount())
	assert.NotNil(t, err)

	_, err = state.NewAccountWrap(mock.NewAddressMock(), nil)
	assert.NotNil(t, err)
}

func TestSimpleAccountWrapGettersSettersValidValsShouldWork(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()

	acntWrap, err := state.NewAccountWrap(adr, state.NewAccount())
	assert.Nil(t, err)

	trie := mock.NewMockTrie()

	acntWrap.SetCode([]byte("aaaa"))
	acntWrap.SetDataTrie(trie)

	assert.Equal(t, adr, acntWrap.AddressContainer())
	assert.Equal(t, []byte("aaaa"), acntWrap.Code())
	assert.Equal(t, trie, acntWrap.DataTrie())
}

func TestAccountWrap_AppendRegistrationDataNonRegAddress_ShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()

	acntWrap, err := state.NewAccountWrap(adr, state.NewAccount())
	assert.Nil(t, err)

	err = acntWrap.AppendRegistrationData(&state.RegistrationData{})
	assert.Equal(t, err, state.ErrNotSupportedAccountsRegistration)
}

func TestAccountWrap_AppendRegistrationDataRegAddress_ShouldWork(t *testing.T) {
	t.Parallel()

	acntWrap, err := state.NewAccountWrap(state.RegistrationAddress, state.NewAccount())
	assert.Nil(t, err)

	err = acntWrap.AppendRegistrationData(&state.RegistrationData{})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(acntWrap.RegistrationData))
}

func TestAccountWrap_TrimLastRegistrationDataNonRegAddress_ShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()

	acntWrap, err := state.NewAccountWrap(adr, state.NewAccount())
	assert.Nil(t, err)

	err = acntWrap.TrimLastRegistrationData()
	assert.Equal(t, err, state.ErrNotSupportedAccountsRegistration)
}

func TestAccountWrap_TrimLastRegistrationDataEmptySlice_ShouldErr(t *testing.T) {
	t.Parallel()

	acntWrap, err := state.NewAccountWrap(state.RegistrationAddress, state.NewAccount())
	assert.Nil(t, err)

	err = acntWrap.TrimLastRegistrationData()
	assert.Equal(t, err, state.ErrTrimOperationNotSupported)
}

func TestAccountWrap_TrimLastRegistrationDataOkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	acntWrap, err := state.NewAccountWrap(state.RegistrationAddress, state.NewAccount())
	assert.Nil(t, err)

	err = acntWrap.AppendRegistrationData(&state.RegistrationData{})
	assert.Nil(t, err)

	err = acntWrap.TrimLastRegistrationData()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(acntWrap.RegistrationData))
}
