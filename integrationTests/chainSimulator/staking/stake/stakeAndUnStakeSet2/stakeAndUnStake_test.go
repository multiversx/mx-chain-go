package stakeAndUnStakeSet2

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking/stake"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/require"
)

// Test description:
// Unstake funds with deactivation of node if below 2500 -> the rest of funds are distributed as topup at epoch change
//
// Internal test scenario #26
func TestChainSimulator_DirectStakingNodes_UnstakeFundsWithDeactivation(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// Test Steps
	//  1. Check the stake amount and number of nodes for the owner of the staked nodes with the vmquery "getTotalStaked", and the account current EGLD balance
	//  2. Create from the owner of staked nodes a transaction to unstake 10 EGLD and send it to the network
	//  3. Check the outcome of the TX & verify new stake state with vmquery "getTotalStaked" and "getUnStakedTokensList"
	//  4. Wait for change of epoch and check the outcome

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:         true,
			TempDir:                        t.TempDir(),
			PathToInitialConfig:            stake.DefaultPathToInitialConfig,
			NumOfShards:                    3,
			RoundDurationInMillis:          stake.RoundDurationInMillis,
			SupernovaRoundDurationInMillis: stake.SupernovaRoundDurationInMillis,
			RoundsPerEpoch:                 stake.RoundsPerEpoch,
			SupernovaRoundsPerEpoch:        stake.SupernovaRoundsPerEpoch,
			ApiInterface:                   api.NewNoApiInterface(),
			MinNodesPerShard:               3,
			MetaChainMinNodes:              3,
			NumNodesWaitingListMeta:        3,
			NumNodesWaitingListShard:       3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 101
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 102

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 102
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivation(t, cs, 1)
	})

	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:         true,
			TempDir:                        t.TempDir(),
			PathToInitialConfig:            stake.DefaultPathToInitialConfig,
			NumOfShards:                    3,
			RoundDurationInMillis:          stake.RoundDurationInMillis,
			SupernovaRoundDurationInMillis: stake.SupernovaRoundDurationInMillis,
			RoundsPerEpoch:                 stake.RoundsPerEpoch,
			SupernovaRoundsPerEpoch:        stake.SupernovaRoundsPerEpoch,
			ApiInterface:                   api.NewNoApiInterface(),
			MinNodesPerShard:               3,
			MetaChainMinNodes:              3,
			NumNodesWaitingListMeta:        3,
			NumNodesWaitingListShard:       3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivation(t, cs, 2)
	})

	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:         true,
			TempDir:                        t.TempDir(),
			PathToInitialConfig:            stake.DefaultPathToInitialConfig,
			NumOfShards:                    3,
			RoundDurationInMillis:          stake.RoundDurationInMillis,
			SupernovaRoundDurationInMillis: stake.SupernovaRoundDurationInMillis,
			RoundsPerEpoch:                 stake.RoundsPerEpoch,
			SupernovaRoundsPerEpoch:        stake.SupernovaRoundsPerEpoch,
			ApiInterface:                   api.NewNoApiInterface(),
			MinNodesPerShard:               3,
			MetaChainMinNodes:              3,
			NumNodesWaitingListMeta:        3,
			NumNodesWaitingListShard:       3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivation(t, cs, 3)
	})

	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:         true,
			TempDir:                        t.TempDir(),
			PathToInitialConfig:            stake.DefaultPathToInitialConfig,
			NumOfShards:                    3,
			RoundDurationInMillis:          stake.RoundDurationInMillis,
			SupernovaRoundDurationInMillis: stake.SupernovaRoundDurationInMillis,
			RoundsPerEpoch:                 stake.RoundsPerEpoch,
			ApiInterface:                   api.NewNoApiInterface(),
			MinNodesPerShard:               3,
			MetaChainMinNodes:              3,
			NumNodesWaitingListMeta:        3,
			NumNodesWaitingListShard:       3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivation(t, cs, 4)
	})
}

