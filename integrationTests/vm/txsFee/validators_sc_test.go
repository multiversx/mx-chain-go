package txsFee

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/txsFee/utils"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	vmAddr "github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	"github.com/stretchr/testify/require"
)

func TestValidatorsSC_DoStakePutInQueueUnStakeAndUnBondShouldRefund(t *testing.T) {
	testContextMeta, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, vm.ArgEnableEpoch{})

	require.Nil(t, err)
	defer testContextMeta.Close()

	saveNodesConfig(t, testContextMeta, 1, 1, 1)

	testContextMeta.BlockchainHook.(*hooks.BlockChainHookImpl).SetCurrentHeader(&block.MetaBlock{Nonce: 10000})

	gasPrice := uint64(10)
	gasLimit := uint64(4000)
	sndAddr := []byte("12345678901234567890123456789012")
	stakeValue, _ := big.NewInt(0).SetString("2500000000000000000000", 10)
	tx := vm.CreateTransaction(0, stakeValue, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("stake@01@4fcdbfce9a3621621d388019353f87aceb0c5ec826256bc5a57220ed8f7d84b5c13f50219dc7a6ef090671f39d398c106f8a36022a04e6c18061ff38629d5b42b125aaea1e94a97b5f1c871bcf93b1f141a49ddb0b4c32b976cd530c1008da86@0b823739887c40e9331f70c5a140623dfaf4558a9138b62f4473b26bbafdd4f58cb5889716a71c561c9e20e7a280e985@b2a11555ce521e4944e09ab17549d85b487dcd26c84b5017a39e31a3670889ba"))

	recCode, err := testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs := testContextMeta.GetIntermediateTransactions(t)
	scrWithMessage := intermediateTxs[1].(*smartContractResult.SmartContractResult)
	// validator is in waiting queue
	require.True(t, strings.Contains(string(scrWithMessage.ReturnMessage), "staking is full key put into waiting list "))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unStake@4fcdbfce9a3621621d388019353f87aceb0c5ec826256bc5a57220ed8f7d84b5c13f50219dc7a6ef090671f39d398c106f8a36022a04e6c18061ff38629d5b42b125aaea1e94a97b5f1c871bcf93b1f141a49ddb0b4c32b976cd530c1008da86"))

	recCode, err = testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 2, len(intermediateTxs))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBond@4fcdbfce9a3621621d388019353f87aceb0c5ec826256bc5a57220ed8f7d84b5c13f50219dc7a6ef090671f39d398c106f8a36022a04e6c18061ff38629d5b42b125aaea1e94a97b5f1c871bcf93b1f141a49ddb0b4c32b976cd530c1008da86"))

	recCode, err = testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 3, len(intermediateTxs))

	scr := intermediateTxs[1].(*smartContractResult.SmartContractResult)
	require.Equal(t, stakeValue, scr.Value)
}

func TestValidatorsSC_DoStakePutInQueueUnStakeAndUnBondTokensShouldRefund(t *testing.T) {
	testContextMeta, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, vm.ArgEnableEpoch{})

	require.Nil(t, err)
	defer testContextMeta.Close()

	saveNodesConfig(t, testContextMeta, 1, 1, 1)

	testContextMeta.BlockchainHook.(*hooks.BlockChainHookImpl).SetCurrentHeader(&block.MetaBlock{Nonce: 10000})

	gasPrice := uint64(10)
	gasLimit := uint64(4000)
	sndAddr := []byte("12345678901234567890123456789012")
	stakeValue, _ := big.NewInt(0).SetString("2500000000000000000000", 10)
	tx := vm.CreateTransaction(0, stakeValue, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("stake@01@4fcdbfce9a3621621d388019353f87aceb0c5ec826256bc5a57220ed8f7d84b5c13f50219dc7a6ef090671f39d398c106f8a36022a04e6c18061ff38629d5b42b125aaea1e94a97b5f1c871bcf93b1f141a49ddb0b4c32b976cd530c1008da86@0b823739887c40e9331f70c5a140623dfaf4558a9138b62f4473b26bbafdd4f58cb5889716a71c561c9e20e7a280e985@b2a11555ce521e4944e09ab17549d85b487dcd26c84b5017a39e31a3670889ba"))

	recCode, err := testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs := testContextMeta.GetIntermediateTransactions(t)
	scrWithMessage := intermediateTxs[1].(*smartContractResult.SmartContractResult)
	// validator is in waiting queue
	require.True(t, strings.Contains(string(scrWithMessage.ReturnMessage), "staking is full key put into waiting list "))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unStake@4fcdbfce9a3621621d388019353f87aceb0c5ec826256bc5a57220ed8f7d84b5c13f50219dc7a6ef090671f39d398c106f8a36022a04e6c18061ff38629d5b42b125aaea1e94a97b5f1c871bcf93b1f141a49ddb0b4c32b976cd530c1008da86"))

	recCode, err = testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 2, len(intermediateTxs))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBondTokens@"+hex.EncodeToString(stakeValue.Bytes())))

	recCode, err = testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, recCode)
	require.Nil(t, err)

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediateTxs))

	scrWithMessage = intermediateTxs[0].(*smartContractResult.SmartContractResult)
	require.True(t, strings.Contains(string(scrWithMessage.ReturnMessage), "cannot unBond tokens, the validator would remain without min deposit, nodes are still active"))
}

