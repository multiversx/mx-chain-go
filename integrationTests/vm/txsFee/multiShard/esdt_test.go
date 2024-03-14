package multiShard

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestESDTTransferShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789010")
	require.Equal(t, uint32(0), testContext.ShardCoordinator.ComputeId(sndAddr))

	rcvAddr := []byte("12345678901234567890123456789011")
	require.Equal(t, uint32(1), testContext.ShardCoordinator.ComputeId(rcvAddr))

	egldBalance := big.NewInt(100000000)
	esdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, egldBalance, token, 0, esdtBalance)

	gasPrice := uint64(10)
	gasLimit := uint64(40)
	tx := utils.CreateESDTTransferTx(0, sndAddr, rcvAddr, token, big.NewInt(100), gasPrice, gasLimit)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedReceiverBalance := big.NewInt(100)
	utils.CheckESDTBalance(t, testContext, rcvAddr, token, expectedReceiverBalance)
}

func TestMultiESDTNFTTransferViaRelayedV2(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	tokenID1 := []byte("MYNFT1")
	tokenID2 := []byte("MYNFT2")
	sh0Addr := []byte("12345678901234567890123456789010")
	sh1Addr := []byte("12345678901234567890123456789011")

	relayerSh0 := []byte("12345678901234567890123456789110")
	relayerSh1 := []byte("12345678901234567890123456789111")
	sh0Context, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer sh0Context.Close()

	sh1Context, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer sh1Context.Close()
	_, _ = vm.CreateAccount(sh1Context.Accounts, sh1Addr, 0, big.NewInt(10000000000))
	_, _ = vm.CreateAccount(sh0Context.Accounts, relayerSh0, 0, big.NewInt(1000000000))
	_, _ = vm.CreateAccount(sh1Context.Accounts, relayerSh1, 0, big.NewInt(1000000000))

	// create the nfts, add the liquidity to the system accounts and check for balances
	utils.CreateAccountWithESDTBalance(t, sh0Context.Accounts, sh0Addr, big.NewInt(100000000), tokenID1, 1, big.NewInt(1))
	utils.CreateAccountWithESDTBalance(t, sh0Context.Accounts, sh0Addr, big.NewInt(100000000), tokenID2, 1, big.NewInt(1))

	sh0Accnt, _ := sh0Context.Accounts.LoadAccount(sh0Addr)
	sh1Accnt, _ := sh1Context.Accounts.LoadAccount(sh1Addr)

	transfers := []*utils.TransferESDTData{
		{
			Token: tokenID1,
			Nonce: 1,
			Value: big.NewInt(1),
		},
		{
			Token: tokenID2,
			Nonce: 1,
			Value: big.NewInt(1),
		},
	}

	//
	// Step 1: transfer the NFTs sh0->sh1 via multi transfer with a shard 0 relayer
	//

	innerTx := utils.CreateMultiTransferTX(sh0Accnt.GetNonce(), sh0Addr, sh1Addr, 10, 10000000, transfers...)
	relayedTx := createRelayedV2FromInnerTx(0, relayerSh0, innerTx)

	retCode, err := sh0Context.TxProcessor.ProcessTransaction(relayedTx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.NoError(t, err)

	scrs := sh0Context.GetIntermediateTransactions(t)

	shard1Scr := scrs[0]
	for _, scr := range scrs {
		if scr.GetRcvAddr()[len(scr.GetRcvAddr())-1] == byte(0) {
			shard1Scr = scr
			break
		}
	}
	// check the balances after the transfer, as well as the liquidity
	utils.ProcessSCRResult(t, sh1Context, shard1Scr, vmcommon.Ok, nil)
	utils.CheckESDTNFTBalance(t, sh0Context, sh0Addr, tokenID1, 1, big.NewInt(0))
	utils.CheckESDTNFTBalance(t, sh0Context, core.SystemAccountAddress, tokenID1, 1, big.NewInt(0))
	utils.CheckESDTNFTBalance(t, sh1Context, sh1Addr, tokenID1, 1, big.NewInt(1))
	utils.CheckESDTNFTBalance(t, sh1Context, core.SystemAccountAddress, tokenID1, 1, big.NewInt(1))

	//
	// Step 2: transfer the NFTs sh1->sh0 via multi transfer with a shard 1 relayer
	//

	sh0Context.CleanIntermediateTransactions(t)
	sh1Context.CleanIntermediateTransactions(t)

	innerTx = utils.CreateMultiTransferTX(sh1Accnt.GetNonce(), sh1Addr, sh0Addr, 10, 10000000, transfers...)
	relayedTx = createRelayedV2FromInnerTx(0, relayerSh1, innerTx)

	retCode, err = sh1Context.TxProcessor.ProcessTransaction(relayedTx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.NoError(t, err)

	scrs = sh1Context.GetIntermediateTransactions(t)
	shard0Scr := scrs[0]
	for _, scr := range scrs {
		if scr.GetRcvAddr()[len(scr.GetRcvAddr())-1] == byte(0) {
			shard0Scr = scr
			break
		}
	}
	// check the balances after the transfer, as well as the liquidity
	utils.ProcessSCRResult(t, sh0Context, shard0Scr, vmcommon.Ok, nil)
	utils.CheckESDTNFTBalance(t, sh0Context, sh0Addr, tokenID1, 1, big.NewInt(1))
	utils.CheckESDTNFTBalance(t, sh0Context, core.SystemAccountAddress, tokenID1, 1, big.NewInt(1))
	utils.CheckESDTNFTBalance(t, sh1Context, sh1Addr, tokenID1, 1, big.NewInt(0))
	utils.CheckESDTNFTBalance(t, sh1Context, core.SystemAccountAddress, tokenID1, 1, big.NewInt(0))

	//
	// Step 3: transfer the NFTs sh0->s1 via multi transfer with a shard 1 relayer
	//

	sh0Context.CleanIntermediateTransactions(t)
	sh1Context.CleanIntermediateTransactions(t)

	innerTx = utils.CreateMultiTransferTX(sh0Accnt.GetNonce()+1, sh0Addr, sh1Addr, 10, 10000000, transfers...)
	relayedTx = createRelayedV2FromInnerTx(1, relayerSh1, innerTx)

	retCode, err = sh0Context.TxProcessor.ProcessTransaction(relayedTx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.NoError(t, err)

	scrs = sh0Context.GetIntermediateTransactions(t)
	shard1Scr = scrs[0]
	for _, scr := range scrs {
		if scr.GetRcvAddr()[len(scr.GetRcvAddr())-1] == byte(1) {
			shard1Scr = scr
			break
		}
	}
	// check the balances after the transfer, as well as the liquidity
	utils.ProcessSCRResult(t, sh1Context, shard1Scr, vmcommon.Ok, nil)
	utils.CheckESDTNFTBalance(t, sh0Context, sh0Addr, tokenID1, 1, big.NewInt(0))
	utils.CheckESDTNFTBalance(t, sh0Context, core.SystemAccountAddress, tokenID1, 1, big.NewInt(0))
	utils.CheckESDTNFTBalance(t, sh1Context, sh1Addr, tokenID1, 1, big.NewInt(1))
	utils.CheckESDTNFTBalance(t, sh1Context, core.SystemAccountAddress, tokenID1, 1, big.NewInt(1))
}

func createRelayedV2FromInnerTx(relayerNonce uint64, relayer []byte, innerTx *transaction.Transaction) *transaction.Transaction {
	nonceHex := "00"
	if innerTx.Nonce > 0 {
		nonceHex = hex.EncodeToString(big.NewInt(int64(innerTx.Nonce)).Bytes())
	}
	data := strings.Join([]string{"relayedTxV2", hex.EncodeToString(innerTx.RcvAddr), nonceHex, hex.EncodeToString(innerTx.Data), hex.EncodeToString(innerTx.Signature)}, "@")
	return &transaction.Transaction{
		Nonce:    relayerNonce,
		Value:    big.NewInt(0),
		SndAddr:  relayer,
		RcvAddr:  innerTx.SndAddr,
		GasPrice: innerTx.GasPrice,
		GasLimit: innerTx.GasLimit + 1_000_000,
		Data:     []byte(data),
	}
}
