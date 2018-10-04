package state

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testHasher = mock.HasherMock{}
var testMarshalizer = mock.MarshalizerMock{}

func createAccountsDB(_ *testing.T) *AccountsDB {
	adb := NewAccountsDB(mock.NewMockTrie(), &testHasher, &testMarshalizer)

	return adb
}

func createAddress(t *testing.T, buff []byte) *Address {
	adr, err := FromPubKeyBytes(buff)
	assert.Nil(t, err)

	return adr
}

func createAccountState(_ *testing.T, address Address) *AccountState {
	return NewAccountState(address, Account{}, testHasher)
}

func TestAccountsDB_RetrieveCode_NilCodeHash_ShouldRetNil(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)

	adb := createAccountsDB(t)

	//since code hash is nil, result should be nil and code should be nil
	err := adb.RetrieveCode(state)
	assert.Nil(t, err)
	assert.Nil(t, state.Code)
}

func TestAccountsDB_RetrieveCode_NilTrie_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.CodeHash = []byte("12345")

	adb := createAccountsDB(t)

	adb.mainTrie = nil

	err := adb.RetrieveCode(state)
	assert.NotNil(t, err)
}

func TestAccountsDB_RetrieveCode_BadLength_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.CodeHash = []byte("12345")

	adb := createAccountsDB(t)

	//should return error
	err := adb.RetrieveCode(state)
	assert.NotNil(t, err)
}

func TestAccountsDB_RetrieveCode_MalfunctionTrie_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.CodeHash = testHasher.Compute("12345")

	adb := createAccountsDB(t)
	adb.mainTrie.(*mock.TrieMock).Fail = true

	//should return error
	err := adb.RetrieveCode(state)
	assert.NotNil(t, err)
}

func TestAccountsDB_RetrieveCode_NotFound_ShouldRetNil(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.CodeHash = testHasher.Compute("12345")

	adb := createAccountsDB(t)

	//should return nil
	err := adb.RetrieveCode(state)
	assert.Nil(t, err)
	assert.Nil(t, state.Code)
}

func TestAccountsDB_RetrieveCode_Found_ShouldWork(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.CodeHash = testHasher.Compute("12345")

	adb := createAccountsDB(t)
	//manually put it in the trie
	adb.mainTrie.Update(state.CodeHash, []byte{68, 69, 70})

	//should return nil
	err := adb.RetrieveCode(state)
	assert.Nil(t, err)
	assert.Equal(t, state.Code, []byte{68, 69, 70})
}

func TestAccountsDB_RetrieveData_NilRoot_ShouldRetNil(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)

	adb := createAccountsDB(t)

	//since root is nil, result should be nil and data trie should be nil
	err := adb.RetrieveData(state)
	assert.Nil(t, err)
	assert.Nil(t, state.Data)
}

func TestAccountsDB_RetrieveData_NilTrie_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = []byte("12345")

	adb := createAccountsDB(t)

	adb.mainTrie = nil

	err := adb.RetrieveData(state)
	assert.NotNil(t, err)
}

func TestAccountsDB_RetrieveData_BadLength_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = []byte("12345")

	adb := createAccountsDB(t)

	//should return error
	err := adb.RetrieveData(state)
	assert.NotNil(t, err)
}

func TestAccountsDB_RetrieveData_MalfunctionTrie_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)
	adb.mainTrie.(*mock.TrieMock).Fail = true

	//should return error
	err := adb.RetrieveData(state)
	assert.NotNil(t, err)
}

func TestAccountsDB_RetrieveData_NotFoundRoot_ShouldReturnEmpty(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)

	//should return error
	err := adb.RetrieveData(state)
	assert.Nil(t, err)
	assert.Equal(t, make([]byte, encoding.HashLength), state.Data.Root())
}

func TestAccountsDB_RetrieveData_WithSomeValues_ShouldWork(t *testing.T) {
	//test simulates creation of a new account, data trie retrieval,
	//adding a (key, value) pair in that data trie, commiting changes
	//and then reloading the data trie based on the root hash generated before
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)

	//should not return error
	err := adb.RetrieveData(state)
	assert.Nil(t, err)
	assert.NotNil(t, state.Data)

	fmt.Printf("Root: %v\n", state.Root)

	//add something to the data trie
	state.Data.Update([]byte{65, 66, 67}, []byte{68, 69, 70})
	state.Root = state.Data.Root()
	fmt.Printf("state.Root is %v\n", state.Root)
	hashResult, err := state.Data.Commit(nil)

	assert.Equal(t, hashResult, state.Root)

	//make data trie null and then retrieve the trie and search the data
	state.Data = nil
	err = adb.RetrieveData(state)
	assert.Nil(t, err)
	assert.NotNil(t, state.Data)

	val, err := state.Data.Get([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{68, 69, 70}, val)
}

