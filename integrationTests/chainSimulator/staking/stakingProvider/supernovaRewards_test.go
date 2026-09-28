package stakingProvider

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/vm"
)

const (
	rewardsStakingV4Step1Epoch = uint32(1)
	rewardsStakingV4Step3Epoch = rewardsStakingV4Step1Epoch + 2
	rewardsSupernovaEpoch      = uint32(6)
	rewardsFixActivationEpoch  = rewardsSupernovaEpoch + 3
	rewardsEpochLength         = uint64(60)
	numInitialRewardDelegators = 7
)

const (
	untouchedRewardDelegatorIndex = iota
	claimedBeforeSupernovaDelegatorIndex
	claimedAfterOverwriteRewardDelegatorIndex
	claimedBeforeFixRewardDelegatorIndex
	delegatedBeforeFixRewardDelegatorIndex
	undelegatedBeforeFixRewardDelegatorIndex
	redelegatedBeforeFixRewardDelegatorIndex
	newBeforeFixRewardDelegatorIndex = numInitialRewardDelegators
)

func TestChainSimulator_DelegationRewardsAcrossDelayedSupernovaFix(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	context := newChainSimulatorRewardsContext(t)
	defer context.cs.Close()

	context.reachEpoch(t, rewardsSupernovaEpoch-1)
	context.requireRewardRecord(t, rewardsSupernovaEpoch-1)

	context.reachEpoch(t, rewardsSupernovaEpoch)
	context.requireRewardRecord(t, rewardsSupernovaEpoch)
	context.claim(t, claimedBeforeSupernovaDelegatorIndex, true)

	context.reachEpoch(t, rewardsSupernovaEpoch+1)
	context.requireRewardRecord(t, rewardsSupernovaEpoch)
	context.claim(t, claimedAfterOverwriteRewardDelegatorIndex, true)

	context.reachEpoch(t, rewardsSupernovaEpoch+2)
	context.requireRewardRecord(t, rewardsSupernovaEpoch+1)
	context.claim(t, claimedBeforeFixRewardDelegatorIndex, true)

	context.delegate(t, delegatedBeforeFixRewardDelegatorIndex, context.delegationValue)
	context.unDelegate(t, undelegatedBeforeFixRewardDelegatorIndex, context.delegationValue)
	context.reDelegateRewards(t, redelegatedBeforeFixRewardDelegatorIndex)
	context.addDelegator(t)

	context.reachEpoch(t, rewardsFixActivationEpoch)
	context.requireMissingRewardRecord(t, rewardsFixActivationEpoch-1)
	context.requireRewardRecord(t, rewardsFixActivationEpoch)
	context.claimProviderOwner(t, true)
	context.claim(t, untouchedRewardDelegatorIndex, true)
	context.unDelegate(t, claimedAfterOverwriteRewardDelegatorIndex, context.delegationValue)
	context.delegate(t, claimedBeforeSupernovaDelegatorIndex, context.delegationValue)
	context.reDelegateRewards(t, claimedBeforeFixRewardDelegatorIndex)
	partialUnDelegateValue := new(big.Int).Div(new(big.Int).Set(context.delegationValue), big.NewInt(2))
	context.unDelegate(t, delegatedBeforeFixRewardDelegatorIndex, partialUnDelegateValue)
	for _, index := range []int{
		delegatedBeforeFixRewardDelegatorIndex,
		undelegatedBeforeFixRewardDelegatorIndex,
		redelegatedBeforeFixRewardDelegatorIndex,
		newBeforeFixRewardDelegatorIndex,
	} {
		context.claim(t, index, true)
	}

	context.reachEpoch(t, rewardsFixActivationEpoch+1)
	context.requireRewardRecord(t, rewardsFixActivationEpoch+1)
	context.claimProviderOwner(t, true)
	for index := range context.delegators {
		context.claim(t, index, index != undelegatedBeforeFixRewardDelegatorIndex)
	}
	context.claim(t, untouchedRewardDelegatorIndex, false)

	context.reachEpoch(t, rewardsFixActivationEpoch+2)
	context.withdraw(t, undelegatedBeforeFixRewardDelegatorIndex, context.delegationValue)
}

