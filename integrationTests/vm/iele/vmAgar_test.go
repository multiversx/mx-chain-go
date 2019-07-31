package mockVM

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/stretchr/testify/assert"
)

var agarioFile = "agarioV2.hex"

func TestDeployAgarioContract(t *testing.T) {
	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(1000000)

	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithIeleVM(t, senderNonce, senderAddressBytes, senderBalance)
	deployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		string(scCode),
		round,
		txProc,
		accnts,
	)

	destinationAddressBytes, _ := hex.DecodeString("000000000000000000002ad210b548f26776b8859b1fabdf8298d9ce0d973132")
	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
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
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(1000000)

	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithIeleVM(t, senderNonce, senderAddressBytes, senderBalance)
	deployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		string(scCode),
		round,
		txProc,
		accnts,
	)
	scAddressBytes, _ := hex.DecodeString("000000000000000000002ad210b548f26776b8859b1fabdf8298d9ce0d973132")

	userAddress := []byte("10000000000000000000000000000000")
	userNonce := uint64(10)
	userBalance := big.NewInt(100000000)
	_ = vm.CreateAccount(accnts, userAddress, userNonce, userBalance)
	_, _ = accnts.Commit()

	//balanceOf should return 0 for userAddress
	assert.Equal(t, big.NewInt(0), getIntValueFromSC(accnts, scAddressBytes, "balanceOf", userAddress))

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

	err = txProc.ProcessTransaction(txRun, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	assert.Equal(t, transfer, getIntValueFromSC(accnts, scAddressBytes, "balanceOf", userAddress))
}

