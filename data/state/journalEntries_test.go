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
	_, _ = rand.Read(buff)

	addr, err := state.NewAddress(buff)
	if err != nil {
		panic(err)
	}

	return addr
}

func jeCreateAccountsDB() *state.AccountsDB {
	marsh := mock.MarshalizerMock{}
	adb := state.NewAccountsDB(mock.NewMockTrie(), mock.HasherMock{}, &marsh)

	return adb
}

func TestJournalEntryCreation_Revert_OkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jeCreateAccountsDB()
	adr1 := jeCreateRandomAddress()
	//attach a journal
	j := state.NewJournal(adb)

	hashEmptyRoot := adb.MainTrie.Root()

	//create account for address
	_, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	j.AddEntry(state.NewJournalEntryCreation(adr1))

	//get the hash root for the new account
	hashEmptyAcnt := adb.MainTrie.Root()

	//test if empty account root hash is different from empty root hash
	assert.NotEqual(t, hashEmptyAcnt, hashEmptyRoot)

	//revert and test that current trie hash is equal to emptyRootHash
	err = j.RevertFromSnapshot(0)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyRoot, adb.MainTrie.Root())

}

func TestJournalEntryCreation_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jec := state.NewJournalEntryCreation(nil)
	j := state.NewJournal(nil)
	j.AddEntry(jec)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	j = state.NewJournal(jeCreateAccountsDB())
	j.AddEntry(jec)

	//error as nil addresses are not permited
	err = j.RevertFromSnapshot(0)
	assert.NotNil(t, err)
}

func TestJournalEntryNonce_Revert_OkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jeCreateAccountsDB()
	adr1 := jeCreateRandomAddress()

	//create account for address
	acnt, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	//attach a journal
	j := state.NewJournal(adb)

	//get the hash root for account with balance 0
	hashEmptyAcnt := adb.MainTrie.Root()
	//add balance journal entry
	j.AddEntry(state.NewJournalEntryNonce(adr1, acnt, 0))

	//modify the nonce and save
	err = acnt.SetNonce(adb, 50)
	assert.Nil(t, err)
	err = adb.SaveAccountState(acnt)
	assert.Nil(t, err)
	//get the new hash and test it is different from empty account hash
	hashAcnt := adb.MainTrie.Root()
	assert.NotEqual(t, hashEmptyAcnt, hashAcnt)

	//revert and test that hash is equal
	err = j.RevertFromSnapshot(0)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyAcnt, adb.MainTrie.Root())

}

func TestJournalEntryNonce_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jeb := state.NewJournalEntryNonce(nil, nil, 0)
	j := state.NewJournal(nil)
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	j = state.NewJournal(jeCreateAccountsDB())
	j.AddEntry(jeb)

	//error as nil addresses are not permited
	err = j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	jeb = state.NewJournalEntryNonce(jeCreateRandomAddress(), nil, 0)
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err = j.RevertFromSnapshot(1)
	assert.NotNil(t, err)
}

func TestJournalEntryBalance_Revert_OkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jeCreateAccountsDB()
	adr1 := jeCreateRandomAddress()

	//create account for address
	acnt, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	//attach a journal
	j := state.NewJournal(adb)

	//get the hash root for account with balance 0
	hashEmptyAcnt := adb.MainTrie.Root()
	//add balance journal entry
	j.AddEntry(state.NewJournalEntryBalance(adr1, acnt, big.NewInt(0)))

	//modify the balance and save
	err = acnt.SetBalance(adb, big.NewInt(50))
	assert.Nil(t, err)
	err = adb.SaveAccountState(acnt)
	assert.Nil(t, err)
	//get the new hash and test it is different from empty account hash
	hashAcnt := adb.MainTrie.Root()
	assert.NotEqual(t, hashEmptyAcnt, hashAcnt)

	//revert and test that hash is equal
	err = j.RevertFromSnapshot(0)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyAcnt, adb.MainTrie.Root())
}

func TestJournalEntryBalance_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jeb := state.NewJournalEntryBalance(nil, nil, big.NewInt(0))
	j := state.NewJournal(nil)
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	j = state.NewJournal(jeCreateAccountsDB())
	j.AddEntry(jeb)

	//error as nil addresses are not permited
	err = j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	jeb = state.NewJournalEntryBalance(jeCreateRandomAddress(), nil, big.NewInt(0))
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err = j.RevertFromSnapshot(1)
	assert.NotNil(t, err)
}

