package staking

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/validator"
	dataVm "github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	mclsig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/singlesig"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	chainSimulatorProcess "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mockBLSSignature = "010101"
const gasLimitForStakeOperation = 50_000_000
const gasLimitForConvertOperation = 510_000_000
const gasLimitForDelegationContractCreationOperation = 500_000_000
const gasLimitForAddNodesOperation = 500_000_000
const gasLimitForUndelegateOperation = 500_000_000
const gasLimitForDelegate = 12_000_000
const minGasPrice = 1000000000
const txVersion = 1
const mockTxSignature = "sig"
const queuedStatus = "queued"
const stakedStatus = "staked"
const auctionStatus = "auction"
const okReturnCode = "ok"
const maxCap = "00"       // no cap
const serviceFee = "0ea1" // 37.45%

var zeroValue = big.NewInt(0)
var oneEGLD = big.NewInt(1000000000000000000)
var minimumStakeValue = big.NewInt(0).Mul(oneEGLD, big.NewInt(2500))
var minimumCreateDelegationStakeValue = big.NewInt(0).Mul(oneEGLD, big.NewInt(1250))

// Test description
//  Test that delegation contract created with MakeNewContractFromValidatorData works properly
//  Also check that delegate and undelegate works properly and the top-up remain the same if every delegator undelegates.
//  Test that the top-up from normal stake will be transferred after creating the contract and will be used in auction list computing

