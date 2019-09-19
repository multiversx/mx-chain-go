package wasm

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	vmLib "github.com/ElrondNetwork/elrond-go/core/libLocator"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestVmDeployWithFibbonacciAndExecute_WABT_Metered_Off(t *testing.T) {
	runWASMVMBenchmark2(t, "./fibonacci_ewasmified.wasm", "wabt", 100, 1, 1000, 1, "")
}

func TestVmDeployWithFibbonacciAndExecute_WABT_Metered_Uniform(t *testing.T) {
	runWASMVMBenchmark2(t, "./fibonacci_ewasmified.wasm", "wabt", 100, 1, 1000, 1, "wabt_uniform_1")
}

func TestVmDeployWithFibbonacciAndExecute_WABT_Metered_TestValues(t *testing.T) {
	runWASMVMBenchmark2(t, "./fibonacci_ewasmified.wasm", "wabt", 100, 1, 1000, 1, "wabt_testing")
}

func runWASMVMBenchmark2(
	t *testing.T,
	fileSC string,
	engine string,
	numRun int,
	testingValue uint64,
	gasLimit uint64,
	gasPrice uint64,
	gasCostsTable string,
) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(100000000)
	round := uint64(444)
	transferOnCalls := big.NewInt(1)

	scCode, err := ioutil.ReadFile(fileSC)
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)

	tx := vm.CreateTx(
		t,
		ownerAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		ownerNonce,
		transferOnCalls,
		1,
		2000,
		scCodeString+"@"+hex.EncodeToString(factory.HeraWABTVirtualMachine),
	)

	config := vmLib.WASMLibLocation() + ",engine=" + engine
	if gasCostsTable != "" {
		config += ",gas_costs_table_name=" + gasCostsTable
	}

	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithWASMVM(t, ownerNonce, ownerAddressBytes, ownerBalance, config)

	err = txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	scAddress, _ := hex.DecodeString("000000000000000002001a2983b179a480a60c4308da48f13b4480dbb4d33132")

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(10000000000))

	for i := 0; i < numRun; i++ {
		tx = &transaction.Transaction{
			Nonce:     aliceNonce,
			Value:     big.NewInt(0).SetUint64(testingValue),
			RcvAddr:   scAddress,
			SndAddr:   alice,
			GasPrice:  gasPrice,
			GasLimit:  gasLimit,
			Data:      "benchmark",
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
