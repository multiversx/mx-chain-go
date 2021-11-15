package arwenvm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestStorageForDistribution(t *testing.T) {
	// Only a test to benchmark storage on distribution contract
	//t.Skip()

	ownerAddressBytes := []byte("12345678901234567890123456789011")
	ownerNonce := uint64(0)
	ownerBalance := big.NewInt(0).Mul(big.NewInt(999999999999999999), big.NewInt(999999999999999999))
	gasPrice := uint64(1)
	gasLimit := uint64(500000000)

	scCode := arwen.GetSCCode("../testdata/distribution.wasm")

	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV4.toml")
	testContext, err := vm.CreateTxProcessorArwenVMWithGasSchedule(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		gasSchedule,
		vm.ArgEnableEpoch{},
	)
	require.Nil(t, err)
	defer testContext.Close()

	anotherAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, 100, factory.ArwenVirtualMachine)
	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		arwen.CreateDeployTxData(scCode)+"@"+hex.EncodeToString([]byte("TOKEN-010101"))+"@"+hex.EncodeToString(anotherAddress),
	)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)
	ownerNonce++

	tx = vm.CreateTransaction(ownerNonce, big.NewInt(0), ownerAddressBytes, scAddress, gasPrice, 10000000, []byte("startGlobalOperation"))
	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)
	ownerNonce++

	tx = vm.CreateTransaction(ownerNonce, big.NewInt(0), ownerAddressBytes, scAddress, gasPrice, 10000000, []byte("setCommunityDistribution@0781A4DA9A1E3D1C0007271E30@01ce"))
	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)
	ownerNonce++

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	numAddresses := 250000
	testAddresses := createTestAddresses(uint64(numAddresses))
	fmt.Println("SETUP DONE")

	valueToDistribute := big.NewInt(0).Mul(big.NewInt(100000000000), big.NewInt(100000000000))

	totalSteps := numAddresses / 250
	fmt.Printf("Need to process %d transactions \n", totalSteps)
	for i := 0; i < numAddresses/250; i++ {
		start := time.Now()
		txData := "setPerUserDistributedLockedAssets@01ce"

		for j := i * 250; j < (i+1)*250; j++ {
			txData += "@" + hex.EncodeToString(testAddresses[j]) + "@" + hex.EncodeToString(valueToDistribute.Bytes())
		}

		tx = vm.CreateTransaction(ownerNonce, big.NewInt(0), ownerAddressBytes, scAddress, gasPrice, 1000000000, []byte(txData))

		returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, err)
		require.Equal(t, returnCode, vmcommon.Ok)
		ownerNonce++

		elapsedTime := time.Since(start)
		fmt.Printf("ID: %d, time elapsed to process 300 distribution %s \n", i, elapsedTime.String())

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)
	}

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)
}