func TestJournalEntryCodeHash_Revert_OkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jeCreateAccountsDB()
	adr1 := jeCreateRandomAddress()

	//create account for address
	acnt, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	//attach a journal
	j := state.NewJournal(adb)

	//get the hash root for empty account
	hashEmptyAcnt := adb.MainTrie.Root()
	//add code hash journal entry
	j.AddEntry(state.NewJournalEntryCodeHash(adr1, acnt, nil))

	//modify the code hash and save
	err = acnt.SetCodeHash(adb, []byte{65, 66, 67})
	assert.Nil(t, err)
	err = adb.SaveAccountState(acnt)
	assert.Nil(t, err)
	//get the new hash and test it is different from empty account hash
	hashAcnt := adb.MainTrie.Root()
	assert.NotEqual(t, hashEmptyAcnt, hashAcnt)

	//revert and test that hash is equal
	err = j.RevertFromSnapshot(0)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyAcnt, adb.MainTrie.Root())

}

func TestJournalEntryCodeHash_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jech := state.NewJournalEntryCodeHash(nil, nil, make([]byte, 0))
	j := state.NewJournal(nil)
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	j = state.NewJournal(jeCreateAccountsDB())
	j.AddEntry(jech)

	//error as nil addresses are not permited
	err = j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	jech = state.NewJournalEntryCodeHash(jeCreateRandomAddress(), nil, make([]byte, 0))
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err = j.RevertFromSnapshot(1)
	assert.NotNil(t, err)
}

func TestJournalEntryCode_Revert_OkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts
	adb := jeCreateAccountsDB()
	//attach a journal
	j := state.NewJournal(adb)

	//get the hash root for empty root
	hashEmptyRoot := adb.MainTrie.Root()

	code := []byte{65, 66, 67}
	codeHash := mock.HasherMock{}.Compute(string(code))

	//add journal entry for code addition
	j.AddEntry(state.NewJournalEntryCode(codeHash))

	err := adb.PutCode(nil, code)
	assert.Nil(t, err)

	//get the hash root for the trie with code inside
	hashRootCode := adb.MainTrie.Root()

	//test hashes to be different
	assert.NotEqual(t, hashEmptyRoot, hashRootCode)

	//revert
	err = j.RevertFromSnapshot(0)
	assert.Nil(t, err)

	//test root hash to be empty root hash
	assert.Equal(t, hashEmptyRoot, adb.MainTrie.Root())
}

func TestJournalEntryCode_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jech := state.NewJournalEntryCode(nil)
	j := state.NewJournal(nil)
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(0)
	assert.NotNil(t, err)
}

func TestJournalEntryRoot_Revert_OkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jeCreateAccountsDB()
	adr1 := jeCreateRandomAddress()

	//create account for address
	acnt, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	//attach a journal
	j := state.NewJournal(adb)

	//get the hash root for empty account
	hashEmptyAcnt := adb.MainTrie.Root()
	//add code hash journal entry
	j.AddEntry(state.NewJournalEntryRoot(adr1, acnt, nil))

	//modify the root and save
	err = acnt.SetRoot(adb, []byte{65, 66, 67})
	assert.Nil(t, err)
	err = adb.SaveAccountState(acnt)
	assert.Nil(t, err)
	//get the new hash and test it is different from empty account hash
	hashAcnt := adb.MainTrie.Root()
	assert.NotEqual(t, hashEmptyAcnt, hashAcnt)

	//revert and test that hash is equal
	err = j.RevertFromSnapshot(0)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyAcnt, adb.MainTrie.Root())
}

func TestJournalEntryRoot_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jech := state.NewJournalEntryRoot(nil, nil, make([]byte, 0))
	j := state.NewJournal(nil)
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	j = state.NewJournal(jeCreateAccountsDB())
	j.AddEntry(jech)

	//error as nil addresses are not permited
	err = j.RevertFromSnapshot(0)
	assert.NotNil(t, err)

	jech = state.NewJournalEntryRoot(jeCreateRandomAddress(), nil, make([]byte, 0))
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err = j.RevertFromSnapshot(1)
	assert.NotNil(t, err)
}

func TestJournalEntryData_Revert_OkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jeCreateAccountsDB()
	adr1 := jeCreateRandomAddress()

	//create account for address
	acnt, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	//attach a journal
	j := state.NewJournal(adb)

	acnt.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})

	trie := mock.NewMockTrie()
	jed := state.NewJournalEntryData(trie, acnt)

	j.AddEntry(jed)

	err = j.RevertFromSnapshot(0)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(acnt.DirtyData()))
	assert.Equal(t, trie, jed.Trie())
}
