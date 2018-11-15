package state_test

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

func jeCreateRandomAddress() *state.Address {
	buff := make([]byte, state.AdrLen)
	rand.Read(buff)

	addr := state.Address{}
	addr.SetBytes(buff)

	return &addr
}

func jeCreateAccountsDB() *state.AccountsDB {
	marsh := mock.MarshalizerMock{}
	adb := state.NewAccountsDB(mock.NewMockTrie(), mock.HasherMock{}, &marsh)

	return adb
}

func TestJurnalEntryCreation_Revert_OkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jeCreateAccountsDB()
	adr1 := jeCreateRandomAddress()
	//attach a jurnal
	j := state.NewJurnal(adb)

	hashEmptyRoot := adb.MainTrie.Root()

	//create account for address
	acnt, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	j.AddEntry(state.NewJurnalEntryCreation(adr1, acnt))

	//get the hash root for the new account
	hashEmptyAcnt := adb.MainTrie.Root()

	//test if empty account root hash is different from empty root hash
	assert.NotEqual(t, hashEmptyAcnt, hashEmptyRoot)

	//revert and test that current trie hash is equal to emptyRootHash
	err = j.RevertFromSnapshot(0)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyRoot, adb.MainTrie.Root())

}

func TestJurnalEntryCreation_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jec := state.NewJurnalEntryCreation(nil, nil)
	j := state.NewJurnal(nil)
	j.AddEntry(jec)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	j = state.NewJurnal(jeCreateAccountsDB())
	j.AddEntry(jec)

	//error as nil addresses are not permited
	err = j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	jec = state.NewJurnalEntryCreation(jeCreateRandomAddress(), nil)
	j.AddEntry(jec)

	//error as nil accounts are not permited
	err = j.RevertFromSnapshot(1)
	assert.NotNil(t, err)
}

func TestJurnalEntryNonce_Revert_OkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jeCreateAccountsDB()
	adr1 := jeCreateRandomAddress()

	//create account for address
	acnt, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	//attach a jurnal
	j := state.NewJurnal(adb)

	//get the hash root for account with balance 0
	hashEmptyAcnt := adb.MainTrie.Root()
	//add balance jurnal entry
	j.AddEntry(state.NewJurnalEntryNonce(adr1, acnt, 0))

	//modify the nonce and save
	acnt.SetNonce(adb, 50)
	adb.SaveAccountState(acnt)
	//get the new hash and test it is different from empty account hash
	hashAcnt := adb.MainTrie.Root()
	assert.NotEqual(t, hashEmptyAcnt, hashAcnt)

	//revert and test that hash is equal
	err = j.RevertFromSnapshot(0)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyAcnt, adb.MainTrie.Root())

}

func TestJurnalEntryNonce_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jeb := state.NewJurnalEntryNonce(nil, nil, 0)
	j := state.NewJurnal(nil)
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	j = state.NewJurnal(jeCreateAccountsDB())
	j.AddEntry(jeb)

	//error as nil addresses are not permited
	err = j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	jeb = state.NewJurnalEntryNonce(jeCreateRandomAddress(), nil, 0)
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err = j.RevertFromSnapshot(1)
	assert.NotNil(t, err)
}

func TestJurnalEntryBalance_Revert_OkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jeCreateAccountsDB()
	adr1 := jeCreateRandomAddress()

	//create account for address
	acnt, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	//attach a jurnal
	j := state.NewJurnal(adb)

	//get the hash root for account with balance 0
	hashEmptyAcnt := adb.MainTrie.Root()
	//add balance jurnal entry
	j.AddEntry(state.NewJurnalEntryBalance(adr1, acnt, big.NewInt(0)))

	//modify the balance and save
	acnt.SetBalance(adb, big.NewInt(50))
	adb.SaveAccountState(acnt)
	//get the new hash and test it is different from empty account hash
	hashAcnt := adb.MainTrie.Root()
	assert.NotEqual(t, hashEmptyAcnt, hashAcnt)

	//revert and test that hash is equal
	err = j.RevertFromSnapshot(0)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyAcnt, adb.MainTrie.Root())
}

