package state_test

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

func adbCreateAccountsDB() *state.AccountsDB {
	marsh := mock.MarshalizerMock{}
	adb := state.NewAccountsDB(mock.NewMockTrie(), mock.HasherMock{}, &marsh)

	return adb
}

func adbCreateAddress(t *testing.T, buff []byte) *state.Address {
	adr, err := state.FromPubKeyBytes(buff)
	assert.Nil(t, err)

	return adr
}

func adbCreateAccountState(_ *testing.T, address state.Address) *state.AccountState {
	return state.NewAccountState(address, state.Account{}, mock.HasherMock{})
}

func TestAccountsDB_RetrieveCode_NilCodeHash_ShouldRetNil(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()

	//since code hash is nil, result should be nil and code should be nil
	err := adb.RetrieveCode(as)
	assert.Nil(t, err)
	assert.Nil(t, as.Code)
}

func TestAccountsDB_RetrieveCode_NilTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)
	as.Account().CodeHash = []byte("12345")

	adb := adbCreateAccountsDB()

	adb.MainTrie = nil

	err := adb.RetrieveCode(as)
	assert.NotNil(t, err)
}

func TestAccountsDB_RetrieveCode_BadLength_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)
	as.Account().CodeHash = []byte("12345")

	adb := adbCreateAccountsDB()

	//should return error
	err := adb.RetrieveCode(as)
	assert.NotNil(t, err)
}

func TestAccountsDB_RetrieveCode_MalfunctionTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)
	as.Account().CodeHash = hasher.Compute("12345")

	adb := adbCreateAccountsDB()
	adb.MainTrie.(*mock.TrieMock).Fail = true

	//should return error
	err := adb.RetrieveCode(as)
	assert.NotNil(t, err)
}

func TestAccountsDB_RetrieveCode_NotFound_ShouldRetNil(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)
	as.Account().CodeHash = hasher.Compute("12345")

	adb := adbCreateAccountsDB()

	//should return nil
	err := adb.RetrieveCode(as)
	assert.Nil(t, err)
	assert.Nil(t, as.Code)
}

func TestAccountsDB_RetrieveCode_Found_ShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)
	as.Account().CodeHash = hasher.Compute("12345")

	adb := adbCreateAccountsDB()
	//manually put it in the trie
	adb.MainTrie.Update(as.CodeHash(), []byte{68, 69, 70})

	//should return nil
	err := adb.RetrieveCode(as)
	assert.Nil(t, err)
	assert.Equal(t, as.Code, []byte{68, 69, 70})
}

