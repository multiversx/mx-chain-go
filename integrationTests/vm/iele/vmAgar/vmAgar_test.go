package vmAgar

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/iele"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

var agarioFile = "../../../agarioV3.hex"

func TestDeployAgarioContract(t *testing.T) {
	// TODO: Use from utils.go
	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := uint64(1000000)

	testContext := vm.CreatePreparedTxProcessorAndAccountsWithVMs(senderNonce, senderAddressBytes, senderBalance)
	defer testContext.Close()

	_ = iele.DeployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		iele.CreateDeployTxData(scCode),
		testContext.TxProcessor,
		testContext.Accounts,
	)

	destinationAddressBytes, _ := testContext.BlockchainHook.NewAddress(senderAddressBytes, senderNonce, factory.IELEVirtualMachine)
	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		testContext.Accounts,
		big.NewInt(0),
		string(scCode),
		make(map[string]*big.Int))
}

func TestAgarioContractTopUpShouldWork(t *testing.T) {
	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := uint64(1000000)

	testContext := vm.CreatePreparedTxProcessorAndAccountsWithVMs(senderNonce, senderAddressBytes, senderBalance)
	defer testContext.Close()

	_ = iele.DeployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		iele.CreateDeployTxData(scCode),
		testContext.TxProcessor,
		testContext.Accounts,
	)

	scAddressBytes, _ := testContext.BlockchainHook.NewAddress(senderAddressBytes, senderNonce, factory.IELEVirtualMachine)

	userAddress := []byte("10000000000000000000000000000000")
	userNonce := uint64(10)
	userBalance := big.NewInt(100000000)
	_, _ = vm.CreateAccount(testContext.Accounts, userAddress, userNonce, userBalance)
	_, _ = testContext.Accounts.Commit()

	//balanceOf should return 0 for userAddress
	assert.Equal(t, big.NewInt(0), vm.GetIntValueFromSC(nil, testContext.Accounts, scAddressBytes, "balanceOf", userAddress))

	transfer := big.NewInt(123456)
	data := "topUp"
	//contract call tx
	txRun := vm.CreateTx(
		t,
		userAddress,
		scAddressBytes,
		userNonce,
		transfer,
		gasPrice,
		gasLimit,
		data,
	)

	_, err = testContext.TxProcessor.ProcessTransaction(txRun)
	assert.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	assert.Nil(t, err)

	assert.Equal(t, transfer, vm.GetIntValueFromSC(nil, testContext.Accounts, scAddressBytes, "balanceOf", userAddress))
}

func TestAgarioContractTopUpAnfWithdrawShouldWork(t *testing.T) {
	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)

	testContext := vm.CreatePreparedTxProcessorAndAccountsWithVMs(senderNonce, senderAddressBytes, senderBalance)
	defer testContext.Close()

	_ = iele.DeployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		iele.CreateDeployTxData(scCode),
		testContext.TxProcessor,
		testContext.Accounts,
	)

	scAddressBytes, _ := testContext.BlockchainHook.NewAddress(senderAddressBytes, senderNonce, factory.IELEVirtualMachine)

	userAddress := []byte("10000000000000000000000000000000")
	userNonce := uint64(10)
	userBalance := big.NewInt(100000000)
	_, _ = vm.CreateAccount(testContext.Accounts, userAddress, userNonce, userBalance)
	_, _ = testContext.Accounts.Commit()

	//balanceOf should return 0 for userAddress
	assert.Equal(t, big.NewInt(0), vm.GetIntValueFromSC(nil, testContext.Accounts, scAddressBytes, "balanceOf", userAddress))

	transfer := big.NewInt(123456)
	data := "topUp"
	//contract call tx
	txRun := vm.CreateTx(
		t,
		userAddress,
		scAddressBytes,
		userNonce,
		transfer,
		gasPrice,
		gasLimit,
		data,
	)

	userNonce++
	_, err = testContext.TxProcessor.ProcessTransaction(txRun)
	assert.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	assert.Nil(t, err)

	assert.Equal(t, transfer, vm.GetIntValueFromSC(nil, testContext.Accounts, scAddressBytes, "balanceOf", userAddress))

	//withdraw
	withdraw := uint64(4999)
	data = fmt.Sprintf("withdraw@%X", withdraw)
	//contract call tx
	txRun = vm.CreateTx(
		t,
		userAddress,
		scAddressBytes,
		userNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		data,
	)

	_, err = testContext.TxProcessor.ProcessTransaction(txRun)
	assert.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	assert.Nil(t, err)

	newValue := big.NewInt(0).Set(transfer)
	newValue.Sub(newValue, big.NewInt(0).SetUint64(withdraw))
	assert.Equal(t, newValue, vm.GetIntValueFromSC(nil, testContext.Accounts, scAddressBytes, "balanceOf", userAddress))
}

