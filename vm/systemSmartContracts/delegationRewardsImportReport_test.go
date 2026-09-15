package systemSmartContracts

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestDelegationRewardsOperationUsers(t *testing.T) {
	t.Parallel()

	caller := []byte("caller")
	other := []byte("other")
	testCases := []struct {
		name              string
		function          string
		arguments         [][]byte
		expectedAddresses [][]byte
	}{
		{
			name:              "caller operation",
			function:          claimRewards,
			expectedAddresses: [][]byte{caller},
		},
		{
			name:              "validator conversion",
			function:          initFromValidatorData,
			arguments:         [][]byte{other},
			expectedAddresses: [][]byte{other},
		},
		{
			name:              "owner change",
			function:          changeOwner,
			arguments:         [][]byte{other},
			expectedAddresses: [][]byte{caller, other},
		},
		{
			name:              "owner change deduplicates address",
			function:          changeOwner,
			arguments:         [][]byte{caller},
			expectedAddresses: [][]byte{caller},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			input := &vmcommon.ContractCallInput{
				VMInput: vmcommon.VMInput{
					CallerAddr: caller,
					Arguments:  testCase.arguments,
					CallValue:  big.NewInt(0),
				},
				Function: testCase.function,
			}

			require.Equal(t, testCase.expectedAddresses, delegationRewardsOperationUsers(input))
		})
	}
}

func TestDelegationRewardRecordSnapshot(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	delegationContract, err := NewDelegationSystemSC(args)
	require.NoError(t, err)

	record := &RewardComputationData{
		RewardsToDistribute: big.NewInt(123),
		TotalActive:         big.NewInt(456),
		ServiceFee:          789,
	}
	raw, err := args.Marshalizer.Marshal(record)
	require.NoError(t, err)

	snapshot := delegationContract.delegationRewardRecordSnapshot(raw)
	require.True(t, snapshot.Exists)
	require.Equal(t, hex.EncodeToString(raw), snapshot.Raw)
	require.Empty(t, snapshot.DecodeError)
	require.Equal(t, "123", snapshot.RewardsToDistribute)
	require.Equal(t, "456", snapshot.TotalActive)
	require.Equal(t, uint64(789), snapshot.ServiceFee)
}

func TestDelegationRewardsReportIsDisabledOutsideImportDB(t *testing.T) {
	t.Parallel()

	delegationContract := &delegation{}
	input := &vmcommon.ContractCallInput{Function: claimRewards}

	require.Nil(t, delegationContract.startDelegationRewardsOperationReport(input))
	require.False(t, delegationContract.shouldReportDelegationRewardCalculation(input))
}

func TestDelegationRewardsReportExcludesReadOnlyOperations(t *testing.T) {
	t.Parallel()

	for _, function := range []string{
		core.SCDeployInitFunctionName,
		initFromValidatorData,
		mergeValidatorDataToDelegation,
		delegate,
		"unDelegate",
		withdraw,
		claimRewards,
		reDelegateRewards,
		changeOwner,
	} {
		require.True(t, isDelegationRewardsOperation(function), function)
	}

	require.False(t, isDelegationRewardsOperation("getClaimableRewards"))
	require.False(t, isDelegationRewardsOperation("getDelegatorFundsData"))
	require.False(t, isDelegationRewardsOperation("getTotalCumulatedRewardsForUser"))
}

func TestDelegationRewardsProviderSnapshotUsesExplicitBalanceViews(t *testing.T) {
	t.Parallel()

	provider := []byte("provider")
	eei := createDefaultEei()
	eei.SetSCAddress(provider)
	eei.outputAccounts[string(provider)] = &vmcommon.OutputAccount{
		Balance:      big.NewInt(17),
		BalanceDelta: big.NewInt(-3),
	}
	args := createMockArgumentsForDelegation()
	args.Eei = eei
	delegationContract, err := NewDelegationSystemSC(args)
	require.NoError(t, err)

	snapshot := delegationContract.delegationRewardsProviderSnapshot(provider)
	require.Equal(t, "0", snapshot.PersistedBalance)
	require.True(t, snapshot.HasOutputAccount)
	require.Equal(t, "17", snapshot.OutputAccountBalance)
	require.Equal(t, "-3", snapshot.OutputAccountBalanceDelta)

	payload, err := json.Marshal(snapshot)
	require.NoError(t, err)
	require.NotContains(t, string(payload), "executionBalance")
	require.NotContains(t, string(payload), "delegationStatusRaw")
}