//func TestAccountsDB_RetrieveData_NilRoot_ShouldRetNil(t *testing.T) {
//	t.Parallel()
//
//	hasher := mock.HasherMock{}
//
//	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")
//
//	adr := adbCreateAddress(t, testHash)
//	state := adbCreateAccountState(t, *adr)
//
//	adb := adbCreateAccountsDB()
//
//	//since root is nil, result should be nil and data trie should be nil
//	err := adb.RetrieveData(state)
//	assert.Nil(t, err)
//	assert.Nil(t, state.Data)
//}
//
//func TestAccountsDB_RetrieveData_NilTrie_ShouldErr(t *testing.T) {
//	testHash := testHasher.Compute("ABCDEFGHIJKLMNOP")
//
//	adr := createAddress(t, testHash)
//	state := createAccountState(t, *adr)
//	state.Root = []byte("12345")
//
//	adb := createAccountsDB()
//
//	adb.MainTrie = nil
//
//	err := adb.RetrieveData(state)
//	assert.NotNil(t, err)
//}
//
//func TestAccountsDB_RetrieveData_BadLength_ShouldErr(t *testing.T) {
//	testHash := testHasher.Compute("ABCDEFGHIJKLMNOP")
//
//	adr := createAddress(t, testHash)
//	state := createAccountState(t, *adr)
//	state.Root = []byte("12345")
//
//	adb := createAccountsDB()
//
//	//should return error
//	err := adb.RetrieveData(state)
//	assert.NotNil(t, err)
//}
//
//func TestAccountsDB_RetrieveData_MalfunctionTrie_ShouldErr(t *testing.T) {
//	testHash := testHasher.Compute("ABCDEFGHIJKLMNOP")
//
//	adr := createAddress(t, testHash)
//	state := createAccountState(t, *adr)
//	state.Root = testHasher.Compute("12345")
//
//	adb := createAccountsDB()
//	adb.MainTrie.(*mock.TrieMock).Fail = true
//
//	//should return error
//	err := adb.RetrieveData(state)
//	assert.NotNil(t, err)
//}
//
//func TestAccountsDB_RetrieveData_NotFoundRoot_ShouldReturnEmpty(t *testing.T) {
//	testHash := testHasher.Compute("ABCDEFGHIJKLMNOP")
//
//	adr := createAddress(t, testHash)
//	state := createAccountState(t, *adr)
//	state.Root = testHasher.Compute("12345")
//
//	adb := createAccountsDB()
//
//	//should return error
//	err := adb.RetrieveData(state)
//	assert.Nil(t, err)
//	assert.Equal(t, make([]byte, encoding.HashLength), state.Data.Root())
//}
//
//func TestAccountsDB_RetrieveData_WithSomeValues_ShouldWork(t *testing.T) {
//	//test simulates creation of a new account, data trie retrieval,
//	//adding a (key, value) pair in that data trie, commiting changes
//	//and then reloading the data trie based on the root hash generated before
//	testHash := testHasher.Compute("ABCDEFGHIJKLMNOP")
//
//	adr := createAddress(t, testHash)
//	state := createAccountState(t, *adr)
//	state.Root = testHasher.Compute("12345")
//
//	adb := createAccountsDB()
//
//	//should not return error
//	err := adb.RetrieveData(state)
//	assert.Nil(t, err)
//	assert.NotNil(t, state.Data)
//
//	fmt.Printf("Root: %v\n", state.Root)
//
//	//add something to the data trie
//	state.Data.Update([]byte{65, 66, 67}, []byte{68, 69, 70})
//	state.Root = state.Data.Root()
//	fmt.Printf("state.Root is %v\n", state.Root)
//	hashResult, err := state.Data.Commit(nil)
//
//	assert.Equal(t, hashResult, state.Root)
//
//	//make data trie null and then retrieve the trie and search the data
//	state.Data = nil
//	err = adb.RetrieveData(state)
//	assert.Nil(t, err)
//	assert.NotNil(t, state.Data)
//
//	val, err := state.Data.Get([]byte{65, 66, 67})
//	assert.Nil(t, err)
//	assert.Equal(t, []byte{68, 69, 70}, val)
//}

func TestAccountsDB_PutCode_NilCodeHash_ShouldRetNil(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()

	err := adb.PutCode(as, nil)
	assert.Nil(t, err)
	assert.Nil(t, as.CodeHash())
	assert.Nil(t, as.Code)
}

func TestAccountsDB_PutCode_EmptyCodeHash_ShouldRetNil(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()

	err := adb.PutCode(as, make([]byte, 0))
	assert.Nil(t, err)
	assert.Nil(t, as.CodeHash())
	assert.Nil(t, as.Code)
}

func TestAccountsDB_PutCode_NilTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()

	adb.MainTrie = nil

	err := adb.PutCode(as, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDB_PutCode_MalfunctionTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()
	adb.MainTrie.(*mock.TrieMock).Fail = true

	//should return error
	err := adb.PutCode(as, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDB_PutCode_Malfunction2Trie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()
	adb.MainTrie.(*mock.TrieMock).FailUpdate = true

	//should return error
	err := adb.PutCode(as, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDB_PutCode_WithSomeValues_ShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()

	snapshotRoot := adb.MainTrie.Root()

	err := adb.PutCode(as, []byte("Smart contract code"))
	assert.Nil(t, err)
	assert.NotNil(t, as.CodeHash)
	assert.Equal(t, []byte("Smart contract code"), as.Code)

	fmt.Printf("SC code is at address: %v\n", as.CodeHash())

	//retrieve directly from the trie
	data, err := adb.MainTrie.Get(as.CodeHash())
	assert.Nil(t, err)
	assert.Equal(t, data, as.Code)

	fmt.Printf("SC code is: %v\n", string(data))

	//main root trie should have been modified
	assert.NotEqual(t, snapshotRoot, adb.MainTrie.Root())
}

func TestAccountsDB_HasAccount_NilTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)

	adb := adbCreateAccountsDB()
	adb.MainTrie = nil

	val, err := adb.HasAccount(*adr)
	assert.NotNil(t, err)
	assert.False(t, val)
}