func testChainSimulatorDirectStakedUnstakeFundsWithDeactivation(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	privateKeys, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(2)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKeys)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	mintValue := big.NewInt(5010)
	mintValue = mintValue.Mul(chainSimulatorIntegrationTests.OneEGLD, mintValue)

	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	stakeValue := big.NewInt(0).Set(chainSimulatorIntegrationTests.MinimumStakeValue)
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	stake.TestBLSKeyStaked(t, metachainNode, blsKeys[0])

	stakeValue = big.NewInt(0).Set(chainSimulatorIntegrationTests.MinimumStakeValue)
	txDataField = fmt.Sprintf("stake@01@%s@%s", blsKeys[1], staking.MockBLSSignature)
	txStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 1, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	stake.TestBLSKeyStaked(t, metachainNode, blsKeys[1])

	stake.Log.Info("Step 1. Check the stake amount for the owner of the staked nodes")
	stake.CheckExpectedStakedValue(t, metachainNode, validatorOwner.Bytes, 5000)

	stake.Log.Info("Step 2. Create from the owner of staked nodes a transaction to unstake 10 EGLD and send it to the network")

	unStakeValue := big.NewInt(10)
	unStakeValue = unStakeValue.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue.Bytes()))
	txUnStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 2, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	stake.Log.Info("Step 3. Check the outcome of the TX & verify new stake state with vmquery getTotalStaked and getUnStakedTokensList")
	stake.CheckExpectedStakedValue(t, metachainNode, validatorOwner.Bytes, 4990)

	unStakedTokensAmount := stake.GetUnStakedTokensList(t, metachainNode, validatorOwner.Bytes)

	expectedUnStaked := big.NewInt(10)
	expectedUnStaked = expectedUnStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedUnStaked)
	require.Equal(t, expectedUnStaked.String(), big.NewInt(0).SetBytes(unStakedTokensAmount).String())

	stake.Log.Info("Step 4. Wait for change of epoch and check the outcome")
	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch + 1)
	require.Nil(t, err)

	stake.CheckOneOfTheNodesIsUnstaked(t, metachainNode, blsKeys[:2])
}

// Test description:
// Unstake funds with deactivation of node, followed by stake with sufficient ammount does not unstake node at end of epoch
//
// Internal test scenario #27
func TestChainSimulator_DirectStakingNodes_UnstakeFundsWithDeactivation_WithReactivation(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// Test Steps
	// 1. Check the stake amount and number of nodes for the owner of the staked nodes with the vmquery "getTotalStaked", and the account current EGLD balance
	// 2. Create from the owner of staked nodes a transaction to unstake 10 EGLD and send it to the network
	// 3. Check the outcome of the TX & verify new stake state with vmquery
	// 4. Create from the owner of staked nodes a transaction to stake 10 EGLD and send it to the network
	// 5. Check the outcome of the TX & verify new stake state with vmquery
	// 6. Wait for change of epoch and check the outcome

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:         true,
			TempDir:                        t.TempDir(),
			PathToInitialConfig:            stake.DefaultPathToInitialConfig,
			NumOfShards:                    3,
			RoundDurationInMillis:          stake.RoundDurationInMillis,
			SupernovaRoundDurationInMillis: stake.SupernovaRoundDurationInMillis,
			RoundsPerEpoch:                 stake.RoundsPerEpoch,
			SupernovaRoundsPerEpoch:        stake.SupernovaRoundsPerEpoch,
			ApiInterface:                   api.NewNoApiInterface(),
			MinNodesPerShard:               3,
			MetaChainMinNodes:              3,
			NumNodesWaitingListMeta:        3,
			NumNodesWaitingListShard:       3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 101
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 102

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 102
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivationAndReactivation(t, cs, 1)
	})

	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:         true,
			TempDir:                        t.TempDir(),
			PathToInitialConfig:            stake.DefaultPathToInitialConfig,
			NumOfShards:                    3,
			RoundDurationInMillis:          stake.RoundDurationInMillis,
			SupernovaRoundDurationInMillis: stake.SupernovaRoundDurationInMillis,
			RoundsPerEpoch:                 stake.RoundsPerEpoch,
			SupernovaRoundsPerEpoch:        stake.SupernovaRoundsPerEpoch,
			ApiInterface:                   api.NewNoApiInterface(),
			MinNodesPerShard:               3,
			MetaChainMinNodes:              3,
			NumNodesWaitingListMeta:        3,
			NumNodesWaitingListShard:       3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivationAndReactivation(t, cs, 2)
	})

	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:         true,
			TempDir:                        t.TempDir(),
			PathToInitialConfig:            stake.DefaultPathToInitialConfig,
			NumOfShards:                    3,
			RoundDurationInMillis:          stake.RoundDurationInMillis,
			SupernovaRoundDurationInMillis: stake.SupernovaRoundDurationInMillis,
			RoundsPerEpoch:                 stake.RoundsPerEpoch,
			SupernovaRoundsPerEpoch:        stake.SupernovaRoundsPerEpoch,
			ApiInterface:                   api.NewNoApiInterface(),
			MinNodesPerShard:               3,
			MetaChainMinNodes:              3,
			NumNodesWaitingListMeta:        3,
			NumNodesWaitingListShard:       3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivationAndReactivation(t, cs, 3)
	})

	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:         true,
			TempDir:                        t.TempDir(),
			PathToInitialConfig:            stake.DefaultPathToInitialConfig,
			NumOfShards:                    3,
			RoundDurationInMillis:          stake.RoundDurationInMillis,
			SupernovaRoundDurationInMillis: stake.SupernovaRoundDurationInMillis,
			RoundsPerEpoch:                 stake.RoundsPerEpoch,
			SupernovaRoundsPerEpoch:        stake.SupernovaRoundsPerEpoch,
			ApiInterface:                   api.NewNoApiInterface(),
			MinNodesPerShard:               3,
			MetaChainMinNodes:              3,
			NumNodesWaitingListMeta:        3,
			NumNodesWaitingListShard:       3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivationAndReactivation(t, cs, 4)
	})
}

