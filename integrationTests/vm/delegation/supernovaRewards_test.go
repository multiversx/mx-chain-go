package delegation

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
	ssc "github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
)

const (
	supernovaTransitionEpoch = uint32(2)
	delegationValue          = int64(1000)
)

const (
	untouchedDelegatorIndex = iota
	claimedLegacyDelegatorIndex
	claimedAfterOverwriteDelegatorIndex
	claimedAfterPreFixDelegatorIndex
	delegatedBeforeFixDelegatorIndex
	undelegatedBeforeFixDelegatorIndex
	redelegatedBeforeFixDelegatorIndex
	numInitialRolloutDelegators
	newBeforeFixDelegatorIndex = numInitialRolloutDelegators
)

func TestDelegationRewardsAcrossSupernovaTransitions(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("fix after multiple V3 epoch changes", func(t *testing.T) {
		context := newDelegationRewardsTransitionContext(t, supernovaTransitionEpoch+3, numInitialRolloutDelegators)
		defer context.node.Close()

		claimed := make([]int64, len(context.delegators))
		context.addReward(t, &block.MetaBlock{Epoch: supernovaTransitionEpoch}, 100)
		context.requireRewardRecord(t, supernovaTransitionEpoch)
		claimed[claimedLegacyDelegatorIndex] += context.claim(t, claimedLegacyDelegatorIndex)

		context.addReward(t, &block.MetaBlockV3{Epoch: supernovaTransitionEpoch, EpochChangeProposed: true}, 200)
		context.requireRewardRecord(t, supernovaTransitionEpoch)
		context.setCurrentHeader(t, &block.MetaBlockV3{Epoch: supernovaTransitionEpoch + 1})
		claimed[claimedAfterOverwriteDelegatorIndex] += context.claim(t, claimedAfterOverwriteDelegatorIndex)

		context.addReward(t, &block.MetaBlockV3{Epoch: supernovaTransitionEpoch + 1, EpochChangeProposed: true}, 300)
		context.requireRewardRecord(t, supernovaTransitionEpoch+1)
		context.setCurrentHeader(t, &block.MetaBlockV3{Epoch: supernovaTransitionEpoch + 2})
		claimed[claimedAfterPreFixDelegatorIndex] += context.claim(t, claimedAfterPreFixDelegatorIndex)

		context.delegate(t, delegatedBeforeFixDelegatorIndex, delegationValue)
		context.unDelegate(t, undelegatedBeforeFixDelegatorIndex, delegationValue)
		context.reDelegateRewards(t, redelegatedBeforeFixDelegatorIndex)
		claimed[redelegatedBeforeFixDelegatorIndex] += 500

		lateDelegator := getAddresses(1)[0]
		context.delegateAddress(t, lateDelegator, delegationValue)
		context.delegators = append(context.delegators, lateDelegator)
		claimed = append(claimed, 0)
		require.Equal(t, newBeforeFixDelegatorIndex+1, len(context.delegators))

		statesBeforeActivation := []delegatorRewardsExpectation{
			{checkpoint: supernovaTransitionEpoch},
			{checkpoint: supernovaTransitionEpoch + 1, totalCumulated: 100},
			{checkpoint: supernovaTransitionEpoch + 2, totalCumulated: 200},
			{checkpoint: supernovaTransitionEpoch + 3, totalCumulated: 500},
			{checkpoint: supernovaTransitionEpoch + 3, unclaimed: 500},
			{checkpoint: supernovaTransitionEpoch + 3, unclaimed: 500},
			{checkpoint: supernovaTransitionEpoch + 3, totalCumulated: 500},
			{checkpoint: supernovaTransitionEpoch + 3},
		}
		context.requireDelegatorStates(t, statesBeforeActivation)

		context.addReward(t, &block.MetaBlockV3{Epoch: supernovaTransitionEpoch + 2, EpochChangeProposed: true}, 400)
		context.requireMissingRewardRecord(t, supernovaTransitionEpoch+2)
		context.requireRewardRecord(t, supernovaTransitionEpoch+3)
		context.setCurrentHeader(t, &block.MetaBlockV3{Epoch: supernovaTransitionEpoch + 3})
		claimed[untouchedDelegatorIndex] += context.claim(t, untouchedDelegatorIndex)
		context.unDelegate(t, claimedAfterOverwriteDelegatorIndex, delegationValue)
		context.delegate(t, claimedLegacyDelegatorIndex, delegationValue)
		context.reDelegateRewards(t, claimedAfterPreFixDelegatorIndex)
		claimed[claimedAfterPreFixDelegatorIndex] += 400
		for _, index := range []int{
			delegatedBeforeFixDelegatorIndex,
			undelegatedBeforeFixDelegatorIndex,
			redelegatedBeforeFixDelegatorIndex,
			newBeforeFixDelegatorIndex,
		} {
			claimed[index] += context.claim(t, index)
		}

		statesAfterActivation := []delegatorRewardsExpectation{
			{checkpoint: supernovaTransitionEpoch + 4, totalCumulated: 900},
			{checkpoint: supernovaTransitionEpoch + 4, unclaimed: 700, totalCumulated: 100},
			{checkpoint: supernovaTransitionEpoch + 4, unclaimed: 400, totalCumulated: 200},
			{checkpoint: supernovaTransitionEpoch + 4, totalCumulated: 900},
			{checkpoint: supernovaTransitionEpoch + 4, totalCumulated: 1300},
			{checkpoint: supernovaTransitionEpoch + 4, totalCumulated: 500},
			{checkpoint: supernovaTransitionEpoch + 4, totalCumulated: 1100},
			{checkpoint: supernovaTransitionEpoch + 4, totalCumulated: 400},
		}
		context.requireDelegatorStates(t, statesAfterActivation)
		context.requireActiveStake(t, claimedLegacyDelegatorIndex, 2*delegationValue)
		context.requireActiveStake(t, claimedAfterOverwriteDelegatorIndex, 0)
		context.requireActiveStake(t, claimedAfterPreFixDelegatorIndex, delegationValue+400)
		context.requireActiveStake(t, delegatedBeforeFixDelegatorIndex, 2*delegationValue)
		context.requireActiveStake(t, undelegatedBeforeFixDelegatorIndex, 0)
		context.requireActiveStake(t, redelegatedBeforeFixDelegatorIndex, delegationValue+500)
		context.requireActiveStake(t, newBeforeFixDelegatorIndex, delegationValue)

		context.addReward(t, &block.MetaBlockV3{Epoch: supernovaTransitionEpoch + 3, EpochChangeProposed: true}, 500)
		context.setCurrentHeader(t, &block.MetaBlockV3{Epoch: supernovaTransitionEpoch + 4})
		for index := range context.delegators {
			claimed[index] += context.claim(t, index)
		}
		require.Zero(t, context.claim(t, untouchedDelegatorIndex))

		require.Equal(t, []int64{1400, 1800, 600, 1600, 2300, 500, 1850, 900}, claimed)
		context.requireClaimedDelegatorStates(t, supernovaTransitionEpoch+5, claimed)
		context.requireRewardRecord(t, supernovaTransitionEpoch)
		context.requireRewardRecord(t, supernovaTransitionEpoch+1)
		context.requireMissingRewardRecord(t, supernovaTransitionEpoch+2)
		context.requireRewardRecord(t, supernovaTransitionEpoch+3)
		context.requireRewardRecord(t, supernovaTransitionEpoch+4)
	})

	t.Run("fix at first V3 epoch change", func(t *testing.T) {
		context := newDelegationRewardsTransitionContext(t, supernovaTransitionEpoch+1, 4)
		defer context.node.Close()

		claimed := make([]int64, len(context.delegators))
		context.addReward(t, &block.MetaBlock{Epoch: supernovaTransitionEpoch}, 100)
		claimed[0] += context.claim(t, 0)

		context.addReward(t, &block.MetaBlockV3{Epoch: supernovaTransitionEpoch, EpochChangeProposed: true}, 200)
		context.setCurrentHeader(t, &block.MetaBlockV3{Epoch: supernovaTransitionEpoch + 1})
		claimed[0] += context.claim(t, 0)
		claimed[1] += context.claim(t, 1)

		context.addReward(t, &block.MetaBlockV3{Epoch: supernovaTransitionEpoch + 1, EpochChangeProposed: true}, 300)
		context.setCurrentHeader(t, &block.MetaBlockV3{Epoch: supernovaTransitionEpoch + 2})
		claimed[0] += context.claim(t, 0)
		claimed[1] += context.claim(t, 1)
		claimed[2] += context.claim(t, 2)

		context.addReward(t, &block.MetaBlockV3{Epoch: supernovaTransitionEpoch + 2, EpochChangeProposed: true}, 400)
		context.setCurrentHeader(t, &block.MetaBlockV3{Epoch: supernovaTransitionEpoch + 3})
		for index := range context.delegators {
			claimed[index] += context.claim(t, index)
		}

		require.Equal(t, []int64{1000, 1000, 1000, 1000}, claimed)
		context.requireRewardRecord(t, supernovaTransitionEpoch)
		context.requireRewardRecord(t, supernovaTransitionEpoch+1)
		context.requireRewardRecord(t, supernovaTransitionEpoch+2)
		context.requireRewardRecord(t, supernovaTransitionEpoch+3)
	})
}

