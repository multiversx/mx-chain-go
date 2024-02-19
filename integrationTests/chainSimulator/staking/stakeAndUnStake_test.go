package staking

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/validator"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig             = "../../../cmd/node/config/"
	maxNumOfBlockToGenerateWhenExecutingTx = 7
)

var log = logger.GetOrCreate("integrationTests/chainSimulator")

type ValidatorExpectations struct {
	ExpectedTopUp    *big.Int
	NumExpectedNodes int64
	Address          string
}

// Internal test scenario #22
func TestChainSimulator_MultipleStakeUnstakeIsCalculatedProperly(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	// Test scenario done in staking v4 phase step 3
	// Check that if you do multiple stake and unstake actions, it will be calculated live and the values stand
	// until the end of the epoch
	// Attributes on API are changing after every transaction that impacts the API
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

		testChainSimulatorMultipleStakeUnstakeIsCalculatedProperly(t, cs, 4)
	})
}

// Internal test scenario #10
func TestChainSimulator_UnstakeFundsHappyFlow(t *testing.T) {
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

		testChainSimulatorUnstakeFundsHappyFlow(t, cs, 1)
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

		testChainSimulatorUnstakeFundsHappyFlow(t, cs, 2)
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

		testChainSimulatorUnstakeFundsHappyFlow(t, cs, 3)
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

		testChainSimulatorUnstakeFundsHappyFlow(t, cs, 4)
	})
}

