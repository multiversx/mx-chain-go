//go:build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

package txsFee

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestRelayedTxDnsTransaction_ShouldWork(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeployDNS(t, testContext, "../../multiShard/smartContract/dns/dns.wasm")
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789112")
	rcvAddr := []byte("12345678901234567890123456789110")
	gasLimit := uint64(500000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, rcvAddr, 0, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(100000000))

	sndAddrUserName := utils.GenerateUserNameForDefaultDNSContract()
	txData := []byte("register@" + hex.EncodeToString(sndAddrUserName))
	// create user name for sender
	innerTx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, txData)

	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	ret := vm.GetVmOutput(nil, testContext.Accounts, scAddress, "resolve", sndAddrUserName)
	dnsUserNameAddr := ret.ReturnData[0]
	require.Equal(t, sndAddr, dnsUserNameAddr)

	rcvAddrUserName := utils.GenerateUserNameForDefaultDNSContract()
	txData = []byte("register@" + hex.EncodeToString(rcvAddrUserName))
	// create user name for receiver
	innerTx = vm.CreateTransaction(0, big.NewInt(0), rcvAddr, scAddress, gasPrice, gasLimit, txData)

	rtxData = integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit = 1 + gasLimit + uint64(len(rtxData))
	rtx = vm.CreateTransaction(1, innerTx.Value, relayerAddr, rcvAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err = testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	ret = vm.GetVmOutput(nil, testContext.Accounts, scAddress, "resolve", rcvAddrUserName)
	dnsUserNameAddr = ret.ReturnData[0]
	require.Equal(t, rcvAddr, dnsUserNameAddr)

	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	gasLimit = 10
	innerTx = vm.CreateTransaction(1, big.NewInt(0), sndAddr, rcvAddr, gasPrice, gasLimit, nil)
	innerTx.SndUserName = sndAddrUserName
	innerTx.RcvUserName = rcvAddrUserName

	rtxData = integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit = 1 + gasLimit + uint64(len(rtxData))
	rtx = vm.CreateTransaction(2, innerTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err = testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)
}