// Internal test scenario #10
func TestChainSimulator_MakeNewContractFromValidatorData(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	// Test scenario done in staking 3.5 phase (staking v4 is not active)
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Do a stake transaction for the validator key and test that the new key is on queue and topup is 500
	// 4. Execute the MakeNewContractFromValidatorData transaction and test that the key is on queue and topup is 500
	// 5. Execute 2 delegation operations of 100 EGLD each, check the topup is 700
	// 6. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 500
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

		defer cs.Close()

		testChainSimulatorMakeNewContractFromValidatorData(t, cs, 1)
	})

	// Test scenario done in staking v4 phase step 1
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Do a stake transaction for the validator key and test that the new key is on auction list and topup is 500
	// 4. Execute the MakeNewContractFromValidatorData transaction and test that the key is on auction list and topup is 500
	// 5. Execute 2 delegation operations of 100 EGLD each, check the topup is 700
	// 6. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 500
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

		defer cs.Close()

		testChainSimulatorMakeNewContractFromValidatorData(t, cs, 2)
	})

	// Test scenario done in staking v4 phase step 2
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Do a stake transaction for the validator key and test that the new key is on auction list and topup is 500
	// 4. Execute the MakeNewContractFromValidatorData transaction and test that the key is on auction list and topup is 500
	// 5. Execute 2 delegation operations of 100 EGLD each, check the topup is 700
	// 6. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 500
	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
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

		defer cs.Close()

		testChainSimulatorMakeNewContractFromValidatorData(t, cs, 3)
	})

	// Test scenario done in staking v4 phase step 3
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Do a stake transaction for the validator key and test that the new key is on auction list and topup is 500
	// 4. Execute the MakeNewContractFromValidatorData transaction and test that the key is on auction list and topup is 500
	// 5. Execute 2 delegation operations of 100 EGLD each, check the topup is 700
	// 6. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 500
	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
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

		defer cs.Close()

		testChainSimulatorMakeNewContractFromValidatorData(t, cs, 4)
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
	mintValue := big.NewInt(3010)
	mintValue = mintValue.Mul(oneEGLD, mintValue)

	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	delegator1, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	delegator2, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	log.Info("working with the following addresses",
		"newValidatorOwner", validatorOwner.Bech32, "delegator1", delegator1.Bech32, "delegator2", delegator2.Bech32)

	log.Info("Step 3. Do a stake transaction for the validator key and test that the new key is on queue / auction list and the correct topup")
	stakeValue := big.NewInt(0).Set(minimumStakeValue)
	addedStakedValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(500))
	stakeValue.Add(stakeValue, addedStakedValue)
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], mockBLSSignature)
	txStake := generateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, stakeValue, txDataField, gasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	assert.Nil(t, err)

	testBLSKeyIsInQueueOrAuction(t, metachainNode, validatorOwner.Bytes, blsKeys[0], addedStakedValue)

	log.Info("Step 4. Execute the MakeNewContractFromValidatorData transaction and test that the key is on queue / auction list and the correct topup")
	txDataField = fmt.Sprintf("makeNewContractFromValidatorData@%s@%s", maxCap, serviceFee)
	txConvert := generateTransaction(validatorOwner.Bytes, 1, vm.DelegationManagerSCAddress, zeroValue, txDataField, gasLimitForConvertOperation)
	convertTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txConvert, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, convertTx)

	delegationAddress := convertTx.Logs.Events[0].Topics[1]
	delegationAddressBech32 := metachainNode.GetCoreComponents().AddressPubKeyConverter().SilentEncode(delegationAddress, log)
	log.Info("generated delegation address", "address", delegationAddressBech32)

	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)

	testBLSKeyIsInQueueOrAuction(t, metachainNode, delegationAddress, blsKeys[0], addedStakedValue)

	log.Info("Step 5. Execute 2 delegation operations of 100 EGLD each, check the topup is 700")
	delegateValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(100))
	txDelegate1 := generateTransaction(delegator1.Bytes, 0, delegationAddress, delegateValue, "delegate", gasLimitForDelegate)
	delegate1Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txDelegate1, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegate1Tx)

	txDelegate2 := generateTransaction(delegator2.Bytes, 0, delegationAddress, delegateValue, "delegate", gasLimitForDelegate)
	delegate2Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txDelegate2, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegate2Tx)

	expectedTopUp := big.NewInt(0).Mul(oneEGLD, big.NewInt(700))
	testBLSKeyIsInQueueOrAuction(t, metachainNode, delegationAddress, blsKeys[0], expectedTopUp)

	log.Info("6. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 500")
	unDelegateValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(100))
	txDataField = fmt.Sprintf("unDelegate@%s", hex.EncodeToString(unDelegateValue.Bytes()))
	txUnDelegate1 := generateTransaction(delegator1.Bytes, 1, delegationAddress, zeroValue, txDataField, gasLimitForDelegate)
	unDelegate1Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnDelegate1, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unDelegate1Tx)

	txDataField = fmt.Sprintf("unDelegate@%s", hex.EncodeToString(unDelegateValue.Bytes()))
	txUnDelegate2 := generateTransaction(delegator2.Bytes, 1, delegationAddress, zeroValue, txDataField, gasLimitForDelegate)
	unDelegate2Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnDelegate2, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unDelegate2Tx)

	expectedTopUp = big.NewInt(0).Mul(oneEGLD, big.NewInt(500))
	testBLSKeyIsInQueueOrAuction(t, metachainNode, delegationAddress, blsKeys[0], expectedTopUp)
}

func testBLSKeyIsInQueueOrAuction(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, address []byte, blsKey string, expectedTopUp *big.Int) {
	decodedBLSKey, _ := hex.DecodeString(blsKey)
	err := metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	statistics, err := metachainNode.GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)
	require.Equal(t, expectedTopUp.String(), getBLSTopUpValue(t, metachainNode, address).String())

	activationEpoch := metachainNode.GetCoreComponents().EnableEpochsHandler().GetActivationEpoch(common.StakingV4Step1Flag)
	if activationEpoch <= metachainNode.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch() {
		testBLSKeyIsInAuction(t, metachainNode, address, decodedBLSKey, blsKey, expectedTopUp, statistics)
		return
	}

	// in staking ph 2/3.5 we do not find the bls key on the validator statistics
	_, found := statistics[blsKey]
	require.False(t, found)
	require.Equal(t, queuedStatus, getBLSKeyStatus(t, metachainNode, decodedBLSKey))
}