type rewardRecordExpectation struct {
	rewardsToDistribute int64
	totalActive         int64
}

type delegatorRewardsExpectation struct {
	checkpoint     uint32
	unclaimed      int64
	totalCumulated int64
}

type delegationRewardsTransitionContext struct {
	node            *integrationTests.TestProcessorNode
	delegationSC    []byte
	delegators      [][]byte
	rewardRecords   map[uint32]rewardRecordExpectation
	totalActive     int64
	lastHeaderNonce uint64
}

func newDelegationRewardsTransitionContext(
	t *testing.T,
	fixActivationEpoch uint32,
	numDelegators int,
) *delegationRewardsTransitionContext {
	enableEpochs := config.EnableEpochs{
		SupernovaEnableEpoch:                          supernovaTransitionEpoch,
		FixEpochChangeProposedCurrentEpochEnableEpoch: fixActivationEpoch,
	}
	node := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
		EpochsConfig:         &enableEpochs,
	})
	node.InitDelegationManager()

	context := &delegationRewardsTransitionContext{
		node:          node,
		delegators:    getAddresses(numDelegators),
		rewardRecords: make(map[uint32]rewardRecordExpectation),
		totalActive:   int64(numDelegators+1) * delegationValue,
	}
	context.setCurrentHeader(t, &block.MetaBlock{Epoch: supernovaTransitionEpoch - 1})
	context.delegationSC = deployNewSc(t, node, big.NewInt(10000), big.NewInt(0), big.NewInt(delegationValue), node.OwnAccount.Address)
	require.NotEmpty(t, context.delegationSC)
	processMultipleTransactions(t, node, context.delegators, context.delegationSC, "delegate", big.NewInt(delegationValue))

	return context
}

