package txsFee

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	mclsig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/singlesig"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/state"
	vmAddr "github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

const (
	stakingIsFullMessage      = "staking is full key put into waiting list"
	validatorPrivateKey       = "4f0bdd5a2237cb61e495763d93c3a85577420faf51e4bbb40d5a03e915f1e21f"
	validatorBLSKey           = "cb66c844dc64e0cab854997df9b3fd1b1c071ee25e6634e6b95bdb15f0acb38e7c5ac20d02f231a03fde3abb8220951328a7f550915ff9da4a1960c8005d6dfabc50f776bd17433f29e8d0566871380b439024a70dfade593173c6462ad49318"
	validatorStakeData        = "stake@01@" + validatorBLSKey + "@7aaf0222069b5ac98c59c407411e00ec0d826ef323b0325da4e0dd98bc54eeaf042865c2856a4ef88ee54fe24552dd85@b2a11555ce521e4944e09ab17549d85b487dcd26c84b5017a39e31a3670889ba"
	cannotUnBondTokensMessage = "cannot unBond tokens, the validator would remain without min deposit, nodes are still active"
	noTokensToUnBondMessage   = "no tokens that can be unbond at this time"
)

var (
	value2700EGLD, _   = big.NewInt(0).SetString("2700000000000000000000", 10)
	value2500EGLD, _   = big.NewInt(0).SetString("2500000000000000000000", 10)
	value1250EGLD, _   = big.NewInt(0).SetString("1250000000000000000000", 10)
	value200EGLD, _    = big.NewInt(0).SetString("200000000000000000000", 10)
	valueUnJailEGLD, _ = big.NewInt(0).SetString("2500000000000000000", 10)
	blsSigner          = &mclsig.BlsSingleSigner{}
	keyGen             = signing.NewKeyGenerator(&mcl.SuiteBLS12{})
)

const delegationManagementKey = "delegationManagement"
const delegationContractsList = "delegationContracts"

func saveDelegationManagerConfig(testContext *vm.VMTestContext) {
	acc, _ := testContext.Accounts.LoadAccount(vmAddr.DelegationManagerSCAddress)
	userAcc, _ := acc.(state.UserAccountHandler)

	managementData := &systemSmartContracts.DelegationManagement{
		MinDelegationAmount: big.NewInt(1),
		MinDeposit:          big.NewInt(1),
		LastAddress:         vmAddr.FirstDelegationSCAddress,
	}
	marshaledData, _ := testContext.Marshalizer.Marshal(managementData)
	_ = userAcc.SaveKeyValue([]byte(delegationManagementKey), marshaledData)

	list := &systemSmartContracts.DelegationContractList{
		Addresses: [][]byte{vmAddr.FirstDelegationSCAddress},
	}
	marshaledContractList, _ := testContext.Marshalizer.Marshal(list)
	_ = userAcc.SaveKeyValue([]byte(delegationContractsList), marshaledContractList)

	_ = testContext.Accounts.SaveAccount(userAcc)
}

func TestValidatorsSC_DoStakePutInQueueUnStakeAndUnBondShouldRefund(t *testing.T) {
	testContextMeta, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, config.EnableEpochs{})

	require.Nil(t, err)
	defer testContextMeta.Close()

	saveNodesConfig(t, testContextMeta, 1, 1, 1)
	testContextMeta.BlockchainHook.(*hooks.BlockChainHookImpl).SetCurrentHeader(&block.MetaBlock{Epoch: 1})
	saveDelegationManagerConfig(testContextMeta)

	gasLimit := uint64(4000)
	sndAddr := []byte("12345678901234567890123456789012")
	tx := vm.CreateTransaction(0, value2500EGLD, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte(validatorStakeData))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unStake@"+validatorBLSKey))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	intermediateTxs := testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediateTxs))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBond@"+validatorBLSKey))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 2, len(intermediateTxs))

	scr := intermediateTxs[1].(*smartContractResult.SmartContractResult)
	require.Equal(t, value2500EGLD, scr.Value)
}

func checkReturnLog(t *testing.T, testContextMeta *vm.VMTestContext, subStr string, isError bool) {
	allLogs := testContextMeta.TxsLogsProcessor.GetAllCurrentLogs()
	testContextMeta.TxsLogsProcessor.Clean()

	identifierStr := "writeLog"
	if isError {
		identifierStr = "signalError"
	}

	found := false
	for _, eventLog := range allLogs {
		for _, event := range eventLog.GetLogEvents() {
			if string(event.GetIdentifier()) == identifierStr {
				require.True(t, strings.Contains(string(event.GetTopics()[1]), subStr))
				found = true
			}
		}
	}
	require.True(t, found)
}

