package txsFee

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/stretchr/testify/require"
)

func TestRelayedScDeployShouldWork(t *testing.T) {
	t.Parallel()

	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789012")

	senderNonce := uint64(0)
	senderBalance := big.NewInt(0)
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(50000))

	scCode := arwen.GetSCCode("../arwen/testdata/misc/fib_arwen/output/fib_arwen.wasm")
	userTx := vm.CreateTransaction(senderNonce, big.NewInt(0), sndAddr, vm.CreateEmptyAddress(), gasPrice, gasLimit, []byte(arwen.CreateDeployTxData(scCode)))

	rtxData := prepareRelayerTxData(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	_, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalanceRelayer := big.NewInt(28440)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	// check balance inner tx sender
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, big.NewInt(0))
}