func (context *delegationRewardsTransitionContext) addReward(
	t *testing.T,
	header data.HeaderHandler,
	rewardPerDelegationValue int64,
) {
	context.lastHeaderNonce++
	require.NoError(t, header.SetNonce(context.lastHeaderNonce))
	context.node.EpochNotifier.CheckEpoch(header)
	value := big.NewInt(rewardPerDelegationValue * context.totalActive / delegationValue)
	addRewardsToDelegationForHeader(t, context.node, context.delegationSC, value, header)
	context.rewardRecords[context.node.BlockchainHook.CurrentEpoch()] = rewardRecordExpectation{
		rewardsToDistribute: value.Int64(),
		totalActive:         context.totalActive,
	}
}

func (context *delegationRewardsTransitionContext) setCurrentHeader(t *testing.T, header data.HeaderHandler) {
	context.lastHeaderNonce++
	require.NoError(t, header.SetNonce(context.lastHeaderNonce))
	addEpochStartHeader(t, context.node, header.GetEpoch(), 0)
	context.node.EpochNotifier.CheckEpoch(header)
	require.NoError(t, context.node.BlockchainHook.SetCurrentHeader(header))
}

func (context *delegationRewardsTransitionContext) delegate(t *testing.T, delegatorIndex int, value int64) {
	context.delegateAddress(t, context.delegators[delegatorIndex], value)
}

func (context *delegationRewardsTransitionContext) delegateAddress(t *testing.T, delegator []byte, value int64) {
	returnCode, err := processTransaction(context.node, delegator, context.delegationSC, "delegate", big.NewInt(value))
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	context.totalActive += value
}

