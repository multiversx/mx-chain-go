package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

type delegationRewardsImportFormatter struct{}

func (drif *delegationRewardsImportFormatter) Output(line logger.LogLineHandler) []byte {
	if line.GetLoggerName() != "vm/systemsmartcontracts/rewardsimport" ||
		line.GetMessage() != "delegation rewards import report" {
		return nil
	}

	arguments := line.GetArgs()
	for index := 0; index+1 < len(arguments); index += 2 {
		if arguments[index] == "payload" {
			return append([]byte(arguments[index+1]), '\n')
		}
	}

	return nil
}

func (drif *delegationRewardsImportFormatter) IsInterfaceNil() bool {
	return drif == nil
}

func captureDelegationRewardsImportEvents(t *testing.T) *bytes.Buffer {
	t.Helper()

	originalLogPattern := logger.GetLogLevelPattern()
	require.NoError(t, logger.SetLogLevel("*:INFO"))

	buffer := &bytes.Buffer{}
	require.NoError(t, logger.AddLogObserver(buffer, &delegationRewardsImportFormatter{}))
	t.Cleanup(func() {
		require.NoError(t, logger.RemoveLogObserver(buffer))
		require.NoError(t, logger.SetLogLevel(originalLogPattern))
	})

	return buffer
}

func decodeDelegationRewardsImportEvents(t *testing.T, buffer *bytes.Buffer) []delegationRewardsImportEvent {
	t.Helper()

	decoder := json.NewDecoder(bytes.NewReader(buffer.Bytes()))
	events := make([]delegationRewardsImportEvent, 0)
	for {
		event := delegationRewardsImportEvent{}
		err := decoder.Decode(&event)
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		events = append(events, event)
	}

	return events
}