func testBLSKeyIsInAuction(
	t *testing.T,
	metachainNode chainSimulatorProcess.NodeHandler,
	address []byte,
	blsKeyBytes []byte,
	blsKey string,
	topUpInAuctionList *big.Int,
	validatorStatistics map[string]*validator.ValidatorStatistics,
) {
	require.Equal(t, stakedStatus, getBLSKeyStatus(t, metachainNode, blsKeyBytes))

	err := metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	auctionList, err := metachainNode.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)

	expectedNodesInAuctionList := 1
	expectedAuctionListOwnersSize := 1
	currentEpoch := metachainNode.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch()
	if currentEpoch == metachainNode.GetCoreComponents().EnableEpochsHandler().GetActivationEpoch(common.StakingV4Step2Flag) {
		// starting from step 2, we have the shuffled out nodes from the previous epoch in the action list
		expectedAuctionListOwnersSize += 1
		expectedNodesInAuctionList += 8
	}
	if currentEpoch >= metachainNode.GetCoreComponents().EnableEpochsHandler().GetActivationEpoch(common.StakingV4Step3Flag) {
		// starting from step 3, we have the shuffled out nodes from the previous epoch in the action list
		expectedAuctionListOwnersSize += 1
		expectedNodesInAuctionList += 4
	}

	require.Equal(t, expectedAuctionListOwnersSize, len(auctionList))
	nodesInAuctionList := 0
	addressBech32 := metachainNode.GetCoreComponents().AddressPubKeyConverter().SilentEncode(address, log)
	ownerFound := false
	for i := 0; i < len(auctionList); i++ {
		nodesInAuctionList += len(auctionList[i].Nodes)
		if auctionList[i].Owner == addressBech32 {
			ownerFound = true
			require.Equal(t, topUpInAuctionList.String(), auctionList[i].TopUpPerNode)
		}
	}
	require.True(t, ownerFound)
	require.Equal(t, expectedNodesInAuctionList, nodesInAuctionList)

	// in staking ph 4 we should find the key in the validators statics
	validatorInfo, found := validatorStatistics[blsKey]
	require.True(t, found)
	require.Equal(t, auctionStatus, validatorInfo.ValidatorStatus)
}

// Test description
//  Test the creation of a new delegation contract, adding nodes to it, delegating, and undelegating.

// Test scenario
// 1. Initialize the chain simulator
// 2. Generate blocks to activate staking phases
// 3. Create a new delegation contract
// 4. Add validator nodes to the delegation contract
// 5. Perform delegation operations
// 6. Perform undelegation operations
// 7. Validate the results at each step
func TestChainSimulator_CreateNewDelegationContract(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	// Test scenario done in staking 3.5 phase (staking v4 is not active)
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Create a new delegation contract with 1250 egld
	// 3. Add node to the delegation contract
	// 4. Execute 2 delegation operations of 1250 EGLD each, check the topup is 3750
	// 5. Stake node, check the topup is 1250, check the node is staked
	// 5. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 1250
	// 6. Check the node is unstaked in the next epoch
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

		defer cs.Close()

		testChainSimulatorCreateNewDelegationContract(t, cs, 1)
	})

	// Test scenario done in staking v4 phase step 1
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Create a new delegation contract with 1250 egld
	// 3. Add node to the delegation contract
	// 4. Execute 2 delegation operations of 1250 EGLD each, check the topup is 3750
	// 5. Stake node, check the topup is 1250, check the node is in action list
	// 5. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 1250
	// 6. Check the node is unstaked in the next epoch
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

		defer cs.Close()

		testChainSimulatorCreateNewDelegationContract(t, cs, 2)
	})

	// Test scenario done in staking v4 phase step 2
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Create a new delegation contract with 1250 egld
	// 3. Add node to the delegation contract
	// 4. Execute 2 delegation operations of 1250 EGLD each, check the topup is 3750
	// 5. Stake node, check the topup is 1250, check the node is in action list
	// 5. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 1250
	// 6. Check the node is unstaked in the next epoch
	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
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

		defer cs.Close()

		testChainSimulatorCreateNewDelegationContract(t, cs, 3)
	})

	// Test scenario done in staking v4 phase step 3
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Create a new delegation contract with 1250 egld
	// 3. Add node to the delegation contract
	// 4. Execute 2 delegation operations of 1250 EGLD each, check the topup is 3750
	// 5. Stake node, check the topup is 1250, check the node is in action list
	// 5. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 1250
	// 6. Check the node is unstaked in the next epoch
	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
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

		defer cs.Close()

		testChainSimulatorCreateNewDelegationContract(t, cs, 4)
	})

}