func TestValidatorsSC_DoStakePutInQueueUnStakeAndUnBondTokensShouldRefund(t *testing.T) {
	testContextMeta, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, config.EnableEpochs{})

	require.Nil(t, err)
	defer testContextMeta.Close()

	saveNodesConfig(t, testContextMeta, 1, 1, 1)
	saveDelegationManagerConfig(testContextMeta)
	testContextMeta.BlockchainHook.(*hooks.BlockChainHookImpl).SetCurrentHeader(&block.MetaBlock{Epoch: 1})

	gasLimit := uint64(4000)
	sndAddr := []byte("12345678901234567890123456789012")
	tx := vm.CreateTransaction(0, value2500EGLD, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte(validatorStakeData))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	checkReturnLog(t, testContextMeta, "staking is full key put into waiting list ", false)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unStake@"+validatorBLSKey))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBondTokens@"+hex.EncodeToString(value2500EGLD.Bytes())))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.UserError, nil)

	checkReturnLog(t, testContextMeta, cannotUnBondTokensMessage, true)
}

func TestValidatorsSC_DoStakeWithTopUpValueTryToUnStakeTokensAndUnBondTokens(t *testing.T) {
	argUnbondTokensV1 := config.EnableEpochs{
		UnbondTokensV2EnableEpoch: 20000,
	}
	testValidatorsSCDoStakeWithTopUpValueTryToUnStakeTokensAndUnBondTokens(t, argUnbondTokensV1)

	argUnbondTokensV2 := config.EnableEpochs{
		UnbondTokensV2EnableEpoch: 0,
	}
	testValidatorsSCDoStakeWithTopUpValueTryToUnStakeTokensAndUnBondTokens(t, argUnbondTokensV2)
}

func testValidatorsSCDoStakeWithTopUpValueTryToUnStakeTokensAndUnBondTokens(t *testing.T, enableEpochs config.EnableEpochs) {
	testContextMeta, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, enableEpochs)

	require.Nil(t, err)
	defer testContextMeta.Close()

	saveNodesConfig(t, testContextMeta, 1, 1, 1)
	saveDelegationManagerConfig(testContextMeta)
	testContextMeta.BlockchainHook.(*hooks.BlockChainHookImpl).SetCurrentHeader(&block.MetaBlock{Epoch: 0})

	gasLimit := uint64(4000)
	sndAddr := []byte("12345678901234567890123456789012")
	tx := vm.CreateTransaction(0, value2700EGLD, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte(validatorStakeData))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	checkReturnLog(t, testContextMeta, stakingIsFullMessage, false)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unStakeTokens@"+hex.EncodeToString(value200EGLD.Bytes())))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)
	testContextMeta.TxsLogsProcessor.Clean()

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBondTokens@"+hex.EncodeToString(value200EGLD.Bytes())))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	checkReturnLog(t, testContextMeta, noTokensToUnBondMessage, false)
}

func TestValidatorsSC_ToStakePutInQueueUnStakeAndUnBondShouldRefundUnBondTokens(t *testing.T) {
	testContextMeta, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, config.EnableEpochs{})

	require.Nil(t, err)
	defer testContextMeta.Close()

	saveNodesConfig(t, testContextMeta, 1, 1, 1)
	saveDelegationManagerConfig(testContextMeta)
	testContextMeta.BlockchainHook.(*hooks.BlockChainHookImpl).SetCurrentHeader(&block.MetaBlock{Epoch: 1})

	gasLimit := uint64(4000)
	sndAddr := []byte("12345678901234567890123456789012")
	tx := vm.CreateTransaction(0, value2700EGLD, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte(validatorStakeData))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	checkReturnLog(t, testContextMeta, stakingIsFullMessage, false)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unStake@"+validatorBLSKey))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBond@"+validatorBLSKey))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	intermediateTxs := testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 2, len(intermediateTxs))

	scr := intermediateTxs[1].(*smartContractResult.SmartContractResult)
	require.Equal(t, value2500EGLD, scr.Value)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unStakeTokens@"+hex.EncodeToString(value200EGLD.Bytes())))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBondTokens@"+hex.EncodeToString(value200EGLD.Bytes())))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediateTxs))

	scrWithMessage := intermediateTxs[0].(*smartContractResult.SmartContractResult)
	require.Equal(t, value200EGLD, scrWithMessage.Value)
}

