package state_test

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

func accountStateCreateRandomAddress() *state.Address {
	buff := make([]byte, state.AdrLen)
	rand.Read(buff)

	addr, err := state.NewAddress(buff)
	if err != nil {
		panic(err)
	}

	return addr
}

func accountStateCreateAccountsDB() *state.AccountsDB {
	marsh := mock.MarshalizerMock{}
	adb := state.NewAccountsDB(mock.NewMockTrie(), mock.HasherMock{}, &marsh)

	return adb
}

func TestAccountStateSetNonceNilAccountsShouldRetErr(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()

	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetNonce(nil, 0)

	assert.NotNil(t, err)
}

func TestAccountStateSetNonceWithValsShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()
	adb := accountStateCreateAccountsDB()

	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetNonce(adb, 5)
	assert.Nil(t, err)
	assert.Equal(t, uint64(5), as.Nonce())
	assert.Equal(t, 1, adb.JournalLen())

}

func TestAccountStateSetBalanceNilValuesShouldRetErr(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()

	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetBalance(nil, big.NewInt(0))
	assert.NotNil(t, err)

	err = as.SetBalance(accountStateCreateAccountsDB(), nil)
	assert.NotNil(t, err)
}

func TestAccountStateSetBalanceWithValsShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()
	adb := accountStateCreateAccountsDB()

	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetBalance(adb, big.NewInt(65))
	assert.Nil(t, err)
	assert.Equal(t, *big.NewInt(65), as.Balance())
	assert.Equal(t, 1, adb.JournalLen())

}

func TestAccountStateSetCodeHashNilAccountsShouldRetErr(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()

	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetCodeHash(nil, make([]byte, 0))
	assert.NotNil(t, err)
}

func TestAccountStateSetCodeHashWithValsShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()
	adb := accountStateCreateAccountsDB()

	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetCodeHash(adb, []byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{65, 66, 67}, as.CodeHash())
	assert.Equal(t, 1, adb.JournalLen())

}

func TestAccountStateSetRootHashNilAccountsShouldRetErr(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()

	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetRootHash(nil, make([]byte, 0))
	assert.NotNil(t, err)
}

func TestAccountStateSetRootHashWithValsShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()
	adb := accountStateCreateAccountsDB()

	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetRootHash(adb, []byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{65, 66, 67}, as.RootHash())
	assert.Equal(t, 1, adb.JournalLen())

}

func TestAccountStateRetrieveValueNilDataTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()

	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})

	_, err := as.RetrieveValue([]byte{65, 66, 67})
	assert.NotNil(t, err)
}

func TestAccountStateRetrieveValueFoundInDirtyShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()
	adb := accountStateCreateAccountsDB()

	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})
	as.SetDataTrie(adb.MainTrie)
	as.DirtyData()["ABC"] = []byte{32, 33, 34}

	val, err := as.RetrieveValue([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{32, 33, 34}, val)
}

func TestAccountStateRetrieveValueFoundInOriginalShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()
	adb := accountStateCreateAccountsDB()

	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})
	as.SetDataTrie(adb.MainTrie)
	as.DirtyData()["ABC"] = []byte{32, 33, 34}
	as.OriginalData()["ABD"] = []byte{35, 36, 37}

	val, err := as.RetrieveValue([]byte{65, 66, 68})
	assert.Nil(t, err)
	assert.Equal(t, []byte{35, 36, 37}, val)
}

func TestAccountStateRetrieveValueFoundInTrieShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()
	adb := accountStateCreateAccountsDB()

	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})
	as.SetDataTrie(adb.MainTrie)
	err := as.DataTrie().Update([]byte{65, 66, 69}, []byte{38, 39, 40})
	assert.Nil(t, err)
	as.DirtyData()["ABC"] = []byte{32, 33, 34}
	as.OriginalData()["ABD"] = []byte{35, 36, 37}

	val, err := as.RetrieveValue([]byte{65, 66, 69})
	assert.Nil(t, err)
	assert.Equal(t, []byte{38, 39, 40}, val)
}

func TestAccountStateRetrieveValueMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()
	adb := accountStateCreateAccountsDB()
	adb.MainTrie.(*mock.TrieMock).FailGet = true

	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})
	as.SetDataTrie(adb.MainTrie)
	err := as.DataTrie().Update([]byte{65, 66, 69}, []byte{38, 39, 40})
	assert.Nil(t, err)
	as.DirtyData()["ABC"] = []byte{32, 33, 34}
	as.OriginalData()["ABD"] = []byte{35, 36, 37}

	val, err := as.RetrieveValue([]byte{65, 66, 69})
	assert.NotNil(t, err)
	assert.Nil(t, val)
}

func TestAccountStateSaveKeyValueShouldSaveOnlyInDirty(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()
	adb := accountStateCreateAccountsDB()

	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})
	as.SetDataTrie(adb.MainTrie)
	as.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})

	//test in dirty
	assert.Equal(t, []byte{32, 33, 34}, as.DirtyData()["ABC"])
	//test in original
	assert.Nil(t, as.OriginalData()["ABC"])
	//test in trie
	val, err := as.DataTrie().Get([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestAccountStateCodeSetGetShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()
	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})

	assert.Nil(t, as.Code())
	as.SetCode([]byte{65, 66, 67})
	assert.Equal(t, []byte{65, 66, 67}, as.Code())
}

func TestAccountStateClearDataCachesValidDataShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := accountStateCreateRandomAddress()
	as := state.NewAccountState(adr1, state.NewAccount(), mock.HasherMock{})

	assert.Equal(t, 0, len(as.DirtyData()))

	//add something
	as.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	assert.Equal(t, 1, len(as.DirtyData()))

	//clear
	as.ClearDataCaches()
	assert.Equal(t, 0, len(as.DirtyData()))
}