func testChainSimulatorCreateNewDelegationContract(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	initialFunds := big.NewInt(0).Mul(oneEGLD, big.NewInt(10000)) // 10000 EGLD for each
	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, initialFunds)
	require.Nil(t, err)

	delegator1, err := cs.GenerateAndMintWalletAddress(core.AllShardId, initialFunds)
	require.Nil(t, err)

	delegator2, err := cs.GenerateAndMintWalletAddress(core.AllShardId, initialFunds)
	require.Nil(t, err)

	maxDelegationCap := big.NewInt(0).Mul(oneEGLD, big.NewInt(51000)) // 51000 EGLD cap
	txCreateDelegationContract := generateTransaction(validatorOwner.Bytes, 0, vm.DelegationManagerSCAddress, minimumCreateDelegationStakeValue,
		fmt.Sprintf("createNewDelegationContract@%s@%s", hex.EncodeToString(maxDelegationCap.Bytes()), serviceFee),
		gasLimitForDelegationContractCreationOperation)
	createDelegationContractTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txCreateDelegationContract, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, createDelegationContractTx)

	// Check delegation contract creation was successful
	data := createDelegationContractTx.SmartContractResults[0].Data
	parts := strings.Split(data, "@")
	require.Equal(t, 3, len(parts))

	require.Equal(t, hex.EncodeToString([]byte("ok")), parts[1])
	delegationContractAddressHex, _ := hex.DecodeString(parts[2])
	delegationContractAddress, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(delegationContractAddressHex)

	output, err := executeQuery(cs, core.MetachainShardId, vm.DelegationManagerSCAddress, "getAllContractAddresses", nil)
	require.Nil(t, err)
	returnAddress, err := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(output.ReturnData[0])
	require.Nil(t, err)
	require.Equal(t, delegationContractAddress, returnAddress)
	delegationContractAddressBytes := output.ReturnData[0]

	// Step 2: Add validator nodes to the delegation contract
	// This step requires generating BLS keys for validators, signing messages, and sending the "addNodes" transaction.
	// Add checks to verify nodes are added successfully.
	validatorSecretKeysBytes, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(validatorSecretKeysBytes)
	require.Nil(t, err)

	signatures := getSignatures(delegationContractAddressBytes, validatorSecretKeysBytes)
	txAddNodes := generateTransaction(validatorOwner.Bytes, 1, delegationContractAddressBytes, zeroValue, addNodesTxData(blsKeys, signatures), gasLimitForAddNodesOperation)
	addNodesTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txAddNodes, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, addNodesTx)

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys := getNodesFromContract(output.ReturnData)
	require.Equal(t, 0, len(stakedKeys))
	require.Equal(t, 1, len(notStakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(notStakedKeys[0]))
	require.Equal(t, 0, len(unStakedKeys))

	expectedTopUp := big.NewInt(0).Set(minimumCreateDelegationStakeValue)
	expectedTotalStaked := big.NewInt(0).Set(minimumCreateDelegationStakeValue)
	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationContractAddressBytes))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getUserActiveStake", [][]byte{validatorOwner.Bytes})
	require.Nil(t, err)
	require.Equal(t, minimumCreateDelegationStakeValue, big.NewInt(0).SetBytes(output.ReturnData[0]))

	// Step 3: Perform delegation operations
	txDelegate1 := generateTransaction(delegator1.Bytes, 0, delegationContractAddressBytes, minimumCreateDelegationStakeValue, "delegate", gasLimitForDelegate)
	delegate1Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txDelegate1, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegate1Tx)

	expectedTopUp = expectedTopUp.Add(expectedTopUp, minimumCreateDelegationStakeValue)
	expectedTotalStaked = expectedTotalStaked.Add(expectedTotalStaked, minimumCreateDelegationStakeValue)
	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationContractAddressBytes))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getUserActiveStake", [][]byte{delegator1.Bytes})
	require.Nil(t, err)
	require.Equal(t, minimumCreateDelegationStakeValue, big.NewInt(0).SetBytes(output.ReturnData[0]))

	txDelegate2 := generateTransaction(delegator2.Bytes, 0, delegationContractAddressBytes, minimumCreateDelegationStakeValue, "delegate", gasLimitForDelegate)
	delegate2Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txDelegate2, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegate2Tx)

	expectedTopUp = expectedTopUp.Add(expectedTopUp, minimumCreateDelegationStakeValue)
	expectedTotalStaked = expectedTotalStaked.Add(expectedTotalStaked, minimumCreateDelegationStakeValue)
	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationContractAddressBytes))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getUserActiveStake", [][]byte{delegator2.Bytes})
	require.Nil(t, err)
	require.Equal(t, minimumCreateDelegationStakeValue, big.NewInt(0).SetBytes(output.ReturnData[0]))

	// Step 4: Perform stakeNodes

	txStakeNodes := generateTransaction(validatorOwner.Bytes, 2, delegationContractAddressBytes, zeroValue, fmt.Sprintf("stakeNodes@%s", blsKeys[0]), gasLimitForStakeOperation)
	stakeNodesTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStakeNodes, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeNodesTx)

	expectedTopUp = big.NewInt(0).Set(minimumCreateDelegationStakeValue)
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationContractAddressBytes))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys = getNodesFromContract(output.ReturnData)
	require.Equal(t, 1, len(stakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(stakedKeys[0]))
	require.Equal(t, 0, len(notStakedKeys))
	require.Equal(t, 0, len(unStakedKeys))

	// Make block finalized
	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	testBLSKeyIsInQueueOrAuction(t, metachainNode, delegationContractAddressBytes, blsKeys[0], expectedTopUp)

	// Step 5: Perform unDelegate from 1 user
	// The nodes should remain in the staked state
	// The total active stake should be reduced by the amount undelegated

	txUndelegate1 := generateTransaction(delegator1.Bytes, 1, delegationContractAddressBytes, zeroValue, fmt.Sprintf("unDelegate@%s", hex.EncodeToString(minimumCreateDelegationStakeValue.Bytes())), gasLimitForUndelegateOperation)
	undelegate1Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUndelegate1, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, undelegate1Tx)

	expectedTopUp = expectedTopUp.Sub(expectedTopUp, minimumCreateDelegationStakeValue)
	expectedTotalStaked = expectedTotalStaked.Sub(expectedTotalStaked, minimumCreateDelegationStakeValue)
	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp.String(), getBLSTopUpValue(t, metachainNode, delegationContractAddressBytes).String())

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getUserActiveStake", [][]byte{delegator1.Bytes})
	require.Nil(t, err)
	require.Equal(t, zeroValue, big.NewInt(0).SetBytes(output.ReturnData[0]))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys = getNodesFromContract(output.ReturnData)
	require.Equal(t, 1, len(stakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(stakedKeys[0]))
	require.Equal(t, 0, len(notStakedKeys))
	require.Equal(t, 0, len(unStakedKeys))

	// Step 6: Perform unDelegate from last user
	// The nodes should change to unStaked state
	// The total active stake should be reduced by the amount undelegated

	txUndelegate2 := generateTransaction(delegator2.Bytes, 1, delegationContractAddressBytes, zeroValue, fmt.Sprintf("unDelegate@%s", hex.EncodeToString(minimumCreateDelegationStakeValue.Bytes())), gasLimitForUndelegateOperation)
	undelegate2Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUndelegate2, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, undelegate2Tx)

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, "1250000000000000000000", big.NewInt(0).SetBytes(output.ReturnData[0]).String())
	require.Equal(t, zeroValue, getBLSTopUpValue(t, metachainNode, delegationContractAddressBytes))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getUserActiveStake", [][]byte{delegator2.Bytes})
	require.Nil(t, err)
	require.Equal(t, "0", big.NewInt(0).SetBytes(output.ReturnData[0]).String())

	// still staked until epoch change
	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys = getNodesFromContract(output.ReturnData)
	require.Equal(t, 1, len(stakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(stakedKeys[0]))
	require.Equal(t, 0, len(notStakedKeys))
	require.Equal(t, 0, len(unStakedKeys))

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch + 1)
	require.Nil(t, err)

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys = getNodesFromContract(output.ReturnData)
	require.Equal(t, 0, len(stakedKeys))
	require.Equal(t, 0, len(notStakedKeys))
	require.Equal(t, 1, len(unStakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(unStakedKeys[0]))
}