func testChainSimulatorUnstakeFundsHappyFlow(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	//create a validator
	validatorBytes := generateWalletAddressBytes()
	validator, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(validatorBytes)
	validatorShardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(validatorBytes)

	validatorContractAddress := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l"
	validatorContractAddressBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(validatorContractAddress)

	// Set the initial state for the owner
	initialFunds := big.NewInt(0).Mul(oneEGLD, big.NewInt(10000)) // 10000 EGLD

	addresses := []*dtos.AddressState{
		{Address: validator, Balance: initialFunds.String()},
	}

	err = cs.SetStateMultiple(addresses)
	require.Nil(t, err)

	// Step 1 --- add a new validator key in the chain simulator
	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(2)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)

	// Stake one node per validator owner
	stakeValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(5200)) // 2600*2 EGLD
	validatorOwnerAStatus, _, err := cs.GetNodeHandler(validatorShardID).GetFacadeHandler().GetAccount(validator, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	stakeNodes(t, cs, validatorOwnerAStatus.Nonce, blsKeys[:2], validatorBytes, validatorContractAddressBytes, 2, stakeValue)

	// Finalize block
	err = cs.GenerateBlocks(1)
	require.Nil(t, err)
	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)

	// Verify user active stake
	output, err := executeQuery(cs, core.MetachainShardId, validatorContractAddressBytes, "getTotalStaked", [][]byte{validatorBytes})
	require.Nil(t, err)
	//check that total staked is 5200
	require.Equal(t, "5200000000000000000000", string(output.ReturnData[0]))

	// Unstake 50 EGLD
	unstakeValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(50))
	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	validatorOwnerAStatus, _, err = cs.GetNodeHandler(validatorShardID).GetFacadeHandler().GetAccount(validator, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	unstakeTokens(t, cs, validatorOwnerAStatus.Nonce, unstakeValue, validatorBytes, validatorContractAddressBytes)

	// check balance of the user is smaller than before sending the tx
	validatorOwnerAStatusAfter, _, err := cs.GetNodeHandler(validatorShardID).GetFacadeHandler().GetAccount(validator, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	// validatorOwnerStatus has to be greater than validatorOwnerStatusAfter
	require.True(t, validatorOwnerAStatus.Balance > validatorOwnerAStatusAfter.Balance)

	// Finalize block
	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	// Verify user active stake
	output, err = executeQuery(cs, core.MetachainShardId, validatorContractAddressBytes, "getTotalStaked", [][]byte{validatorBytes})
	require.Nil(t, err)
	//check that total staked is 5199
	require.Equal(t, "5150000000000000000000", string(output.ReturnData[0]))

	// Verify user active stake
	output, err = executeQuery(cs, core.MetachainShardId, validatorContractAddressBytes, "getUnStakedTokensList", [][]byte{validatorBytes})
	require.Nil(t, err)
}

func testChainSimulatorMultipleStakeUnstakeIsCalculatedProperly(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	//create 2 validator owners and 2 delegators
	validatorOwnerABytes := generateWalletAddressBytes()
	validatorOwnerA, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(validatorOwnerABytes)
	validatorOwnerAShardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(validatorOwnerABytes)
	validatorOwnerBBytes := generateWalletAddressBytes()
	validatorOwnerB, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(validatorOwnerBBytes)
	validatorOwnerBShardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(validatorOwnerBBytes)

	validatorContractAddress := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l"
	validatorContractAddressBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(validatorContractAddress)

	// Set the initial state for the owner and the 2 delegators
	initialFunds := big.NewInt(0).Mul(oneEGLD, big.NewInt(10000)) // 10000 EGLD for each

	addresses := []*dtos.AddressState{
		{Address: validatorOwnerA, Balance: initialFunds.String()},
		{Address: validatorOwnerB, Balance: initialFunds.String()},
	}
	err = cs.SetStateMultiple(addresses)
	require.Nil(t, err)

	// Step 1 --- add a new validator key in the chain simulator
	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(2)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)

	// Stake one node per validator owner
	stakeValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(3000)) // 3000 EGLD
	validatorOwnerAStatus, _, err := cs.GetNodeHandler(validatorOwnerAShardID).GetFacadeHandler().GetAccount(validatorOwnerA, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	stakeNodes(t, cs, validatorOwnerAStatus.Nonce, blsKeys[:1], validatorOwnerABytes, validatorContractAddressBytes, 1, stakeValue)
	validatorOwnerBStatus, _, err := cs.GetNodeHandler(validatorOwnerBShardID).GetFacadeHandler().GetAccount(validatorOwnerB, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	stakeNodes(t, cs, validatorOwnerBStatus.Nonce, blsKeys[1:], validatorOwnerBBytes, validatorContractAddressBytes, 1, stakeValue)

	// Finalize block
	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	//get auction list
	auctionList, err := metachainNode.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)

	// 1st check - both validators should have 500 EGLD topUp
	validatorOwnersExpectations := map[string]ValidatorExpectations{
		validatorOwnerA: {
			ExpectedTopUp:    big.NewInt(0).Mul(oneEGLD, big.NewInt(500)),
			NumExpectedNodes: 1,
		},
		validatorOwnerB: {
			ExpectedTopUp:    big.NewInt(0).Mul(oneEGLD, big.NewInt(500)),
			NumExpectedNodes: 1,
		},
	}
	checkAuctionList(t, auctionList, validatorOwnersExpectations)

	//2nd check after unstaking 100 EGLD from A and nothing from B ------------------------------------
	validatorOwnerAStatus, _, err = cs.GetNodeHandler(validatorOwnerAShardID).GetFacadeHandler().GetAccount(validatorOwnerA, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	unstakeValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(100)) // 100 EGLD
	unstakeTokens(t, cs, validatorOwnerAStatus.Nonce, unstakeValue, validatorOwnerABytes, validatorContractAddressBytes)

	// Finalize block
	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	// check position of validator A in auction list is below validator B
	//get auction list
	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	auctionList, err = metachainNode.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)

	validatorOwnersExpectations = map[string]ValidatorExpectations{
		validatorOwnerA: {
			ExpectedTopUp:    big.NewInt(0).Mul(oneEGLD, big.NewInt(400)),
			NumExpectedNodes: 1,
		},
		validatorOwnerB: {
			ExpectedTopUp:    big.NewInt(0).Mul(oneEGLD, big.NewInt(500)),
			NumExpectedNodes: 1,
		},
	}

	checkAuctionList(t, auctionList, validatorOwnersExpectations)

	//check if delegation contract: A is above B in auction list
	validatorAIndex, validatorBIndex := checkValidatorPositions(auctionList, blsKeys[0], blsKeys[1])
	require.True(t, validatorAIndex > validatorBIndex)

	//3rd check after staking 50 EGLD from A and nothing from B ------------------------------------
	validatorOwnerAStatus, _, err = cs.GetNodeHandler(validatorOwnerAShardID).GetFacadeHandler().GetAccount(validatorOwnerA, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	stakeValue = big.NewInt(0).Mul(oneEGLD, big.NewInt(50)) // 50 EGLD
	stakeNodes(t, cs, validatorOwnerAStatus.Nonce, blsKeys[:1], validatorOwnerABytes, validatorContractAddressBytes, 1, stakeValue)

	// Finalize block
	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	// check position of validator A in auction list is below validator B
	//get auction list
	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	auctionList, err = metachainNode.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)

	validatorOwnersExpectations = map[string]ValidatorExpectations{
		validatorOwnerA: {
			ExpectedTopUp:    big.NewInt(0).Mul(oneEGLD, big.NewInt(450)),
			NumExpectedNodes: 1,
		},
		validatorOwnerB: {
			ExpectedTopUp:    big.NewInt(0).Mul(oneEGLD, big.NewInt(500)),
			NumExpectedNodes: 1,
		},
	}
	checkAuctionList(t, auctionList, validatorOwnersExpectations)

	//check if delegation contract: A is above B in auction list
	validatorAIndex, validatorBIndex = checkValidatorPositions(auctionList, blsKeys[0], blsKeys[1])

	require.True(t, validatorAIndex > validatorBIndex)

	// 4th step - stake 1000 EGLD to A and check if A is now on upper position
	validatorOwnerAStatus, _, err = cs.GetNodeHandler(validatorOwnerAShardID).GetFacadeHandler().GetAccount(validatorOwnerA, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	stakeValue = big.NewInt(0).Mul(oneEGLD, big.NewInt(1000)) // 1000 EGLD
	stakeNodes(t, cs, validatorOwnerAStatus.Nonce, blsKeys[:1], validatorOwnerABytes, validatorContractAddressBytes, 1, stakeValue)

	// Finalize block
	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	// check position of validator A in auction list is below validator B
	//get auction list
	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	auctionList, err = metachainNode.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)

	validatorOwnersExpectations = map[string]ValidatorExpectations{
		validatorOwnerA: {
			ExpectedTopUp:    big.NewInt(0).Mul(oneEGLD, big.NewInt(1450)),
			NumExpectedNodes: 1,
		},
		validatorOwnerB: {
			ExpectedTopUp:    big.NewInt(0).Mul(oneEGLD, big.NewInt(500)),
			NumExpectedNodes: 1,
		},
	}
	checkAuctionList(t, auctionList, validatorOwnersExpectations)

	//check if delegation contract: A is above B in auction list
	validatorAIndex, validatorBIndex = checkValidatorPositions(auctionList, blsKeys[0], blsKeys[1])

	require.True(t, validatorAIndex < validatorBIndex)

	//5th check after unstaking 950 EGLD from A and nothing from B ------------------------------------
	validatorOwnerAStatus, _, err = cs.GetNodeHandler(validatorOwnerAShardID).GetFacadeHandler().GetAccount(validatorOwnerA, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	unstakeValue = big.NewInt(0).Mul(oneEGLD, big.NewInt(950)) // 950 EGLD
	unstakeTokens(t, cs, validatorOwnerAStatus.Nonce, unstakeValue, validatorOwnerABytes, validatorContractAddressBytes)

	// Finalize block
	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	//get auction list
	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	auctionList, err = metachainNode.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)

	validatorOwnersExpectations = map[string]ValidatorExpectations{
		validatorOwnerA: {
			ExpectedTopUp:    big.NewInt(0).Mul(oneEGLD, big.NewInt(500)),
			NumExpectedNodes: 1,
		},
		validatorOwnerB: {
			ExpectedTopUp:    big.NewInt(0).Mul(oneEGLD, big.NewInt(500)),
			NumExpectedNodes: 1,
		},
	}

	//checks also that both validators have same amount of top-up
	checkAuctionList(t, auctionList, validatorOwnersExpectations)
}

