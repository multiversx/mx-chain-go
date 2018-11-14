package state

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

func jeCreateRandomAddress() *Address {
	buff := make([]byte, AdrLen)
	rand.Read(buff)

	addr := Address{}
	addr.SetBytes(buff)

	return &addr
}

func jeCreateAccountsDB() *AccountsDB {
	adb := NewAccountsDB(mock.NewMockTrie(), &testHasher, &testMarshalizer)

	return adb
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
	j := NewJurnal(adb)

	//get the hash root for account with balance 0
	hashEmptyAcnt := adb.MainTrie.Root()
	//add balance jurnal entry
	j.AddEntry(NewJurnalEntryBalance(adr1, acnt, big.NewInt(0)))

	//modify the balance and save
	acnt.Balance = big.NewInt(50)
	adb.SaveAccountState(acnt)
	//get the new hash and test it is different from empty account hash
	hashAcnt := adb.MainTrie.Root()
	assert.NotEqual(t, hashEmptyAcnt, hashAcnt)

	//revert and test that hash is equal
	err = j.RevertFromSnapshot(uint32(1))
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyAcnt, adb.MainTrie.Root())

}

func TestJurnalEntryBalance_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jeb := NewJurnalEntryBalance(nil, nil, big.NewInt(0))
	j := NewJurnal(nil)
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(uint32(1))
	assert.NotNil(t, err)

	j = NewJurnal(jeCreateAccountsDB())
	j.AddEntry(jeb)

	//error as nil addresses are not permited
	err = j.RevertFromSnapshot(uint32(1))
	assert.NotNil(t, err)

	jeb = NewJurnalEntryBalance(jeCreateRandomAddress(), nil, big.NewInt(0))
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err = j.RevertFromSnapshot(uint32(2))
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
	j := NewJurnal(adb)

	//get the hash root for account with balance 0
	hashEmptyAcnt := adb.MainTrie.Root()
	//add balance jurnal entry
	j.AddEntry(NewJurnalEntryNonce(adr1, acnt, 0))

	//modify the nonce and save
	acnt.Nonce = 50
	adb.SaveAccountState(acnt)
	//get the new hash and test it is different from empty account hash
	hashAcnt := adb.MainTrie.Root()
	assert.NotEqual(t, hashEmptyAcnt, hashAcnt)

	//revert and test that hash is equal
	err = j.RevertFromSnapshot(uint32(1))
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyAcnt, adb.MainTrie.Root())

}

func TestJurnalEntryNonce_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jeb := NewJurnalEntryNonce(nil, nil, 0)
	j := NewJurnal(nil)
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(uint32(1))
	assert.NotNil(t, err)

	j = NewJurnal(jeCreateAccountsDB())
	j.AddEntry(jeb)

	//error as nil addresses are not permited
	err = j.RevertFromSnapshot(uint32(1))
	assert.NotNil(t, err)

	jeb = NewJurnalEntryNonce(jeCreateRandomAddress(), nil, 0)
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err = j.RevertFromSnapshot(uint32(2))
	assert.NotNil(t, err)
}

func TestJurnalEntryCreation_Revert_OkVals_ShouldWork(t *testing.T) {
	t.Parallel()

	//create accounts and address
	adb := jeCreateAccountsDB()
	adr1 := jeCreateRandomAddress()

	//get the hash root for account with balance 0
	hashEmptyRoot := adb.MainTrie.Root()

	//create account for address
	acnt, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	//attach a jurnal
	j := NewJurnal(adb)

	//add creation jurnal entry
	j.AddEntry(NewJurnalEntryCreation(adr1, acnt))

	//get the new hash and test it is different from empty root hash
	hashAcnt := adb.MainTrie.Root()
	assert.NotEqual(t, hashEmptyRoot, hashAcnt)

	//revert and test that hash is equal
	err = j.RevertFromSnapshot(uint32(1))
	assert.Nil(t, err)
	assert.Equal(t, hashEmptyRoot, adb.MainTrie.Root())

}

func TestJurnalEntryCreation_Revert_InvalidVals_ShouldErr(t *testing.T) {
	t.Parallel()

	jeb := NewJurnalEntryCreation(nil, nil)
	j := NewJurnal(nil)
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err := j.RevertFromSnapshot(uint32(1))
	assert.NotNil(t, err)

	j = NewJurnal(jeCreateAccountsDB())
	j.AddEntry(jeb)

	//error as nil addresses are not permited
	err = j.RevertFromSnapshot(uint32(1))
	assert.NotNil(t, err)

	jeb = NewJurnalEntryCreation(jeCreateRandomAddress(), nil)
	j.AddEntry(jeb)

	//error as nil accounts are not permited
	err = j.RevertFromSnapshot(uint32(2))
	assert.NotNil(t, err)
}
