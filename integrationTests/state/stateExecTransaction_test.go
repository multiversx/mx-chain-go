package state

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	transaction2 "github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/stretchr/testify/assert"
)

func TestExecTransaction_SelfTransactionShouldWork(t *testing.T) {
	accnts := adbCreateAccountsDB()

	pubKeyBuff := createDummyHexAddress(64)

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}
	addrConv, _ := state.NewPlainAddressConverter(32, "0x")

	txProcessor, _ := transaction.NewTxProcessor(accnts, hasher, addrConv, marshalizer)

	nonce := uint64(6)
	balance := big.NewInt(10000)

	//Step 1. create account with a nonce and a balance
	address, _ := addrConv.CreateAddressFromHex(string(pubKeyBuff))
	account, _ := accnts.GetJournalizedAccount(address)
	account.SetNonceWithJournal(nonce)
	account.SetBalanceWithJournal(*balance)

	hashCreated, _ := accnts.Commit()

	//Step 2. create a tx moving 1 from pubKeyBuff to pubKeyBuff
	tx := &transaction2.Transaction{
		Nonce:   nonce,
		Value:   *big.NewInt(1),
		SndAddr: address.Bytes(),
		RcvAddr: address.Bytes(),
	}

	err := txProcessor.ProcessTransaction(tx, 0)
	assert.Nil(t, err)

	hashAfterExec, _ := accnts.Commit()
	assert.NotEqual(t, hashCreated, hashAfterExec)

	accountAfterExec, _ := accnts.GetJournalizedAccount(address)
	assert.Equal(t, nonce+1, accountAfterExec.BaseAccount().Nonce)
	assert.Equal(t, *balance, accountAfterExec.BaseAccount().Balance)
}

func TestExecTransaction_SelfTransactionWithRevertShouldWork(t *testing.T) {
	accnts := adbCreateAccountsDB()

	pubKeyBuff := createDummyHexAddress(64)

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}
	addrConv, _ := state.NewPlainAddressConverter(32, "0x")

	txProcessor, _ := transaction.NewTxProcessor(accnts, hasher, addrConv, marshalizer)

	nonce := uint64(6)
	balance := big.NewInt(10000)

	//Step 1. create account with a nonce and a balance
	address, _ := addrConv.CreateAddressFromHex(string(pubKeyBuff))
	account, _ := accnts.GetJournalizedAccount(address)
	account.SetNonceWithJournal(nonce)
	account.SetBalanceWithJournal(*balance)

	accnts.Commit()

	//Step 2. create a tx moving 1 from pubKeyBuff to pubKeyBuff
	tx := &transaction2.Transaction{
		Nonce:   nonce,
		Value:   *big.NewInt(1),
		SndAddr: address.Bytes(),
		RcvAddr: address.Bytes(),
	}

	err := txProcessor.ProcessTransaction(tx, 0)
	assert.Nil(t, err)

	_ = accnts.RevertToSnapshot(0)

	accountAfterExec, _ := accnts.GetJournalizedAccount(address)
	assert.Equal(t, nonce, accountAfterExec.BaseAccount().Nonce)
	assert.Equal(t, *balance, accountAfterExec.BaseAccount().Balance)
}