func delegationRewardsImportEventsByType(
	events []delegationRewardsImportEvent,
	eventType string,
) []delegationRewardsImportEvent {
	filtered := make([]delegationRewardsImportEvent, 0)
	for _, event := range events {
		if event.Event == eventType {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

func requireIncreasingDelegationRewardsImportSequences(t *testing.T, events []delegationRewardsImportEvent) {
	t.Helper()

	var previousSequence uint64
	for _, event := range events {
		require.Equal(t, uint32(delegationRewardsImportReportVersion), event.Version)
		require.Greater(t, event.Sequence, previousSequence)
		previousSequence = event.Sequence
	}
}

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

func TestDelegationRewardsImportReportRecordsManagerOperation(t *testing.T) {
	buffer := captureDelegationRewardsImportEvents(t)
	manager, eei := createTestEEIAndDelegationFormMergeValidator()
	manager.isImportDBMode = true
	provider := bytes.Repeat([]byte{2}, 32)
	input := getDefaultVmInputForDelegationManager(
		"mergeValidatorToDelegationWithWhitelist",
		[][]byte{provider, []byte("extra argument")},
	)
	input.CurrentTxHash = []byte("manager operation")
	input.GasProvided = manager.gasCost.MetaChainSystemSCsCost.ValidatorToDelegation
	eei.gasRemaining = input.GasProvided

	require.Equal(t, vmcommon.UserError, manager.Execute(input))

	successfulInput := prepareVmInputContextAndDelegationManager(manager, eei)
	successfulInput.CurrentTxHash = []byte("successful manager operation")
	require.NoError(t, eei.SetSystemSCContainer(&mock.SystemSCContainerStub{
		GetCalled: func(_ []byte) (vm.SystemSmartContract, error) {
			return &mock.SystemSCStub{
				ExecuteCalled: func(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
					return vmcommon.Ok
				},
			}, nil
		},
	}))
	require.Equal(t, vmcommon.Ok, manager.Execute(successfulInput))

	events := decodeDelegationRewardsImportEvents(t, buffer)
	require.Len(t, events, 2)
	requireIncreasingDelegationRewardsImportSequences(t, events)
	event := events[0]
	require.Equal(t, "manager_operation", event.Event)
	require.NotNil(t, event.ManagerOperation)
	require.Equal(t, int(vmcommon.UserError), event.ManagerOperation.ReturnCode)
	require.Equal(t, vmcommon.UserError.String(), event.ManagerOperation.ReturnCodeName)
	require.Equal(t, []string{hex.EncodeToString(provider)}, event.ManagerOperation.Providers)
	require.Equal(t, hex.EncodeToString(input.CurrentTxHash), event.Context.CurrentTxHash)
	require.Equal(t, encodeByteSlices(input.Arguments), event.Context.Arguments)

	successfulEvent := events[1]
	require.Equal(t, "manager_operation", successfulEvent.Event)
	require.NotNil(t, successfulEvent.ManagerOperation)
	require.Equal(t, int(vmcommon.Ok), successfulEvent.ManagerOperation.ReturnCode)
	require.Equal(t, vmcommon.Ok.String(), successfulEvent.ManagerOperation.ReturnCodeName)
	require.Equal(t, encodeByteSlices(successfulInput.Arguments), successfulEvent.ManagerOperation.Providers)
	require.Equal(t, hex.EncodeToString(successfulInput.CurrentTxHash), successfulEvent.Context.CurrentTxHash)
}

func TestDelegationRewardsImportReportRecordsMultiOperationResults(t *testing.T) {
	buffer := captureDelegationRewardsImportEvents(t)
	manager, eei := createTestEEIAndDelegationFormMergeValidator()
	manager.isImportDBMode = true
	require.NoError(t, eei.SetSystemSCContainer(
		&mock.SystemSCContainerStub{
			GetCalled: func(_ []byte) (vm.SystemSmartContract, error) {
				return &mock.SystemSCStub{
					ExecuteCalled: func(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
						return vmcommon.Ok
					},
				}, nil
			},
		},
	))

	firstProvider := bytes.Repeat([]byte{2}, 32)
	secondProvider := bytes.Repeat([]byte{3}, 32)
	successfulInput := getDefaultVmInputForDelegationManager(
		"claimMulti",
		[][]byte{firstProvider, secondProvider},
	)
	successfulInput.CallerAddr = bytes.Repeat([]byte{1}, 32)
	successfulInput.RecipientAddr = vm.DelegationManagerSCAddress
	successfulInput.CurrentTxHash = []byte("successful multi operation")
	require.Equal(t, vmcommon.Ok, manager.Execute(successfulInput))

	failedInput := getDefaultVmInputForDelegationManager(
		"reDelegateMulti",
		[][]byte{firstProvider, firstProvider},
	)
	failedInput.CallerAddr = bytes.Repeat([]byte{1}, 32)
	failedInput.RecipientAddr = vm.DelegationManagerSCAddress
	failedInput.CurrentTxHash = []byte("failed multi operation")
	require.Equal(t, vmcommon.UserError, manager.Execute(failedInput))

	events := decodeDelegationRewardsImportEvents(t, buffer)
	require.Len(t, events, 2)
	requireIncreasingDelegationRewardsImportSequences(t, events)
	successfulEvent := events[0]
	require.Equal(t, "multi_operation", successfulEvent.Event)
	require.NotNil(t, successfulEvent.MultiOperation)
	require.Equal(t, claimRewards, successfulEvent.MultiOperation.TargetFunction)
	require.Equal(t, encodeByteSlices(successfulInput.Arguments), successfulEvent.MultiOperation.Providers)
	require.Equal(t, int(vmcommon.Ok), successfulEvent.MultiOperation.ReturnCode)
	require.Equal(t, hex.EncodeToString(successfulInput.CurrentTxHash), successfulEvent.Context.CurrentTxHash)

	failedEvent := events[1]
	require.Equal(t, "multi_operation", failedEvent.Event)
	require.NotNil(t, failedEvent.MultiOperation)
	require.Equal(t, reDelegateRewards, failedEvent.MultiOperation.TargetFunction)
	require.Equal(t, encodeByteSlices(failedInput.Arguments), failedEvent.MultiOperation.Providers)
	require.Equal(t, int(vmcommon.UserError), failedEvent.MultiOperation.ReturnCode)
	require.Equal(t, hex.EncodeToString(failedInput.CurrentTxHash), failedEvent.Context.CurrentTxHash)
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
	const currentEpoch = uint32(15)
	const currentNonce = uint64(123)
	const currentRound = uint64(456)
	buffer := captureDelegationRewardsImportEvents(t)
	args := createMockArgumentsForDelegation()
	args.IsImportDBMode = true
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return currentEpoch
		},
		CurrentNonceCalled: func() uint64 {
			return currentNonce
		},
		CurrentRoundCalled: func() uint64 {
			return currentRound
		},
	}
	args.Eei = eei
	delegationContract, err := NewDelegationSystemSC(args)
	require.NoError(t, err)

	previousRaw := marshalRewardRecord(t, args, 10)
	followingRaw := marshalRewardRecord(t, args, 30)
	delegationContract.eei.SetStorage(rewardKeyForEpoch(currentEpoch), previousRaw)
	delegationContract.eei.SetStorage(rewardKeyForEpoch(currentEpoch+1), followingRaw)
	delegationContract.eei.SetStorage([]byte(totalActiveKey), big.NewInt(200).Bytes())
	delegationContract.eei.SetStorage([]byte(serviceFeeKey), big.NewInt(100).Bytes())
	input := getDefaultVmInputForFunc("updateRewards", nil)
	input.CallerAddr = vm.EndOfEpochAddress
	input.CallValue = big.NewInt(20)
	input.OriginalCallerAddr = []byte("original caller")
	input.OriginalTxHash = []byte("original hash")
	input.CurrentTxHash = []byte("current hash")
	input.PrevTxHash = []byte("previous hash")

	require.Equal(t, vmcommon.Ok, delegationContract.Execute(input))
	found, storedRecord, err := delegationContract.getRewardComputationData(currentEpoch)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, big.NewInt(20), storedRecord.RewardsToDistribute)
	require.Equal(t, big.NewInt(200), storedRecord.TotalActive)
	require.Equal(t, uint64(100), storedRecord.ServiceFee)

	events := decodeDelegationRewardsImportEvents(t, buffer)
	require.Len(t, events, 1)
	requireIncreasingDelegationRewardsImportSequences(t, events)
	event := events[0]
	require.Equal(t, "reward_write", event.Event)
	require.NotNil(t, event.RewardWrite)
	require.Nil(t, event.Operation)
	require.Equal(t, currentEpoch, event.Context.ObservedEpoch)
	require.Equal(t, currentNonce, event.Context.BlockNonce)
	require.Equal(t, currentRound, event.Context.BlockRound)
	require.Equal(t, "updateRewards", event.Context.Function)
	require.Equal(t, hex.EncodeToString(input.RecipientAddr), event.Context.Provider)
	require.Equal(t, hex.EncodeToString(input.CallerAddr), event.Context.Caller)
	require.Equal(t, hex.EncodeToString(input.OriginalCallerAddr), event.Context.OriginalCaller)
	require.Equal(t, "20", event.Context.CallValue)
	require.Equal(t, hex.EncodeToString(input.OriginalTxHash), event.Context.OriginalTxHash)
	require.Equal(t, hex.EncodeToString(input.CurrentTxHash), event.Context.CurrentTxHash)
	require.Equal(t, hex.EncodeToString(input.PrevTxHash), event.Context.PreviousTxHash)
	require.Equal(t, currentEpoch, event.RewardWrite.Epoch)
	require.Equal(t, hex.EncodeToString(rewardKeyForEpoch(currentEpoch)), event.RewardWrite.Key)
	require.Equal(t, hex.EncodeToString(previousRaw), event.RewardWrite.Previous.Raw)
	require.Equal(t, "10", event.RewardWrite.Previous.RewardsToDistribute)
	require.Equal(t, "20", event.RewardWrite.Current.RewardsToDistribute)
	require.Equal(t, "200", event.RewardWrite.Current.TotalActive)
	require.Equal(t, uint64(100), event.RewardWrite.Current.ServiceFee)
	require.Equal(t, currentEpoch+1, event.RewardWrite.FollowingSlot.Epoch)
	require.Equal(t, hex.EncodeToString(followingRaw), event.RewardWrite.FollowingSlot.Record.Raw)
	require.Equal(t, "30", event.RewardWrite.FollowingSlot.Record.RewardsToDistribute)
	require.Equal(t, "200", event.RewardWrite.ProviderState.TotalActive)
	require.Equal(t, uint64(100), event.RewardWrite.ProviderState.ServiceFee)
	require.Contains(t, buffer.String(), `"rewardWrite"`)
	require.Contains(t, buffer.String(), `"followingSlot"`)
}