func TestDelegationRewardWriteIncludesFollowingSlot(t *testing.T) {
	t.Parallel()

	const epoch = uint32(12)
	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	provider := []byte("provider")
	eei.SetSCAddress(provider)
	args.Eei = eei
	delegationContract, err := NewDelegationSystemSC(args)
	require.NoError(t, err)

	previousRaw := marshalRewardRecord(t, args, 10)
	currentRaw := marshalRewardRecord(t, args, 20)
	followingRaw := marshalRewardRecord(t, args, 30)
	eei.SetStorage(rewardKeyForEpoch(epoch), currentRaw)

	rewardWrite := delegationContract.newDelegationRewardWrite(provider, epoch, previousRaw, followingRaw)
	require.Equal(t, epoch, rewardWrite.Epoch)
	require.Equal(t, "10", rewardWrite.Previous.RewardsToDistribute)
	require.Equal(t, "20", rewardWrite.Current.RewardsToDistribute)
	require.Equal(t, epoch+1, rewardWrite.FollowingSlot.Epoch)
	require.Equal(t, hex.EncodeToString(rewardKeyForEpoch(epoch+1)), rewardWrite.FollowingSlot.Key)
	require.Equal(t, "30", rewardWrite.FollowingSlot.Record.RewardsToDistribute)

	rewardWrite = delegationContract.newDelegationRewardWrite(provider, epoch, previousRaw, nil)
	require.False(t, rewardWrite.FollowingSlot.Record.Exists)
}

func TestDelegationRewardCalculationEntryIncludesCalculationParameters(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	delegationContract, err := NewDelegationSystemSC(args)
	require.NoError(t, err)
	recordRaw := marshalRewardRecord(t, args, 50)

	entry := delegationContract.newDelegationRewardCalculationEntry(
		[]byte("user"),
		7,
		recordRaw,
		big.NewInt(100),
		true,
		big.NewInt(5),
		big.NewInt(45),
		big.NewInt(50),
	)

	require.Equal(t, "5", entry.ProviderOwnerShare)
	require.Equal(t, "45", entry.UserStakeShare)
	require.Equal(t, "50", entry.UserReward)
	require.Equal(t, args.DelegationSCConfig.MaxServiceFee, entry.ServiceFeeDenominator)
	require.True(t, entry.UsesTrimmedPercentage)
	payload, err := json.Marshal(entry)
	require.NoError(t, err)
	require.NotContains(t, string(payload), "ownerReward")
	require.NotContains(t, string(payload), "stakeReward")
	require.Contains(t, string(payload), "providerOwnerShare")
	require.Contains(t, string(payload), "userStakeShare")

	args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub).RemoveActiveFlags(common.StakingV2FlagAfterEpoch)
	entry = delegationContract.newDelegationRewardCalculationEntry(
		[]byte("user"),
		7,
		recordRaw,
		big.NewInt(100),
		true,
		big.NewInt(5),
		big.NewInt(45),
		big.NewInt(50),
	)
	require.False(t, entry.UsesTrimmedPercentage)
}

func TestDelegationRewardsManagerOperationIncludesOuterResult(t *testing.T) {
	t.Parallel()

	manager := &delegationManager{}
	capture := &delegationRewardsManagerOperationCapture{
		context: delegationRewardsImportContext{
			Function:      "mergeValidatorToDelegationSameOwner",
			CurrentTxHash: "abcd",
		},
		providers: [][]byte{[]byte("provider")},
	}

	event := manager.newDelegationRewardsManagerOperationEvent(capture, vmcommon.UserError)
	require.Equal(t, "manager_operation", event.Event)
	require.Equal(t, "abcd", event.Context.CurrentTxHash)
	require.Equal(t, []string{hex.EncodeToString([]byte("provider"))}, event.ManagerOperation.Providers)
	require.Equal(t, int(vmcommon.UserError), event.ManagerOperation.ReturnCode)
}

