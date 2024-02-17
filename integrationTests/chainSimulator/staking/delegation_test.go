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
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
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
const maxCap = "00"

// no cap
var serviceFee = big.NewInt(3745) // 37.45%
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

	testBLSKeyIsInQueueOrAuction(t, cs, validatorOwner.Bytes, blsKeys[0], addedStakedValue)

	log.Info("Step 4. Execute the MakeNewContractFromValidatorData transaction and test that the key is on queue / auction list and the correct topup")
	txDataField = fmt.Sprintf("makeNewContractFromValidatorData@%s@%s", maxCap, hex.EncodeToString(serviceFee.Bytes()))
	txConvert := generateTransaction(validatorOwner.Bytes, 1, vm.DelegationManagerSCAddress, zeroValue, txDataField, gasLimitForConvertOperation)
	convertTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txConvert, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, convertTx)

	delegationAddressHex := convertTx.Logs.Events[0].Topics[1]
	delegationAddressBech32 := metachainNode.GetCoreComponents().AddressPubKeyConverter().SilentEncode(delegationAddressHex, log)
	delegationAddress := dtos.WalletAddress{
		Bytes:  delegationAddressHex,
		Bech32: delegationAddressBech32,
	}
	log.Info("generated delegation address", "address", delegationAddressBech32)

	testBLSKeyIsInQueueOrAuction(t, cs, delegationAddress.Bytes, blsKeys[0], addedStakedValue)

	log.Info("Step 5. Execute 2 delegation operations of 100 EGLD each, check the topup is 700")
	delegateValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(100))
	expectedTopUp := big.NewInt(0).Mul(oneEGLD, big.NewInt(600))
	expectedTotalStaked := big.NewInt(0).Mul(oneEGLD, big.NewInt(3100))
	delegate(t, cs, delegationAddress, delegator1, 0, delegateValue, delegateValue, expectedTotalStaked, expectedTopUp)

	expectedTopUp = big.NewInt(0).Mul(oneEGLD, big.NewInt(700))
	expectedTotalStaked = big.NewInt(0).Mul(oneEGLD, big.NewInt(3200))
	delegate(t, cs, delegationAddress, delegator2, 0, delegateValue, delegateValue, expectedTotalStaked, expectedTopUp)
	testBLSKeyIsInQueueOrAuction(t, cs, delegationAddress.Bytes, blsKeys[0], expectedTopUp)

	log.Info("6. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 500")
	unDelegateValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(100))
	txDataField = fmt.Sprintf("unDelegate@%s", hex.EncodeToString(unDelegateValue.Bytes()))
	txUnDelegate1 := generateTransaction(delegator1.Bytes, 1, delegationAddress.Bytes, zeroValue, txDataField, gasLimitForDelegate)
	unDelegate1Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnDelegate1, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unDelegate1Tx)

	txDataField = fmt.Sprintf("unDelegate@%s", hex.EncodeToString(unDelegateValue.Bytes()))
	txUnDelegate2 := generateTransaction(delegator2.Bytes, 1, delegationAddress.Bytes, zeroValue, txDataField, gasLimitForDelegate)
	unDelegate2Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnDelegate2, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unDelegate2Tx)

	expectedTopUp = big.NewInt(0).Mul(oneEGLD, big.NewInt(500))
	testBLSKeyIsInQueueOrAuction(t, cs, delegationAddress.Bytes, blsKeys[0], expectedTopUp)
}

func testBLSKeyIsInQueueOrAuction(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, address []byte, blsKey string, expectedTopUp *big.Int) {
	decodedBLSKey, _ := hex.DecodeString(blsKey)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	err := metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	statistics, err := metachainNode.GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)
	require.Equal(t, expectedTopUp.String(), getBLSTopUpValue(t, cs, address).String())

	activationEpoch := metachainNode.GetCoreComponents().EnableEpochsHandler().GetActivationEpoch(common.StakingV4Step1Flag)
	if activationEpoch <= metachainNode.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch() {
		testBLSKeyIsInAuction(t, cs, address, decodedBLSKey, blsKey, expectedTopUp, statistics)
		return
	}

	// in staking ph 2/3.5 we do not find the bls key on the validator statistics
	_, found := statistics[blsKey]
	require.False(t, found)
	require.Equal(t, queuedStatus, getBLSKeyStatus(t, cs, decodedBLSKey))
}