func TestDelegationRewardsImportReportPreservesClaim(t *testing.T) {
	buffer := captureDelegationRewardsImportEvents(t)
	args := createMockArgumentsForDelegation()
	args.IsImportDBMode = true
	args.DelegationSCConfig.MaxServiceFee = 10000
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
		CurrentNonceCalled: func() uint64 {
			return 11
		},
		CurrentRoundCalled: func() uint64 {
			return 12
		},
		GetUserAccountCalled: func(_ []byte) (vmcommon.UserAccountHandler, error) {
			return nil, errors.New("account unavailable")
		},
	}
	args.Eei = eei
	delegationContract, err := NewDelegationSystemSC(args)
	require.NoError(t, err)

	input := getDefaultVmInputForFunc(claimRewards, nil)
	input.OriginalTxHash = []byte("original hash")
	input.CurrentTxHash = []byte("current hash")
	input.PrevTxHash = []byte("previous hash")
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

	events := decodeDelegationRewardsImportEvents(t, buffer)
	require.Len(t, events, 6)
	requireIncreasingDelegationRewardsImportSequences(t, events)
	for _, event := range events {
		require.Equal(t, uint32(2), event.Context.ObservedEpoch)
		require.Equal(t, uint64(11), event.Context.BlockNonce)
		require.Equal(t, uint64(12), event.Context.BlockRound)
		require.Equal(t, claimRewards, event.Context.Function)
		require.Equal(t, hex.EncodeToString(input.CurrentTxHash), event.Context.CurrentTxHash)
	}

	calculationEntries := delegationRewardsImportEventsByType(events, "calculation_entry")
	require.Len(t, calculationEntries, 3)
	for epoch, event := range calculationEntries {
		require.NotNil(t, event.CalculationEntry)
		require.Equal(t, uint32(epoch), event.CalculationEntry.Epoch)
	}
	require.True(t, calculationEntries[0].CalculationEntry.Record.Exists)
	require.Equal(t, "100", calculationEntries[0].CalculationEntry.Record.RewardsToDistribute)
	require.Equal(t, "90", calculationEntries[0].CalculationEntry.UserStakeShare)
	require.Equal(t, "90", calculationEntries[0].CalculationEntry.UserReward)
	for _, event := range calculationEntries[1:] {
		require.False(t, event.CalculationEntry.Record.Exists)
		require.Equal(t, "0", event.CalculationEntry.ProviderOwnerShare)
		require.Equal(t, "0", event.CalculationEntry.UserStakeShare)
		require.Equal(t, "0", event.CalculationEntry.UserReward)
	}

	calculationResults := delegationRewardsImportEventsByType(events, "calculation_result")
	require.Len(t, calculationResults, 1)
	calculationResult := calculationResults[0].CalculationResult
	require.NotNil(t, calculationResult)
	require.Equal(t, uint32(0), calculationResult.CheckpointBefore)
	require.Equal(t, uint32(3), calculationResult.CheckpointAfter)
	require.Equal(t, "0", calculationResult.UnclaimedRewardsBefore)
	require.Equal(t, "90", calculationResult.UnclaimedRewardsAfter)
	require.Equal(t, "90", calculationResult.TotalCalculatedRewards)
	require.Empty(t, calculationResult.CalculationError)

	amountEvents := delegationRewardsImportEventsByType(events, "claimed_rewards")
	require.Len(t, amountEvents, 1)
	require.Equal(t, "90", amountEvents[0].Amount.Value)
	require.False(t, amountEvents[0].Amount.Deleted)

	operationEvents := delegationRewardsImportEventsByType(events, "operation")
	require.Len(t, operationEvents, 1)
	operation := operationEvents[0].Operation
	require.NotNil(t, operation)
	require.Equal(t, int(vmcommon.Ok), operation.ReturnCode)
	require.Equal(t, "account unavailable", operation.Before.BalanceError)
	require.Equal(t, "account unavailable", operation.After.BalanceError)
	require.Len(t, operation.DelegatorsBefore, 1)
	require.Len(t, operation.DelegatorsAfter, 1)
	require.Equal(t, uint32(0), operation.DelegatorsBefore[0].RewardsCheckpoint)
	require.Equal(t, uint32(3), operation.DelegatorsAfter[0].RewardsCheckpoint)
	require.Equal(t, "90", operation.DelegatorsAfter[0].TotalCumulatedRewards)
	require.Contains(t, buffer.String(), `"calculationEntry"`)
	require.Contains(t, buffer.String(), `"calculationResult"`)
}