func TestChainSimulator_MaxDelegationCap(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	// Test scenario done in staking 3.5 phase (staking v4 is not active)
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 3 delegators
	// 3. Create a new delegation contract with 1250 egld and maximum delegation cap of 3000 EGLD
	// 4. Add node to the delegation contract
	// 5. Delegate from user A 1250 EGLD, check the topup is 2500
	// 6. Delegate from user B 501 EGLD, check it fails
	// 7. Stake node, check the topup is 0, check the node is staked
	// 8. Delegate from user B 501 EGLD, check it fails
	// 9. Delegate from user B 500 EGLD, check the topup is 500
	// 10. Delegate from user B 20 EGLD, check it fails
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

		defer cs.Close()

		testChainSimulatorMaxDelegationCap(t, cs, 1)
	})

	// Test scenario done in staking v4 phase step 1
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 3 delegators
	// 3. Create a new delegation contract with 1250 egld and maximum delegation cap of 3000 EGLD
	// 4. Add node to the delegation contract
	// 5. Delegate from user A 1250 EGLD, check the topup is 2500
	// 6. Delegate from user B 501 EGLD, check it fails
	// 7. Stake node, check the topup is 0, check the node is staked, check the node is in action list
	// 8. Delegate from user B 501 EGLD, check it fails
	// 9. Delegate from user B 500 EGLD, check the topup is 500
	// 10. Delegate from user B 20 EGLD, check it fails
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

		defer cs.Close()

		testChainSimulatorMaxDelegationCap(t, cs, 2)
	})

	// Test scenario done in staking v4 phase step 2
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 3 delegators
	// 3. Create a new delegation contract with 1250 egld and maximum delegation cap of 3000 EGLD
	// 4. Add node to the delegation contract
	// 5. Delegate from user A 1250 EGLD, check the topup is 2500
	// 6. Delegate from user B 501 EGLD, check it fails
	// 7. Stake node, check the topup is 0, check the node is staked, check the node is in action list
	// 8. Delegate from user B 501 EGLD, check it fails
	// 9. Delegate from user B 500 EGLD, check the topup is 500
	// 10. Delegate from user B 20 EGLD, check it fails
	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
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

		defer cs.Close()

		testChainSimulatorMaxDelegationCap(t, cs, 3)
	})

	// Test scenario done in staking v4 phase step 3
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 3 delegators
	// 3. Create a new delegation contract with 1250 egld
	// 4. Add node to the delegation contract
	// 5. Delegate from user A 1250 EGLD, check the topup is 2500
	// 6. Delegate from user B 501 EGLD, check it fails
	// 7. Stake node, check the topup is 0, check the node is staked, check the node is in action list
	// 8. Delegate from user B 501 EGLD, check it fails
	// 9. Delegate from user B 500 EGLD, check the topup is 500
	// 10. Delegate from user B 20 EGLD, check it fails
	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
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

		defer cs.Close()

		testChainSimulatorMaxDelegationCap(t, cs, 4)
	})

}