func TestChainSimulator_StakingProviderNodeOperationsAfterDelayedSupernovaFix(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := newStakingV4SupernovaSimulator(t, 4, func(cfg *config.Configs) {
		newNumNodes := cfg.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake + 8
		configs.SetMaxNumberOfNodesInConfigs(cfg, uint32(newNumNodes), 0, 3)
		configs.SetQuickJailRatingConfig(cfg)
		cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriod = 1
		cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodSupernova = 1
	})
	defer cs.Close()

	err := cs.GenerateBlocksUntilEpochIsReached(int32(rewardsStakingV4Step3Epoch))
	require.NoError(t, err)
	delegationSC, providerOwner := convertGenesisValidatorToDelegationProvider(t, cs)
	context := &chainSimulatorRewardsContext{
		cs:            cs,
		delegationSC:  delegationSC,
		providerOwner: providerOwner,
	}

	_, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(2)
	require.NoError(t, err)
	context.addNode(t, blsKeys[0])
	context.sendOperation(t, providerOwner, "delegate", chainSimulatorIntegrationTests.MinimumStakeValue, gasLimitForDelegate)
	context.stakeNode(t, blsKeys[0])

	context.reachEpoch(t, rewardsFixActivationEpoch+2)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	decodedJailedKey, err := hex.DecodeString(blsKeys[0])
	require.NoError(t, err)
	require.Equal(t, "jailed", staking.GetBLSKeyStatus(t, metachainNode, decodedJailedKey))

	unJailValue, ok := new(big.Int).SetString("2500000000000000000", 10)
	require.True(t, ok)
	context.sendOperation(t, providerOwner, "unJailNodes@"+blsKeys[0], unJailValue, staking.GasLimitForStakeOperation)
	err = cs.GenerateBlocks(1)
	require.NoError(t, err)
	require.Equal(t, staking.StakedStatus, staking.GetBLSKeyStatus(t, metachainNode, decodedJailedKey))

	context.addNode(t, blsKeys[1])
	context.sendOperation(t, providerOwner, "delegate", chainSimulatorIntegrationTests.MinimumStakeValue, gasLimitForDelegate)
	context.stakeNode(t, blsKeys[1])
	context.requireProviderNodeStatus(t, blsKeys[1], staking.StakedStatus)

	context.sendOperation(t, providerOwner, "unStakeNodes@"+blsKeys[1], chainSimulatorIntegrationTests.ZeroValue, staking.GasLimitForStakeOperation)
	err = cs.GenerateBlocks(2)
	require.NoError(t, err)
	context.requireProviderNodeStatus(t, blsKeys[1], staking.UnStakedStatus)

	context.sendOperation(t, providerOwner, "unBondNodes@"+blsKeys[1], chainSimulatorIntegrationTests.ZeroValue, staking.GasLimitForStakeOperation)
	context.requireProviderNodeStatus(t, blsKeys[1], staking.UnStakedStatus)

	context.reachEpoch(t, rewardsFixActivationEpoch+3)
	context.sendOperation(t, providerOwner, "unBondNodes@"+blsKeys[1], chainSimulatorIntegrationTests.ZeroValue, staking.GasLimitForStakeOperation)
	context.requireProviderNodeStatus(t, blsKeys[1], staking.NotStakedStatus)
}

type chainSimulatorRewardsContext struct {
	cs              chainSimulatorIntegrationTests.ChainSimulator
	delegationSC    []byte
	providerOwner   dtos.WalletAddress
	delegators      []dtos.WalletAddress
	delegationValue *big.Int
}