func TestDelegationRewardsImportReportRecordsCalculationFailure(t *testing.T) {
	buffer := captureDelegationRewardsImportEvents(t)
	args := createMockArgumentsForDelegation()
	args.IsImportDBMode = true
	args.DelegationSCConfig.MaxServiceFee = 10000
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 0
		},
	}
	args.Eei = eei
	delegationContract, err := NewDelegationSystemSC(args)
	require.NoError(t, err)

	input := getDefaultVmInputForFunc(claimRewards, nil)
	input.CurrentTxHash = []byte("failed claim")
	fundKey := []byte{1}
	require.NoError(t, delegationContract.saveDelegatorData(input.CallerAddr, &DelegatorData{
		ActiveFund:            fundKey,
		RewardsCheckpoint:     0,
		UnClaimedRewards:      big.NewInt(7),
		TotalCumulatedRewards: big.NewInt(11),
	}))
	require.NoError(t, delegationContract.saveFund(fundKey, &Fund{Value: big.NewInt(1000)}))
	malformedRewardRecord := []byte("malformed reward record")
	malformedGlobalFund := []byte("malformed global fund")
	delegationContract.eei.SetStorage(rewardKeyForEpoch(0), malformedRewardRecord)
	delegationContract.eei.SetStorage([]byte(globalFundKey), malformedGlobalFund)

	require.Equal(t, vmcommon.UserError, delegationContract.Execute(input))

	events := decodeDelegationRewardsImportEvents(t, buffer)
	require.Len(t, events, 3)
	requireIncreasingDelegationRewardsImportSequences(t, events)
	require.Empty(t, delegationRewardsImportEventsByType(events, "claimed_rewards"))

	calculationEntries := delegationRewardsImportEventsByType(events, "calculation_entry")
	require.Len(t, calculationEntries, 1)
	entry := calculationEntries[0].CalculationEntry
	require.NotNil(t, entry)
	require.True(t, entry.Record.Exists)
	require.Equal(t, hex.EncodeToString(malformedRewardRecord), entry.Record.Raw)
	require.NotEmpty(t, entry.Record.DecodeError)

	calculationResults := delegationRewardsImportEventsByType(events, "calculation_result")
	require.Len(t, calculationResults, 1)
	result := calculationResults[0].CalculationResult
	require.NotNil(t, result)
	require.NotEmpty(t, result.CalculationError)
	require.Equal(t, uint32(0), result.CheckpointBefore)
	require.Equal(t, uint32(0), result.CheckpointAfter)
	require.Equal(t, "7", result.UnclaimedRewardsBefore)
	require.Equal(t, "7", result.UnclaimedRewardsAfter)
	require.Empty(t, result.TotalCalculatedRewards)

	operationEvents := delegationRewardsImportEventsByType(events, "operation")
	require.Len(t, operationEvents, 1)
	operation := operationEvents[0].Operation
	require.Equal(t, int(vmcommon.UserError), operation.ReturnCode)
	require.Equal(t, hex.EncodeToString(malformedGlobalFund), operation.Before.GlobalFundRaw)
	require.NotEmpty(t, operation.Before.GlobalFundError)
	require.Equal(t, hex.EncodeToString(malformedGlobalFund), operation.After.GlobalFundRaw)
	require.NotEmpty(t, operation.After.GlobalFundError)
	require.Equal(t, uint32(0), operation.DelegatorsAfter[0].RewardsCheckpoint)
	require.Equal(t, "7", operation.DelegatorsAfter[0].UnclaimedRewards)
}

