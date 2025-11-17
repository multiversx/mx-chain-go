package stakeAndUnStakeSet4

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
// Unstaking funds in different batches allows correct withdrawal for each batch
// at the corresponding epoch.
//
// Internal test scenario #30
func TestChainSimulator_DirectStakingNodes_WithdrawUnstakedInBatches(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// Test Steps
	// 1. Create 3 transactions for unstaking: first one unstaking 11 egld each, second one unstaking 12 egld and third one unstaking 13 egld.
	// 2. Send the transactions in consecutive epochs, one TX in each epoch.
	// 3. Wait for the epoch when first tx unbonding period ends.
	// 4. Create a transaction for withdraw and send it to the network
	// 5. Wait for an epoch
	// 6. Create another transaction for withdraw and send it to the network
	// 7. Wait for an epoch
	// 8. Create another transasction for withdraw and send it to the network

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

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 6
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInBatches(t, cs, 1)
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

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 6
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInBatches(t, cs, 2)
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

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 6
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInBatches(t, cs, 3)
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

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 6
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInBatches(t, cs, 4)
	})
}

func testChainSimulatorDirectStakedWithdrawUnstakedFundsInBatches(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
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

	stake.Log.Info("Step 1. Create 3 transactions for unstaking: first one unstaking 11 egld, second one unstaking 12 egld and third one unstaking 13 egld.")
	stake.Log.Info("Step 2. Send the transactions in consecutive epochs, one TX in each epoch.")

	unStakeValue1 := big.NewInt(11)
	unStakeValue1 = unStakeValue1.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue1)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue1.Bytes()))
	txUnStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 1, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	unStakeTxFee1, _ := big.NewInt(0).SetString(unStakeTx.Fee, 10)

	testEpoch := targetEpoch + 1
	err = cs.GenerateBlocksUntilEpochIsReached(testEpoch)
	require.Nil(t, err)

	unStakeValue2 := big.NewInt(12)
	unStakeValue2 = unStakeValue2.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue2)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue2.Bytes()))
	txUnStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 2, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	unStakeTxFee2, _ := big.NewInt(0).SetString(unStakeTx.Fee, 10)

	testEpoch++
	err = cs.GenerateBlocksUntilEpochIsReached(testEpoch)
	require.Nil(t, err)

	unStakeValue3 := big.NewInt(13)
	unStakeValue3 = unStakeValue3.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue3)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue3.Bytes()))
	txUnStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 3, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	unStakeTxFee3, _ := big.NewInt(0).SetString(unStakeTx.Fee, 10)

	testEpoch++
	err = cs.GenerateBlocksUntilEpochIsReached(testEpoch)
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

	expectedUnStaked := big.NewInt(11)
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

	testEpoch += 3
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

	// the owner balance should increase with the (11 EGLD - tx fee)
	accountValidatorOwner, _, err = cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceAfterUnbonding, _ := big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue1)

	txsFee := big.NewInt(0)

	txsFee.Add(txsFee, stakeTxFee)
	txsFee.Add(txsFee, unBondTxFee)
	txsFee.Add(txsFee, unStakeTxFee1)
	txsFee.Add(txsFee, unStakeTxFee2)
	txsFee.Add(txsFee, unStakeTxFee3)

	balanceAfterUnbonding.Add(balanceAfterUnbonding, txsFee)

	require.Equal(t, 1, balanceAfterUnbonding.Cmp(balanceBeforeUnbonding))

	stake.Log.Info("Step 4.2. Create from the owner of staked nodes a transaction to withdraw the unstaked funds")

	testEpoch++
	err = cs.GenerateBlocksUntilEpochIsReached(testEpoch)
	require.Nil(t, err)

	txDataField = fmt.Sprintf("unBondTokens@%s", blsKeys[0])
	txUnBond = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 5, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForUnBond)
	unBondTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnBond, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unBondTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	// the owner balance should increase with the (11+12 EGLD - tx fee)
	accountValidatorOwner, _, err = cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceAfterUnbonding, _ = big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue1)
	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue2)

	txsFee.Add(txsFee, unBondTxFee)
	balanceAfterUnbonding.Add(balanceAfterUnbonding, txsFee)

	require.Equal(t, 1, balanceAfterUnbonding.Cmp(balanceBeforeUnbonding))

	stake.Log.Info("Step 4.3. Create from the owner of staked nodes a transaction to withdraw the unstaked funds")

	testEpoch++
	err = cs.GenerateBlocksUntilEpochIsReached(testEpoch)
	require.Nil(t, err)

	txDataField = fmt.Sprintf("unBondTokens@%s", blsKeys[0])
	txUnBond = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 6, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForUnBond)
	unBondTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnBond, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unBondTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	// the owner balance should increase with the (11+12+13 EGLD - tx fee)
	accountValidatorOwner, _, err = cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceAfterUnbonding, _ = big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue1)
	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue2)
	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue3)

	txsFee.Add(txsFee, unBondTxFee)
	balanceAfterUnbonding.Add(balanceAfterUnbonding, txsFee)

	require.Equal(t, 1, balanceAfterUnbonding.Cmp(balanceBeforeUnbonding))
}