func testChainSimulatorDirectStakedUnstakeFundsWithDeactivationAndReactivation(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	privateKeys, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(2)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKeys)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	mintValue := big.NewInt(6000)
	mintValue = mintValue.Mul(chainSimulatorIntegrationTests.OneEGLD, mintValue)

	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	stakeValue := big.NewInt(0).Set(chainSimulatorIntegrationTests.MinimumStakeValue)
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	stake.TestBLSKeyStaked(t, metachainNode, blsKeys[0])

	stakeValue = big.NewInt(0).Set(chainSimulatorIntegrationTests.MinimumStakeValue)
	txDataField = fmt.Sprintf("stake@01@%s@%s", blsKeys[1], staking.MockBLSSignature)
	txStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 1, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	stake.TestBLSKeyStaked(t, metachainNode, blsKeys[1])

	stake.Log.Info("Step 1. Check the stake amount for the owner of the staked nodes")
	stake.CheckExpectedStakedValue(t, metachainNode, validatorOwner.Bytes, 5000)

	stake.Log.Info("Step 2. Create from the owner of staked nodes a transaction to unstake 10 EGLD and send it to the network")

	unStakeValue := big.NewInt(10)
	unStakeValue = unStakeValue.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue.Bytes()))
	txUnStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 2, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	stake.Log.Info("Step 3. Check the outcome of the TX & verify new stake state with vmquery getTotalStaked and getUnStakedTokensList")
	stake.CheckExpectedStakedValue(t, metachainNode, validatorOwner.Bytes, 4990)

	unStakedTokensAmount := stake.GetUnStakedTokensList(t, metachainNode, validatorOwner.Bytes)

	expectedUnStaked := big.NewInt(10)
	expectedUnStaked = expectedUnStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedUnStaked)
	require.Equal(t, expectedUnStaked.String(), big.NewInt(0).SetBytes(unStakedTokensAmount).String())

	stake.Log.Info("Step 4. Create from the owner of staked nodes a transaction to stake 10 EGLD and send it to the network")

	newStakeValue := big.NewInt(10)
	newStakeValue = newStakeValue.Mul(chainSimulatorIntegrationTests.OneEGLD, newStakeValue)
	txDataField = fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 3, vm.ValidatorSCAddress, newStakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	stake.Log.Info("5. Check the outcome of the TX & verify new stake state with vmquery")
	stake.CheckExpectedStakedValue(t, metachainNode, validatorOwner.Bytes, 5000)

	stake.Log.Info("Step 6. Wait for change of epoch and check the outcome")
	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch + 1)
	require.Nil(t, err)

	stake.TestBLSKeyStaked(t, metachainNode, blsKeys[0])
	stake.TestBLSKeyStaked(t, metachainNode, blsKeys[1])
}