func testBLSKeyIsInAuction(
	t *testing.T,
	cs chainSimulatorIntegrationTests.ChainSimulator,
	address []byte,
	blsKeyBytes []byte,
	blsKey string,
	topUpInAuctionList *big.Int,
	validatorStatistics map[string]*validator.ValidatorStatistics,
) {
	require.Equal(t, stakedStatus, getBLSKeyStatus(t, cs, blsKeyBytes))
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	err := metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	auctionList, err := metachainNode.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)

	expectedNodesInAuctionList := 1
	expectedAuctionListOwnersSize := 1
	currentEpoch := metachainNode.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch()
	if currentEpoch == metachainNode.GetCoreComponents().EnableEpochsHandler().GetActivationEpoch(common.StakingV4Step2Flag) {
		// starting from phase 2, we have the shuffled out nodes from the previous epoch in the action list
		expectedAuctionListOwnersSize += 1
		expectedNodesInAuctionList += 8
	}
	if currentEpoch >= metachainNode.GetCoreComponents().EnableEpochsHandler().GetActivationEpoch(common.StakingV4Step3Flag) {
		// starting from phase 2, we have the shuffled out nodes from the previous epoch in the action list
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

	initialFunds := big.NewInt(0).Mul(oneEGLD, big.NewInt(10000)) // 10000 EGLD for each
	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, initialFunds)
	require.Nil(t, err)

	delegator1, err := cs.GenerateAndMintWalletAddress(core.AllShardId, initialFunds)
	require.Nil(t, err)

	delegator2, err := cs.GenerateAndMintWalletAddress(core.AllShardId, initialFunds)
	require.Nil(t, err)

	maxDelegationCap := big.NewInt(0).Mul(oneEGLD, big.NewInt(51000)) // 51000 EGLD cap
	delegationContract := createDelegationContract(t, cs, validatorOwner, maxDelegationCap, serviceFee, minimumCreateDelegationStakeValue)

	// Step 2: Add validator nodes to the delegation contract
	// This step requires generating BLS keys for validators, signing messages, and sending the "addNodes" transaction.
	// Add checks to verify nodes are added successfully.
	validatorSecretKeysBytes, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(validatorSecretKeysBytes)
	require.Nil(t, err)

	signatures := getSignatures(delegationContract.Bytes, validatorSecretKeysBytes)
	txAddNodes := generateTransaction(validatorOwner.Bytes, 1, delegationContract.Bytes, zeroValue, addNodesTxData(blsKeys, signatures), gasLimitForAddNodesOperation)
	addNodesTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txAddNodes, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, addNodesTx)

	output, err := executeQuery(cs, core.MetachainShardId, delegationContract.Bytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys := getNodesFromContract(output.ReturnData)
	require.Equal(t, 0, len(stakedKeys))
	require.Equal(t, 1, len(notStakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(notStakedKeys[0]))
	require.Equal(t, 0, len(unStakedKeys))

	expectedTopUp := big.NewInt(0).Set(minimumCreateDelegationStakeValue)
	expectedTotalStaked := big.NewInt(0).Set(minimumCreateDelegationStakeValue)
	output, err = executeQuery(cs, core.MetachainShardId, delegationContract.Bytes, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, cs, delegationContract.Bytes))

	// Step 3: Perform delegation operations
	expectedTopUp = expectedTopUp.Add(expectedTopUp, minimumCreateDelegationStakeValue)
	expectedTotalStaked = expectedTotalStaked.Add(expectedTotalStaked, minimumCreateDelegationStakeValue)
	delegateValue := big.NewInt(0).Set(minimumCreateDelegationStakeValue)
	delegate(t, cs, delegationContract, delegator1, 0, delegateValue, delegateValue, expectedTotalStaked, expectedTopUp)

	expectedTopUp = expectedTopUp.Add(expectedTopUp, minimumCreateDelegationStakeValue)
	expectedTotalStaked = expectedTotalStaked.Add(expectedTotalStaked, minimumCreateDelegationStakeValue)
	delegate(t, cs, delegationContract, delegator2, 0, delegateValue, delegateValue, expectedTotalStaked, expectedTopUp)

	// Step 4: Perform stakeNodes

	txStakeNodes := generateTransaction(validatorOwner.Bytes, 2, delegationContract.Bytes, zeroValue, fmt.Sprintf("stakeNodes@%s", blsKeys[0]), gasLimitForStakeOperation)
	stakeNodesTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStakeNodes, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeNodesTx)

	expectedTopUp = big.NewInt(0).Set(minimumCreateDelegationStakeValue)
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, cs, delegationContract.Bytes))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContract.Bytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys = getNodesFromContract(output.ReturnData)
	require.Equal(t, 1, len(stakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(stakedKeys[0]))
	require.Equal(t, 0, len(notStakedKeys))
	require.Equal(t, 0, len(unStakedKeys))

	// Make block finalized
	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	testBLSKeyIsInQueueOrAuction(t, cs, delegationContract.Bytes, blsKeys[0], expectedTopUp)

	// Step 5: Perform unDelegate from 1 user
	// The nodes should remain in the staked state
	// The total active stake should be reduced by the amount undelegated

	txUndelegate1 := generateTransaction(delegator1.Bytes, 1, delegationContract.Bytes, zeroValue, fmt.Sprintf("unDelegate@%s", hex.EncodeToString(minimumCreateDelegationStakeValue.Bytes())), gasLimitForUndelegateOperation)
	undelegate1Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUndelegate1, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, undelegate1Tx)

	expectedTopUp = expectedTopUp.Sub(expectedTopUp, minimumCreateDelegationStakeValue)
	expectedTotalStaked = expectedTotalStaked.Sub(expectedTotalStaked, minimumCreateDelegationStakeValue)
	output, err = executeQuery(cs, core.MetachainShardId, delegationContract.Bytes, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp.String(), getBLSTopUpValue(t, cs, delegationContract.Bytes).String())

	output, err = executeQuery(cs, core.MetachainShardId, delegationContract.Bytes, "getUserActiveStake", [][]byte{delegator1.Bytes})
	require.Nil(t, err)
	require.Equal(t, zeroValue, big.NewInt(0).SetBytes(output.ReturnData[0]))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContract.Bytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys = getNodesFromContract(output.ReturnData)
	require.Equal(t, 1, len(stakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(stakedKeys[0]))
	require.Equal(t, 0, len(notStakedKeys))
	require.Equal(t, 0, len(unStakedKeys))

	// Step 6: Perform unDelegate from last user
	// The nodes should change to unStaked state
	// The total active stake should be reduced by the amount undelegated

	txUndelegate2 := generateTransaction(delegator2.Bytes, 1, delegationContract.Bytes, zeroValue, fmt.Sprintf("unDelegate@%s", hex.EncodeToString(minimumCreateDelegationStakeValue.Bytes())), gasLimitForUndelegateOperation)
	undelegate2Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUndelegate2, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, undelegate2Tx)

	output, err = executeQuery(cs, core.MetachainShardId, delegationContract.Bytes, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, "1250000000000000000000", big.NewInt(0).SetBytes(output.ReturnData[0]).String())
	require.Equal(t, zeroValue, getBLSTopUpValue(t, cs, delegationContract.Bytes))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContract.Bytes, "getUserActiveStake", [][]byte{delegator2.Bytes})
	require.Nil(t, err)
	require.Equal(t, "0", big.NewInt(0).SetBytes(output.ReturnData[0]).String())

	// still staked until epoch change
	output, err = executeQuery(cs, core.MetachainShardId, delegationContract.Bytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys = getNodesFromContract(output.ReturnData)
	require.Equal(t, 1, len(stakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(stakedKeys[0]))
	require.Equal(t, 0, len(notStakedKeys))
	require.Equal(t, 0, len(unStakedKeys))

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch + 1)
	require.Nil(t, err)

	output, err = executeQuery(cs, core.MetachainShardId, delegationContract.Bytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys = getNodesFromContract(output.ReturnData)
	require.Equal(t, 0, len(stakedKeys))
	require.Equal(t, 0, len(notStakedKeys))
	require.Equal(t, 1, len(unStakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(unStakedKeys[0]))
}

func createDelegationContract(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, validatorOwner dtos.WalletAddress, maxDelegationCap *big.Int, serviceFee *big.Int, initialStakeAmount *big.Int) dtos.WalletAddress {
	txCreateDelegationContract := generateTransaction(validatorOwner.Bytes, 0, vm.DelegationManagerSCAddress, initialStakeAmount,
		fmt.Sprintf("createNewDelegationContract@%s@%s", hex.EncodeToString(maxDelegationCap.Bytes()), hex.EncodeToString(serviceFee.Bytes())),
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

	delegationDTO := dtos.WalletAddress{
		Bytes:  output.ReturnData[0],
		Bech32: delegationContractAddress,
	}
	checkUserDelegatedAmount(t, cs, delegationDTO, validatorOwner, initialStakeAmount)
	return delegationDTO
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
	delegationContract := createDelegationContract(t, cs, validatorOwner, maxDelegationCap, serviceFee, minimumCreateDelegationStakeValue)

	// Step 2: Add validator nodes to the delegation contract
	// This step requires generating BLS keys for validators, signing messages, and sending the "addNodes" transaction.
	// Add checks to verify nodes are added successfully.
	validatorSecretKeysBytes, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(validatorSecretKeysBytes)
	require.Nil(t, err)

	signatures := getSignatures(delegationContract.Bytes, validatorSecretKeysBytes)
	txAddNodes := generateTransaction(validatorOwner.Bytes, 1, delegationContract.Bytes, zeroValue, addNodesTxData(blsKeys, signatures), gasLimitForAddNodesOperation)
	addNodesTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txAddNodes, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, addNodesTx)

	expectedTopUp := big.NewInt(0).Set(minimumCreateDelegationStakeValue)
	expectedTotalStaked := big.NewInt(0).Set(minimumCreateDelegationStakeValue)
	checkDelegationContractTotalStaked(t, cs, delegationContract, expectedTotalStaked, expectedTopUp)

	// Step 3: Perform delegation operations

	delegateValue := big.NewInt(0).Set(minimumCreateDelegationStakeValue)
	expectedTopUp = expectedTopUp.Add(expectedTopUp, minimumCreateDelegationStakeValue)
	expectedTotalStaked = expectedTotalStaked.Add(expectedTotalStaked, minimumCreateDelegationStakeValue)
	delegate(t, cs, delegationContract, delegatorA, 0, delegateValue, delegateValue, expectedTotalStaked, expectedTopUp)

	delegateValue = big.NewInt(0).Mul(oneEGLD, big.NewInt(501)) // 501 EGLD
	delegateFailWithMessage(t, cs, delegationContract, delegatorB, 0, delegateValue, "total delegation cap reached")
	checkDelegationContractTotalStaked(t, cs, delegationContract, expectedTotalStaked, expectedTopUp)
	checkUserNotDelegator(t, cs, delegationContract, delegatorB)

	// Step 4: Perform stakeNodes

	txStakeNodes := generateTransaction(validatorOwner.Bytes, 2, delegationContract.Bytes, zeroValue, fmt.Sprintf("stakeNodes@%s", blsKeys[0]), gasLimitForDelegate)
	stakeNodesTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStakeNodes, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeNodesTx)

	require.Equal(t, zeroValue.String(), getBLSTopUpValue(t, cs, delegationContract.Bytes).String())

	output, err := executeQuery(cs, core.MetachainShardId, delegationContract.Bytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys := getNodesFromContract(output.ReturnData)
	require.Equal(t, 1, len(stakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(stakedKeys[0]))
	require.Equal(t, 0, len(notStakedKeys))
	require.Equal(t, 0, len(unStakedKeys))

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	testBLSKeyIsInQueueOrAuction(t, cs, delegationContract.Bytes, blsKeys[0], zeroValue)

	delegateFailWithMessage(t, cs, delegationContract, delegatorB, 1, delegateValue, "total delegation cap reached")
	checkDelegationContractTotalStaked(t, cs, delegationContract, expectedTotalStaked, zeroValue)
	checkUserNotDelegator(t, cs, delegationContract, delegatorB)

	delegateValue = big.NewInt(0).Mul(oneEGLD, big.NewInt(500)) // 500 EGLD
	expectedTopUp = big.NewInt(0).Set(delegateValue)
	expectedTotalStaked = expectedTotalStaked.Add(expectedTotalStaked, delegateValue)
	delegate(t, cs, delegationContract, delegatorB, 2, delegateValue, delegateValue, expectedTotalStaked, expectedTopUp)

	delegateValue = big.NewInt(0).Mul(oneEGLD, big.NewInt(20)) // 20 EGLD
	delegateFailWithMessage(t, cs, delegationContract, delegatorC, 0, delegateValue, "total delegation cap reached")
	checkDelegationContractTotalStaked(t, cs, delegationContract, expectedTotalStaked, expectedTopUp)
	checkUserNotDelegator(t, cs, delegationContract, delegatorC)
}

func delegate(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, delegationContract dtos.WalletAddress, delegator dtos.WalletAddress, nonce uint64, delegateValue *big.Int, expectedDelegated *big.Int, expectedTotalStaked *big.Int, expectedTopUp *big.Int) {
	tx3delegatorB := generateTransaction(delegator.Bytes, nonce, delegationContract.Bytes, delegateValue, "delegate", gasLimitForDelegate)
	delegatorBTx3, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx3delegatorB, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegatorBTx3)

	//TODO: check tx successfull

	checkDelegationContractTotalStaked(t, cs, delegationContract, expectedTotalStaked, expectedTopUp)
	checkUserDelegatedAmount(t, cs, delegationContract, delegator, expectedDelegated)
}

func delegateFailWithMessage(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, delegationContract dtos.WalletAddress, delegator dtos.WalletAddress, nonce uint64, delegateValue *big.Int, message string) {
	tx := generateTransaction(delegator.Bytes, nonce, delegationContract.Bytes, delegateValue, "delegate", gasLimitForDelegate)
	txSent, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txSent)
	require.Equal(t, txSent.SmartContractResults[0].ReturnMessage, message)
}

func checkDelegationContractTotalStaked(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, delegationContract dtos.WalletAddress, expectedTotalStaked *big.Int, expectedTopUp *big.Int) {
	output, err := executeQuery(cs, core.MetachainShardId, delegationContract.Bytes, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, cs, delegationContract.Bytes))
}