func TestDelegationRewardsImportReportPreservesMalformedDelegator(t *testing.T) {
	buffer := captureDelegationRewardsImportEvents(t)
	args := createMockArgumentsForDelegation()
	args.IsImportDBMode = true
	eei := createDefaultEei()
	args.Eei = eei
	delegationContract, err := NewDelegationSystemSC(args)
	require.NoError(t, err)

	input := getDefaultVmInputForFunc(claimRewards, nil)
	malformedDelegator := []byte("malformed delegator")
	delegationContract.eei.SetStorage(input.CallerAddr, malformedDelegator)

	require.Equal(t, vmcommon.UserError, delegationContract.Execute(input))

	events := decodeDelegationRewardsImportEvents(t, buffer)
	require.Len(t, events, 1)
	requireIncreasingDelegationRewardsImportSequences(t, events)
	operation := events[0].Operation
	require.NotNil(t, operation)
	require.Equal(t, int(vmcommon.UserError), operation.ReturnCode)
	require.Len(t, operation.DelegatorsBefore, 1)
	require.Equal(t, hex.EncodeToString(malformedDelegator), operation.DelegatorsBefore[0].Raw)
	require.NotEmpty(t, operation.DelegatorsBefore[0].DecodeError)
	require.Len(t, operation.DelegatorsAfter, 1)
	require.Equal(t, hex.EncodeToString(malformedDelegator), operation.DelegatorsAfter[0].Raw)
	require.NotEmpty(t, operation.DelegatorsAfter[0].DecodeError)
}