func testChainSimulatorMaxDelegationCap(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	initialFunds := big.NewInt(0).Mul(oneEGLD, big.NewInt(10000)) // 10000 EGLD for each
	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, initialFunds)
	require.Nil(t, err)

	delegatorA, err := cs.GenerateAndMintWalletAddress(core.AllShardId, initialFunds)
	require.Nil(t, err)

	delegatorB, err := cs.GenerateAndMintWalletAddress(core.AllShardId, initialFunds)
	require.Nil(t, err)

	delegatorC, err := cs.GenerateAndMintWalletAddress(core.AllShardId, initialFunds)
	require.Nil(t, err)

	// Step 3: Create a new delegation contract

	maxDelegationCap := big.NewInt(0).Mul(oneEGLD, big.NewInt(3000)) // 3000 EGLD cap
	txCreateDelegationContract := generateTransaction(validatorOwner.Bytes, 0, vm.DelegationManagerSCAddress, minimumCreateDelegationStakeValue,
		fmt.Sprintf("createNewDelegationContract@%s@%s", hex.EncodeToString(maxDelegationCap.Bytes()), serviceFee),
		gasLimitForDelegationContractCreationOperation)
	createDelegationContractTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txCreateDelegationContract, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, createDelegationContractTx)

	output, err := executeQuery(cs, core.MetachainShardId, vm.DelegationManagerSCAddress, "getAllContractAddresses", nil)
	require.Nil(t, err)
	delegationContractAddress := output.ReturnData[0]

	// Step 2: Add validator nodes to the delegation contract
	// This step requires generating BLS keys for validators, signing messages, and sending the "addNodes" transaction.
	// Add checks to verify nodes are added successfully.
	validatorSecretKeysBytes, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(validatorSecretKeysBytes)
	require.Nil(t, err)

	signatures := getSignatures(delegationContractAddress, validatorSecretKeysBytes)
	txAddNodes := generateTransaction(validatorOwner.Bytes, 1, delegationContractAddress, zeroValue, addNodesTxData(blsKeys, signatures), gasLimitForAddNodesOperation)
	addNodesTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txAddNodes, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, addNodesTx)

	expectedTopUp := big.NewInt(0).Set(minimumCreateDelegationStakeValue)
	expectedTotalStaked := big.NewInt(0).Set(minimumCreateDelegationStakeValue)
	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddress, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationContractAddress))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddress, "getUserActiveStake", [][]byte{validatorOwner.Bytes})
	require.Nil(t, err)
	require.Equal(t, minimumCreateDelegationStakeValue, big.NewInt(0).SetBytes(output.ReturnData[0]))

	// Step 3: Perform delegation operations
	tx1delegatorA := generateTransaction(delegatorA.Bytes, 0, delegationContractAddress, minimumCreateDelegationStakeValue, "delegate", gasLimitForDelegate)
	delegatorATx1, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx1delegatorA, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegatorATx1)

	expectedTopUp = expectedTopUp.Add(expectedTopUp, minimumCreateDelegationStakeValue)
	expectedTotalStaked = expectedTotalStaked.Add(expectedTotalStaked, minimumCreateDelegationStakeValue)
	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddress, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationContractAddress))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddress, "getUserActiveStake", [][]byte{delegatorA.Bytes})
	require.Nil(t, err)
	require.Equal(t, minimumCreateDelegationStakeValue, big.NewInt(0).SetBytes(output.ReturnData[0]))

	delegateValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(501)) // 501 EGLD
	tx1delegatorB := generateTransaction(delegatorB.Bytes, 0, delegationContractAddress, delegateValue, "delegate", gasLimitForDelegate)
	delegatorBTx1, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx1delegatorB, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegatorBTx1)
	require.Equal(t, delegatorBTx1.SmartContractResults[0].ReturnMessage, "total delegation cap reached")

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddress, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationContractAddress))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddress, "getUserActiveStake", [][]byte{delegatorB.Bytes})
	require.Nil(t, err)
	require.Zero(t, len(output.ReturnData))
	require.Equal(t, "view function works only for existing delegators", output.ReturnMessage)

	// Step 4: Perform stakeNodes

	txStakeNodes := generateTransaction(validatorOwner.Bytes, 2, delegationContractAddress, zeroValue, fmt.Sprintf("stakeNodes@%s", blsKeys[0]), gasLimitForDelegate)
	stakeNodesTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStakeNodes, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeNodesTx)

	require.Equal(t, zeroValue.String(), getBLSTopUpValue(t, metachainNode, delegationContractAddress).String())

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddress, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys := getNodesFromContract(output.ReturnData)
	require.Equal(t, 1, len(stakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(stakedKeys[0]))
	require.Equal(t, 0, len(notStakedKeys))
	require.Equal(t, 0, len(unStakedKeys))

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	testBLSKeyIsInQueueOrAuction(t, metachainNode, delegationContractAddress, blsKeys[0], zeroValue)

	tx2delegatorB := generateTransaction(delegatorB.Bytes, 1, delegationContractAddress, delegateValue, "delegate", gasLimitForDelegate)
	delegatorBTx2, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx2delegatorB, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegatorBTx2)
	require.Equal(t, delegatorBTx2.SmartContractResults[0].ReturnMessage, "total delegation cap reached")

	// check the tx failed

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddress, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, zeroValue.String(), getBLSTopUpValue(t, metachainNode, delegationContractAddress).String())

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddress, "getUserActiveStake", [][]byte{delegatorB.Bytes})
	require.Nil(t, err)
	require.Zero(t, len(output.ReturnData))
	require.Equal(t, "view function works only for existing delegators", output.ReturnMessage)

	delegateValue = big.NewInt(0).Mul(oneEGLD, big.NewInt(500)) // 500 EGLD
	tx3delegatorB := generateTransaction(delegatorB.Bytes, 2, delegationContractAddress, delegateValue, "delegate", gasLimitForDelegate)
	delegatorBTx3, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx3delegatorB, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegatorBTx3)

	expectedTopUp = big.NewInt(0).Set(delegateValue)
	expectedTotalStaked = expectedTotalStaked.Add(expectedTotalStaked, delegateValue)
	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddress, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationContractAddress))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddress, "getUserActiveStake", [][]byte{delegatorB.Bytes})
	require.Nil(t, err)
	require.Equal(t, delegateValue, big.NewInt(0).SetBytes(output.ReturnData[0]))

	delegateValue = big.NewInt(0).Mul(oneEGLD, big.NewInt(20)) // 20 EGLD
	tx1DelegatorC := generateTransaction(delegatorC.Bytes, 0, delegationContractAddress, delegateValue, "delegate", gasLimitForDelegate)
	delegatorCTx1, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx1DelegatorC, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegatorCTx1)
	require.Equal(t, delegatorBTx2.SmartContractResults[0].ReturnMessage, "total delegation cap reached")

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddress, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationContractAddress))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddress, "getUserActiveStake", [][]byte{delegatorC.Bytes})
	require.Nil(t, err)
	require.Zero(t, len(output.ReturnData))
	require.Equal(t, "view function works only for existing delegators", output.ReturnMessage)
}