func checkUserDelegatedAmount(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, delegationContract dtos.WalletAddress, delegator dtos.WalletAddress, expectedDelegated *big.Int) {
	output, err := executeQuery(cs, core.MetachainShardId, delegationContract.Bytes, "getUserActiveStake", [][]byte{delegator.Bytes})
	require.Nil(t, err)
	require.Equal(t, expectedDelegated, big.NewInt(0).SetBytes(output.ReturnData[0]))
}

func checkUserNotDelegator(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, delegationContract dtos.WalletAddress, delegatorC dtos.WalletAddress) {
	output, err := executeQuery(cs, core.MetachainShardId, delegationContract.Bytes, "getUserActiveStake", [][]byte{delegatorC.Bytes})
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

func getBLSKeyStatus(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, blsKey []byte) string {
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
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

func getBLSTopUpValue(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, address []byte) *big.Int {
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
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

func TestChainSimulator_DelegationContractInteractions(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
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

		defer cs.Close()

		testChainSimulatorDelegationContractInteractions(t, cs, 1)
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

		defer cs.Close()

		testChainSimulatorDelegationContractInteractions(t, cs, 2)
	})

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

		testChainSimulatorDelegationContractInteractions(t, cs, 3)
	})

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

		testChainSimulatorDelegationContractInteractions(t, cs, 4)
	})
}

