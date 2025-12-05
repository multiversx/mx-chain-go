package stakeAndUnStakeSet3

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking/stake"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
)

// Test description:
// Withdraw unstaked funds in first available withdraw epoch
//
// Internal test scenario #29
func TestChainSimulator_DirectStakingNodes_WithdrawUnstakedInWithdrawEpoch(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// Test Steps
	// 1. Wait for the unbonding epoch to start
	// 2. Create from the owner of staked nodes a transaction to withdraw the unstaked funds
	// 3. Check the outcome of the TX & verify new stake state with vmquery ("getUnStakedTokensList")

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
				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInFirstEpoch(t, cs, 1)
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
				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInFirstEpoch(t, cs, 2)
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
				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 1
				integrationTests.DeactivateSupernovaInConfig(cfg)
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInFirstEpoch(t, cs, 3)
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
				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 1
				integrationTests.DeactivateSupernovaInConfig(cfg)
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInFirstEpoch(t, cs, 4)
	})
}

func testChainSimulatorDirectStakedWithdrawUnstakedFundsInFirstEpoch(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
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

	stake.Log.Info("Step 1. Wait for the unbonding epoch to start")

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch + 1)
	require.Nil(t, err)

	stake.Log.Info("Step 2. Create from the owner of staked nodes a transaction to withdraw the unstaked funds")

	txDataField = fmt.Sprintf("unBondTokens@%s", blsKeys[0])
	txUnBond := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 2, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForUnBond)
	unBondTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnBond, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unBondTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	stake.Log.Info("Step 3. Check the outcome of the TX & verify new stake state with vmquery (`getUnStakedTokensList`)")

	scQuery = &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getTotalStaked",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{validatorOwner.Bytes},
	}
	result, _, err = metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, result.ReturnCode)

	expectedStaked := big.NewInt(2590)
	expectedStaked = expectedStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedStaked)
	require.Equal(t, expectedStaked.String(), string(result.ReturnData[0]))

	// the owner balance should increase with the (10 EGLD - tx fee)
	accountValidatorOwner, _, err = cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceAfterUnbonding, _ := big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	// substract unbonding value
	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue)

	txsFee, _ := big.NewInt(0).SetString(unBondTx.Fee, 10)
	balanceAfterUnbondingWithFee := big.NewInt(0).Add(balanceAfterUnbonding, txsFee)

	txsFee, _ = big.NewInt(0).SetString(unStakeTx.Fee, 10)
	balanceAfterUnbondingWithFee.Add(balanceAfterUnbondingWithFee, txsFee)

	txsFee, _ = big.NewInt(0).SetString(stakeTx.Fee, 10)
	balanceAfterUnbondingWithFee.Add(balanceAfterUnbondingWithFee, txsFee)

	require.Equal(t, 1, balanceAfterUnbondingWithFee.Cmp(balanceBeforeUnbonding))
}

// Test description:
// Unstake funds in different batches in the same epoch allows correct withdrawal in the correct epoch
//
// Internal test scenario #31
func TestChainSimulator_DirectStakingNodes_WithdrawUnstakedInEpoch(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// Test Steps
	// 1. Create 3 transactions for unstaking: first one unstaking 11 egld each, second one unstaking 12 egld and third one unstaking 13 egld.
	// 2. Send the transactions consecutively in the same epoch
	// 3. Wait for the epoch when unbonding period ends.
	// 4. Create a transaction for withdraw and send it to the network

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

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 3
				integrationTests.DeactivateSupernovaInConfig(cfg)
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInEpoch(t, cs, 1)
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

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 3
				integrationTests.DeactivateSupernovaInConfig(cfg)
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInEpoch(t, cs, 2)
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

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 3
				integrationTests.DeactivateSupernovaInConfig(cfg)
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInEpoch(t, cs, 3)
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

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 3
				integrationTests.DeactivateSupernovaInConfig(cfg)
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInEpoch(t, cs, 4)
	})
}