// Test description:
// Withdraw unstaked funds before unbonding period should return error
//
// Internal test scenario #28
func TestChainSimulator_DirectStakingNodes_WithdrawUnstakedFundsBeforeUnbonding(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// Test Steps
	// 1. Create from the owner of staked nodes a transaction to withdraw the unstaked funds
	// 2. Check the outcome of the TX & verify new stake state with vmquery ("getUnStakedTokensList")

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:         true,
			TempDir:                        t.TempDir(),
			PathToInitialConfig:            stake.DefaultPathToInitialConfig,
			NumOfShards:                    3,
			RoundDurationInMillis:          stake.RoundDurationInMillis,
			SupernovaRoundDurationInMillis: stake.SupernovaRoundDurationInMillis,
			RoundsPerEpoch:                 stake.RoundsPerEpoch,
			SupernovaRoundsPerEpoch:        stake.SupernovaRoundsPerEpoch,
			ApiInterface:                   api.NewNoApiInterface(),
			MinNodesPerShard:               3,
			MetaChainMinNodes:              3,
			NumNodesWaitingListMeta:        3,
			NumNodesWaitingListShard:       3,
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

		testChainSimulatorDirectStakedWithdrawUnstakedFundsBeforeUnbonding(t, cs, 1)
	})

	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:         true,
			TempDir:                        t.TempDir(),
			PathToInitialConfig:            stake.DefaultPathToInitialConfig,
			NumOfShards:                    3,
			RoundDurationInMillis:          stake.RoundDurationInMillis,
			SupernovaRoundDurationInMillis: stake.SupernovaRoundDurationInMillis,
			RoundsPerEpoch:                 stake.RoundsPerEpoch,
			SupernovaRoundsPerEpoch:        stake.SupernovaRoundsPerEpoch,
			ApiInterface:                   api.NewNoApiInterface(),
			MinNodesPerShard:               3,
			MetaChainMinNodes:              3,
			NumNodesWaitingListMeta:        3,
			NumNodesWaitingListShard:       3,
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

		testChainSimulatorDirectStakedWithdrawUnstakedFundsBeforeUnbonding(t, cs, 2)
	})

	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:         true,
			TempDir:                        t.TempDir(),
			PathToInitialConfig:            stake.DefaultPathToInitialConfig,
			NumOfShards:                    3,
			RoundDurationInMillis:          stake.RoundDurationInMillis,
			SupernovaRoundDurationInMillis: stake.SupernovaRoundDurationInMillis,
			RoundsPerEpoch:                 stake.RoundsPerEpoch,
			SupernovaRoundsPerEpoch:        stake.SupernovaRoundsPerEpoch,
			ApiInterface:                   api.NewNoApiInterface(),
			MinNodesPerShard:               3,
			MetaChainMinNodes:              3,
			NumNodesWaitingListMeta:        3,
			NumNodesWaitingListShard:       3,
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

		testChainSimulatorDirectStakedWithdrawUnstakedFundsBeforeUnbonding(t, cs, 3)
	})

	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:         true,
			TempDir:                        t.TempDir(),
			PathToInitialConfig:            stake.DefaultPathToInitialConfig,
			NumOfShards:                    3,
			RoundDurationInMillis:          stake.RoundDurationInMillis,
			SupernovaRoundDurationInMillis: stake.SupernovaRoundDurationInMillis,
			RoundsPerEpoch:                 stake.RoundsPerEpoch,
			SupernovaRoundsPerEpoch:        stake.SupernovaRoundsPerEpoch,
			ApiInterface:                   api.NewNoApiInterface(),
			MinNodesPerShard:               3,
			MetaChainMinNodes:              3,
			NumNodesWaitingListMeta:        3,
			NumNodesWaitingListShard:       3,
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

		testChainSimulatorDirectStakedWithdrawUnstakedFundsBeforeUnbonding(t, cs, 4)
	})
}

func testChainSimulatorDirectStakedWithdrawUnstakedFundsBeforeUnbonding(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	mintValue := big.NewInt(10000)
	mintValue = mintValue.Mul(chainSimulatorIntegrationTests.OneEGLD, mintValue)

	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	stakeValue := big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(2600))
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	stake.TestBLSKeyStaked(t, metachainNode, blsKeys[0])

	shardIDValidatorOwner := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(validatorOwner.Bytes)
	accountValidatorOwner, _, err := cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceBeforeUnbonding, _ := big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	stake.Log.Info("Step 1. Create from the owner of staked nodes a transaction to withdraw the unstaked funds")

	unStakeValue := big.NewInt(10)
	unStakeValue = unStakeValue.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue.Bytes()))
	txUnStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 1, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	// check bls key is still staked
	stake.TestBLSKeyStaked(t, metachainNode, blsKeys[0])

	txDataField = fmt.Sprintf("unBondTokens@%s", blsKeys[0])
	txUnBond := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 2, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForUnBond)
	unBondTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnBond, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unBondTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	stake.Log.Info("Step 2. Check the outcome of the TX & verify new stake state with vmquery (`getUnStakedTokensList`)")

	scQuery := &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getUnStakedTokensList",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{validatorOwner.Bytes},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, result.ReturnCode)

	expectedUnStaked := big.NewInt(10)
	expectedUnStaked = expectedUnStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedUnStaked)
	require.Equal(t, expectedUnStaked.String(), big.NewInt(0).SetBytes(result.ReturnData[0]).String())

	// the owner balance should decrease only with the txs fee
	accountValidatorOwner, _, err = cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceAfterUnbonding, _ := big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	txsFee, _ := big.NewInt(0).SetString(unBondTx.Fee, 10)
	balanceAfterUnbondingWithFee := big.NewInt(0).Add(balanceAfterUnbonding, txsFee)

	txsFee, _ = big.NewInt(0).SetString(unStakeTx.Fee, 10)
	balanceAfterUnbondingWithFee.Add(balanceAfterUnbondingWithFee, txsFee)

	txsFee, _ = big.NewInt(0).SetString(stakeTx.Fee, 10)
	balanceAfterUnbondingWithFee.Add(balanceAfterUnbondingWithFee, txsFee)

	require.Equal(t, 1, balanceAfterUnbondingWithFee.Cmp(balanceBeforeUnbonding))
}