func TestValidatorsSC_DoStakeWithTopUpValueTryToUnStakeTokensAndUnBondTokens(t *testing.T) {
	testContextMeta, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, vm.ArgEnableEpoch{})

	require.Nil(t, err)
	defer testContextMeta.Close()

	saveNodesConfig(t, testContextMeta, 1, 1, 1)

	testContextMeta.BlockchainHook.(*hooks.BlockChainHookImpl).SetCurrentHeader(&block.MetaBlock{Nonce: 10000})

	gasPrice := uint64(10)
	gasLimit := uint64(4000)
	sndAddr := []byte("12345678901234567890123456789012")
	stakeValue, _ := big.NewInt(0).SetString("2700000000000000000000", 10)
	tx := vm.CreateTransaction(0, stakeValue, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("stake@01@4fcdbfce9a3621621d388019353f87aceb0c5ec826256bc5a57220ed8f7d84b5c13f50219dc7a6ef090671f39d398c106f8a36022a04e6c18061ff38629d5b42b125aaea1e94a97b5f1c871bcf93b1f141a49ddb0b4c32b976cd530c1008da86@0b823739887c40e9331f70c5a140623dfaf4558a9138b62f4473b26bbafdd4f58cb5889716a71c561c9e20e7a280e985@b2a11555ce521e4944e09ab17549d85b487dcd26c84b5017a39e31a3670889ba"))

	recCode, err := testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs := testContextMeta.GetIntermediateTransactions(t)
	scrWithMessage := intermediateTxs[1].(*smartContractResult.SmartContractResult)
	// validator is in waiting queue
	require.True(t, strings.Contains(string(scrWithMessage.ReturnMessage), "staking is full key put into waiting list "))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	topUpValueStaked, _ := big.NewInt(0).SetString("200000000000000000000", 10)
	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unStakeTokens@"+hex.EncodeToString(topUpValueStaked.Bytes())))

	recCode, err = testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediateTxs))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBondTokens@"+hex.EncodeToString(topUpValueStaked.Bytes())))

	recCode, err = testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediateTxs))

	scrWithMessage = intermediateTxs[0].(*smartContractResult.SmartContractResult)
	require.True(t, strings.Contains(string(scrWithMessage.ReturnMessage), "no tokens that can be unbond at this time"))
}

func TestValidatorsSC_ToStakePutInQueueUnStakeAndUnBondShouldRefund_UnBondTokens(t *testing.T) {
	testContextMeta, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, vm.ArgEnableEpoch{})

	require.Nil(t, err)
	defer testContextMeta.Close()

	saveNodesConfig(t, testContextMeta, 1, 1, 1)

	testContextMeta.BlockchainHook.(*hooks.BlockChainHookImpl).SetCurrentHeader(&block.MetaBlock{Nonce: 10000})

	gasPrice := uint64(10)
	gasLimit := uint64(4000)
	sndAddr := []byte("12345678901234567890123456789012")
	stakeValue, _ := big.NewInt(0).SetString("2700000000000000000000", 10)
	tx := vm.CreateTransaction(0, stakeValue, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("stake@01@4fcdbfce9a3621621d388019353f87aceb0c5ec826256bc5a57220ed8f7d84b5c13f50219dc7a6ef090671f39d398c106f8a36022a04e6c18061ff38629d5b42b125aaea1e94a97b5f1c871bcf93b1f141a49ddb0b4c32b976cd530c1008da86@0b823739887c40e9331f70c5a140623dfaf4558a9138b62f4473b26bbafdd4f58cb5889716a71c561c9e20e7a280e985@b2a11555ce521e4944e09ab17549d85b487dcd26c84b5017a39e31a3670889ba"))

	recCode, err := testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs := testContextMeta.GetIntermediateTransactions(t)
	scrWithMessage := intermediateTxs[1].(*smartContractResult.SmartContractResult)
	// validator is in waiting queue
	require.True(t, strings.Contains(string(scrWithMessage.ReturnMessage), "staking is full key put into waiting list "))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unStake@4fcdbfce9a3621621d388019353f87aceb0c5ec826256bc5a57220ed8f7d84b5c13f50219dc7a6ef090671f39d398c106f8a36022a04e6c18061ff38629d5b42b125aaea1e94a97b5f1c871bcf93b1f141a49ddb0b4c32b976cd530c1008da86"))

	recCode, err = testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 2, len(intermediateTxs))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBond@4fcdbfce9a3621621d388019353f87aceb0c5ec826256bc5a57220ed8f7d84b5c13f50219dc7a6ef090671f39d398c106f8a36022a04e6c18061ff38629d5b42b125aaea1e94a97b5f1c871bcf93b1f141a49ddb0b4c32b976cd530c1008da86"))

	recCode, err = testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 3, len(intermediateTxs))

	scr := intermediateTxs[1].(*smartContractResult.SmartContractResult)
	stakeValue, _ = big.NewInt(0).SetString("2500000000000000000000", 10)
	require.Equal(t, stakeValue, scr.Value)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	topUpValueStaked, _ := big.NewInt(0).SetString("200000000000000000000", 10)
	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unStakeTokens@"+hex.EncodeToString(topUpValueStaked.Bytes())))

	recCode, err = testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediateTxs))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBondTokens@"+hex.EncodeToString(topUpValueStaked.Bytes())))

	recCode, err = testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 2, len(intermediateTxs))

	scrWithMessage = intermediateTxs[0].(*smartContractResult.SmartContractResult)
	require.Equal(t, topUpValueStaked, scrWithMessage.Value)
}