func TestAccountsDB_HasAccount_MalfunctionTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)

	adb := adbCreateAccountsDB()
	adb.MainTrie.(*mock.TrieMock).Fail = true

	//should return error
	val, err := adb.HasAccount(*adr)
	assert.NotNil(t, err)
	assert.False(t, val)
}

func TestAccountsDB_HasAccount_NotFound_ShouldRetFalse(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)

	adb := adbCreateAccountsDB()

	//should return false
	val, err := adb.HasAccount(*adr)
	assert.Nil(t, err)
	assert.False(t, val)
}

func TestAccountsDB_HasAccount_Found_ShouldRetTrue(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)

	adb := adbCreateAccountsDB()
	adb.MainTrie.Update(hasher.Compute(string(adr.Bytes())), []byte{65})

	//should return true
	val, err := adb.HasAccount(*adr)
	assert.Nil(t, err)
	assert.True(t, val)
}

func TestAccountsDB_SaveAccountState_NilTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()
	adb.MainTrie = nil

	err := adb.SaveAccountState(as)
	assert.NotNil(t, err)
}

func TestAccountsDB_SaveAccountState_NilState_ShouldErr(t *testing.T) {
	t.Parallel()

	adb := adbCreateAccountsDB()
	adb.MainTrie = nil

	err := adb.SaveAccountState(nil)
	assert.NotNil(t, err)
}

func TestAccountsDB_SaveAccountState_MalfunctionTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()
	adb.MainTrie.(*mock.TrieMock).Fail = true

	//should return error
	err := adb.SaveAccountState(as)
	assert.NotNil(t, err)
}

func TestAccountsDB_SaveAccountState_MalfunctionMarshalizer_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()
	adb.Marsh().(*mock.MarshalizerMock).Fail = true
	defer func() {
		//call defer to reset the fail pointer so other tests will work as expected
		adb.Marsh().(*mock.MarshalizerMock).Fail = false
	}()

	//should return error
	err := adb.SaveAccountState(as)

	assert.NotNil(t, err)
}

func TestAccountsDB_SaveAccountState_WithSomeValues_ShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()

	//should return error
	err := adb.SaveAccountState(as)
	assert.Nil(t, err)
}

func TestAccountsDB_GetOrCreateAccount_NilTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)

	adb := adbCreateAccountsDB()
	adb.MainTrie = nil

	acnt, err := adb.GetOrCreateAccount(*adr)
	assert.NotNil(t, err)
	assert.Nil(t, acnt)

}

func TestAccountsDB_GetOrCreateAccount_MalfunctionTrie_ShouldErr(t *testing.T) {
	//test when there is no such account

	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)

	adb := adbCreateAccountsDB()
	adb.MainTrie.(*mock.TrieMock).FailGet = true
	adb.MainTrie.(*mock.TrieMock).TTFGet = 1
	adb.MainTrie.(*mock.TrieMock).FailUpdate = true
	adb.MainTrie.(*mock.TrieMock).TTFUpdate = 0

	acnt, err := adb.GetOrCreateAccount(*adr)
	assert.NotNil(t, err)
	assert.Nil(t, acnt)

}

func TestAccountsDB_GetOrCreateAccount_Malfunction2Trie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()
	adb.SaveAccountState(as)

	adb.MainTrie.(*mock.TrieMock).FailGet = true
	adb.MainTrie.(*mock.TrieMock).TTFGet = 1

	acnt, err := adb.GetOrCreateAccount(*adr)
	assert.NotNil(t, err)
	assert.Nil(t, acnt)

}

func TestAccountsDB_GetOrCreateAccount_MalfunctionMarshalizer_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()
	adb.SaveAccountState(as)

	adb.Marsh().(*mock.MarshalizerMock).Fail = true
	defer func() {
		adb.Marsh().(*mock.MarshalizerMock).Fail = false
	}()

	acnt, err := adb.GetOrCreateAccount(*adr)
	assert.NotNil(t, err)
	assert.Nil(t, acnt)

}

