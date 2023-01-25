package vmDeploy

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVmDeployWithoutTransferShouldDeploySCCode(t *testing.T) {
	vmOpGas := uint64(1)
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000000)
	gasPrice := uint64(1)
	gasLimit := vmOpGas + 100
	transferOnCalls := big.NewInt(0)

	initialValueForInternalVariable := uint64(45)
	scCode := fmt.Sprintf("aaaa@%s@0000@%X", hex.EncodeToString(factory.InternalTestingVM), initialValueForInternalVariable)

	tx := vm.CreateTx(
		senderAddressBytes,
		vm.CreateEmptyAddress(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
	)

	txProc, accnts, err := vm.CreatePreparedTxProcessorAndAccountsWithMockedVM(
		vmOpGas,
		senderNonce,
		senderAddressBytes,
		senderBalance,
		config.EnableEpochs{},
		&sync.RWMutex{},
	)
	require.Nil(t, err)

	_, err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		vm.ComputeExpectedBalance(senderBalance, transferOnCalls, gasLimit, gasPrice))
	destinationAddressBytes, _ := hex.DecodeString("0000000000000000ffff1a2983b179a480a60c4308da48f13b4480dbb4d33132")

	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		scCode,
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable)})
}

func TestVmDeployWithTransferShouldDeploySCCode(t *testing.T) {
	vmOpGas := uint64(1)
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := vmOpGas + 1000
	transferOnCalls := big.NewInt(50)

	initialValueForInternalVariable := uint64(45)
	scCode := fmt.Sprintf("aaaa@%s@0000@%X", hex.EncodeToString(factory.InternalTestingVM), initialValueForInternalVariable)

	tx := vm.CreateTx(
		senderAddressBytes,
		vm.CreateEmptyAddress(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
	)

	txProc, accnts, err := vm.CreatePreparedTxProcessorAndAccountsWithMockedVM(
		vmOpGas,
		senderNonce,
		senderAddressBytes,
		senderBalance,
		config.EnableEpochs{},
		&sync.RWMutex{},
	)
	require.Nil(t, err)

	_, err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		vm.ComputeExpectedBalance(senderBalance, transferOnCalls, gasLimit, gasPrice))
	destinationAddressBytes, _ := hex.DecodeString("0000000000000000ffff1a2983b179a480a60c4308da48f13b4480dbb4d33132")
	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		scCode,
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable)})
}

func TestVmDeployWithTransferAndGasShouldDeploySCCode(t *testing.T) {
	vmOpGas := uint64(1)
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	// equal with requirement
	gasLimit := vmOpGas + 100
	transferOnCalls := big.NewInt(50)

	initialValueForInternalVariable := uint64(45)
	scCode := fmt.Sprintf("aaaa@%s@0000@%X", hex.EncodeToString(factory.InternalTestingVM), initialValueForInternalVariable)

	tx := vm.CreateTx(
		senderAddressBytes,
		vm.CreateEmptyAddress(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
	)

	txProc, accnts, err := vm.CreatePreparedTxProcessorAndAccountsWithMockedVM(
		vmOpGas,
		senderNonce,
		senderAddressBytes,
		senderBalance,
		config.EnableEpochs{},
		&sync.RWMutex{},
	)
	require.Nil(t, err)

	_, err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		vm.ComputeExpectedBalance(senderBalance, transferOnCalls, gasLimit, gasPrice))

	destinationAddressBytes, _ := hex.DecodeString("0000000000000000ffff1a2983b179a480a60c4308da48f13b4480dbb4d33132")
	vm.TestDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		scCode,
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable)})

}

func TestVMDeployWithTransferWithInsufficientGasShouldReturnErr(t *testing.T) {
	vmOpGas := uint64(1000)
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	// slightly less than requirement
	gasLimit := vmOpGas - 1
	transferOnCalls := big.NewInt(50)

	initialValueForInternalVariable := uint64(45)
	scCode := fmt.Sprintf("aaaa@%s@@0000%X", hex.EncodeToString(factory.InternalTestingVM), initialValueForInternalVariable)

	tx := vm.CreateTx(
		senderAddressBytes,
		vm.CreateEmptyAddress(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
	)

	txProc, accnts, err := vm.CreatePreparedTxProcessorAndAccountsWithMockedVM(
		vmOpGas,
		senderNonce,
		senderAddressBytes,
		senderBalance,
		config.EnableEpochs{},
		&sync.RWMutex{},
	)
	require.Nil(t, err)

	_, err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		// the transfer should get back to the sender as the tx failed
		vm.ComputeExpectedBalance(senderBalance, big.NewInt(0), gasLimit, gasPrice))
	destinationAddressBytes, _ := hex.DecodeString("0000000000000000ffff1a2983b179a480a60c4308da48f13b4480dbb4d33132")

	assert.False(t, vm.AccountExists(accnts, destinationAddressBytes))
}
