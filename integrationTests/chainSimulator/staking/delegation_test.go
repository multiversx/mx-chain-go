package staking

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	chainSimulatorProcess "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const walletAddressBytesLen = 32
const mockBLSSignature = "010101"
const gasLimitForStakeOperation = 50_000_000
const gasLimitForConvertOperation = 510_000_000
const gasLimitForDelegate = 12_000_000
const minGasPrice = 1000000000
const txVersion = 1
const mockTxSignature = "sig"
const queuedStatus = "queued"
const stakedStatus = "staked"
const okReturnCode = "ok"
const maxCap = "00"       // no cap
const serviceFee = "0ea1" // 37.45%

var zeroValue = big.NewInt(0)
var oneEGLD = big.NewInt(1000000000000000000)
var minimumStakeValue = big.NewInt(0).Mul(oneEGLD, big.NewInt(2500))

// Test description
//  Test that delegation contract created with MakeNewContractFromValidatorData works properly
//  Also check that delegate and undelegate works properly and the top-up remain the same if every delegator undelegates.
//  Test that the top-up from normal stake will be transferred after creating the contract and will be used in auction list computing

// Test scenario
// 1. Add a new validator private key in the multi key handler
// 2. Set the initial state for the owner and the 2 delegators
// 3. Do a stake transaction for the validator key and test that the new key is on queue / auction list and topup is 500
// 4. Execute the MakeNewContractFromValidatorData transaction and test that the key is on queue / auction list and topup is 500
// 5. Execute 2 delegation operations of 100 EGLD each, check the topup is 700
// 6. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 500

func TestChainSimulator_MakeNewContractFromValidatorData(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 101
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 102

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 102
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		testChainSimulatorMakeNewContractFromValidatorData(t, cs, 1)
	})
	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		testChainSimulatorMakeNewContractFromValidatorData(t, cs, 2)
	})
}

func testChainSimulatorMakeNewContractFromValidatorData(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	log.Info("Step 1. Add a new validator private key in the multi key handler")
	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	log.Info("Step 2. Set the initial state for the owner and the 2 delegators")
	validatorOwner := generateWalletAddressBytes()
	validatorOwnerBech32 := metachainNode.GetCoreComponents().AddressPubKeyConverter().SilentEncode(validatorOwner, log)
	mintValue := big.NewInt(3010)
	mintValue = mintValue.Mul(oneEGLD, mintValue)

	delegator1 := generateWalletAddressBytes()
	delegator1Bech32 := metachainNode.GetCoreComponents().AddressPubKeyConverter().SilentEncode(delegator1, log)
	delegator2 := generateWalletAddressBytes()
	delegator2Bech32 := metachainNode.GetCoreComponents().AddressPubKeyConverter().SilentEncode(delegator2, log)

	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: validatorOwnerBech32,
			Balance: mintValue.String(),
		},
		{
			Address: delegator1Bech32,
			Balance: mintValue.String(),
		},
		{
			Address: delegator2Bech32,
			Balance: mintValue.String(),
		},
	})
	require.Nil(t, err)

	log.Info("working with the following addresses",
		"newValidatorOwner", validatorOwnerBech32, "delegator1", delegator1Bech32, "delegator2", delegator2Bech32)

	log.Info("Step 3. Do a stake transaction for the validator key and test that the new key is on queue / auction list and the correct topup")
	stakeValue := big.NewInt(0).Set(minimumStakeValue)
	addedStakedValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(500))
	stakeValue.Add(stakeValue, addedStakedValue)
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], mockBLSSignature)
	txStake := generateTransaction(validatorOwner, 0, vm.ValidatorSCAddress, stakeValue, txDataField, gasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(5)
	assert.Nil(t, err)

	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	statistics, err := metachainNode.GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)

	_, found := statistics[blsKeys[0]]
	require.False(t, found)

	decodedBLSKey, _ := hex.DecodeString(blsKeys[0])
	testBLSKeyIsInQueueOrAuction(t, metachainNode, decodedBLSKey, addedStakedValue)
	assert.Equal(t, addedStakedValue, getBLSTopUpValue(t, metachainNode, validatorOwner))

	log.Info("Step 4. Execute the MakeNewContractFromValidatorData transaction and test that the key is on queue / auction list and the correct topup")
	txDataField = fmt.Sprintf("makeNewContractFromValidatorData@%s@%s", maxCap, serviceFee)
	txConvert := generateTransaction(validatorOwner, 1, vm.DelegationManagerSCAddress, zeroValue, txDataField, gasLimitForConvertOperation)
	convertTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txConvert, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, convertTx)

	delegationAddress := convertTx.Logs.Events[0].Topics[1]
	delegationAddressBech32 := metachainNode.GetCoreComponents().AddressPubKeyConverter().SilentEncode(delegationAddress, log)
	log.Info("generated delegation address", "address", delegationAddressBech32)

	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	statistics, err = metachainNode.GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)

	_, found = statistics[blsKeys[0]]
	require.False(t, found)

	testBLSKeyIsInQueueOrAuction(t, metachainNode, decodedBLSKey, addedStakedValue)
	assert.Equal(t, addedStakedValue, getBLSTopUpValue(t, metachainNode, delegationAddress))

	log.Info("Step 5. Execute 2 delegation operations of 100 EGLD each, check the topup is 700")
	delegateValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(100))
	txDelegate1 := generateTransaction(delegator1, 0, delegationAddress, delegateValue, "delegate", gasLimitForDelegate)
	delegate1Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txDelegate1, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegate1Tx)

	txDelegate2 := generateTransaction(delegator2, 0, delegationAddress, delegateValue, "delegate", gasLimitForDelegate)
	delegate2Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txDelegate2, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegate2Tx)

	expectedTopUp := big.NewInt(0).Mul(oneEGLD, big.NewInt(700))
	testBLSKeyIsInQueueOrAuction(t, metachainNode, decodedBLSKey, expectedTopUp)
	assert.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationAddress))

	log.Info("6. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 500")
	unDelegateValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(100))
	txDataField = fmt.Sprintf("unDelegate@%s", hex.EncodeToString(unDelegateValue.Bytes()))
	txUnDelegate1 := generateTransaction(delegator1, 1, delegationAddress, zeroValue, txDataField, gasLimitForDelegate)
	unDelegate1Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnDelegate1, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unDelegate1Tx)

	txDataField = fmt.Sprintf("unDelegate@%s", hex.EncodeToString(unDelegateValue.Bytes()))
	txUnDelegate2 := generateTransaction(delegator2, 1, delegationAddress, zeroValue, txDataField, gasLimitForDelegate)
	unDelegate2Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnDelegate2, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unDelegate2Tx)

	expectedTopUp = big.NewInt(0).Mul(oneEGLD, big.NewInt(500))
	testBLSKeyIsInQueueOrAuction(t, metachainNode, decodedBLSKey, expectedTopUp)
	assert.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationAddress))

}