func TestAccountsDB_GetOrCreateAccount_ReturnExistingAccnt_ShouldWork(t *testing.T) {
	//test when the account exists
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adb := adbCreateAccountsDB()

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)
	as.SetBalance(adb, big.NewInt(40))

	adb.SaveAccountState(as)

	acnt, err := adb.GetOrCreateAccount(*adr)
	assert.Nil(t, err)
	assert.NotNil(t, acnt)
	assert.Equal(t, acnt.Balance(), *big.NewInt(40))
}

func TestAccountsDB_GetOrCreateAccount_ReturnNotFoundAccnt_ShouldWork(t *testing.T) {
	//test when the account does not exists
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adb := adbCreateAccountsDB()

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)
	as.SetBalance(adb, big.NewInt(40))

	//same address of the unsaved account
	acnt, err := adb.GetOrCreateAccount(*adr)
	assert.Nil(t, err)
	assert.NotNil(t, acnt)
	assert.Equal(t, acnt.Balance(), *big.NewInt(0))
}

//func TestAccountsDB_Commit_MalfunctionTrie_ShouldErr(t *testing.T) {
//	//test creates 2 accounts (one with a data root)
//	//verifies that commit saves the new tries and that can be loaded back
//	//second data trie malfunctions
//
//	testHash1 := testHasher.Compute("ABCDEFGHIJKLMNOP")
//	adr1 := createAddress(t, testHash1)
//
//	testHash2 := testHasher.Compute("ABCDEFGHIJKLMNOPQ")
//	adr2 := createAddress(t, testHash2)
//
//	adb := createAccountsDB()
//
//	//first account has the balance of 40
//	state1, err := adb.GetOrCreateAccount(*adr1)
//	assert.Nil(t, err)
//	state1.Balance = big.NewInt(40)
//
//	//second account has the balance of 50 and something in data root
//	state2, err := adb.GetOrCreateAccount(*adr2)
//	assert.Nil(t, err)
//
//	state2.Balance = big.NewInt(50)
//	state2.Root = make([]byte, encoding.HashLength)
//	err = adb.RetrieveData(state2)
//	assert.Nil(t, err)
//
//	state2.Data.Update([]byte{65, 66, 67}, []byte{68, 69, 70})
//
//	//states are now prepared, committing
//
//	state2.Data.(*mock.TrieMock).Fail = true
//
//	_, err = adb.Commit()
//	assert.NotNil(t, err)
//}
//
//func TestAccountsDB_Commit_MalfunctionMainTrie_ShouldErr(t *testing.T) {
//	//test creates 2 accounts (one with a data root)
//	//verifies that commit saves the new tries and that can be loaded back
//	//main trie malfunctions
//
//	testHash1 := testHasher.Compute("ABCDEFGHIJKLMNOP")
//	adr1 := createAddress(t, testHash1)
//
//	testHash2 := testHasher.Compute("ABCDEFGHIJKLMNOPQ")
//	adr2 := createAddress(t, testHash2)
//
//	adb := createAccountsDB()
//
//	//first account has the balance of 40
//	state1, err := adb.GetOrCreateAccount(*adr1)
//	assert.Nil(t, err)
//	state1.Balance = big.NewInt(40)
//
//	//second account has the balance of 50 and something in data root
//	state2, err := adb.GetOrCreateAccount(*adr2)
//	assert.Nil(t, err)
//
//	state2.Balance = big.NewInt(50)
//	state2.Root = make([]byte, encoding.HashLength)
//	err = adb.RetrieveData(state2)
//	assert.Nil(t, err)
//
//	state2.Data.Update([]byte{65, 66, 67}, []byte{68, 69, 70})
//
//	//states are now prepared, committing
//
//	adb.MainTrie.(*mock.TrieMock).Fail = true
//
//	_, err = adb.Commit()
//	assert.NotNil(t, err)
//}
//
//func TestAccountsDB_Commit_MalfunctionMarshalizer_ShouldErr(t *testing.T) {
//	//test creates 2 accounts (one with a data root)
//	//verifies that commit saves the new tries and that can be loaded back
//	//marshalizer fails
//
//	testHash1 := testHasher.Compute("ABCDEFGHIJKLMNOP")
//	adr1 := createAddress(t, testHash1)
//
//	testHash2 := testHasher.Compute("ABCDEFGHIJKLMNOPQ")
//	adr2 := createAddress(t, testHash2)
//
//	adb := createAccountsDB()
//
//	//first account has the balance of 40
//	state1, err := adb.GetOrCreateAccount(*adr1)
//	assert.Nil(t, err)
//	state1.Balance = big.NewInt(40)
//
//	//second account has the balance of 50 and something in data root
//	state2, err := adb.GetOrCreateAccount(*adr2)
//	assert.Nil(t, err)
//
//	state2.Balance = big.NewInt(50)
//	state2.Root = make([]byte, encoding.HashLength)
//	err = adb.RetrieveData(state2)
//	assert.Nil(t, err)
//
//	state2.Data.Update([]byte{65, 66, 67}, []byte{68, 69, 70})
//
//	//states are now prepared, committing
//
//	adb.marsh.(*mock.MarshalizerMock).Fail = true
//	defer func() {
//		adb.marsh.(*mock.MarshalizerMock).Fail = false
//	}()
//
//	_, err = adb.Commit()
//	assert.NotNil(t, err)
//}
//
//func TestAccountsDB_Commit_2okAccounts_ShouldWork(t *testing.T) {
//	//test creates 2 accounts (one with a data root)
//	//verifies that commit saves the new tries and that can be loaded back
//
//	testHash1 := testHasher.Compute("ABCDEFGHIJKLMNOP")
//	adr1 := createAddress(t, testHash1)
//
//	testHash2 := testHasher.Compute("ABCDEFGHIJKLMNOPQ")
//	adr2 := createAddress(t, testHash2)
//
//	adb := createAccountsDB()
//
//	//first account has the balance of 40
//	state1, err := adb.GetOrCreateAccount(*adr1)
//	assert.Nil(t, err)
//	state1.Balance = big.NewInt(40)
//
//	//second account has the balance of 50 and something in data root
//	state2, err := adb.GetOrCreateAccount(*adr2)
//	assert.Nil(t, err)
//
//	state2.Balance = big.NewInt(50)
//	state2.Root = make([]byte, encoding.HashLength)
//	err = adb.RetrieveData(state2)
//	assert.Nil(t, err)
//
//	state2.Data.Update([]byte{65, 66, 67}, []byte{68, 69, 70})
//
//	//states are now prepared, committing
//
//	h, err := adb.Commit()
//	assert.Nil(t, err)
//	fmt.Printf("Result hash: %v\n", base64.StdEncoding.EncodeToString(h))
//
//	fmt.Printf("Data committed! Root: %v\n", base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))
//
//	//reloading a new trie to test if data is inside
//	newTrie, err := adb.MainTrie.Recreate(adb.MainTrie.Root(), adb.MainTrie.DBW())
//	assert.Nil(t, err)
//	newAdb := NewAccountsDB(newTrie, testHasher, &testMarshalizer)
//
//	//checking state1
//	newState1, err := newAdb.GetOrCreateAccount(*adr1)
//	assert.Nil(t, err)
//	assert.Equal(t, newState1.Balance, big.NewInt(40))
//
//	//checking state2
//	newState2, err := newAdb.GetOrCreateAccount(*adr2)
//	assert.Nil(t, err)
//	assert.Equal(t, newState2.Balance, big.NewInt(50))
//	assert.NotNil(t, newState2.Root)
//	//get data
//	err = adb.RetrieveData(newState2)
//	assert.Nil(t, err)
//	assert.NotNil(t, newState2.Data)
//	val, err := newState2.Data.Get([]byte{65, 66, 67})
//	assert.Nil(t, err)
//	assert.Equal(t, val, []byte{68, 69, 70})
//}
//
//func TestAccountsDB_Commit_AccountData_ShouldWork(t *testing.T) {
//	testHash1 := testHasher.Compute("ABCDEFGHIJKLMNOP")
//	adr1 := createAddress(t, testHash1)
//
//	adb := createAccountsDB()
//	hrEmpty := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
//	fmt.Printf("State root - empty: %v\n", hrEmpty)
//
//	state1, err := adb.GetOrCreateAccount(*adr1)
//	assert.Nil(t, err)
//
//	state1.Balance = big.NewInt(40)
//	hrCreated := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
//	fmt.Printf("State root - created account: %v\n", hrCreated)
//
//	adb.SaveAccountState(state1)
//	hrWithBalance := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
//	fmt.Printf("State root - account with balance 40: %v\n", hrWithBalance)
//
//	adb.Commit()
//	hrCommit := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
//	fmt.Printf("State root - commited: %v\n", hrCommit)
//
//	//commit hash == account with balance
//	assert.Equal(t, hrCommit, hrWithBalance)
//
//	state1.Balance = big.NewInt(0)
//	adb.SaveAccountState(state1)
//
//	//root hash == hrCreated
//	assert.Equal(t, hrCreated, base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))
//	fmt.Printf("State root - account with balance 0: %v\n", base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))
//
//	adb.RemoveAccount(*adr1)
//	//root hash == hrEmpty
//	assert.Equal(t, hrEmpty, base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))
//	fmt.Printf("State root - empty: %v\n", base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))
//}

