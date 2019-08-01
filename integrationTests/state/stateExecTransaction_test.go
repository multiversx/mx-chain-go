package state

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/stretchr/testify/assert"
)

func TestExecTransaction_SelfTransactionShouldWork(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("this is not a short test")
	}

	accnts, _, _ := integrationTests.CreateAccountsDB(nil)
	txProcessor := integrationTests.CreateSimpleTxProcessor(accnts)
	nonce := uint64(6)
	balance := big.NewInt(10000)

	//Step 1. create account with a nonce and a balance
	address := integrationTests.CreateAccount(accnts, nonce, balance)
	hashCreated, _ := accnts.Commit()

	//Step 2. create a tx moving 1 from address to address
	tx := &transaction.Transaction{
		Nonce:   nonce,
		Value:   big.NewInt(1),
		SndAddr: address.Bytes(),
		RcvAddr: address.Bytes(),
	}

	err := txProcessor.ProcessTransaction(tx, 0)
	assert.Nil(t, err)

	hashAfterExec, _ := accnts.Commit()
	assert.NotEqual(t, hashCreated, hashAfterExec)

	accountAfterExec, _ := accnts.GetAccountWithJournal(address)
	assert.Equal(t, nonce+1, accountAfterExec.(*state.Account).Nonce)
	assert.Equal(t, balance, accountAfterExec.(*state.Account).Balance)
}

func TestExecTransaction_SelfTransactionWithRevertShouldWork(t *testing.T) {
	t.Parallel()

	accnts, _, _ := integrationTests.CreateAccountsDB(nil)
	txProcessor := integrationTests.CreateSimpleTxProcessor(accnts)
	nonce := uint64(6)
	balance := big.NewInt(10000)

	//Step 1. create account with a nonce and a balance
	address := integrationTests.CreateAccount(accnts, nonce, balance)
	_, _ = accnts.Commit()

	//Step 2. create a tx moving 1 from pubKeyBuff to pubKeyBuff
	tx := &transaction.Transaction{
		Nonce:   nonce,
		Value:   big.NewInt(1),
		SndAddr: address.Bytes(),
		RcvAddr: address.Bytes(),
	}

	err := txProcessor.ProcessTransaction(tx, 0)
	assert.Nil(t, err)

	_ = accnts.RevertToSnapshot(0)

	accountAfterExec, _ := accnts.GetAccountWithJournal(address)
	assert.Equal(t, nonce, accountAfterExec.(*state.Account).Nonce)
	assert.Equal(t, balance, accountAfterExec.(*state.Account).Balance)
}

func TestExecTransaction_MoreTransactionsWithRevertShouldWork(t *testing.T) {
	t.Parallel()

	accnts, _, _ := integrationTests.CreateAccountsDB(nil)

	nonce := uint64(6)
	initialBalance := int64(100000)
	balance := big.NewInt(initialBalance)

	sender := integrationTests.CreateAccount(accnts, nonce, balance)
	receiver := integrationTests.CreateAddressFromHex()

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
	initialBalance int64,
) {

	txProcessor := integrationTests.CreateSimpleTxProcessor(accnts)
	txToGenerate := 15000

	//Step 1. execute a lot moving transactions from pubKeyBuff to another pubKeyBuff
	for i := 0; i < txToGenerate; i++ {
		tx := &transaction.Transaction{
			Nonce:   initialNonce + uint64(i),
			Value:   big.NewInt(1),
			SndAddr: sender.Bytes(),
			RcvAddr: receiver.Bytes(),
		}

		err := txProcessor.ProcessTransaction(tx, 0)
		assert.Nil(t, err)
	}

	modifiedHash, err := accnts.RootHash()
	assert.Nil(t, err)
	fmt.Printf("Modified hash: %s\n", base64.StdEncoding.EncodeToString(modifiedHash))

	//Step 2. test that accounts have correct nonces and balances
	newAccount, _ := accnts.GetAccountWithJournal(receiver)
	account, _ := accnts.GetAccountWithJournal(sender)

	assert.Equal(t, account.(*state.Account).Balance, big.NewInt(initialBalance-int64(txToGenerate)))
	assert.Equal(t, account.(*state.Account).Nonce, uint64(txToGenerate)+initialNonce)

	assert.Equal(t, newAccount.(*state.Account).Balance, big.NewInt(int64(txToGenerate)))
	assert.Equal(t, newAccount.(*state.Account).Nonce, uint64(0))

	assert.NotEqual(t, initialHash, modifiedHash)

	fmt.Printf("Journalized: %d modifications to the state\n", accnts.JournalLen())

	//Step 3. Revert and test again nonces, balances and root hash
	err = accnts.RevertToSnapshot(0)
	assert.Nil(t, err)

	revertedHash, err := accnts.RootHash()
	assert.Nil(t, err)
	fmt.Printf("Reverted hash: %s\n", base64.StdEncoding.EncodeToString(revertedHash))

	receiver2, _ := accnts.GetExistingAccount(receiver)
	account, _ = accnts.GetAccountWithJournal(sender)

	assert.Equal(t, account.(*state.Account).Balance, big.NewInt(initialBalance))
	assert.Equal(t, account.(*state.Account).Nonce, initialNonce)

	assert.Nil(t, receiver2)

	assert.Equal(t, initialHash, revertedHash)
}

func TestExecTransaction_MoreTransactionsMoreIterationsWithRevertShouldWork(t *testing.T) {
	t.Parallel()

	accnts, _, _ := integrationTests.CreateAccountsDB(nil)

	nonce := uint64(6)
	initialBalance := int64(100000)
	balance := big.NewInt(initialBalance)

	sender := integrationTests.CreateAccount(accnts, nonce, balance)
	receiver := integrationTests.CreateAddressFromHex()

	initialHash, _ := accnts.Commit()
	fmt.Printf("Initial hash: %s\n", base64.StdEncoding.EncodeToString(initialHash))

	for i := 0; i < 10; i++ {
		fmt.Printf("Iteration: %d\n", i)

		testExecTransactionsMoreTxWithRevert(t, accnts, sender, receiver, initialHash, nonce, initialBalance)
	}
}
