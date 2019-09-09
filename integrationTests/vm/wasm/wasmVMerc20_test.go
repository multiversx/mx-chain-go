package wasm

import (
	"encoding/hex"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/stretchr/testify/assert"
)

var erc20wasm = "./main_ewasmified.wasm"
var benchmarks = "./stringconcat_ewasmified.wasm"
var agarioFile = "../../agar_v1_min.hex"

func TestVmDeployWithTransferAndGasShouldDeploySCCode(t *testing.T) {
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	scCode, err := ioutil.ReadFile(benchmarks)
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)

	tx := vm.CreateTx(
		t,
		senderAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString,
	)

	config := "./libhera.so,engine=wavm"
	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithWASMVM(t, senderNonce, senderAddressBytes, senderBalance, config)

	err = txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	expectedBalance := big.NewInt(99999139)
	fmt.Printf("%s \n", hex.EncodeToString(expectedBalance.Bytes()))

	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		expectedBalance)

	fmt.Printf("%s \n", hex.EncodeToString(keccak.Keccak{}.Compute("transfer")))
	fmt.Printf("%s \n", hex.EncodeToString(keccak.Keccak{}.Compute("balance")))
	fmt.Printf("%s \n", hex.EncodeToString(keccak.Keccak{}.Compute("topUp")))
	fmt.Printf("%s \n", hex.EncodeToString(keccak.Keccak{}.Compute("joinGame")))
	fmt.Printf("%s \n", hex.EncodeToString(keccak.Keccak{}.Compute("rewardAndSendToWallet")))
}

func TestVmDeployWithTransferAndExecute(t *testing.T) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(1)

	scCode, err := ioutil.ReadFile(benchmarks)
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)

	tx := vm.CreateTx(
		t,
		ownerAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString,
	)

	config := "./libhera.so,engine=wavm"
	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithWASMVM(t, ownerNonce, ownerAddressBytes, ownerBalance, config)

	err = txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	scAddress, _ := hex.DecodeString("0000000000000000000061356266303466386236353138373434383437633132")

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(10000000000))

	bob := []byte("12345678901234567890123456789222")
	_ = vm.CreateAccount(accnts, bob, 0, big.NewInt(1000000))

	for i := 0; i < 1000; i++ {
		tx = &transaction.Transaction{
			Nonce:     aliceNonce,
			Value:     big.NewInt(10000),
			RcvAddr:   scAddress,
			SndAddr:   alice,
			GasPrice:  0,
			GasLimit:  5000,
			Data:      "joinGame@111",
			Signature: nil,
			Challenge: nil,
		}

		startTime := time.Now()
		err = txProc.ProcessTransaction(tx, round)
		elapsedTime := time.Since(startTime)
		fmt.Printf("time elapsed full process %s \n", elapsedTime.String())
		assert.Nil(t, err)

		_, err = accnts.Commit()
		assert.Nil(t, err)

		aliceNonce++
	}
}

func TestVmDeployWithTransferAndExecuteIele(t *testing.T) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(500000)

	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	scCodeString := string(scCode)

	tx := vm.CreateTx(
		t,
		ownerAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString,
	)

	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithIeleVM(t, ownerNonce, ownerAddressBytes, ownerBalance)

	err = txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	scAddress, _ := hex.DecodeString("000000000000000000002ad210b548f26776b8859b1fabdf8298d9ce0d973132")

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(1000000))

	bob := []byte("12345678901234567890123456789222")
	_ = vm.CreateAccount(accnts, bob, 0, big.NewInt(1000000))

	startTime := time.Now()
	for i := 0; i < 10000; i++ {
		tx = &transaction.Transaction{
			Nonce:     aliceNonce,
			Value:     big.NewInt(50),
			RcvAddr:   scAddress,
			SndAddr:   alice,
			GasPrice:  0,
			GasLimit:  5000,
			Data:      "joinGame@111",
			Signature: nil,
			Challenge: nil,
		}

		err = txProc.ProcessTransaction(tx, round)
		assert.Nil(t, err)

		_, err = accnts.Commit()
		assert.Nil(t, err)

		aliceNonce++
	}

	elapsedTime := time.Since(startTime)
	fmt.Printf("time elapsed %s \n", elapsedTime.String())
}

func TestVmDeployWithTransferAndExecuteERC20(t *testing.T) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(5)

	scCode, err := ioutil.ReadFile(erc20wasm)
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)

	tx := vm.CreateTx(
		t,
		ownerAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString,
	)

	config := "./libhera.so,engine=wabt"
	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithWASMVM(t, ownerNonce, ownerAddressBytes, ownerBalance, config)

	err = txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	scAddress, _ := hex.DecodeString("0000000000000000000061356266303466386236353138373434383437633132")

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(1000000))

	bob := []byte("12345678901234567890123456789222")
	_ = vm.CreateAccount(accnts, bob, 0, big.NewInt(1000000))

	tx = &transaction.Transaction{
		Nonce:     aliceNonce,
		Value:     big.NewInt(100000),
		RcvAddr:   scAddress,
		SndAddr:   alice,
		GasPrice:  0,
		GasLimit:  5000,
		Data:      "topUp@50",
		Signature: nil,
		Challenge: nil,
	}
	start := time.Now()
	err = txProc.ProcessTransaction(tx, round)
	elapsedTime := time.Since(start)
	fmt.Printf("time elapsed to process topup %s \n", elapsedTime.String())
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	aliceNonce++

	start = time.Now()
	for i := 0; i < 100000; i++ {
		tx = &transaction.Transaction{
			Nonce:     aliceNonce,
			Value:     big.NewInt(5),
			RcvAddr:   scAddress,
			SndAddr:   alice,
			GasPrice:  0,
			GasLimit:  5000,
			Data:      "transfer@" + string(bob) + "@5",
			Signature: nil,
			Challenge: nil,
		}

		err = txProc.ProcessTransaction(tx, round)
		assert.Nil(t, err)

		aliceNonce++
	}

	_, err = accnts.Commit()
	assert.Nil(t, err)

	elapsedTime = time.Since(start)
	fmt.Printf("time elapsed to process 100000 ERC20 transfers %s \n", elapsedTime.String())
}