func TestAccountsDB_RevertBalanceStepByStep_AccountData_ShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash1 := hasher.Compute("ABCDEFGHIJKLMNOP")
	testHash2 := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adr1 := adbCreateAddress(t, testHash1)
	adr2 := adbCreateAddress(t, testHash2)

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()
	hrEmpty := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.Jurnal().Len()
	hrCreated1 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)

	state2, err := adb.GetOrCreateAccount(*adr2)
	assert.Nil(t, err)
	snapshotCreated2 := adb.Jurnal().Len()
	hrCreated2 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Save the preset snapshot id
	snapshotPreSet := adb.Jurnal().Len()

	//Step 3. Set balances and save data
	state1.SetBalance(adb, big.NewInt(40))
	adb.SaveAccountState(state1)
	hrWithBalance1 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - account with balance 40: %v\n", hrWithBalance1)

	state2.SetBalance(adb, big.NewInt(50))
	adb.SaveAccountState(state2)
	hrWithBalance2 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - account with balance 50: %v\n", hrWithBalance2)

	//Test 3.1. current root hash shall not match created root hash hrCreated2
	assert.NotEqual(t, hrCreated2, adb.MainTrie.Root())

	//Step 4. Revert account balances and test
	adb.Jurnal().RevertFromSnapshot(snapshotPreSet)

	//Test 4.1. current root hash shall match created root hash hrCreated
	hrFinal := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	assert.Equal(t, hrCreated2, hrFinal)
	fmt.Printf("State root - reverted last 2 balance set: %v\n", hrFinal)
}

