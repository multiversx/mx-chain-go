package state_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/stretchr/testify/assert"
)

func TestBaseAccountLoadFromDbAccountInvalidDataShouldErr(t *testing.T) {
	t.Parallel()

	baseAccnt := state.NewAccount()

	//nil source
	err := baseAccnt.LoadFromDbAccount(nil)
	assert.NotNil(t, err)

	//decode error
	dbAccnt := state.NewDbAccount()
	dbAccnt.SetNonce(1)
	dbAccnt.SetBalance([]byte("bbb"))
	dbAccnt.SetCodeHash([]byte("ccc"))
	dbAccnt.SetRootHash([]byte("rrr"))
	err = baseAccnt.LoadFromDbAccount(dbAccnt)
	assert.NotNil(t, err)
}

func TestBaseAccountLoadFromDbAccountValidDataShouldWork(t *testing.T) {
	t.Parallel()

	baseAccnt := state.NewAccount()

	//decode error
	balance := big.NewInt(40)
	buff, err := balance.GobEncode()
	assert.Nil(t, err)

	dbAccnt := state.NewDbAccount()
	dbAccnt.SetNonce(1)
	dbAccnt.SetBalance(buff)
	dbAccnt.SetCodeHash([]byte("ccc"))
	dbAccnt.SetRootHash([]byte("rrr"))

	err = baseAccnt.LoadFromDbAccount(dbAccnt)
	assert.Nil(t, err)

	assert.Equal(t, dbAccnt.Nonce(), baseAccnt.Nonce())
	assert.Equal(t, *balance, baseAccnt.Balance())
	assert.Equal(t, dbAccnt.CodeHash(), baseAccnt.CodeHash())
	assert.Equal(t, dbAccnt.RootHash(), baseAccnt.RootHash())
}

func TestBaseAccountSaveToDbAccountInvalidDataShouldErr(t *testing.T) {
	t.Parallel()

	baseAccnt := state.NewAccount()

	//nil target
	err := baseAccnt.SaveToDbAccount(nil)
	assert.NotNil(t, err)
}

func TestBaseAccountSaveToDbAccountValidDataShouldWork(t *testing.T) {
	t.Parallel()

	baseAccnt := state.NewAccount()
	baseAccnt.SetNonce(2)
	baseAccnt.SetBalance(*big.NewInt(40))
	baseAccnt.SetCodeHash([]byte("aaaa"))
	baseAccnt.SetRootHash([]byte("bbbb"))

	dbAccnt := state.NewDbAccount()
	err := baseAccnt.SaveToDbAccount(dbAccnt)
	assert.Nil(t, err)

	buff, err := big.NewInt(40).GobEncode()
	assert.Nil(t, err)

	assert.Equal(t, baseAccnt.Nonce(), dbAccnt.Nonce())
	assert.Equal(t, dbAccnt.Balance(), buff)
	assert.Equal(t, baseAccnt.CodeHash(), dbAccnt.CodeHash())
	assert.Equal(t, baseAccnt.RootHash(), dbAccnt.RootHash())
}