func TestDelegationRewardsImportReportDoesNotRecordReadOnlyCalculation(t *testing.T) {
	buffer := captureDelegationRewardsImportEvents(t)
	args := createMockArgumentsForDelegation()
	args.IsImportDBMode = true
	eei := createDefaultEei()
	args.Eei = eei
	delegationContract, err := NewDelegationSystemSC(args)
	require.NoError(t, err)

	input := getDefaultVmInputForFunc("getClaimableRewards", [][]byte{[]byte("missing delegator")})
	require.Equal(t, vmcommon.UserError, delegationContract.Execute(input))
	require.Empty(t, buffer.String())
}

func TestDelegationRewardsImportReportRecordsDeletedDelegator(t *testing.T) {
	buffer := captureDelegationRewardsImportEvents(t)
	args := createMockArgumentsForDelegation()
	args.IsImportDBMode = true
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 10
		},
	}
	args.Eei = eei
	delegationContract, err := NewDelegationSystemSC(args)
	require.NoError(t, err)

	input := getDefaultVmInputForFunc(claimRewards, nil)
	require.NoError(t, delegationContract.saveDelegatorData(input.CallerAddr, &DelegatorData{
		RewardsCheckpoint:     0,
		UnClaimedRewards:      big.NewInt(135),
		TotalCumulatedRewards: big.NewInt(0),
	}))
	require.NoError(t, delegationContract.saveDelegationStatus(&DelegationContractStatus{NumUsers: 10}))

	require.Equal(t, vmcommon.Ok, delegationContract.Execute(input))

	events := decodeDelegationRewardsImportEvents(t, buffer)
	requireIncreasingDelegationRewardsImportSequences(t, events)
	amountEvents := delegationRewardsImportEventsByType(events, "claimed_rewards")
	require.Len(t, amountEvents, 1)
	require.Equal(t, "135", amountEvents[0].Amount.Value)
	require.True(t, amountEvents[0].Amount.Deleted)

	operationEvents := delegationRewardsImportEventsByType(events, "operation")
	require.Len(t, operationEvents, 1)
	operation := operationEvents[0].Operation
	require.Len(t, operation.DelegatorsBefore, 1)
	require.True(t, operation.DelegatorsBefore[0].Exists)
	require.Len(t, operation.DelegatorsAfter, 1)
	require.False(t, operation.DelegatorsAfter[0].Exists)
}