func (context *delegationRewardsTransitionContext) unDelegate(t *testing.T, delegatorIndex int, value int64) {
	txData := "unDelegate@" + hex.EncodeToString(big.NewInt(value).Bytes())
	returnCode, err := processTransaction(context.node, context.delegators[delegatorIndex], context.delegationSC, txData, big.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	context.totalActive -= value
}

func (context *delegationRewardsTransitionContext) reDelegateRewards(t *testing.T, delegatorIndex int) {
	claimable := context.claimable(t, delegatorIndex)
	returnCode, err := processTransaction(context.node, context.delegators[delegatorIndex], context.delegationSC, "reDelegateRewards", big.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	context.totalActive += claimable
}

func (context *delegationRewardsTransitionContext) claim(t *testing.T, delegatorIndex int) int64 {
	delegator := context.delegators[delegatorIndex]
	claimable := context.claimable(t, delegatorIndex)
	returnCode, err := processTransaction(context.node, delegator, context.delegationSC, "claimRewards", big.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	require.Empty(t, viewFuncSingleResult(t, context.node, context.delegationSC, "getClaimableRewards", [][]byte{delegator}))

	return claimable
}

func (context *delegationRewardsTransitionContext) claimable(t *testing.T, delegatorIndex int) int64 {
	delegator := context.delegators[delegatorIndex]
	result := viewFuncSingleResult(t, context.node, context.delegationSC, "getClaimableRewards", [][]byte{delegator})

	return big.NewInt(0).SetBytes(result).Int64()
}

func (context *delegationRewardsTransitionContext) requireActiveStake(t *testing.T, delegatorIndex int, expected int64) {
	delegator := context.delegators[delegatorIndex]
	result := viewFuncSingleResult(t, context.node, context.delegationSC, "getUserActiveStake", [][]byte{delegator})
	require.Equal(t, big.NewInt(expected), big.NewInt(0).SetBytes(result))
}

func (context *delegationRewardsTransitionContext) requireDelegatorStates(
	t *testing.T,
	expected []delegatorRewardsExpectation,
) {
	require.Len(t, expected, len(context.delegators))
	for index := range context.delegators {
		context.requireDelegatorState(t, index, expected[index])
	}
}

func (context *delegationRewardsTransitionContext) requireClaimedDelegatorStates(
	t *testing.T,
	checkpoint uint32,
	totalCumulated []int64,
) {
	require.Len(t, totalCumulated, len(context.delegators))
	for index := range context.delegators {
		context.requireDelegatorState(t, index, delegatorRewardsExpectation{
			checkpoint:     checkpoint,
			totalCumulated: totalCumulated[index],
		})
	}
}

func (context *delegationRewardsTransitionContext) requireDelegatorState(
	t *testing.T,
	delegatorIndex int,
	expected delegatorRewardsExpectation,
) {
	delegator := context.delegators[delegatorIndex]
	marshaledData, _, err := context.node.BlockchainHook.GetStorageData(context.delegationSC, delegator)
	require.NoError(t, err)
	require.NotEmpty(t, marshaledData)

	delegatorData := &ssc.DelegatorData{}
	require.NoError(t, integrationTests.TestMarshalizer.Unmarshal(delegatorData, marshaledData))
	require.Equal(t, expected.checkpoint, delegatorData.RewardsCheckpoint)
	require.Equal(t, big.NewInt(expected.unclaimed), delegatorData.UnClaimedRewards)
	require.Equal(t, big.NewInt(expected.totalCumulated), delegatorData.TotalCumulatedRewards)
}

func (context *delegationRewardsTransitionContext) requireRewardRecord(
	t *testing.T,
	epoch uint32,
) {
	expected, ok := context.rewardRecords[epoch]
	require.True(t, ok)

	output := context.queryRewardRecord(t, epoch)
	require.Equal(t, vmcommon.Ok, output.ReturnCode)
	require.Len(t, output.ReturnData, 3)
	require.Equal(t, big.NewInt(expected.rewardsToDistribute), big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, big.NewInt(expected.totalActive), big.NewInt(0).SetBytes(output.ReturnData[1]))
	require.Empty(t, output.ReturnData[2])
}

func (context *delegationRewardsTransitionContext) requireMissingRewardRecord(t *testing.T, epoch uint32) {
	require.NotContains(t, context.rewardRecords, epoch)

	output := context.queryRewardRecord(t, epoch)
	require.Equal(t, vmcommon.UserError, output.ReturnCode)
	require.Equal(t, "reward not found", output.ReturnMessage)
}

func (context *delegationRewardsTransitionContext) queryRewardRecord(t *testing.T, epoch uint32) *vmcommon.VMOutput {
	query := &process.SCQuery{
		ScAddress:  context.delegationSC,
		FuncName:   "getRewardData",
		CallerAddr: vm.EndOfEpochAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{big.NewInt(0).SetUint64(uint64(epoch)).Bytes()},
	}
	output, _, err := context.node.SCQueryService.ExecuteQuery(query)
	require.NoError(t, err)

	return output
}