func testChainSimulatorDirectStakedWithdrawUnstakedFundsInEpoch(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	mintValue := big.NewInt(2700)
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

	stakeTxFee, _ := big.NewInt(0).SetString(stakeTx.Fee, 10)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	stake.TestBLSKeyStaked(t, metachainNode, blsKeys[0])

	shardIDValidatorOwner := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(validatorOwner.Bytes)
	accountValidatorOwner, _, err := cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceBeforeUnbonding, _ := big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	stake.Log.Info("Step 1. Create 3 transactions for unstaking: first one unstaking 11 egld each, second one unstaking 12 egld and third one unstaking 13 egld.")
	stake.Log.Info("Step 2. Send the transactions in consecutively in same epoch.")

	unStakeValue1 := big.NewInt(11)
	unStakeValue1 = unStakeValue1.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue1)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue1.Bytes()))
	txUnStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 1, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	unStakeTxFee, _ := big.NewInt(0).SetString(unStakeTx.Fee, 10)

	unStakeValue2 := big.NewInt(12)
	unStakeValue2 = unStakeValue2.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue2)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue2.Bytes()))
	txUnStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 2, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	unStakeValue3 := big.NewInt(13)
	unStakeValue3 = unStakeValue3.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue3)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue3.Bytes()))
	txUnStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 3, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	// check bls key is still staked
	stake.TestBLSKeyStaked(t, metachainNode, blsKeys[0])

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

	expectedUnStaked := big.NewInt(11 + 12 + 13)
	expectedUnStaked = expectedUnStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedUnStaked)
	require.Equal(t, expectedUnStaked.String(), big.NewInt(0).SetBytes(result.ReturnData[0]).String())

	scQuery = &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getTotalStaked",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{validatorOwner.Bytes},
	}
	result, _, err = metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, result.ReturnCode)

	expectedStaked := big.NewInt(2600 - 11 - 12 - 13)
	expectedStaked = expectedStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedStaked)
	require.Equal(t, expectedStaked.String(), string(result.ReturnData[0]))

	stake.Log.Info("Step 3. Wait for the unbonding epoch to start")

	testEpoch := targetEpoch + 3
	err = cs.GenerateBlocksUntilEpochIsReached(testEpoch)
	require.Nil(t, err)

	stake.Log.Info("Step 4.1. Create from the owner of staked nodes a transaction to withdraw the unstaked funds")

	txDataField = fmt.Sprintf("unBondTokens@%s", blsKeys[0])
	txUnBond := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 4, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForUnBond)
	unBondTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnBond, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unBondTx)

	unBondTxFee, _ := big.NewInt(0).SetString(unBondTx.Fee, 10)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	// the owner balance should increase with the (11+12+13 EGLD - tx fee)
	accountValidatorOwner, _, err = cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceAfterUnbonding, _ := big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue1)
	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue2)
	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue3)

	txsFee := big.NewInt(0)

	txsFee.Add(txsFee, stakeTxFee)
	txsFee.Add(txsFee, unBondTxFee)
	txsFee.Add(txsFee, unStakeTxFee)
	txsFee.Add(txsFee, unStakeTxFee)
	txsFee.Add(txsFee, unStakeTxFee)

	balanceAfterUnbonding.Add(balanceAfterUnbonding, txsFee)

	require.Equal(t, 1, balanceAfterUnbonding.Cmp(balanceBeforeUnbonding))
}