func TestJurnalEntryBalance_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jeb := state.NewJurnalEntryBalance(nil, nil, big.NewInt(0))
	j := state.NewJurnal(nil)
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	j = state.NewJurnal(jeCreateAccountsDB())
	j.AddEntry(jeb)

	//error as nil addresses are not permited
	err = j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	jeb = state.NewJurnalEntryBalance(jeCreateRandomAddress(), nil, big.NewInt(0))
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err = j.RevertFromSnapshot(1)
	assert.NotNil(t, err)
}

func TestJurnalEntryCodeHash_Revert_OkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jeCreateAccountsDB()
	adr1 := jeCreateRandomAddress()

	//create account for address
	acnt, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	//attach a jurnal
	j := state.NewJurnal(adb)

	//get the hash root for empty account
	hashEmptyAcnt := adb.MainTrie.Root()
	//add code hash jurnal entry
	j.AddEntry(state.NewJurnalEntryCodeHash(adr1, acnt, nil))

	//modify the code hash and save
	acnt.SetCodeHash(adb, []byte{65, 66, 67})
	adb.SaveAccountState(acnt)
	//get the new hash and test it is different from empty account hash
	hashAcnt := adb.MainTrie.Root()
	assert.NotEqual(t, hashEmptyAcnt, hashAcnt)

	//revert and test that hash is equal
	err = j.RevertFromSnapshot(0)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyAcnt, adb.MainTrie.Root())

}

func TestJurnalEntryCodeHash_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jech := state.NewJurnalEntryCodeHash(nil, nil, make([]byte, 0))
	j := state.NewJurnal(nil)
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	j = state.NewJurnal(jeCreateAccountsDB())
	j.AddEntry(jech)

	//error as nil addresses are not permited
	err = j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	jech = state.NewJurnalEntryCodeHash(jeCreateRandomAddress(), nil, make([]byte, 0))
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err = j.RevertFromSnapshot(1)
	assert.NotNil(t, err)
}

func TestJurnalEntryCode_Revert_OkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts
	adb := jeCreateAccountsDB()
	//attach a jurnal
	j := state.NewJurnal(adb)

	//get the hash root for empty root
	hashEmptyRoot := adb.MainTrie.Root()

	code := []byte{65, 66, 67}
	codeHash := mock.HasherMock{}.Compute(string(code))

	//add jurnal entry for code addition
	j.AddEntry(state.NewJurnalEntryCode(codeHash))

	adb.PutCode(nil, code)

	//get the hash root for the trie with code inside
	hashRootCode := adb.MainTrie.Root()

	//test hashes to be different
	assert.NotEqual(t, hashEmptyRoot, hashRootCode)

	//revert
	j.RevertFromSnapshot(0)

	//test root hash to be empty root hash
	assert.Equal(t, hashEmptyRoot, adb.MainTrie.Root())
}

func TestJurnalEntryCode_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jech := state.NewJurnalEntryCode(nil)
	j := state.NewJurnal(nil)
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(0)
	assert.NotNil(t, err)
}

func TestJurnalEntryRoot_Revert_OkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jeCreateAccountsDB()
	adr1 := jeCreateRandomAddress()

	//create account for address
	acnt, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	//attach a jurnal
	j := state.NewJurnal(adb)

	//get the hash root for empty account
	hashEmptyAcnt := adb.MainTrie.Root()
	//add code hash jurnal entry
	j.AddEntry(state.NewJurnalEntryRoot(adr1, acnt, nil))

	//modify the root and save
	acnt.SetRoot(adb, []byte{65, 66, 67})
	adb.SaveAccountState(acnt)
	//get the new hash and test it is different from empty account hash
	hashAcnt := adb.MainTrie.Root()
	assert.NotEqual(t, hashEmptyAcnt, hashAcnt)

	//revert and test that hash is equal
	err = j.RevertFromSnapshot(0)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyAcnt, adb.MainTrie.Root())
}

func TestJurnalEntryRoot_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jech := state.NewJurnalEntryRoot(nil, nil, make([]byte, 0))
	j := state.NewJurnal(nil)
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	j = state.NewJurnal(jeCreateAccountsDB())
	j.AddEntry(jech)

	//error as nil addresses are not permited
	err = j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	jech = state.NewJurnalEntryRoot(jeCreateRandomAddress(), nil, make([]byte, 0))
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err = j.RevertFromSnapshot(1)
	assert.NotNil(t, err)
}