func newChainSimulatorRewardsContext(t *testing.T) *chainSimulatorRewardsContext {
	cs := newStakingV4SupernovaSimulator(t, 3, nil)

	err := cs.GenerateBlocksUntilEpochIsReached(int32(rewardsStakingV4Step3Epoch))
	require.NoError(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	require.True(t, metachainNode.GetCoreComponents().EnableEpochsHandler().IsFlagEnabled(common.StakingV4Step3Flag))

	delegationSC, providerOwner := convertGenesisValidatorToDelegationProvider(t, cs)
	delegationValue := new(big.Int).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(100))
	delegators := make([]dtos.WalletAddress, numInitialRewardDelegators)
	delegateTxs := make([]*transaction.Transaction, numInitialRewardDelegators)
	initialBalance := new(big.Int).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(1000))
	for index := range delegators {
		delegators[index], err = cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)
		delegateTxs[index] = chainSimulatorIntegrationTests.GenerateTransaction(
			delegators[index].Bytes,
			0,
			delegationSC,
			delegationValue,
			"delegate",
			gasLimitForDelegate,
		)
	}

	err = cs.GenerateBlocks(1)
	require.NoError(t, err)
	results, err := cs.SendTxsAndGenerateBlocksTilAreExecuted(delegateTxs, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.NoError(t, err)
	require.Len(t, results, len(delegateTxs))
	for _, result := range results {
		require.Equal(t, transaction.TxStatusSuccess, result.Status)
	}

	return &chainSimulatorRewardsContext{
		cs:              cs,
		delegationSC:    delegationSC,
		providerOwner:   providerOwner,
		delegators:      delegators,
		delegationValue: delegationValue,
	}
}

func newStakingV4SupernovaSimulator(
	t *testing.T,
	minNodesPerShard uint32,
	alterConfigs func(cfg *config.Configs),
) chainSimulatorIntegrationTests.ChainSimulator {
	roundsPerEpoch := core.OptionalUint64{HasValue: true, Value: rewardsEpochLength}
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    3,
		RoundDurationInMillis:          roundDurationInMillis,
		SupernovaRoundDurationInMillis: supernovaRoundDurationInMillis,
		RoundsPerEpoch:                 roundsPerEpoch,
		SupernovaRoundsPerEpoch:        roundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               minNodesPerShard,
		MetaChainMinNodes:              minNodesPerShard,
		AlterConfigsFunction: func(cfg *config.Configs) {
			configs.SetStakingV4ActivationEpochs(cfg, rewardsStakingV4Step1Epoch)
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = rewardsSupernovaEpoch
			cfg.EpochConfig.EnableEpochs.FixEpochChangeProposedCurrentEpochEnableEpoch = rewardsFixActivationEpoch
			supernovaRoundConfig := cfg.RoundConfig.RoundActivations[string(common.SupernovaRoundFlag)]
			supernovaRoundConfig.Round = fmt.Sprintf("%d", uint64(rewardsSupernovaEpoch)*rewardsEpochLength+10)
			cfg.RoundConfig.RoundActivations[string(common.SupernovaRoundFlag)] = supernovaRoundConfig
			if alterConfigs != nil {
				alterConfigs(cfg)
			}
		},
	})
	require.NoError(t, err)
	require.NotNil(t, cs)

	return cs
}

func convertGenesisValidatorToDelegationProvider(
	t *testing.T,
	cs chainSimulatorIntegrationTests.ChainSimulator,
) ([]byte, dtos.WalletAddress) {
	owner := cs.GetInitialWalletKeys().StakeWallets[0].Address
	ownerBalance := new(big.Int).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(10000))
	err := cs.SetStateMultiple([]*dtos.AddressState{{
		Address: owner.Bech32,
		Balance: ownerBalance.String(),
	}})
	require.NoError(t, err)
	err = cs.GenerateBlocks(1)
	require.NoError(t, err)

	account, err := cs.GetAccount(owner)
	require.NoError(t, err)
	txData := fmt.Sprintf("makeNewContractFromValidatorData@%s@%s", maxCap, hexServiceFee)
	tx := chainSimulatorIntegrationTests.GenerateTransaction(
		owner.Bytes,
		account.Nonce,
		vm.DelegationManagerSCAddress,
		chainSimulatorIntegrationTests.ZeroValue,
		txData,
		gasLimitForConvertOperation,
	)
	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, transaction.TxStatusSuccess, result.Status)
	require.NotNil(t, result.Logs)
	require.NotEmpty(t, result.Logs.Events)
	require.GreaterOrEqual(t, len(result.Logs.Events[0].Topics), 2)

	return result.Logs.Events[0].Topics[1], owner
}

func (context *chainSimulatorRewardsContext) reachEpoch(t *testing.T, epoch uint32) {
	err := context.cs.GenerateBlocksUntilEpochIsReached(int32(epoch))
	require.NoError(t, err)
}