func checkValidatorPositions(auctionList []*common.AuctionListValidatorAPIResponse, blsKeyA, blsKeyB string) (int, int) {
	validatorAPosition, validatorBPosition := -1, -1
	for i, position := range auctionList {
		if strings.EqualFold(position.Nodes[0].BlsKey, blsKeyA) {
			validatorAPosition = i
		}
		if strings.EqualFold(position.Nodes[0].BlsKey, blsKeyB) {
			validatorBPosition = i
		}

		// Once both validators are found, no need to continue the loop
		if validatorAPosition != -1 && validatorBPosition != -1 {
			break
		}
	}
	// Return the positions in auction list
	return validatorAPosition, validatorBPosition
}

func checkAuctionList(t *testing.T, auctionList []*common.AuctionListValidatorAPIResponse, validatorOwnersExpectations map[string]ValidatorExpectations) {
	for _, validator := range auctionList {
		expectations, ok := validatorOwnersExpectations[validator.Owner]
		if !ok {
			// Skip if validators are not among the specified validatorOwners
			continue
		}

		// Verify the aggregate data for each validator
		require.Equal(t, expectations.NumExpectedNodes, int64(len(validator.Nodes)), "The number of nodes does not match the expectations for validator: "+validator.Owner)

		require.Equal(t, expectations.ExpectedTopUp.String(), validator.TotalTopUp)
	}
}