func TestAccountsDB_PutCode_NilCodeHash_ShouldRetNil(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = state.AddrHash

	adb := createAccountsDB(t)

	err := adb.PutCode(state, nil)
	assert.Nil(t, err)
	assert.Nil(t, state.CodeHash)
	assert.Nil(t, state.Code)
}

func TestAccountsDB_PutCode_EmptyCodeHash_ShouldRetNil(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = state.AddrHash

	adb := createAccountsDB(t)

	err := adb.PutCode(state, make([]byte, 0))
	assert.Nil(t, err)
	assert.Nil(t, state.CodeHash)
	assert.Nil(t, state.Code)
}

func TestAccountsDB_PutCode_NilTrie_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = []byte("12345")

	adb := createAccountsDB(t)

	adb.mainTrie = nil

	err := adb.PutCode(state, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDB_PutCode_MalfunctionTrie_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)
	adb.mainTrie.(*mock.TrieMock).Fail = true

	//should return error
	err := adb.PutCode(state, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDB_PutCode_Malfunction2Trie_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)
	adb.mainTrie.(*mock.TrieMock).FailRecreate = true

	//should return error
	err := adb.PutCode(state, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDB_PutCode_WithSomeValues_ShouldWork(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)

	snapshotRoot := adb.mainTrie.Root()

	err := adb.PutCode(state, []byte("Smart contract code"))
	assert.Nil(t, err)
	assert.NotNil(t, state.CodeHash)
	assert.Equal(t, []byte("Smart contract code"), state.Code)

	fmt.Printf("SC code is at address: %v\n", state.CodeHash)

	//retrieve directly from the trie
	data, err := adb.mainTrie.Get(state.CodeHash)
	assert.Nil(t, err)
	assert.Equal(t, data, state.Code)

	fmt.Printf("SC code is: %v\n", string(data))

	//main root trie should have been modified
	assert.NotEqual(t, snapshotRoot, adb.mainTrie.Root())
}

func TestAccountsDB_HasAccount_NilTrie_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = []byte("12345")

	adb := createAccountsDB(t)
	adb.mainTrie = nil

	val, err := adb.HasAccount(*adr)
	assert.NotNil(t, err)
	assert.False(t, val)
}

func TestAccountsDB_HasAccount_MalfunctionTrie_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)
	adb.mainTrie.(*mock.TrieMock).Fail = true

	//should return error
	val, err := adb.HasAccount(*adr)
	assert.NotNil(t, err)
	assert.False(t, val)
}

func TestAccountsDB_HasAccount_NotFound_ShouldRetFalse(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)

	//should return false
	val, err := adb.HasAccount(*adr)
	assert.Nil(t, err)
	assert.False(t, val)
}

func TestAccountsDB_HasAccount_Found_ShouldRetTrue(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)
	adb.mainTrie.Update(testHasher.Compute(string(adr.Bytes())), []byte{65})

	//should return true
	val, err := adb.HasAccount(*adr)
	assert.Nil(t, err)
	assert.True(t, val)
}

func TestAccountsDB_SaveAccountState_NilTrie_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = []byte("12345")

	adb := createAccountsDB(t)
	adb.mainTrie = nil

	err := adb.SaveAccountState(state)
	assert.NotNil(t, err)
}

func TestAccountsDB_SaveAccountState_NilState_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = []byte("12345")

	adb := createAccountsDB(t)
	adb.mainTrie = nil

	err := adb.SaveAccountState(nil)
	assert.NotNil(t, err)
}

func TestAccountsDB_SaveAccountState_MalfunctionTrie_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)
	adb.mainTrie.(*mock.TrieMock).Fail = true

	//should return error
	err := adb.SaveAccountState(state)
	assert.NotNil(t, err)
}

func TestAccountsDB_SaveAccountState_MalfunctionMarshalizer_ShouldErr(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)
	adb.marsh.(*mock.MarshalizerMock).Fail = true
	defer func() {
		//call defer to reset the fail pointer so other tests will work as expected
		adb.marsh.(*mock.MarshalizerMock).Fail = false
	}()

	//should return error
	err := adb.SaveAccountState(state)

	assert.NotNil(t, err)
}

func TestAccountsDB_SaveAccountState_WithSomeValues_ShouldWork(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHJIJKLMNOP")

	adr := createAddress(t, testHash)
	state := createAccountState(t, *adr)
	state.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)

	//should return error
	err := adb.SaveAccountState(state)
	assert.Nil(t, err)
}

//func TestPutCode(t *testing.T) {
//	testHash := DefHasher.Compute("ABCDEFGHJIJKLMNOP")
//
//	adr := createAddress(t, testHash)
//	state := createAccountState(t, *adr)
//	state.Root = state.AddrHash
//
//	tr := createMemTrie(t)
//
//}