func TestValidatorsSC_ToStakePutInQueueUnStakeNodesAndUnBondNodesShouldRefund(t *testing.T) {
	testContextMeta, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, vm.ArgEnableEpoch{})

	require.Nil(t, err)
	defer testContextMeta.Close()

	saveNodesConfig(t, testContextMeta, 1, 1, 1)

	testContextMeta.BlockchainHook.(*hooks.BlockChainHookImpl).SetCurrentHeader(&block.MetaBlock{Nonce: 10000})

	gasPrice := uint64(10)
	gasLimit := uint64(4000)
	sndAddr := []byte("12345678901234567890123456789012")
	stakeValue, _ := big.NewInt(0).SetString("2700000000000000000000", 10)
	tx := vm.CreateTransaction(0, stakeValue, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("stake@01@4fcdbfce9a3621621d388019353f87aceb0c5ec826256bc5a57220ed8f7d84b5c13f50219dc7a6ef090671f39d398c106f8a36022a04e6c18061ff38629d5b42b125aaea1e94a97b5f1c871bcf93b1f141a49ddb0b4c32b976cd530c1008da86@0b823739887c40e9331f70c5a140623dfaf4558a9138b62f4473b26bbafdd4f58cb5889716a71c561c9e20e7a280e985@b2a11555ce521e4944e09ab17549d85b487dcd26c84b5017a39e31a3670889ba"))

	recCode, err := testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs := testContextMeta.GetIntermediateTransactions(t)
	scrWithMessage := intermediateTxs[1].(*smartContractResult.SmartContractResult)
	// validator is in waiting queue
	require.True(t, strings.Contains(string(scrWithMessage.ReturnMessage), "staking is full key put into waiting list "))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unStakeNodes@4fcdbfce9a3621621d388019353f87aceb0c5ec826256bc5a57220ed8f7d84b5c13f50219dc7a6ef090671f39d398c106f8a36022a04e6c18061ff38629d5b42b125aaea1e94a97b5f1c871bcf93b1f141a49ddb0b4c32b976cd530c1008da86"))

	recCode, err = testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 2, len(intermediateTxs))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBondNodes@4fcdbfce9a3621621d388019353f87aceb0c5ec826256bc5a57220ed8f7d84b5c13f50219dc7a6ef090671f39d398c106f8a36022a04e6c18061ff38629d5b42b125aaea1e94a97b5f1c871bcf93b1f141a49ddb0b4c32b976cd530c1008da86"))

	recCode, err = testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 2, len(intermediateTxs))

	scr := intermediateTxs[1].(*smartContractResult.SmartContractResult)
	require.Equal(t, big.NewInt(0), scr.Value)

	stakedValue, _ := big.NewInt(0).SetString("2500000000000000000000", 10)
	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unStakeTokens@"+hex.EncodeToString(stakedValue.Bytes())))

	recCode, err = testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)
	require.Nil(t, testContextMeta.GetLatestError())

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 3, len(intermediateTxs))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBondTokens@"+hex.EncodeToString(stakedValue.Bytes())))

	recCode, err = testContextMeta.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, recCode)
	require.Nil(t, err)

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 2, len(intermediateTxs))

	scrWithMessage = intermediateTxs[0].(*smartContractResult.SmartContractResult)
	require.Equal(t, stakedValue, scrWithMessage.Value)
}

func saveNodesConfig(t *testing.T, testContext *vm.VMTestContext, stakedNodes, minNumNodes, maxNumNodes int64) {
	protoMarshalizer := &marshal.GogoProtoMarshalizer{}

	account, err := testContext.Accounts.LoadAccount(vmAddr.StakingSCAddress)
	require.Nil(t, err)
	userAccount, _ := account.(state.UserAccountHandler)

	nodesConfigData := &systemSmartContracts.StakingNodesConfig{
		StakedNodes: stakedNodes,
		MinNumNodes: minNumNodes,
		MaxNumNodes: maxNumNodes,
	}
	nodesDataBytes, _ := protoMarshalizer.Marshal(nodesConfigData)

	_ = userAccount.DataTrieTracker().SaveKeyValue([]byte("nodesConfig"), nodesDataBytes)
	_ = testContext.Accounts.SaveAccount(account)
	_, _ = testContext.Accounts.Commit()
}