// TODO scenarios
// Make a staking provider with max num of nodes
// DO a merge transaction

// Test scenario
// 1. Add a new validator private key in the multi key handler
// 2. Do a stake transaction for the validator key
// 3. Do an unstake transaction (to make a place for the new validator)
// 4. Check if the new validator has generated rewards
func TestChainSimulator_AddValidatorKey(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:   false,
		TempDir:                  t.TempDir(),
		PathToInitialConfig:      defaultPathToInitialConfig,
		NumOfShards:              numOfShards,
		GenesisTimestamp:         startTime,
		RoundDurationInMillis:    roundDurationInMillis,
		RoundsPerEpoch:           roundsPerEpoch,
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         3,
		MetaChainMinNodes:        3,
		NumNodesWaitingListMeta:  0,
		NumNodesWaitingListShard: 0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			newNumNodes := cfg.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake + 8 // 8 nodes until new nodes will be placed on queue
			configs.SetMaxNumberOfNodesInConfigs(cfg, newNumNodes, numOfShards)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(30)
	require.Nil(t, err)

	// Step 1 --- add a new validator key in the chain simulator
	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)

	newValidatorOwner := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	newValidatorOwnerBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(newValidatorOwner)
	rcv := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l"
	rcvAddrBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(rcv)

	// Step 2 --- set an initial balance for the address that will initialize all the transactions
	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl",
			Balance: "10000000000000000000000",
		},
	})
	require.Nil(t, err)

	// Step 3 --- generate and send a stake transaction with the BLS key of the validator key that was added at step 1
	stakeValue, _ := big.NewInt(0).SetString("2500000000000000000000", 10)
	tx := &transaction.Transaction{
		Nonce:     0,
		Value:     stakeValue,
		SndAddr:   newValidatorOwnerBytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("stake@01@%s@010101", blsKeys[0])),
		GasLimit:  50_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	shardIDValidatorOwner := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(newValidatorOwnerBytes)
	accountValidatorOwner, _, err := cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(newValidatorOwner, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceBeforeActiveValidator := accountValidatorOwner.Balance

	// Step 5 --- create an unStake transaction with the bls key of an initial validator and execute the transaction to make place for the validator that was added at step 3
	firstValidatorKey, err := cs.GetValidatorPrivateKeys()[0].GeneratePublic().ToByteArray()
	require.Nil(t, err)

	initialAddressWithValidators := cs.GetInitialWalletKeys().InitialWalletWithStake.Address
	senderBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(initialAddressWithValidators)
	shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(senderBytes)
	initialAccount, _, err := cs.GetNodeHandler(shardID).GetFacadeHandler().GetAccount(initialAddressWithValidators, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	tx = &transaction.Transaction{
		Nonce:     initialAccount.Nonce,
		Value:     big.NewInt(0),
		SndAddr:   senderBytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("unStake@%s", hex.EncodeToString(firstValidatorKey))),
		GasLimit:  50_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)

	// Step 6 --- generate 8 epochs to get rewards
	err = cs.GenerateBlocksUntilEpochIsReached(8)
	require.Nil(t, err)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	validatorStatistics, err := metachainNode.GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)
	checkValidatorsRating(t, validatorStatistics)

	accountValidatorOwner, _, err = cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(newValidatorOwner, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceAfterActiveValidator := accountValidatorOwner.Balance

	log.Info("balance before validator", "value", balanceBeforeActiveValidator)
	log.Info("balance after validator", "value", balanceAfterActiveValidator)

	balanceBeforeBig, _ := big.NewInt(0).SetString(balanceBeforeActiveValidator, 10)
	balanceAfterBig, _ := big.NewInt(0).SetString(balanceAfterActiveValidator, 10)
	diff := balanceAfterBig.Sub(balanceAfterBig, balanceBeforeBig)
	log.Info("difference", "value", diff.String())

	// Step 7 --- check the balance of the validator owner has been increased
	require.True(t, diff.Cmp(big.NewInt(0)) > 0)
}