// Test that if we unStake one active node(waiting/eligible), the number of qualified nodes will remain the same
// Nodes configuration at genesis consisting of a total of 32 nodes, distributed on 3 shards + meta:
// - 4 eligible nodes/shard
// - 4 waiting nodes/shard
// - 2 nodes to shuffle per shard
// - max num nodes config for stakingV4 step3 = 24 (being downsized from previously 32 nodes)
// - with this config, we should always select 8 nodes from auction list
// We will add one extra node, so auction list size = 9, but will always select 8. Even if we unStake one active node,
// we should still only select 8 nodes.
func TestChainSimulator_UnStakeOneActiveNodeAndCheckAPIAuctionList(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	stakingV4Step1Epoch := uint32(2)
	stakingV4Step2Epoch := uint32(3)
	stakingV4Step3Epoch := uint32(4)

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            stake.DefaultPathToInitialConfig,
		NumOfShards:                    numOfShards,
		RoundDurationInMillis:          stake.RoundDurationInMillis,
		SupernovaRoundDurationInMillis: stake.SupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 stake.RoundsPerEpoch,
		SupernovaRoundsPerEpoch:        stake.SupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               4,
		MetaChainMinNodes:              4,
		NumNodesWaitingListMeta:        4,
		NumNodesWaitingListShard:       4,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = stakingV4Step1Epoch
			cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = stakingV4Step2Epoch
			cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = stakingV4Step3Epoch
			cfg.EpochConfig.EnableEpochs.CleanupAuctionOnLowWaitingListEnableEpoch = stakingV4Step1Epoch

			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[1].MaxNumNodes = 32
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[1].NodesToShufflePerShard = 2

			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = stakingV4Step3Epoch
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].MaxNumNodes = 24
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].NodesToShufflePerShard = 2
			integrationTests.DeactivateSupernovaInConfig(cfg)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(int32(stakingV4Step3Epoch + 1))
	require.Nil(t, err)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	qualified, unQualified := stake.GetQualifiedAndUnqualifiedNodes(t, metachainNode)
	require.Equal(t, 8, len(qualified))
	require.Equal(t, 0, len(unQualified))

	stakeOneNode(t, cs)

	qualified, unQualified = stake.GetQualifiedAndUnqualifiedNodes(t, metachainNode)
	require.Equal(t, 8, len(qualified))
	require.Equal(t, 1, len(unQualified))

	unStakeOneActiveNode(t, cs)

	qualified, unQualified = stake.GetQualifiedAndUnqualifiedNodes(t, metachainNode)
	require.Equal(t, 8, len(qualified))
	require.Equal(t, 1, len(unQualified))
}

// Nodes configuration at genesis consisting of a total of 40 nodes, distributed on 3 shards + meta:
// - 4 eligible nodes/shard
// - 4 waiting nodes/shard
// - 2 nodes to shuffle per shard
// - max num nodes config for stakingV4 step3 = 32 (being downsized from previously 40 nodes)
// - with this config, we should always select max 8 nodes from auction list if there are > 40 nodes in the network
// This test will run with only 32 nodes and check that there are no nodes in the auction list,
// because of the lowWaitingList condition being triggered when in staking v4
func TestChainSimulator_EdgeCaseLowWaitingList(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	stakingV4Step1Epoch := uint32(2)
	stakingV4Step2Epoch := uint32(3)
	stakingV4Step3Epoch := uint32(4)

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            stake.DefaultPathToInitialConfig,
		NumOfShards:                    numOfShards,
		RoundDurationInMillis:          stake.RoundDurationInMillis,
		SupernovaRoundDurationInMillis: stake.SupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 stake.RoundsPerEpoch,
		SupernovaRoundsPerEpoch:        stake.SupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               4,
		MetaChainMinNodes:              4,
		NumNodesWaitingListMeta:        2,
		NumNodesWaitingListShard:       2,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = stakingV4Step1Epoch
			cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = stakingV4Step2Epoch
			cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = stakingV4Step3Epoch

			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[1].MaxNumNodes = 40
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[1].NodesToShufflePerShard = 2

			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = stakingV4Step3Epoch
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].MaxNumNodes = 32
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].NodesToShufflePerShard = 2
			integrationTests.DeactivateSupernovaInConfig(cfg)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	epochToCheck := int32(stakingV4Step3Epoch + 1)
	err = cs.GenerateBlocksUntilEpochIsReached(epochToCheck)
	require.Nil(t, err)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	qualified, unQualified := stake.GetQualifiedAndUnqualifiedNodes(t, metachainNode)
	require.Equal(t, 0, len(qualified))
	require.Equal(t, 0, len(unQualified))

	// we always have 0 in auction list because of the lowWaitingList condition
	epochToCheck += 1
	err = cs.GenerateBlocksUntilEpochIsReached(epochToCheck)
	require.Nil(t, err)

	qualified, unQualified = stake.GetQualifiedAndUnqualifiedNodes(t, metachainNode)
	require.Equal(t, 0, len(qualified))
	require.Equal(t, 0, len(unQualified))

	// stake 16 mode nodes, these will go to auction list
	stakeNodes(t, cs, 17)

	epochToCheck += 1
	err = cs.GenerateBlocksUntilEpochIsReached(epochToCheck)
	require.Nil(t, err)

	qualified, unQualified = stake.GetQualifiedAndUnqualifiedNodes(t, metachainNode)
	// all the previously registered will be selected, as we have 24 nodes in eligible+waiting, 8 will shuffle out,
	// but this time there will be not be lowWaitingList, as there are enough in auction, so we will end up with
	// 24-8 = 16 nodes remaining + 16 from auction, to fill up all 32 positions
	require.Equal(t, 16, len(qualified))
	require.Equal(t, 1, len(unQualified))

	shuffledOutNodesKeys, err := metachainNode.GetProcessComponents().NodesCoordinator().GetShuffledOutToAuctionValidatorsPublicKeys(uint32(epochToCheck))
	require.Nil(t, err)

	checkKeysNotInMap(t, shuffledOutNodesKeys, qualified)
	checkKeysNotInMap(t, shuffledOutNodesKeys, unQualified)
}

