package state

import (
	"encoding/base64"
	"fmt"
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
	account.SetBalanceWithJournal(balance)

	hashCreated, _ := accnts.Commit()

	//Step 2. create a tx moving 1 from pubKeyBuff to pubKeyBuff
	tx := &transaction2.Transaction{
		Nonce:   nonce,
		Value:   big.NewInt(1),
		SndAddr: address.Bytes(),
		RcvAddr: address.Bytes(),
	}

	err := txProcessor.ProcessTransaction(tx, 0)
	assert.Nil(t, err)

	hashAfterExec, _ := accnts.Commit()
	assert.NotEqual(t, hashCreated, hashAfterExec)

	accountAfterExec, _ := accnts.GetJournalizedAccount(address)
	assert.Equal(t, nonce+1, accountAfterExec.BaseAccount().Nonce)
	assert.Equal(t, balance, accountAfterExec.BaseAccount().Balance)
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
	account.SetBalanceWithJournal(balance)

	accnts.Commit()

	//Step 2. create a tx moving 1 from pubKeyBuff to pubKeyBuff
	tx := &transaction2.Transaction{
		Nonce:   nonce,
		Value:   big.NewInt(1),
		SndAddr: address.Bytes(),
		RcvAddr: address.Bytes(),
	}

	err := txProcessor.ProcessTransaction(tx, 0)
	assert.Nil(t, err)

	_ = accnts.RevertToSnapshot(0)

	accountAfterExec, _ := accnts.GetJournalizedAccount(address)
	assert.Equal(t, nonce, accountAfterExec.BaseAccount().Nonce)
	assert.Equal(t, balance, accountAfterExec.BaseAccount().Balance)
}

func TestExecTransaction_MoreTransactionsWithRevertShouldWork(t *testing.T) {
	accnts := adbCreateAccountsDB()

	nonce := uint64(6)
	initialBalance := int64(100000)
	balance := big.NewInt(initialBalance)

	addrConv, _ := state.NewPlainAddressConverter(32, "0x")
	pubKeyBuff := createDummyHexAddress(64)
	sender, _ := addrConv.CreateAddressFromHex(string(pubKeyBuff))

	pubKeyBuff = createDummyHexAddress(64)
	receiver, _ := addrConv.CreateAddressFromHex(string(pubKeyBuff))

	account, _ := accnts.GetJournalizedAccount(sender)
	account.SetNonceWithJournal(nonce)
	account.SetBalanceWithJournal(balance)

	initialHash, _ := accnts.Commit()
	fmt.Printf("Initial hash: %s\n", base64.StdEncoding.EncodeToString(initialHash))

	testExecTransactionsMoreTxWithRevert(t, accnts, sender, receiver, initialHash, nonce, initialBalance)
}

func testExecTransactionsMoreTxWithRevert(
	t *testing.T,
	accnts state.AccountsAdapter,
	sender state.AddressContainer,
	receiver state.AddressContainer,
	initialHash []byte,
	initialNonce uint64,
	initialBalance int64) {

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}
	addrConv, _ := state.NewPlainAddressConverter(32, "0x")

	txProcessor, _ := transaction.NewTxProcessor(accnts, hasher, addrConv, marshalizer)

	txToGenerate := 15000

	//Step 1. execute a lot moving transactions from pubKeyBuff to another pubKeyBuff
	for i := 0; i < txToGenerate; i++ {
		tx := &transaction2.Transaction{
			Nonce:   initialNonce + uint64(i),
			Value:   big.NewInt(1),
			SndAddr: sender.Bytes(),
			RcvAddr: receiver.Bytes(),
		}

		err := txProcessor.ProcessTransaction(tx, 0)
		assert.Nil(t, err)
	}

	modifiedHash := accnts.RootHash()
	fmt.Printf("Modified hash: %s\n", base64.StdEncoding.EncodeToString(modifiedHash))

	//Step 2. test that accounts have correct nonces and balances
	newAccount, _ := accnts.GetJournalizedAccount(receiver)
	account, _ := accnts.GetJournalizedAccount(sender)

	assert.Equal(t, account.BaseAccount().Balance, big.NewInt(initialBalance-int64(txToGenerate)))
	assert.Equal(t, account.BaseAccount().Nonce, uint64(txToGenerate)+initialNonce)

	assert.Equal(t, newAccount.BaseAccount().Balance, big.NewInt(int64(txToGenerate)))
	assert.Equal(t, newAccount.BaseAccount().Nonce, uint64(0))

	assert.NotEqual(t, initialHash, modifiedHash)

	fmt.Printf("Journalized: %d modifications to the state\n", accnts.JournalLen())

	//Step 3. Revert and test again nonces, balances and root hash
	err := accnts.RevertToSnapshot(0)

	assert.Nil(t, err)

	revertedHash := accnts.RootHash()
	fmt.Printf("Reverted hash: %s\n", base64.StdEncoding.EncodeToString(revertedHash))

	receiver2, _ := accnts.GetExistingAccount(receiver)
	account, _ = accnts.GetJournalizedAccount(sender)

	assert.Equal(t, account.BaseAccount().Balance, big.NewInt(initialBalance))
	assert.Equal(t, account.BaseAccount().Nonce, initialNonce)

	assert.Nil(t, receiver2)

	assert.Equal(t, initialHash, revertedHash)
}

func TestExecTransaction_MoreTransactionsMoreIterationsWithRevertShouldWork(t *testing.T) {
	t.Skip("This is a very long test")

	accnts := adbCreateAccountsDB()

	nonce := uint64(6)
	initialBalance := int64(100000)
	balance := big.NewInt(initialBalance)

	addrConv, _ := state.NewPlainAddressConverter(32, "0x")
	pubKeyBuff := createDummyHexAddress(64)
	sender, _ := addrConv.CreateAddressFromHex(string(pubKeyBuff))

	pubKeyBuff = createDummyHexAddress(64)
	receiver, _ := addrConv.CreateAddressFromHex(string(pubKeyBuff))

	account, _ := accnts.GetJournalizedAccount(sender)
	account.SetNonceWithJournal(nonce)
	account.SetBalanceWithJournal(balance)

	initialHash, _ := accnts.Commit()
	fmt.Printf("Initial hash: %s\n", base64.StdEncoding.EncodeToString(initialHash))

	for i := 0; i < 10000; i++ {
		fmt.Printf("Iteration: %d\n", i)

		testExecTransactionsMoreTxWithRevert(t, accnts, sender, receiver, initialHash, nonce, initialBalance)
	}
}