func TestAccountsDB_RevertCodeStepByStep_AccountData_ShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	//adr1 puts code hash + code inside trie. adr2 has the same code hash
	//revert should work

	testHash1 := hasher.Compute("ABCDEFGHIJKLMNOP")
	testHash2 := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adr1 := adbCreateAddress(t, testHash1)
	adr2 := adbCreateAddress(t, testHash2)

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()
	hrEmpty := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetOrCreateAccount(*adr1)
	adb.PutCode(state1, []byte{65, 66, 67})
	assert.Nil(t, err)
	snapshotCreated1 := adb.Jurnal().Len()
	hrCreated1 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)

	state2, err := adb.GetOrCreateAccount(*adr2)
	adb.PutCode(state2, []byte{65, 66, 67})
	assert.Nil(t, err)
	snapshotCreated2 := adb.Jurnal().Len()
	hrCreated2 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Step 3. Revert second account
	adb.Jurnal().RevertFromSnapshot(snapshotCreated1)

	//Test 3.1. current root hash shall match created root hash hrCreated1
	hrCrt := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	assert.Equal(t, hrCreated1, hrCrt)
	fmt.Printf("State root - reverted last account: %v\n", hrCrt)

	//Step 4. Revert first account
	adb.Jurnal().RevertFromSnapshot(0)

	//Test 4.1. current root hash shall match empty root hash
	hrCrt = base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	assert.Equal(t, hrEmpty, hrCrt)
	fmt.Printf("State root - reverted first account: %v\n", hrCrt)
}