func TestDelegationRewardsManagerOperationSelection(t *testing.T) {
	t.Parallel()

	for _, function := range []string{
		"createNewDelegationContract",
		"makeNewContractFromValidatorData",
		"mergeValidatorToDelegationSameOwner",
		"mergeValidatorToDelegationWithWhitelist",
	} {
		require.True(t, isDelegationRewardsManagerOperation(function), function)
	}
	require.False(t, isDelegationRewardsManagerOperation("getContractConfig"))
}

func TestDelegationRewardsManagerOperationResolvesCreatedProvider(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManager()
	lastAddress := []byte{0, 0, 7}
	managementData, err := args.Marshalizer.Marshal(&DelegationManagement{LastAddress: lastAddress})
	require.NoError(t, err)
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if string(key) == delegationManagementKey {
				return managementData
			}

			return nil
		},
	}
	manager, err := NewDelegationManagerSystemSC(args)
	require.NoError(t, err)

	providers := manager.delegationRewardsManagerOperationProviders(&vmcommon.ContractCallInput{
		Function: "createNewDelegationContract",
	})

	require.Equal(t, [][]byte{createNewAddress(lastAddress)}, providers)
}

func marshalRewardRecord(t *testing.T, args ArgsNewDelegation, rewards int64) []byte {
	t.Helper()

	raw, err := args.Marshalizer.Marshal(&RewardComputationData{
		RewardsToDistribute: big.NewInt(rewards),
		TotalActive:         big.NewInt(100),
		ServiceFee:          10,
	})
	require.NoError(t, err)

	return raw
}

func TestDelegationRewardsImportReportPreservesRewardUpdate(t *testing.T) {
	t.Parallel()

	const currentEpoch = uint32(15)
	args := createMockArgumentsForDelegation()
	args.IsImportDBMode = true
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return currentEpoch
		},
	}
	args.Eei = eei
	delegationContract, err := NewDelegationSystemSC(args)
	require.NoError(t, err)

	delegationContract.eei.SetStorage([]byte(totalActiveKey), big.NewInt(200).Bytes())
	delegationContract.eei.SetStorage([]byte(serviceFeeKey), big.NewInt(100).Bytes())
	input := getDefaultVmInputForFunc("updateRewards", nil)
	input.CallerAddr = vm.EndOfEpochAddress
	input.CallValue = big.NewInt(20)

	require.Equal(t, vmcommon.Ok, delegationContract.Execute(input))
	found, storedRecord, err := delegationContract.getRewardComputationData(currentEpoch)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, big.NewInt(20), storedRecord.RewardsToDistribute)
	require.Equal(t, big.NewInt(200), storedRecord.TotalActive)
	require.Equal(t, uint64(100), storedRecord.ServiceFee)
}

func TestDelegationRewardsImportReportPreservesClaim(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	args.IsImportDBMode = true
	args.DelegationSCConfig.MaxServiceFee = 10000
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	args.Eei = eei
	delegationContract, err := NewDelegationSystemSC(args)
	require.NoError(t, err)

	input := getDefaultVmInputForFunc(claimRewards, nil)
	fundKey := []byte{1}
	require.NoError(t, delegationContract.saveDelegatorData(input.CallerAddr, &DelegatorData{
		ActiveFund:            fundKey,
		RewardsCheckpoint:     0,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	}))
	require.NoError(t, delegationContract.saveFund(fundKey, &Fund{Value: big.NewInt(1000)}))
	require.NoError(t, delegationContract.saveRewardData(0, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(1000),
		ServiceFee:          1000,
	}))

	require.Equal(t, vmcommon.Ok, delegationContract.Execute(input))
	_, delegatorData, err := delegationContract.getOrCreateDelegatorData(input.CallerAddr)
	require.NoError(t, err)
	require.Equal(t, uint32(3), delegatorData.RewardsCheckpoint)
	require.Zero(t, delegatorData.UnClaimedRewards.Sign())
	require.Equal(t, big.NewInt(90), delegatorData.TotalCumulatedRewards)
}