func (context *chainSimulatorRewardsContext) addDelegator(t *testing.T) {
	initialBalance := new(big.Int).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(1000))
	delegator, err := context.cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)
	err = context.cs.GenerateBlocks(1)
	require.NoError(t, err)

	context.delegators = append(context.delegators, delegator)
	context.sendOperation(t, delegator, "delegate", context.delegationValue, gasLimitForDelegate)
	require.Zero(t, context.delegationValue.Cmp(context.activeStake(t, newBeforeFixRewardDelegatorIndex)))
}

func (context *chainSimulatorRewardsContext) addNode(t *testing.T, blsKey string) {
	txData := fmt.Sprintf("addNodes@%s@%s", blsKey, staking.MockBLSSignature)
	context.sendOperation(t, context.providerOwner, txData, chainSimulatorIntegrationTests.ZeroValue, staking.GasLimitForStakeOperation)
	context.requireProviderNodeStatus(t, blsKey, staking.NotStakedStatus)
}

func (context *chainSimulatorRewardsContext) stakeNode(t *testing.T, blsKey string) {
	context.sendOperation(t, context.providerOwner, "stakeNodes@"+blsKey, chainSimulatorIntegrationTests.ZeroValue, staking.GasLimitForStakeOperation)
	context.requireProviderNodeStatus(t, blsKey, staking.StakedStatus)
}

func (context *chainSimulatorRewardsContext) requireProviderNodeStatus(t *testing.T, blsKey string, expected string) {
	metachainNode := context.cs.GetNodeHandler(core.MetachainShardId)
	nodeStates := staking.GetAllNodeStates(t, metachainNode, context.delegationSC)
	require.Equal(t, expected, nodeStates[blsKey])
}

func (context *chainSimulatorRewardsContext) delegate(t *testing.T, index int, value *big.Int) {
	activeStakeBefore := context.activeStake(t, index)
	context.sendOperation(t, context.delegators[index], "delegate", value, gasLimitForDelegate)
	expected := new(big.Int).Add(activeStakeBefore, value)
	require.Zero(t, expected.Cmp(context.activeStake(t, index)))
}

func (context *chainSimulatorRewardsContext) unDelegate(t *testing.T, index int, value *big.Int) {
	activeStakeBefore := context.activeStake(t, index)
	txData := "unDelegate@" + hex.EncodeToString(value.Bytes())
	context.sendOperation(t, context.delegators[index], txData, chainSimulatorIntegrationTests.ZeroValue, gasLimitForUndelegateOperation)
	expected := new(big.Int).Sub(activeStakeBefore, value)
	require.Zero(t, expected.Cmp(context.activeStake(t, index)))
}

func (context *chainSimulatorRewardsContext) reDelegateRewards(t *testing.T, index int) {
	claimable := context.claimableRewards(t, index)
	require.Positive(t, claimable.Sign())
	activeStakeBefore := context.activeStake(t, index)
	totalRewardsBefore := context.totalCumulatedRewards(t, index)

	context.sendOperation(t, context.delegators[index], "reDelegateRewards", chainSimulatorIntegrationTests.ZeroValue, gasLimitForDelegate)

	expectedStake := new(big.Int).Add(activeStakeBefore, claimable)
	require.Zero(t, expectedStake.Cmp(context.activeStake(t, index)))
	require.Zero(t, context.claimableRewards(t, index).Sign())
	require.Zero(t, totalRewardsBefore.Cmp(context.totalCumulatedRewards(t, index)))
}

func (context *chainSimulatorRewardsContext) claim(t *testing.T, index int, requirePositive bool) {
	context.claimAddress(t, context.delegators[index], requirePositive)
}

func (context *chainSimulatorRewardsContext) claimProviderOwner(t *testing.T, requirePositive bool) {
	context.claimAddress(t, context.providerOwner, requirePositive)
}

func (context *chainSimulatorRewardsContext) claimAddress(t *testing.T, address dtos.WalletAddress, requirePositive bool) {
	claimable := context.claimableRewardsForAddress(t, address.Bytes)
	if requirePositive {
		require.Positive(t, claimable.Sign())
	}
	totalRewardsBefore := context.totalCumulatedRewardsForAddress(t, address.Bytes)

	context.sendOperation(t, address, "claimRewards", chainSimulatorIntegrationTests.ZeroValue, gasLimitForDelegate)

	require.Zero(t, context.claimableRewardsForAddress(t, address.Bytes).Sign())
	require.Zero(t, totalRewardsBefore.Cmp(context.totalCumulatedRewardsForAddress(t, address.Bytes)))
}