func TestAgarioContractTopUpAnfWithdrawShouldWork(t *testing.T) {
	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(1000000)

	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithIeleVM(t, senderNonce, senderAddressBytes, senderBalance)
	deployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		string(scCode),
		round,
		txProc,
		accnts,
	)
	scAddressBytes, _ := hex.DecodeString("000000000000000000002ad210b548f26776b8859b1fabdf8298d9ce0d973132")

	userAddress := []byte("10000000000000000000000000000000")
	userNonce := uint64(10)
	userBalance := big.NewInt(100000000)
	_ = vm.CreateAccount(accnts, userAddress, userNonce, userBalance)
	_, _ = accnts.Commit()

	//balanceOf should return 0 for userAddress
	assert.Equal(t, big.NewInt(0), getIntValueFromSC(accnts, scAddressBytes, "balanceOf", userAddress))

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

	err = txProc.ProcessTransaction(txRun, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	assert.Equal(t, transfer, getIntValueFromSC(accnts, scAddressBytes, "balanceOf", userAddress))

	//withdraw
	withdraw := big.NewInt(49999)
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

	err = txProc.ProcessTransaction(txRun, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	newValue := big.NewInt(0).Set(transfer)
	newValue.Sub(newValue, withdraw)
	assert.Equal(t, newValue, getIntValueFromSC(accnts, scAddressBytes, "balanceOf", userAddress))
}

func TestAgarioContractJoinGameReward(t *testing.T) {
	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(0)
	gasLimit := uint64(1000000)

	txProc, accnts, blockchainHook := vm.CreatePreparedTxProcessorAndAccountsWithIeleVM(t, senderNonce, senderAddressBytes, senderBalance)
	deployContract(
		t,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		string(scCode),
		round,
		txProc,
		accnts,
	)
	scAddressBytes, _ := hex.DecodeString("000000000000000000002ad210b548f26776b8859b1fabdf8298d9ce0d973132")

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

		_ = vm.CreateAccount(accnts, userAddress, defaultUserNonce, defaultUserBalance)
		_, _ = accnts.Commit()

		usersAddresses[i] = userAddress
	}

	for i := 0; i < noOfUsers; i++ {
		//balanceOf should return 0 for userAddress
		balanceOfUser := getIntValueFromSC(accnts, scAddressBytes, "balanceOf", usersAddresses[i])
		fmt.Printf("balance of user %s: %v\n", hex.EncodeToString(usersAddresses[i]), balanceOfUser)
		assert.Equal(t, big.NewInt(0), balanceOfUser)
	}

	for i := 0; i < noOfUsers; i++ {
		data := "joinGame@aaaa"

		fmt.Printf("==== Balance before: %d\n", vm.GetAccountsBalance(usersAddresses[i], accnts))

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

		err = txProc.ProcessTransaction(txRun, round)
		assert.Nil(t, err)

		newUserBalance := vm.GetAccountsBalance(usersAddresses[i], accnts)
		fmt.Printf("==== Balance after: %d\n", newUserBalance)
		afterJoinUsersBalances[i] = newUserBalance
	}

	_, err = accnts.Commit()
	assert.Nil(t, err)

	balanceOfSC, _ := blockchainHook.GetBalance(scAddressBytes)
	fmt.Printf("balance of SC: %v\n", balanceOfSC)
	computedBalance := big.NewInt(0).Set(transfer)
	computedBalance.Mul(computedBalance, big.NewInt(int64(noOfUsers)))
	assert.Equal(t, computedBalance, balanceOfSC)

	//reward
	prize := big.NewInt(10)
	for i := 0; i < noOfUsers; i++ {
		data := fmt.Sprintf("rewardAndSendToWallet@aaaa@%s@%X", hex.EncodeToString(usersAddresses[i]), prize)
		//contract call tx
		txRun := vm.CreateTx(
			t,
			senderAddressBytes,
			scAddressBytes,
			defaultUserNonce,
			big.NewInt(0),
			gasPrice,
			gasLimit,
			data,
		)

		err = txProc.ProcessTransaction(txRun, round)
		assert.Nil(t, err)
	}

	_, err = accnts.Commit()
	assert.Nil(t, err)

	for i := 0; i < noOfUsers; i++ {
		existingUserBalance := vm.GetAccountsBalance(usersAddresses[i], accnts)
		computedBalance := big.NewInt(0).Set(afterJoinUsersBalances[i])
		computedBalance.Add(computedBalance, prize)

		assert.Equal(t, computedBalance, existingUserBalance)
	}

	transferredBack := big.NewInt(0).Set(prize)
	transferredBack.Mul(transferredBack, big.NewInt(int64(noOfUsers)))
	computedBalance = big.NewInt(0).Set(transfer)
	computedBalance.Mul(computedBalance, big.NewInt(int64(noOfUsers)))
	computedBalance.Sub(computedBalance, transferredBack)
	balanceOfSC, _ = blockchainHook.GetBalance(scAddressBytes)
	fmt.Printf("balance of SC: %v\n", balanceOfSC)
	assert.Equal(t, computedBalance, balanceOfSC)
}

func BenchmarkAgarioJoinGame(b *testing.B) {
	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(b, err)

	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(0)
	gasLimit := uint64(1000000)

	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithIeleVM(b, senderNonce, senderAddressBytes, senderBalance)
	deployContract(
		b,
		senderAddressBytes,
		senderNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		string(scCode),
		round,
		txProc,
		accnts,
	)
	scAddressBytes, _ := hex.DecodeString("000000000000000000002ad210b548f26776b8859b1fabdf8298d9ce0d973132")

	defaultUserNonce := uint64(10)
	defaultUserBalance := big.NewInt(10000000000)
	transfer := big.NewInt(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		userAddress := make([]byte, 32)
		_, _ = rand.Reader.Read(userAddress)
		_ = vm.CreateAccount(accnts, userAddress, defaultUserNonce, defaultUserBalance)
		_, _ = accnts.Commit()

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
		_ = txProc.ProcessTransaction(txRun, round)
	}
}

func getIntValueFromSC(accnts state.AccountsAdapter, scAddressBytes []byte, funcName string, args ...[]byte) *big.Int {
	ieleVM, _ := vm.CreateVMAndBlockchainHook(accnts)
	scgd, _ := smartContract.NewSCDataGetter(ieleVM)

	returnedVals, _ := scgd.Get(scAddressBytes, funcName, args...)
	return big.NewInt(0).SetBytes(returnedVals)
}