func testChainSimulatorDelegationContractInteractions(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	initialFunds := big.NewInt(0).Mul(oneEGLD, big.NewInt(10000)) // 10000 EGLD for each
	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, initialFunds)
	require.Nil(t, err)

	_, err = cs.GenerateAndMintWalletAddress(core.AllShardId, initialFunds)
	require.Nil(t, err)

	// 33 deploy a staking provider with below 1250 egld does not work

	belowMinimumCreateDelegationStakeValue := big.NewInt(0).Sub(minimumCreateDelegationStakeValue, oneEGLD) // 1000 EGLD
	maxDelegationCap := big.NewInt(0).Mul(oneEGLD, big.NewInt(3000))                                        // 3000 EGLD cap
	txCreateDelegationContract := generateTransaction(validatorOwner.Bytes, 0, vm.DelegationManagerSCAddress, belowMinimumCreateDelegationStakeValue,
		fmt.Sprintf("createNewDelegationContract@%s@%s", hex.EncodeToString(maxDelegationCap.Bytes()), hex.EncodeToString(serviceFee.Bytes())),
		gasLimitForDelegationContractCreationOperation)
	createDelegationContractTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txCreateDelegationContract, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, createDelegationContractTx)

	// 32 deploy a staking provider with 1250 egld works
	_ = createDelegationContract(t, cs, validatorOwner, maxDelegationCap, serviceFee, minimumCreateDelegationStakeValue)

	// 35 delegate by staking provider works
	//delegatorB, err := cs.GenerateAndMintWalletAddress(core.AllShardId, initialFunds)
	//require.Nil(t, err)
	//
	//delegatorC, err := cs.GenerateAndMintWalletAddress(core.AllShardId, initialFunds)
	//require.Nil(t, err)

	// 41
	// 42
	// 43
	// 44
	// 45
	// 47
	// undelegate
	// 46
	// 36
	// 37
	// 38
	// 39

}
