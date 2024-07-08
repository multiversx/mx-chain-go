package multiShard

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNFTTransferAndUpdateOnOldTypeToken(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	enableEpochs := config.EnableEpochs{
		CheckCorrectTokenIDForTransferRoleEnableEpoch: 3,
		DisableExecByCallerEnableEpoch:                3,
		RefactorContextEnableEpoch:                    3,
		FailExecutionOnEveryAPIErrorEnableEpoch:       3,
		ManagedCryptoAPIsEnableEpoch:                  3,
		CheckFunctionArgumentEnableEpoch:              3,
		CheckExecuteOnReadOnlyEnableEpoch:             3,
		ESDTMetadataContinuousCleanupEnableEpoch:      3,
		FixOldTokenLiquidityEnableEpoch:               3,
		RuntimeMemStoreLimitEnableEpoch:               3,
		SetSenderInEeiOutputTransferEnableEpoch:       3,
		AlwaysSaveTokenMetaDataEnableEpoch:            3,
	}

	tokenID := []byte("MYNFT")
	sh0Addr := []byte("sender-shard0-567890123456789010")
	sh1Addr := []byte("receiver-shard1-7890123456789011")
	initialAttribute := []byte("initial attribute")
	newAttribute := []byte("new attribute")

	sh0Context, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, enableEpochs)
	require.Nil(t, err)
	defer sh0Context.Close()

	sh1Context, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, enableEpochs)
	require.Nil(t, err)
	defer sh1Context.Close()

	_, _ = vm.CreateAccount(sh1Context.Accounts, sh1Addr, 0, big.NewInt(1000000000))

	// create the nft and ensure that it exists on the account's trie and in the system account on shard 0
	utils.CreateAccountWithNFT(t, sh0Context.Accounts, sh0Addr, big.NewInt(100000000), tokenID, initialAttribute)
	utils.CheckESDTNFTBalance(t, sh0Context, sh0Addr, tokenID, 1, big.NewInt(1))
	utils.CheckESDTNFTBalance(t, sh0Context, core.SystemAccountAddress, tokenID, 1, big.NewInt(1))

	// set special roles
	utils.SetESDTRoles(t, sh0Context.Accounts, sh0Addr, tokenID, [][]byte{[]byte(core.ESDTRoleNFTUpdateAttributes)})

	// transfer sh0Addr -> sh1Addr
	transferToken(t, tokenID, sh0Context, sh0Addr, sh1Context, sh1Addr)

	// transfer sh1Addr -> sh0Addr
	transferToken(t, tokenID, sh1Context, sh1Addr, sh0Context, sh0Addr)

	// update attributes in shard 0
	updateAttributes(t, tokenID, newAttribute, sh0Context, sh0Addr)

	// change to epoch 3 as to activate the update fix
	epoch3Block := &block.Header{
		Epoch: 3,
	}
	sh0Context.EpochNotifier.CheckEpoch(epoch3Block)
	sh1Context.EpochNotifier.CheckEpoch(epoch3Block)

	// transfer sh0Addr -> sh1Addr
	transferToken(t, tokenID, sh0Context, sh0Addr, sh1Context, sh1Addr)

	// both system accounts should have the new attributes
	checkAttributes(t, sh0Context, core.SystemAccountAddress, tokenID, newAttribute)
	checkAttributes(t, sh1Context, core.SystemAccountAddress, tokenID, newAttribute)
}

func transferToken(
	tb testing.TB,
	tokenID []byte,
	senderContext *vm.VMTestContext,
	senderAddr []byte,
	destinationContext *vm.VMTestContext,
	destinationAddr []byte,
) {
	senderContext.CleanIntermediateTransactions(tb)
	destinationContext.CleanIntermediateTransactions(tb)

	senderAccnt, _ := senderContext.Accounts.LoadAccount(senderAddr)
	crossShardTransferTx := utils.CreateESDTNFTTransferTx(senderAccnt.GetNonce(), senderAddr, destinationAddr, tokenID, 1, big.NewInt(1), 10, 1000000, "")
	retCode, err := senderContext.TxProcessor.ProcessTransaction(crossShardTransferTx)
	require.Equal(tb, vmcommon.Ok, retCode)
	require.NoError(tb, err)

	scrs := senderContext.GetIntermediateTransactions(tb)

	// check the balances after the transfer, as well as the liquidity
	utils.ProcessSCRResult(tb, destinationContext, scrs[0], vmcommon.Ok, nil)
	utils.CheckESDTNFTBalance(tb, senderContext, senderAddr, tokenID, 1, big.NewInt(0))
	utils.CheckESDTNFTBalance(tb, destinationContext, destinationAddr, tokenID, 1, big.NewInt(1))
}

func updateAttributes(
	tb testing.TB,
	tokenID []byte,
	newAttributes []byte,
	context *vm.VMTestContext,
	addr []byte,
) {
	context.CleanIntermediateTransactions(tb)

	senderAccnt, _ := context.Accounts.LoadAccount(addr)
	tx := utils.CreateESDTNFTUpdateAttributesTx(senderAccnt.GetNonce(), addr, tokenID, 1, 1000000, newAttributes)
	retCode, err := context.TxProcessor.ProcessTransaction(tx)
	require.Equal(tb, vmcommon.Ok, retCode)
	require.NoError(tb, err)

	//check the attributes after update
	checkAttributes(tb, context, addr, tokenID, newAttributes)
	checkAttributes(tb, context, core.SystemAccountAddress, tokenID, newAttributes)
}

func checkAttributes(tb testing.TB, context *vm.VMTestContext, address []byte, tokenID []byte, expectedAttributes []byte) {
	esdtData, err := context.BlockchainHook.GetESDTToken(address, tokenID, 1)
	require.Nil(tb, err)
	assert.Equal(tb, expectedAttributes, esdtData.TokenMetaData.Attributes)
}