func TestValidatorsSC_ToStakePutInQueueUnStakeNodesAndUnBondNodesShouldRefund(t *testing.T) {
	testContextMeta, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, config.EnableEpochs{})

	require.Nil(t, err)
	defer testContextMeta.Close()

	saveNodesConfig(t, testContextMeta, 1, 1, 1)
	saveDelegationManagerConfig(testContextMeta)
	testContextMeta.BlockchainHook.(*hooks.BlockChainHookImpl).SetCurrentHeader(&block.MetaBlock{Epoch: 1})

	gasLimit := uint64(4000)
	sndAddr := []byte("12345678901234567890123456789012")
	tx := vm.CreateTransaction(0, value2700EGLD, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte(validatorStakeData))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	checkReturnLog(t, testContextMeta, stakingIsFullMessage, false)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unStakeNodes@"+validatorBLSKey))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBondNodes@"+validatorBLSKey))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	intermediateTxs := testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediateTxs))

	scr := intermediateTxs[0].(*smartContractResult.SmartContractResult)
	require.Equal(t, big.NewInt(0), scr.Value)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unStakeTokens@"+hex.EncodeToString(value2500EGLD.Bytes())))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBondTokens@"+hex.EncodeToString(value2500EGLD.Bytes())))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	intermediateTxs = testContextMeta.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediateTxs))

	scrWithMessage := intermediateTxs[0].(*smartContractResult.SmartContractResult)
	require.Equal(t, value2500EGLD, scrWithMessage.Value)
}

func TestValidatorsSC_ToStakeJailUnbondShouldNotWork(t *testing.T) {
	t.Skip()
	testContextMeta, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, config.EnableEpochs{})
	require.Nil(t, err)

	defer testContextMeta.Close()

	saveNodesConfig(t, testContextMeta, 1, 1, 2)
	saveDelegationManagerConfig(testContextMeta)
	testContextMeta.BlockchainHook.(*hooks.BlockChainHookImpl).SetCurrentHeader(&block.MetaBlock{Epoch: 1})

	gasPrice := uint64(10)
	gasLimit := uint64(4000)
	sndAddr := []byte("12345678901234567890123456789012")
	tx := vm.CreateTransaction(0, value2700EGLD, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte(validatorStakeData))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), vmAddr.JailingAddress, vmAddr.StakingSCAddress, gasPrice, gasLimit, []byte("jail@"+validatorBLSKey))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unBondNodes@"+validatorBLSKey))
	executeTxAndCheckResultsNot(t, testContextMeta, tx, vmcommon.Ok, nil)
}

func getDelegationManagement(testContextMeta *vm.VMTestContext) *systemSmartContracts.DelegationManagement {
	acc, _ := testContextMeta.Accounts.LoadAccount(vmAddr.DelegationManagerSCAddress)
	userAcc, _ := acc.(state.UserAccountHandler)

	managementData := &systemSmartContracts.DelegationManagement{}
	marshaledData, _, _ := userAcc.RetrieveValue([]byte(delegationManagementKey))
	_ = testContextMeta.Marshalizer.Unmarshal(managementData, marshaledData)
	return managementData
}

func TestValidatorsSC_ToStakeJailUnJailShouldWork(t *testing.T) {
	testContextMeta, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, config.EnableEpochs{})
	require.Nil(t, err)

	defer testContextMeta.Close()

	saveNodesConfig(t, testContextMeta, 1, 1, 2)
	saveDelegationManagerConfig(testContextMeta)
	testContextMeta.BlockchainHook.(*hooks.BlockChainHookImpl).SetCurrentHeader(&block.MetaBlock{Epoch: 1})

	gasPrice := uint64(10)
	gasLimit := uint64(4000)
	sndAddr := []byte("12345678901234567890123456789012")

	tx := vm.CreateTransaction(0, value2700EGLD, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte(validatorStakeData))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), vmAddr.JailingAddress, vmAddr.StakingSCAddress, gasPrice, gasLimit, []byte("jail@"+validatorBLSKey))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, valueUnJailEGLD, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unJail@"+validatorBLSKey))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)
}

