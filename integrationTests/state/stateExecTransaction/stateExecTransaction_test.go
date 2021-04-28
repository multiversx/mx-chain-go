package stateExecTransaction

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	t.Parallel()

	trieStorage, _ := integrationTests.CreateTrieStorageManager(integrationTests.CreateMemUnit())
	accnts, _ := integrationTests.CreateAccountsDB(0, trieStorage)
	txProcessor := integrationTests.CreateSimpleTxProcessor(accnts)
	nonce := uint64(6)
	balance := big.NewInt(10000)

	//Step 1. create account with a nonce and a balance
	address := integrationTests.CreateAccount(accnts, nonce, balance)
	hashCreated, _ := accnts.Commit()

	//Step 2. create a tx moving 1 from pubKeyBuff to pubKeyBuff
	tx := &transaction.Transaction{
		Nonce:    nonce,
		Value:    big.NewInt(1),
		GasLimit: 2,
		GasPrice: 1,
		SndAddr:  address,
		RcvAddr:  address,
	}

	_, err := txProcessor.ProcessTransaction(tx)
	assert.Nil(t, err)

	hashAfterExec, _ := accnts.Commit()
	assert.NotEqual(t, hashCreated, hashAfterExec)

	balance.Sub(balance, big.NewInt(0).SetUint64(tx.GasPrice*tx.GasLimit))

	accountAfterExec, _ := accnts.LoadAccount(address)
	assert.Equal(t, nonce+1, accountAfterExec.(state.UserAccountHandler).GetNonce())
	assert.Equal(t, balance, accountAfterExec.(state.UserAccountHandler).GetBalance())
}

func TestExecTransaction_SelfTransactionWithRevertShouldWork(t *testing.T) {
	t.Parallel()

	trieStorage, _ := integrationTests.CreateTrieStorageManager(integrationTests.CreateMemUnit())
	accnts, _ := integrationTests.CreateAccountsDB(0, trieStorage)
	txProcessor := integrationTests.CreateSimpleTxProcessor(accnts)

	nonce := uint64(6)
	balance := big.NewInt(10000)

	//Step 1. create account with a nonce and a balance
	address := integrationTests.CreateAccount(accnts, nonce, balance)
	_, err := accnts.Commit()
	assert.Nil(t, err)

	//Step 2. create a tx moving 1 from pubKeyBuff to pubKeyBuff
	tx := &transaction.Transaction{
		Nonce:    nonce,
		Value:    big.NewInt(1),
		SndAddr:  address,
		RcvAddr:  address,
		GasLimit: 2,
		GasPrice: 2,
	}

	_, err = txProcessor.ProcessTransaction(tx)
	assert.Nil(t, err)

	err = accnts.RevertToSnapshot(0)
	assert.Nil(t, err)

	accountAfterExec, _ := accnts.LoadAccount(address)
	assert.Equal(t, nonce, accountAfterExec.(state.UserAccountHandler).GetNonce())
	assert.Equal(t, balance, accountAfterExec.(state.UserAccountHandler).GetBalance())
}

func TestExecTransaction_MoreTransactionsWithRevertShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Parallel()

	trieStorage, _ := integrationTests.CreateTrieStorageManager(integrationTests.CreateMemUnit())
	accnts, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	nonce := uint64(6)
	initialBalance := int64(100000)
	balance := big.NewInt(initialBalance)

	sender := integrationTests.CreateAccount(accnts, nonce, balance)
	receiver := integrationTests.CreateRandomAddress()

	initialHash, _ := accnts.Commit()
	fmt.Printf("Initial hash: %s\n", base64.StdEncoding.EncodeToString(initialHash))

	testExecTransactionsMoreTxWithRevert(t, accnts, sender, receiver, initialHash, nonce, initialBalance)
}

func testExecTransactionsMoreTxWithRevert(
	t *testing.T,
	accnts state.AccountsAdapter,
	sender []byte,
	receiver []byte,
	initialHash []byte,
	initialNonce uint64,
	initialBalance int64,
) {

	txProcessor := integrationTests.CreateSimpleTxProcessor(accnts)

	txToGenerate := 15000
	gasPrice := uint64(2)
	gasLimit := uint64(2)
	value := uint64(1)
	//Step 1. execute a lot moving transactions from pubKeyBuff to another pubKeyBuff
	for i := 0; i < txToGenerate; i++ {
		tx := &transaction.Transaction{
			Nonce:    initialNonce + uint64(i),
			Value:    big.NewInt(int64(value)),
			GasPrice: gasPrice,
			GasLimit: gasLimit,
			SndAddr:  sender,
			RcvAddr:  receiver,
		}

		_, err := txProcessor.ProcessTransaction(tx)
		assert.Nil(t, err)
	}

	modifiedHash, err := accnts.RootHash()
	assert.Nil(t, err)
	fmt.Printf("Modified hash: %s\n", base64.StdEncoding.EncodeToString(modifiedHash))

	//Step 2. test that accounts have correct nonces and balances
	newAccount, _ := accnts.LoadAccount(receiver)
	account, _ := accnts.LoadAccount(sender)

	assert.Equal(t, account.(state.UserAccountHandler).GetBalance(), big.NewInt(initialBalance-int64(uint64(txToGenerate)*(gasPrice*gasLimit+value))))
	assert.Equal(t, account.(state.UserAccountHandler).GetNonce(), uint64(txToGenerate)+initialNonce)

	assert.Equal(t, newAccount.(state.UserAccountHandler).GetBalance(), big.NewInt(int64(txToGenerate)))
	assert.Equal(t, newAccount.(state.UserAccountHandler).GetNonce(), uint64(0))

	assert.NotEqual(t, initialHash, modifiedHash)

	fmt.Printf("Journalized: %d modifications to the state\n", accnts.JournalLen())

	//Step 3. Revert and test again nonces, balances and root hash
	err = accnts.RevertToSnapshot(0)
	assert.Nil(t, err)

	revertedHash, err := accnts.RootHash()
	assert.Nil(t, err)
	fmt.Printf("Reverted hash: %s\n", base64.StdEncoding.EncodeToString(revertedHash))

	receiver2, _ := accnts.GetExistingAccount(receiver)
	account, _ = accnts.LoadAccount(sender)

	assert.Equal(t, account.(state.UserAccountHandler).GetBalance(), big.NewInt(initialBalance))
	assert.Equal(t, account.(state.UserAccountHandler).GetNonce(), initialNonce)

	assert.Nil(t, receiver2)

	assert.Equal(t, initialHash, revertedHash)
}

func TestExecTransaction_MoreTransactionsMoreIterationsWithRevertShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	t.Parallel()

	trieStorage, _ := integrationTests.CreateTrieStorageManager(integrationTests.CreateMemUnit())
	accnts, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	nonce := uint64(6)
	initialBalance := int64(100000)
	balance := big.NewInt(initialBalance)

	sender := integrationTests.CreateAccount(accnts, nonce, balance)
	receiver := integrationTests.CreateRandomAddress()

	initialHash, _ := accnts.Commit()
	fmt.Printf("Initial hash: %s\n", base64.StdEncoding.EncodeToString(initialHash))

	for i := 0; i < 10; i++ {
		fmt.Printf("Iteration: %d\n", i)

		testExecTransactionsMoreTxWithRevert(t, accnts, sender, receiver, initialHash, nonce, initialBalance)
	}
}
