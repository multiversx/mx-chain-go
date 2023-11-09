//go:build !race

package multiShard

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestRelayedSCDeployShouldWork(t *testing.T) {
	testContextRelayer, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(2, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextRelayer.Close()

	testContextInner, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextInner.Close()

	relayerAddr := []byte("12345678901234567890123456789032")
	require.Equal(t, uint32(2), testContextRelayer.ShardCoordinator.ComputeId(relayerAddr))

	sndAddr := []byte("12345678901234567890123456789011")
	require.Equal(t, uint32(1), testContextRelayer.ShardCoordinator.ComputeId(sndAddr))

	gasPrice := uint64(10)
	gasLimit := uint64(1961)

	_, _ = vm.CreateAccount(testContextRelayer.Accounts, relayerAddr, 0, big.NewInt(100000))

	contractPath := "../../wasm/testdata/misc/fib_wasm/output/fib_wasm.wasm"
	scCode := wasm.GetSCCode(contractPath)
	userTx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, vm.CreateEmptyAddress(), gasPrice, gasLimit, []byte(wasm.CreateDeployTxData(scCode)))

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	// execute on relayer shard
	retCode, err := testContextRelayer.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContextRelayer.Accounts.Commit()
	require.Nil(t, err)

	expectedBalanceRelayer := big.NewInt(52520)
	utils.TestAccount(t, testContextRelayer.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	// check accumulated fees
	accumulatedFees := testContextRelayer.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(27870), accumulatedFees)

	// execute on inner tx destination
	retCode, err = testContextInner.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContextInner.Accounts.Commit()
	require.Nil(t, err)

	expectedBalanceRelayer = big.NewInt(0)
	utils.TestAccount(t, testContextInner.Accounts, sndAddr, 1, expectedBalanceRelayer)

	// check accumulated fees
	accumulatedFees = testContextInner.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(19600), accumulatedFees)

	txs := testContextInner.GetIntermediateTransactions(t)

	scr := txs[2]
	utils.ProcessSCRResult(t, testContextRelayer, scr, vmcommon.UserError, nil)
}