func TestValidatorsSC_AddNodes(t *testing.T) {
	testContextMeta, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, config.EnableEpochs{})
	require.Nil(t, err)

	defer testContextMeta.Close()

	saveNodesConfig(t, testContextMeta, 1, 1, 10)
	saveDelegationManagerConfig(testContextMeta)
	testContextMeta.BlockchainHook.(*hooks.BlockChainHookImpl).SetCurrentHeader(&block.MetaBlock{Epoch: 0, Nonce: 1})

	gasPrice := uint64(10)
	gasLimit := uint64(4000)
	sndAddr := []byte("12345678901234567890123456789012")

	tx := vm.CreateTransaction(0, value2700EGLD, sndAddr, vmAddr.DelegationManagerSCAddress, gasPrice, gasLimit, []byte("createNewDelegationContract@00@00"))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	managementData := getDelegationManagement(testContextMeta)
	stakingContractAddress := managementData.LastAddress
	signature, err := signMessage(stakingContractAddress)
	require.Nil(t, err)

	txData := "addNodes@" + validatorBLSKey + "@" + hex.EncodeToString(signature)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, stakingContractAddress, gasPrice, 7000000, []byte(txData))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, stakingContractAddress, gasPrice, 2000000, []byte("setAutomaticActivation@74727565"))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, value1250EGLD, sndAddr, stakingContractAddress, gasPrice, gasLimit, []byte("delegate"))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), vmAddr.JailingAddress, vmAddr.StakingSCAddress, gasPrice, gasLimit, []byte("jail@"+validatorBLSKey))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, valueUnJailEGLD, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unJail@"+validatorBLSKey))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)
}

func TestValidatorsSC_StakeNodes(t *testing.T) {
	testContextMeta, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, config.EnableEpochs{})
	require.Nil(t, err)

	defer testContextMeta.Close()

	saveNodesConfig(t, testContextMeta, 1, 1, 10)
	saveDelegationManagerConfig(testContextMeta)
	testContextMeta.BlockchainHook.(*hooks.BlockChainHookImpl).SetCurrentHeader(&block.MetaBlock{Epoch: 0, Nonce: 1})

	gasPrice := uint64(10)
	gasLimit := uint64(4000)
	sndAddr := []byte("12345678901234567890123456789012")

	tx := vm.CreateTransaction(0, value2700EGLD, sndAddr, vmAddr.DelegationManagerSCAddress, gasPrice, gasLimit, []byte("createNewDelegationContract@00@00"))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	managementData := getDelegationManagement(testContextMeta)
	stakingContractAddress := managementData.LastAddress
	signature, err := signMessage(stakingContractAddress)
	require.Nil(t, err)

	txData := "addNodes@" + validatorBLSKey + "@" + hex.EncodeToString(signature)

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, stakingContractAddress, gasPrice, 7000000, []byte(txData))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	txData = "stakeNodes@" + validatorBLSKey

	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, stakingContractAddress, gasPrice, 7000000, []byte(txData))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	//tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, stakingContractAddress, gasPrice, 2000000, []byte("setAutomaticActivation@74727565"))
	//executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)
	//
	//utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	//tx = vm.CreateTransaction(0, value1250EGLD, sndAddr, stakingContractAddress, gasPrice, gasLimit, []byte("delegate"))
	//executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)
	//
	//utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, big.NewInt(0), vmAddr.JailingAddress, vmAddr.StakingSCAddress, gasPrice, gasLimit, []byte("jail@"+validatorBLSKey))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextMeta)

	tx = vm.CreateTransaction(0, valueUnJailEGLD, sndAddr, vmAddr.ValidatorSCAddress, gasPrice, gasLimit, []byte("unJail@"+validatorBLSKey))
	executeTxAndCheckResults(t, testContextMeta, tx, vmcommon.Ok, nil)
}

func signMessage(message []byte) ([]byte, error) {
	skBuff, err := hex.DecodeString(validatorPrivateKey)
	if err != nil {
		return nil, err
	}

	sk, err := keyGen.PrivateKeyFromByteArray(skBuff)
	if err != nil {
		return nil, err
	}

	signature, err := blsSigner.Sign(sk, message)
	return signature, err
}

func executeTxAndCheckResults(
	t *testing.T,
	testContext *vm.VMTestContext,
	tx *transaction.Transaction,
	vmCodeExpected vmcommon.ReturnCode,
	expectedErr error,
) {
	recCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	assert.Equal(t, vmCodeExpected, recCode)
	assert.Equal(t, expectedErr, err)
}

func executeTxAndCheckResultsNot(
	t *testing.T,
	testContext *vm.VMTestContext,
	tx *transaction.Transaction,
	notVmCodeExpected vmcommon.ReturnCode,
	notExpectedErr error,
) {
	recCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.NotEqual(t, notVmCodeExpected, recCode)
	require.NotEqual(t, notExpectedErr, err)
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

	_ = userAccount.SaveKeyValue([]byte("nodesConfig"), nodesDataBytes)
	_ = testContext.Accounts.SaveAccount(account)
	_, _ = testContext.Accounts.Commit()
}