func TestChainSimulator_AddANewValidatorAfterStakingV4(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            numOfShards,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       100,
		MetaChainMinNodes:      100,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			cfg.GeneralConfig.ValidatorStatistics.CacheRefreshIntervalInSec = 1
			newNumNodes := cfg.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake + 8 // 8 nodes until new nodes will be placed on queue
			configs.SetMaxNumberOfNodesInConfigs(cfg, newNumNodes, numOfShards)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(150)
	require.Nil(t, err)

	// Step 1 --- add a new validator key in the chain simulator
	numOfNodes := 20
	validatorSecretKeysBytes, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(numOfNodes)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(validatorSecretKeysBytes)
	require.Nil(t, err)

	newValidatorOwner := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	newValidatorOwnerBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(newValidatorOwner)
	rcv := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l"
	rcvAddrBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(rcv)

	// Step 2 --- set an initial balance for the address that will initialize all the transactions
	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl",
			Balance: "1000000000000000000000000",
		},
	})
	require.Nil(t, err)

	// Step 3 --- generate and send a stake transaction with the BLS keys of the validators key that were added at step 1
	validatorData := ""
	for _, blsKey := range blsKeys {
		validatorData += fmt.Sprintf("@%s@010101", blsKey)
	}

	numOfNodesHex := hex.EncodeToString(big.NewInt(int64(numOfNodes)).Bytes())
	stakeValue, _ := big.NewInt(0).SetString("51000000000000000000000", 10)
	tx := &transaction.Transaction{
		Nonce:     0,
		Value:     stakeValue,
		SndAddr:   newValidatorOwnerBytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("stake@%s%s", numOfNodesHex, validatorData)),
		GasLimit:  500_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txFromNetwork, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txFromNetwork)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	results, err := metachainNode.GetFacadeHandler().AuctionListApi()
	require.Nil(t, err)
	require.Equal(t, newValidatorOwner, results[0].Owner)
	require.Equal(t, 20, len(results[0].Nodes))
	checkTotalQualified(t, results, 8)

	err = cs.GenerateBlocks(100)
	require.Nil(t, err)

	results, err = cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().AuctionListApi()
	require.Nil(t, err)
	checkTotalQualified(t, results, 0)
}

// Internal test scenario #4 #5 #6
// do stake
// do unStake
// do unBondNodes
// do unBondTokens
func TestChainSimulatorStakeUnStakeUnBond(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		testStakeUnStakeUnBond(t, 1)
	})

	t.Run("staking ph 4 step 1 active", func(t *testing.T) {
		testStakeUnStakeUnBond(t, 4)
	})

	t.Run("staking ph 4 step 2 active", func(t *testing.T) {
		testStakeUnStakeUnBond(t, 5)
	})

	t.Run("staking ph 4 step 3 active", func(t *testing.T) {
		testStakeUnStakeUnBond(t, 6)
	})
}