func TestDelegationRewardsImportReportRecordsRedelegatedAmount(t *testing.T) {
	delegationContract, _ := prepareReDelegateRewardsComponents(t, 1000, big.NewInt(1156))
	delegationContract.isImportDBMode = true
	buffer := captureDelegationRewardsImportEvents(t)
	input := getDefaultVmInputForFunc(reDelegateRewards, nil)
	input.CallerAddr = []byte("stakingProvider")

	require.Equal(t, vmcommon.Ok, delegationContract.Execute(input))

	events := decodeDelegationRewardsImportEvents(t, buffer)
	requireIncreasingDelegationRewardsImportSequences(t, events)
	amountEvents := delegationRewardsImportEventsByType(events, "redelegated_rewards")
	require.Len(t, amountEvents, 1)
	require.Equal(t, "155", amountEvents[0].Amount.Value)
	require.Equal(t, hex.EncodeToString(input.CallerAddr), amountEvents[0].Amount.User)
	require.Empty(t, amountEvents[0].Amount.Recipient)
}

func TestDelegationRewardsImportReportRecordsWithdrawnAmount(t *testing.T) {
	buffer := captureDelegationRewardsImportEvents(t)
	args := createMockArgumentsForDelegation()
	args.IsImportDBMode = true
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 60
		},
	}
	args.Eei = eei
	addValidatorAndStakingScToVmContext(eei)
	delegationContract, err := NewDelegationSystemSC(args)
	require.NoError(t, err)

	input := getDefaultVmInputForFunc(withdraw, nil)
	fundKey := []byte{1}
	require.NoError(t, delegationContract.saveDelegatorData(input.CallerAddr, &DelegatorData{
		UnStakedFunds:         [][]byte{fundKey},
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	}))
	require.NoError(t, delegationContract.saveFund(fundKey, &Fund{
		Value:   big.NewInt(60),
		Address: input.CallerAddr,
		Epoch:   10,
		Type:    unStaked,
	}))
	require.NoError(t, delegationContract.saveDelegationContractConfig(&DelegationConfig{
		UnBondPeriodInEpochs: 50,
	}))
	require.NoError(t, delegationContract.saveGlobalFundData(&GlobalFundData{
		TotalUnStaked: big.NewInt(60),
		TotalActive:   big.NewInt(0),
	}))
	require.NoError(t, delegationContract.saveDelegationStatus(&DelegationContractStatus{NumUsers: 1}))

	require.Equal(t, vmcommon.Ok, delegationContract.Execute(input))

	events := decodeDelegationRewardsImportEvents(t, buffer)
	requireIncreasingDelegationRewardsImportSequences(t, events)
	amountEvents := delegationRewardsImportEventsByType(events, "withdrawn_funds")
	require.Len(t, amountEvents, 1)
	require.Equal(t, "60", amountEvents[0].Amount.Value)
	require.True(t, amountEvents[0].Amount.Deleted)
	operationEvents := delegationRewardsImportEventsByType(events, "operation")
	require.Len(t, operationEvents, 1)
	require.True(t, operationEvents[0].Operation.DelegatorsBefore[0].Exists)
	require.False(t, operationEvents[0].Operation.DelegatorsAfter[0].Exists)
}