func generateWalletAddressBytes() []byte {
	buff := make([]byte, walletAddressBytesLen)
	_, _ = rand.Read(buff)

	return buff
}

func testBLSKeyIsInQueueOrAuction(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, blsKey []byte, topUpInAuctionList *big.Int) {
	activationEpoch := metachainNode.GetCoreComponents().EnableEpochsHandler().GetActivationEpoch(common.StakingV4Step1Flag)
	if activationEpoch <= metachainNode.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch() {
		testBLSKeyIsInAuction(t, metachainNode, blsKey, topUpInAuctionList)
		return
	}

	require.Equal(t, queuedStatus, getBLSKeyStatus(t, metachainNode, blsKey))
}

func testBLSKeyIsInAuction(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, blsKey []byte, topUpInAuctionList *big.Int) {
	require.Equal(t, stakedStatus, getBLSKeyStatus(t, metachainNode, blsKey))

	err := metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	auctionList, err := metachainNode.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)

	require.Equal(t, 1, len(auctionList))
	require.Equal(t, 1, len(auctionList[0].AuctionList))
	require.Equal(t, topUpInAuctionList, auctionList[0].TopUpPerNode)
}

func getBLSKeyStatus(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, blsKey []byte) string {
	scQuery := &process.SCQuery{
		ScAddress:  vm.StakingSCAddress,
		FuncName:   "getBLSKeyStatus",
		CallerAddr: vm.StakingSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{blsKey},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, okReturnCode, result.ReturnCode)

	return string(result.ReturnData[0])
}

func getBLSTopUpValue(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, address []byte) *big.Int {
	scQuery := &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getTotalStakedTopUpStakedBlsKeys",
		CallerAddr: vm.StakingSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{address},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, okReturnCode, result.ReturnCode)

	return big.NewInt(0).SetBytes(result.ReturnData[0])
}

func generateTransaction(sender []byte, nonce uint64, receiver []byte, value *big.Int, data string, gasLimit uint64) *transaction.Transaction {
	return &transaction.Transaction{
		Nonce:     nonce,
		Value:     value,
		SndAddr:   sender,
		RcvAddr:   receiver,
		Data:      []byte(data),
		GasLimit:  gasLimit,
		GasPrice:  minGasPrice,
		ChainID:   []byte(configs.ChainID),
		Version:   txVersion,
		Signature: []byte(mockTxSignature),
	}
}