func (context *chainSimulatorRewardsContext) withdraw(t *testing.T, index int, expectedValue *big.Int) {
	address := context.delegators[index]
	unStakedValue := context.querySingleBigInt(t, "getUserUnStakedValue", address.Bytes)
	unBondableValue := context.querySingleBigInt(t, "getUserUnBondable", address.Bytes)
	require.Zero(t, expectedValue.Cmp(unStakedValue))
	require.Zero(t, expectedValue.Cmp(unBondableValue))

	context.sendOperation(t, address, "withdraw", chainSimulatorIntegrationTests.ZeroValue, gasLimitForUndelegateOperation)

	output, err := executeQuery(context.cs, core.MetachainShardId, context.delegationSC, "isDelegator", [][]byte{address.Bytes})
	require.NoError(t, err)
	require.NotEqual(t, chainSimulatorIntegrationTests.OkReturnCode, output.ReturnCode)
}

func (context *chainSimulatorRewardsContext) sendOperation(
	t *testing.T,
	delegator dtos.WalletAddress,
	data string,
	value *big.Int,
	gasLimit uint64,
) {
	account, err := context.cs.GetAccount(delegator)
	require.NoError(t, err)
	tx := chainSimulatorIntegrationTests.GenerateTransaction(
		delegator.Bytes,
		account.Nonce,
		context.delegationSC,
		value,
		data,
		gasLimit,
	)
	result, err := context.cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, transaction.TxStatusSuccess, result.Status)
}

func (context *chainSimulatorRewardsContext) activeStake(t *testing.T, index int) *big.Int {
	return context.querySingleBigInt(t, "getUserActiveStake", context.delegators[index].Bytes)
}

func (context *chainSimulatorRewardsContext) claimableRewards(t *testing.T, index int) *big.Int {
	return context.claimableRewardsForAddress(t, context.delegators[index].Bytes)
}

func (context *chainSimulatorRewardsContext) totalCumulatedRewards(t *testing.T, index int) *big.Int {
	return context.totalCumulatedRewardsForAddress(t, context.delegators[index].Bytes)
}

func (context *chainSimulatorRewardsContext) claimableRewardsForAddress(t *testing.T, address []byte) *big.Int {
	return context.querySingleBigInt(t, "getClaimableRewards", address)
}

func (context *chainSimulatorRewardsContext) totalCumulatedRewardsForAddress(t *testing.T, address []byte) *big.Int {
	return context.querySingleBigInt(t, "getTotalCumulatedRewardsForUser", address)
}

func (context *chainSimulatorRewardsContext) querySingleBigInt(t *testing.T, function string, argument []byte) *big.Int {
	output, err := executeQuery(context.cs, core.MetachainShardId, context.delegationSC, function, [][]byte{argument})
	require.NoError(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, output.ReturnCode)
	require.Len(t, output.ReturnData, 1)

	return new(big.Int).SetBytes(output.ReturnData[0])
}

func (context *chainSimulatorRewardsContext) requireRewardRecord(t *testing.T, epoch uint32) {
	output, err := executeQuery(
		context.cs,
		core.MetachainShardId,
		context.delegationSC,
		"getRewardData",
		[][]byte{new(big.Int).SetUint64(uint64(epoch)).Bytes()},
	)
	require.NoError(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, output.ReturnCode)
	require.Len(t, output.ReturnData, 3)
	require.Positive(t, new(big.Int).SetBytes(output.ReturnData[0]).Sign())
	require.Positive(t, new(big.Int).SetBytes(output.ReturnData[1]).Sign())
}

func (context *chainSimulatorRewardsContext) requireMissingRewardRecord(t *testing.T, epoch uint32) {
	output, err := executeQuery(
		context.cs,
		core.MetachainShardId,
		context.delegationSC,
		"getRewardData",
		[][]byte{new(big.Int).SetUint64(uint64(epoch)).Bytes()},
	)
	require.NoError(t, err)
	require.NotEqual(t, chainSimulatorIntegrationTests.OkReturnCode, output.ReturnCode)
	require.Equal(t, "reward not found", output.ReturnMessage)
}