func TestAgarioContractJoinGameReward(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(0)
	gasLimit := uint64(100000)

	testContext := vm.CreatePreparedTxProcessorAndAccountsWithVMs(senderNonce, senderAddressBytes, senderBalance)
	defer testContext.Close()

	_ = iele.DeployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		iele.CreateDeployTxData(scCode),
		testContext.TxProcessor,
		testContext.Accounts,
	)

	scAddressBytes, _ := testContext.BlockchainHook.NewAddress(senderAddressBytes, senderNonce, factory.IELEVirtualMachine)

	senderNonce++

	defaultUserNonce := uint64(10)
	defaultUserBalance := big.NewInt(100000000)

	noOfUsers := 10
	usersAddresses := make([][]byte, noOfUsers)
	transfer := big.NewInt(100)

	afterJoinUsersBalances := make([]*big.Int, noOfUsers)

	for i := 0; i < noOfUsers; i++ {
		userAddress := make([]byte, 32)
		_, _ = rand.Reader.Read(userAddress)
		fmt.Printf("Generated user account: %v\n", hex.EncodeToString(userAddress))

		_, _ = vm.CreateAccount(testContext.Accounts, userAddress, defaultUserNonce, defaultUserBalance)
		_, _ = testContext.Accounts.Commit()

		usersAddresses[i] = userAddress
	}

	for i := 0; i < noOfUsers; i++ {
		//balanceOf should return 0 for userAddress
		balanceOfUser := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddressBytes, "balanceOf", usersAddresses[i])
		fmt.Printf("balance of user %s: %v\n", hex.EncodeToString(usersAddresses[i]), balanceOfUser)
		assert.Equal(t, big.NewInt(0), balanceOfUser)
	}

	for i := 0; i < noOfUsers; i++ {
		data := "joinGame@aaaa"

		fmt.Printf("==== Balance before: %d\n", vm.GetAccountsBalance(usersAddresses[i], testContext.Accounts))

		//contract call tx
		txRun := vm.CreateTx(
			t,
			usersAddresses[i],
			scAddressBytes,
			defaultUserNonce,
			transfer,
			gasPrice,
			gasLimit,
			data,
		)

		_, err = testContext.TxProcessor.ProcessTransaction(txRun)
		assert.Nil(t, err)

		newUserBalance := vm.GetAccountsBalance(usersAddresses[i], testContext.Accounts)
		fmt.Printf("==== Balance after: %d\n", newUserBalance)
		afterJoinUsersBalances[i] = newUserBalance
	}

	_, err = testContext.Accounts.Commit()
	assert.Nil(t, err)

	scAccount, _ := testContext.BlockchainHook.GetUserAccount(scAddressBytes)
	balanceOfSC := scAccount.GetBalance()
	fmt.Printf("balance of SC: %v\n", balanceOfSC)
	computedBalance := big.NewInt(0).Set(transfer)
	computedBalance.Mul(computedBalance, big.NewInt(int64(noOfUsers)))
	assert.Equal(t, computedBalance, balanceOfSC)

	//reward
	prize := big.NewInt(10)
	for i := 0; i < noOfUsers; i++ {
		data := "rewardAndSendToWallet@aaaa@" + hex.EncodeToString(usersAddresses[i]) + "@" + hex.EncodeToString(prize.Bytes())
		//contract call tx
		txRun := vm.CreateTx(
			t,
			senderAddressBytes,
			scAddressBytes,
			senderNonce,
			big.NewInt(0),
			gasPrice,
			gasLimit,
			data,
		)

		_, err = testContext.TxProcessor.ProcessTransaction(txRun)
		assert.Nil(t, err)

		senderNonce++
	}

	_, err = testContext.Accounts.Commit()
	assert.Nil(t, err)

	for i := 0; i < noOfUsers; i++ {
		existingUserBalance := vm.GetAccountsBalance(usersAddresses[i], testContext.Accounts)
		computedBalance = big.NewInt(0).Set(afterJoinUsersBalances[i])
		computedBalance.Add(computedBalance, prize)

		assert.Equal(t, computedBalance.Uint64(), existingUserBalance.Uint64())
	}

	transferredBack := big.NewInt(0).Set(prize)
	transferredBack.Mul(transferredBack, big.NewInt(int64(noOfUsers)))
	computedBalance = big.NewInt(0).Set(transfer)
	computedBalance.Mul(computedBalance, big.NewInt(int64(noOfUsers)))
	computedBalance.Sub(computedBalance, transferredBack)

	scAccount, _ = testContext.BlockchainHook.GetUserAccount(scAddressBytes)
	balanceOfSC = scAccount.GetBalance()
	fmt.Printf("balance of SC: %v\n", balanceOfSC)
	assert.Equal(t, computedBalance.Uint64(), balanceOfSC.Uint64())
}

func BenchmarkAgarioJoinGame(b *testing.B) {
	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(b, err)

	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(0)
	gasLimit := uint64(1000000)

	testContext := vm.CreatePreparedTxProcessorAndAccountsWithVMs(senderNonce, senderAddressBytes, senderBalance)
	defer testContext.Close()

	_ = iele.DeployContract(
		b,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		iele.CreateDeployTxData(scCode),
		testContext.TxProcessor,
		testContext.Accounts,
	)

	scAddressBytes, _ := testContext.BlockchainHook.NewAddress(senderAddressBytes, senderNonce, factory.IELEVirtualMachine)

	defaultUserNonce := uint64(10)
	defaultUserBalance := big.NewInt(10000000000)
	transfer := big.NewInt(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		userAddress := make([]byte, 32)
		_, _ = rand.Reader.Read(userAddress)
		_, _ = vm.CreateAccount(testContext.Accounts, userAddress, defaultUserNonce, defaultUserBalance)
		_, _ = testContext.Accounts.Commit()

		data := "joinGame@aaaa"

		txRun := vm.CreateTx(
			b,
			userAddress,
			scAddressBytes,
			defaultUserNonce,
			transfer,
			gasPrice,
			gasLimit,
			data,
		)

		b.StartTimer()
		_, _ = testContext.TxProcessor.ProcessTransaction(txRun)
	}
}