func checkKeysNotInMap(t *testing.T, m map[uint32][][]byte, keys []string) {
	for _, key := range keys {
		for _, v := range m {
			for _, k := range v {
				mapKey := hex.EncodeToString(k)
				require.NotEqual(t, key, mapKey)
			}
		}
	}
}

func stakeNodes(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, numNodesToStake int) {
	txs := make([]*transaction.Transaction, numNodesToStake)
	for i := 0; i < numNodesToStake; i++ {
		txs[i] = createStakeTransaction(t, cs)
	}

	stakeTxs, err := cs.SendTxsAndGenerateBlocksTilAreExecuted(txs, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTxs)
	require.Len(t, stakeTxs, numNodesToStake)

	require.Nil(t, cs.GenerateBlocks(1))
}

func stakeOneNode(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator) {
	txStake := createStakeTransaction(t, cs)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	require.Nil(t, cs.GenerateBlocks(1))
}

func createStakeTransaction(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator) *transaction.Transaction {
	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)

	mintValue := big.NewInt(0).Add(chainSimulatorIntegrationTests.MinimumStakeValue, chainSimulatorIntegrationTests.OneEGLD)
	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	return chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.MinimumStakeValue, txDataField, staking.GasLimitForStakeOperation)
}

func unStakeOneActiveNode(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator) {
	err := cs.ForceResetValidatorStatisticsCache()
	require.Nil(t, err)

	validators, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)

	idx := 0
	keyToUnStake := make([]byte, 0)
	numKeys := len(cs.GetValidatorPrivateKeys())
	for idx = 0; idx < numKeys; idx++ {
		keyToUnStake, err = cs.GetValidatorPrivateKeys()[idx].GeneratePublic().ToByteArray()
		require.Nil(t, err)

		apiValidator, found := validators[hex.EncodeToString(keyToUnStake)]
		require.True(t, found)

		validatorStatus := apiValidator.ValidatorStatus
		if validatorStatus == "waiting" || validatorStatus == "eligible" {
			stake.Log.Info("found active key to unStake", "index", idx, "bls key", keyToUnStake, "list", validatorStatus)
			break
		}

		if idx == numKeys-1 {
			require.Fail(t, "did not find key to unStake")
		}
	}

	rcv := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l"
	rcvAddrBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(rcv)

	validatorWallet := cs.GetInitialWalletKeys().StakeWallets[idx].Address
	shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(validatorWallet.Bytes)
	initialAccount, _, err := cs.GetNodeHandler(shardID).GetFacadeHandler().GetAccount(validatorWallet.Bech32, coreAPI.AccountQueryOptions{})

	require.Nil(t, err)
	tx := &transaction.Transaction{
		Nonce:     initialAccount.Nonce,
		Value:     big.NewInt(0),
		SndAddr:   validatorWallet.Bytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("unStake@%s", hex.EncodeToString(keyToUnStake))),
		GasLimit:  50_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	err = cs.ForceResetValidatorStatisticsCache()
	require.Nil(t, err)
	validators, err = cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)

	apiValidator, found := validators[hex.EncodeToString(keyToUnStake)]
	require.True(t, found)
	require.True(t, strings.Contains(apiValidator.ValidatorStatus, "leaving"))
}