func testStakeUnStakeUnBond(t *testing.T, targetEpoch int32) {
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}
	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            numOfShards,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       3,
		MetaChainMinNodes:      3,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriod = 1
			cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 1
			newNumNodes := cfg.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake + 10
			configs.SetMaxNumberOfNodesInConfigs(cfg, newNumNodes, numOfShards)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	privateKeys, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(privateKeys)
	require.Nil(t, err)

	mintValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(2600))
	walletAddressShardID := uint32(0)
	walletAddress, err := cs.GenerateAndMintWalletAddress(walletAddressShardID, mintValue)
	require.Nil(t, err)

	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], mockBLSSignature)
	txStake := generateTransaction(walletAddress.Bytes, 0, vm.ValidatorSCAddress, minimumStakeValue, txDataField, gasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	bls0, _ := hex.DecodeString(blsKeys[0])
	blsKeyStatus := getBLSKeyStatus(t, metachainNode, bls0)
	require.Equal(t, "staked", blsKeyStatus)

	// do unStake
	txUnStake := generateTransaction(walletAddress.Bytes, 1, vm.ValidatorSCAddress, zeroValue, fmt.Sprintf("unStake@%s", blsKeys[0]), gasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	blsKeyStatus = getBLSKeyStatus(t, metachainNode, bls0)
	require.Equal(t, "unStaked", blsKeyStatus)

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch + 1)
	require.Nil(t, err)

	// do unBond
	txUnBond := generateTransaction(walletAddress.Bytes, 2, vm.ValidatorSCAddress, zeroValue, fmt.Sprintf("unBondNodes@%s", blsKeys[0]), gasLimitForStakeOperation)
	unBondTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnBond, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unBondTx)

	// do claim
	txClaim := generateTransaction(walletAddress.Bytes, 3, vm.ValidatorSCAddress, zeroValue, "unBondTokens", gasLimitForStakeOperation)
	claimTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txClaim, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, claimTx)

	err = cs.GenerateBlocks(5)
	require.Nil(t, err)

	// check tokens are in the wallet balance
	walletAccount, _, err := cs.GetNodeHandler(walletAddressShardID).GetFacadeHandler().GetAccount(walletAddress.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	walletBalanceBig, _ := big.NewInt(0).SetString(walletAccount.Balance, 10)
	require.True(t, walletBalanceBig.Cmp(minimumStakeValue) > 0)
}

func stakeNodes(t *testing.T, cm chainSimulatorIntegrationTests.ChainSimulator, nonce uint64, blsKeys []string, newValidatorOwnerBytes, rcvAddrBytes []byte, numOfNodes int, stakeValue *big.Int) {
	validatorData := ""
	for _, blsKey := range blsKeys {
		validatorData += fmt.Sprintf("@%s@010101", blsKey)
	}

	numOfNodesHex := hex.EncodeToString(big.NewInt(int64(numOfNodes)).Bytes())
	tx := &transaction.Transaction{
		Nonce:     nonce,
		Value:     stakeValue,
		SndAddr:   newValidatorOwnerBytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("stake@%s%s", numOfNodesHex, validatorData)),
		GasLimit:  500_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txFromNetwork, err := cm.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txFromNetwork)
}

func unstakeTokens(t *testing.T, cm chainSimulatorIntegrationTests.ChainSimulator, nonce uint64, unstakeValue *big.Int, senderBytes []byte, rcvAddrBytes []byte) {
	tx := &transaction.Transaction{
		Nonce:     nonce,
		Value:     big.NewInt(0), // Unstake transactions typically does not carry a value
		SndAddr:   senderBytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unstakeValue.Bytes()))),
		GasLimit:  50_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	_, err := cm.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
}

func checkTotalQualified(t *testing.T, auctionList []*common.AuctionListValidatorAPIResponse, expected int) {
	totalQualified := 0
	for _, res := range auctionList {
		for _, node := range res.Nodes {
			if node.Qualified {
				totalQualified++
			}
		}
	}
	require.Equal(t, expected, totalQualified)
}

func checkValidatorsRating(t *testing.T, validatorStatistics map[string]*validator.ValidatorStatistics) {
	countRatingIncreased := 0
	for _, validatorInfo := range validatorStatistics {
		validatorSignedAtLeastOneBlock := validatorInfo.NumValidatorSuccess > 0 || validatorInfo.NumLeaderSuccess > 0
		if !validatorSignedAtLeastOneBlock {
			continue
		}
		countRatingIncreased++
		require.Greater(t, validatorInfo.TempRating, validatorInfo.Rating)
	}
	require.Greater(t, countRatingIncreased, 0)
}
