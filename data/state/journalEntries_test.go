package state_test

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

func jurnalEntriesCreateRandomAddress() *state.Address {
	buff := make([]byte, state.AdrLen)
	_, _ = rand.Read(buff)

	addr, err := state.NewAddress(buff)
	if err != nil {
		panic(err)
	}

	return addr
}

func jurnalEntriesCreateAccountsDB() *state.AccountsDB {
	marsh := mock.MarshalizerMock{}
	adb := state.NewAccountsDB(mock.NewMockTrie(), mock.HasherMock{}, &marsh)

	return adb
}

func TestJournalEntryCreationRevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jurnalEntriesCreateAccountsDB()
	adr1 := jurnalEntriesCreateRandomAddress()
	//attach a journal
	j := state.NewJournal()

	hashEmptyRoot := adb.MainTrie.Root()

	//create account for address
	_, err := adb.GetOrCreateAccountState(adr1)
	assert.Nil(t, err)
	j.AddEntry(state.NewJournalEntryCreation(adr1))

	//get the hash root for the new account
	hashEmptyAcnt := adb.MainTrie.Root()

	//test if empty account root hash is different from empty root hash
	assert.NotEqual(t, hashEmptyAcnt, hashEmptyRoot)

	//revert and test that current trie hash is equal to emptyRootHash
	err = j.RevertToSnapshot(0, adb)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyRoot, adb.MainTrie.Root())
	//check address retention
	assert.Equal(t, state.NewJournalEntryCreation(adr1).DirtiedAddress(), adr1)
}

func TestJournalEntryCreationRevertInvalidValsShouldErr(t *testing.T) {
	t.Parallel()

	jec := state.NewJournalEntryCreation(nil)
	j := state.NewJournal()
	j.AddEntry(jec)

	//error as nil accounts are not permited
	err := j.RevertToSnapshot(0, nil)
	assert.NotNil(t, err)

	j = state.NewJournal()
	j.AddEntry(jec)

	//error as nil addresses are not permited
	err = j.RevertToSnapshot(0, jurnalEntriesCreateAccountsDB())
	assert.NotNil(t, err)
}

func TestJournalEntryNonceRevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jurnalEntriesCreateAccountsDB()
	adr1 := jurnalEntriesCreateRandomAddress()

	//create account for address
	acnt, err := adb.GetOrCreateAccountState(adr1)
	assert.Nil(t, err)
	//attach a journal
	j := state.NewJournal()

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
	err = j.RevertToSnapshot(0, adb)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyAcnt, adb.MainTrie.Root())
	//check address retention
	assert.Equal(t, state.NewJournalEntryNonce(adr1, acnt, 0).DirtiedAddress(), adr1)
}

func TestJournalEntryNonceRevertInvalidValsShouldErr(t *testing.T) {
	t.Parallel()

	jeb := state.NewJournalEntryNonce(nil, nil, 0)
	j := state.NewJournal()
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err := j.RevertToSnapshot(0, nil)
	assert.NotNil(t, err)

	j = state.NewJournal()
	j.AddEntry(jeb)

	//error as nil addresses are not permited
	err = j.RevertToSnapshot(0, jurnalEntriesCreateAccountsDB())
	assert.NotNil(t, err)

	jeb = state.NewJournalEntryNonce(jurnalEntriesCreateRandomAddress(), nil, 0)
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err = j.RevertToSnapshot(1, jurnalEntriesCreateAccountsDB())
	assert.NotNil(t, err)
}

func TestJournalEntryBalanceRevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jurnalEntriesCreateAccountsDB()
	adr1 := jurnalEntriesCreateRandomAddress()

	//create account for address
	acnt, err := adb.GetOrCreateAccountState(adr1)
	assert.Nil(t, err)
	//attach a journal
	j := state.NewJournal()

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
	err = j.RevertToSnapshot(0, adb)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyAcnt, adb.MainTrie.Root())
	//check address retention
	assert.Equal(t, state.NewJournalEntryBalance(adr1, acnt, big.NewInt(0)).DirtiedAddress(), adr1)
}

func TestJournalEntryBalanceRevertInvalidValsShouldErr(t *testing.T) {
	t.Parallel()

	jeb := state.NewJournalEntryBalance(nil, nil, big.NewInt(0))
	j := state.NewJournal()
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err := j.RevertToSnapshot(0, nil)
	assert.NotNil(t, err)

	j = state.NewJournal()
	j.AddEntry(jeb)

	//error as nil addresses are not permited
	err = j.RevertToSnapshot(0, jurnalEntriesCreateAccountsDB())
	assert.NotNil(t, err)

	jeb = state.NewJournalEntryBalance(jurnalEntriesCreateRandomAddress(), nil, big.NewInt(0))
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err = j.RevertToSnapshot(1, jurnalEntriesCreateAccountsDB())
	assert.NotNil(t, err)
}

func TestJournalEntryCodeHashRevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jurnalEntriesCreateAccountsDB()
	adr1 := jurnalEntriesCreateRandomAddress()

	//create account for address
	acnt, err := adb.GetOrCreateAccountState(adr1)
	assert.Nil(t, err)
	//attach a journal
	j := state.NewJournal()

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
	err = j.RevertToSnapshot(0, adb)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyAcnt, adb.MainTrie.Root())
	//check address retention
	assert.Equal(t, state.NewJournalEntryCodeHash(adr1, acnt, nil).DirtiedAddress(), adr1)
}

func TestJournalEntryCodeHashRevertInvalidValsShouldErr(t *testing.T) {
	t.Parallel()

	jech := state.NewJournalEntryCodeHash(nil, nil, make([]byte, 0))
	j := state.NewJournal()
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err := j.RevertToSnapshot(0, nil)
	assert.NotNil(t, err)

	j = state.NewJournal()
	j.AddEntry(jech)

	//error as nil addresses are not permited
	err = j.RevertToSnapshot(0, jurnalEntriesCreateAccountsDB())
	assert.NotNil(t, err)

	jech = state.NewJournalEntryCodeHash(jurnalEntriesCreateRandomAddress(), nil, make([]byte, 0))
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err = j.RevertToSnapshot(1, jurnalEntriesCreateAccountsDB())
	assert.NotNil(t, err)
}

func TestJournalEntryCodeRevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts
	adb := jurnalEntriesCreateAccountsDB()
	//attach a journal
	j := state.NewJournal()

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
	err = j.RevertToSnapshot(0, adb)
	assert.Nil(t, err)

	//test root hash to be empty root hash
	assert.Equal(t, hashEmptyRoot, adb.MainTrie.Root())
	//check address retention
	assert.Equal(t, state.NewJournalEntryCode(codeHash).DirtiedAddress(), nil)
}

func TestJournalEntryCodeRevertInvalidValsShouldErr(t *testing.T) {
	t.Parallel()

	jech := state.NewJournalEntryCode(nil)
	j := state.NewJournal()
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err := j.RevertToSnapshot(0, nil)
	assert.NotNil(t, err)
}

func TestJournalEntryRootHashRevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jurnalEntriesCreateAccountsDB()
	adr1 := jurnalEntriesCreateRandomAddress()

	//create account for address
	acnt, err := adb.GetOrCreateAccountState(adr1)
	assert.Nil(t, err)
	//attach a journal
	j := state.NewJournal()

	//get the hash root for empty account
	hashEmptyAcnt := adb.MainTrie.Root()
	//add code hash journal entry
	j.AddEntry(state.NewJournalEntryRootHash(adr1, acnt, nil))

	//modify the root and save
	err = acnt.SetRootHash(adb, []byte{65, 66, 67})
	assert.Nil(t, err)
	err = adb.SaveAccountState(acnt)
	assert.Nil(t, err)
	//get the new hash and test it is different from empty account hash
	hashAcnt := adb.MainTrie.Root()
	assert.NotEqual(t, hashEmptyAcnt, hashAcnt)

	//revert and test that hash is equal
	err = j.RevertToSnapshot(0, adb)
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyAcnt, adb.MainTrie.Root())
	//check address retention
	assert.Equal(t, state.NewJournalEntryRootHash(adr1, acnt, nil).DirtiedAddress(), adr1)
}

func TestJournalEntryRootHashRevertInvalidValsShouldErr(t *testing.T) {
	t.Parallel()

	jech := state.NewJournalEntryRootHash(nil, nil, make([]byte, 0))
	j := state.NewJournal()
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err := j.RevertToSnapshot(0, nil)
	assert.NotNil(t, err)

	j = state.NewJournal()
	j.AddEntry(jech)

	//error as nil addresses are not permited
	err = j.RevertToSnapshot(0, jurnalEntriesCreateAccountsDB())
	assert.NotNil(t, err)

	jech = state.NewJournalEntryRootHash(jurnalEntriesCreateRandomAddress(), nil, make([]byte, 0))
	j.AddEntry(jech)

	//error as nil accounts are not permited
	err = j.RevertToSnapshot(1, jurnalEntriesCreateAccountsDB())
	assert.NotNil(t, err)
}

func TestJournalEntryDataRevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jurnalEntriesCreateAccountsDB()
	adr1 := jurnalEntriesCreateRandomAddress()

	//create account for address
	acnt, err := adb.GetOrCreateAccountState(adr1)
	assert.Nil(t, err)
	//attach a journal
	j := state.NewJournal()

	acnt.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})

	trie := mock.NewMockTrie()
	jed := state.NewJournalEntryData(trie, acnt)

	j.AddEntry(jed)

	err = j.RevertToSnapshot(0, adb)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(acnt.DirtyData()))
	assert.Equal(t, trie, jed.Trie())
	//check address retention
	assert.Equal(t, state.NewJournalEntryData(trie, acnt).DirtiedAddress(), nil)
}