func executeQuery(cs chainSimulatorIntegrationTests.ChainSimulator, shardID uint32, scAddress []byte, funcName string, args [][]byte) (*dataVm.VMOutputApi, error) {
	output, _, err := cs.GetNodeHandler(shardID).GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress: scAddress,
		FuncName:  funcName,
		Arguments: args,
	})
	return output, err
}

func addNodesTxData(blsKeys []string, sigs [][]byte) string {
	txData := "addNodes"

	for i := range blsKeys {
		txData = txData + "@" + blsKeys[i] + "@" + hex.EncodeToString(sigs[i])
	}

	return txData
}

func getSignatures(msg []byte, blsKeys [][]byte) [][]byte {
	signer := mclsig.NewBlsSigner()

	signatures := make([][]byte, len(blsKeys))
	for i, blsKey := range blsKeys {
		sk, _ := signing.NewKeyGenerator(mcl.NewSuiteBLS12()).PrivateKeyFromByteArray(blsKey)
		signatures[i], _ = signer.Sign(sk, msg)
	}

	return signatures
}

func getNodesFromContract(returnData [][]byte) ([][]byte, [][]byte, [][]byte) {
	var stakedKeys, notStakedKeys, unStakedKeys [][]byte

	for i := 0; i < len(returnData); i += 2 {
		switch string(returnData[i]) {
		case "staked":
			stakedKeys = append(stakedKeys, returnData[i+1])
		case "notStaked":
			notStakedKeys = append(notStakedKeys, returnData[i+1])
		case "unStaked":
			unStakedKeys = append(unStakedKeys, returnData[i+1])
		}
	}
	return stakedKeys, notStakedKeys, unStakedKeys
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

	if len(result.ReturnData[0]) == 0 {
		return big.NewInt(0)
	}

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
