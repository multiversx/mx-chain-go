package state_test

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

func aCreateRandomAddress() *state.Address {
	buff := make([]byte, state.AdrLen)
	rand.Read(buff)

	addr, err := state.NewAddress(buff)
	if err != nil {
		panic(err)
	}

	return addr
}

func aCreateAccountsDB() *state.AccountsDB {
	marsh := mock.MarshalizerMock{}
	adb := state.NewAccountsDB(mock.NewMockTrie(), mock.HasherMock{}, &marsh)

	return adb
}

func TestAccountState_SetNonce_NilAccounts_ShouldRetErr(t *testing.T) {
	t.Parallel()

	adr1 := aCreateRandomAddress()

	as := state.NewAccountState(*adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetNonce(nil, 0)

	assert.NotNil(t, err)
}

func TestAccountState_SetNonce_WithVals_ShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := aCreateRandomAddress()
	adb := aCreateAccountsDB()

	as := state.NewAccountState(*adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetNonce(adb, 5)
	assert.Nil(t, err)
	assert.Equal(t, uint64(5), as.Nonce())
	assert.Equal(t, 1, adb.Journal().Len())

}

func TestAccountState_SetBalance_NilValues_ShouldRetErr(t *testing.T) {
	t.Parallel()

	adr1 := aCreateRandomAddress()

	as := state.NewAccountState(*adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetBalance(nil, big.NewInt(0))
	assert.NotNil(t, err)

	err = as.SetBalance(aCreateAccountsDB(), nil)
	assert.NotNil(t, err)
}

func TestAccountState_SetBalance_WithVals_ShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := aCreateRandomAddress()
	adb := aCreateAccountsDB()

	as := state.NewAccountState(*adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetBalance(adb, big.NewInt(65))
	assert.Nil(t, err)
	assert.Equal(t, *big.NewInt(65), as.Balance())
	assert.Equal(t, 1, adb.Journal().Len())

}

func TestAccountState_SetCodeHash_NilAccounts_ShouldRetErr(t *testing.T) {
	t.Parallel()

	adr1 := aCreateRandomAddress()

	as := state.NewAccountState(*adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetCodeHash(nil, make([]byte, 0))
	assert.NotNil(t, err)
}

func TestAccountState_SetCodeHash_WithVals_ShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := aCreateRandomAddress()
	adb := aCreateAccountsDB()

	as := state.NewAccountState(*adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetCodeHash(adb, []byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{65, 66, 67}, as.CodeHash())
	assert.Equal(t, 1, adb.Journal().Len())

}

func TestAccountState_SetRoot_NilAccounts_ShouldRetErr(t *testing.T) {
	t.Parallel()

	adr1 := aCreateRandomAddress()

	as := state.NewAccountState(*adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetRoot(nil, make([]byte, 0))
	assert.NotNil(t, err)
}

func TestAccountState_SetRoot_WithVals_ShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := aCreateRandomAddress()
	adb := aCreateAccountsDB()

	as := state.NewAccountState(*adr1, state.NewAccount(), mock.HasherMock{})

	err := as.SetRoot(adb, []byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{65, 66, 67}, as.Root())
	assert.Equal(t, 1, adb.Journal().Len())

}

func TestAccountState_RetrieveValue_NilDataTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := aCreateRandomAddress()

	as := state.NewAccountState(*adr1, state.NewAccount(), mock.HasherMock{})

	_, err := as.RetrieveValue([]byte{65, 66, 67})
	assert.NotNil(t, err)
}

func TestAccountState_RetrieveValue_FoundInDirty_ShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := aCreateRandomAddress()
	adb := aCreateAccountsDB()

	as := state.NewAccountState(*adr1, state.NewAccount(), mock.HasherMock{})
	as.DataTrie = adb.MainTrie
	as.DirtyData()["ABC"] = []byte{32, 33, 34}

	val, err := as.RetrieveValue([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{32, 33, 34}, val)
}

func TestAccountState_RetrieveValue_FoundInOriginal_ShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := aCreateRandomAddress()
	adb := aCreateAccountsDB()

	as := state.NewAccountState(*adr1, state.NewAccount(), mock.HasherMock{})
	as.DataTrie = adb.MainTrie
	as.DirtyData()["ABC"] = []byte{32, 33, 34}
	as.OriginalData()["ABD"] = []byte{35, 36, 37}

	val, err := as.RetrieveValue([]byte{65, 66, 68})
	assert.Nil(t, err)
	assert.Equal(t, []byte{35, 36, 37}, val)
}

func TestAccountState_RetrieveValue_FoundInTrie_ShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := aCreateRandomAddress()
	adb := aCreateAccountsDB()

	as := state.NewAccountState(*adr1, state.NewAccount(), mock.HasherMock{})
	as.DataTrie = adb.MainTrie
	as.DataTrie.Update([]byte{65, 66, 69}, []byte{38, 39, 40})
	as.DirtyData()["ABC"] = []byte{32, 33, 34}
	as.OriginalData()["ABD"] = []byte{35, 36, 37}

	val, err := as.RetrieveValue([]byte{65, 66, 69})
	assert.Nil(t, err)
	assert.Equal(t, []byte{38, 39, 40}, val)
}

func TestAccountState_SaveKeyValue_ShouldSaveOnlyInDirty(t *testing.T) {
	t.Parallel()

	adr1 := aCreateRandomAddress()
	adb := aCreateAccountsDB()

	as := state.NewAccountState(*adr1, state.NewAccount(), mock.HasherMock{})
	as.DataTrie = adb.MainTrie
	as.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})

	//test in dirty
	assert.Equal(t, []byte{32, 33, 34}, as.DirtyData()["ABC"])
	//test in original
	assert.Nil(t, as.OriginalData()["ABC"])
	//test in trie
	val, err := as.DataTrie.Get([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestAccountState_CollapseDirty_ValidVals_ShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := aCreateRandomAddress()
	adb := aCreateAccountsDB()

	as := state.NewAccountState(*adr1, state.NewAccount(), mock.HasherMock{})
	as.DataTrie = adb.MainTrie
	//this one wil have its value dirtied
	as.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	//this one will have the same value as in original
	as.SaveKeyValue([]byte{65, 66, 68}, []byte{35, 36, 37})

	as.OriginalData()["ABC"] = make([]byte, 0)
	as.OriginalData()["ABD"] = []byte{35, 36, 37}

	as.CollapseDirty()

	for k, v := range as.DirtyData() {
		//test not to be ABD
		assert.NotEqual(t, k, []byte{65, 66, 68})
		//test that value is not []byte{35, 36, 37}
		assert.NotEqual(t, v, []byte{35, 36, 37})
	}
}
