package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"sync/atomic"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const delegationRewardsImportReportVersion = 2

var logDelegationRewardsImport = logger.GetOrCreate("vm/systemsmartcontracts/rewardsimport")
var delegationRewardsImportSequence atomic.Uint64

type delegationRewardsImportEvent struct {
	Version           uint32                              `json:"version"`
	Sequence          uint64                              `json:"sequence"`
	Event             string                              `json:"event"`
	Context           delegationRewardsImportContext      `json:"context"`
	Operation         *delegationRewardsOperation         `json:"operation,omitempty"`
	RewardWrite       *delegationRewardWrite              `json:"rewardWrite,omitempty"`
	CalculationEntry  *delegationRewardCalculationEntry   `json:"calculationEntry,omitempty"`
	CalculationResult *delegationRewardCalculationResult  `json:"calculationResult,omitempty"`
	Amount            *delegationRewardsImportAmountEvent `json:"amount,omitempty"`
	MultiOperation    *delegationRewardsMultiOperation    `json:"multiOperation,omitempty"`
	ManagerOperation  *delegationRewardsManagerOperation  `json:"managerOperation,omitempty"`
}

type delegationRewardsImportContext struct {
	ObservedEpoch  uint32   `json:"observedEpoch"`
	BlockNonce     uint64   `json:"blockNonce"`
	BlockRound     uint64   `json:"blockRound"`
	Function       string   `json:"function"`
	Provider       string   `json:"provider"`
	Caller         string   `json:"caller"`
	OriginalCaller string   `json:"originalCaller,omitempty"`
	CallValue      string   `json:"callValue"`
	OriginalTxHash string   `json:"originalTxHash,omitempty"`
	CurrentTxHash  string   `json:"currentTxHash,omitempty"`
	PreviousTxHash string   `json:"previousTxHash,omitempty"`
	CallType       int      `json:"callType"`
	Arguments      []string `json:"arguments,omitempty"`
}

type delegationRewardsOperation struct {
	ReturnCode       int                               `json:"returnCode"`
	ReturnCodeName   string                            `json:"returnCodeName"`
	Before           delegationRewardsProviderSnapshot `json:"before"`
	After            delegationRewardsProviderSnapshot `json:"after"`
	DelegatorsBefore []delegationRewardsUserSnapshot   `json:"delegatorsBefore,omitempty"`
	DelegatorsAfter  []delegationRewardsUserSnapshot   `json:"delegatorsAfter,omitempty"`
}

type delegationRewardsProviderSnapshot struct {
	Owner                     string                       `json:"owner,omitempty"`
	PersistedBalance          string                       `json:"persistedBalance"`
	BalanceError              string                       `json:"balanceError,omitempty"`
	HasOutputAccount          bool                         `json:"hasOutputAccount"`
	OutputAccountBalance      string                       `json:"outputAccountBalance,omitempty"`
	OutputAccountBalanceDelta string                       `json:"outputAccountBalanceDelta,omitempty"`
	TotalActive               string                       `json:"totalActive"`
	ServiceFee                uint64                       `json:"serviceFee"`
	GlobalFundRaw             string                       `json:"globalFundRaw,omitempty"`
	GlobalFund                *delegationRewardsGlobalFund `json:"globalFund,omitempty"`
	GlobalFundError           string                       `json:"globalFundError,omitempty"`
}

type delegationRewardsGlobalFund struct {
	TotalActive   string `json:"totalActive"`
	TotalUnstaked string `json:"totalUnstaked"`
}

type delegationRewardsUserSnapshot struct {
	Address               string                         `json:"address"`
	Exists                bool                           `json:"exists"`
	Raw                   string                         `json:"raw,omitempty"`
	DecodeError           string                         `json:"decodeError,omitempty"`
	RewardsCheckpoint     uint32                         `json:"rewardsCheckpoint"`
	UnclaimedRewards      string                         `json:"unclaimedRewards"`
	TotalCumulatedRewards string                         `json:"totalCumulatedRewards"`
	ActiveFundKey         string                         `json:"activeFundKey,omitempty"`
	ActiveFund            *delegationRewardsFundSnapshot `json:"activeFund,omitempty"`
	ActiveFundError       string                         `json:"activeFundError,omitempty"`
	UnstakedFundKeys      []string                       `json:"unstakedFundKeys,omitempty"`
}

type delegationRewardsFundSnapshot struct {
	Raw     string `json:"raw"`
	Value   string `json:"value"`
	Address string `json:"address,omitempty"`
	Epoch   uint32 `json:"epoch"`
	Type    uint32 `json:"type"`
}

type delegationRewardWrite struct {
	Epoch         uint32                            `json:"epoch"`
	Key           string                            `json:"key"`
	Previous      delegationRewardRecordSnapshot    `json:"previous"`
	Current       delegationRewardRecordSnapshot    `json:"current"`
	FollowingSlot delegationRewardSlotSnapshot      `json:"followingSlot"`
	ProviderState delegationRewardsProviderSnapshot `json:"providerState"`
}

type delegationRewardSlotSnapshot struct {
	Epoch  uint32                         `json:"epoch"`
	Key    string                         `json:"key"`
	Record delegationRewardRecordSnapshot `json:"record"`
}

type delegationRewardRecordSnapshot struct {
	Exists              bool   `json:"exists"`
	Raw                 string `json:"raw,omitempty"`
	DecodeError         string `json:"decodeError,omitempty"`
	RewardsToDistribute string `json:"rewardsToDistribute"`
	TotalActive         string `json:"totalActive"`
	ServiceFee          uint64 `json:"serviceFee"`
}

type delegationRewardCalculationEntry struct {
	User                  string                         `json:"user"`
	Epoch                 uint32                         `json:"epoch"`
	Record                delegationRewardRecordSnapshot `json:"record"`
	ActiveValue           string                         `json:"activeValue"`
	IsOwner               bool                           `json:"isOwner"`
	ProviderOwnerShare    string                         `json:"providerOwnerShare"`
	UserStakeShare        string                         `json:"userStakeShare"`
	UserReward            string                         `json:"userReward"`
	ServiceFeeDenominator uint64                         `json:"serviceFeeDenominator"`
	UsesTrimmedPercentage bool                           `json:"usesTrimmedPercentage"`
}

type delegationRewardCalculationResult struct {
	User                   string `json:"user"`
	CurrentEpoch           uint32 `json:"currentEpoch"`
	CheckpointBefore       uint32 `json:"checkpointBefore"`
	CheckpointAfter        uint32 `json:"checkpointAfter"`
	UnclaimedRewardsBefore string `json:"unclaimedRewardsBefore"`
	UnclaimedRewardsAfter  string `json:"unclaimedRewardsAfter"`
	TotalCalculatedRewards string `json:"totalCalculatedRewards"`
	CalculationError       string `json:"calculationError,omitempty"`
}

type delegationRewardsImportAmountEvent struct {
	User      string `json:"user"`
	Recipient string `json:"recipient,omitempty"`
	Value     string `json:"value"`
	Deleted   bool   `json:"deleted,omitempty"`
}

type delegationRewardsMultiOperation struct {
	TargetFunction string   `json:"targetFunction"`
	Providers      []string `json:"providers"`
	ReturnCode     int      `json:"returnCode"`
	ReturnCodeName string   `json:"returnCodeName"`
}

type delegationRewardsManagerOperation struct {
	Providers      []string `json:"providers,omitempty"`
	ReturnCode     int      `json:"returnCode"`
	ReturnCodeName string   `json:"returnCodeName"`
}

type delegationRewardsOperationCapture struct {
	context       delegationRewardsImportContext
	provider      []byte
	userAddresses [][]byte
	providerState delegationRewardsProviderSnapshot
	delegators    []delegationRewardsUserSnapshot
}

type delegationRewardsManagerOperationCapture struct {
	context   delegationRewardsImportContext
	providers [][]byte
}

func (d *delegation) startDelegationRewardsOperationReport(args *vmcommon.ContractCallInput) *delegationRewardsOperationCapture {
	if !d.isImportDBMode || !isDelegationRewardsOperation(args.Function) {
		return nil
	}

	userAddresses := delegationRewardsOperationUsers(args)
	return &delegationRewardsOperationCapture{
		context:       d.delegationRewardsReportContext(args, true),
		provider:      append([]byte(nil), args.RecipientAddr...),
		userAddresses: userAddresses,
		providerState: d.delegationRewardsProviderSnapshot(args.RecipientAddr),
		delegators:    d.delegationRewardsUserSnapshots(userAddresses),
	}
}

func (d *delegation) finishDelegationRewardsOperationReport(
	capture *delegationRewardsOperationCapture,
	returnCode vmcommon.ReturnCode,
) {
	if capture == nil {
		return
	}

	logDelegationRewardsImportEvent(delegationRewardsImportEvent{
		Event:   "operation",
		Context: capture.context,
		Operation: &delegationRewardsOperation{
			ReturnCode:       int(returnCode),
			ReturnCodeName:   returnCode.String(),
			Before:           capture.providerState,
			After:            d.delegationRewardsProviderSnapshot(capture.provider),
			DelegatorsBefore: capture.delegators,
			DelegatorsAfter:  d.delegationRewardsUserSnapshots(capture.userAddresses),
		},
	})
}

func (d *delegation) reportDelegationRewardWrite(
	args *vmcommon.ContractCallInput,
	epoch uint32,
	previousRaw []byte,
	followingRaw []byte,
) {
	if !d.isImportDBMode {
		return
	}

	logDelegationRewardsImportEvent(delegationRewardsImportEvent{
		Event:       "reward_write",
		Context:     d.delegationRewardsReportContext(args, false),
		RewardWrite: d.newDelegationRewardWrite(args.RecipientAddr, epoch, previousRaw, followingRaw),
	})
}

func (d *delegation) newDelegationRewardWrite(
	provider []byte,
	epoch uint32,
	previousRaw []byte,
	followingRaw []byte,
) *delegationRewardWrite {
	key := rewardKeyForEpoch(epoch)
	followingEpoch := epoch + 1
	followingKey := rewardKeyForEpoch(followingEpoch)

	return &delegationRewardWrite{
		Epoch:    epoch,
		Key:      hex.EncodeToString(key),
		Previous: d.delegationRewardRecordSnapshot(previousRaw),
		Current:  d.delegationRewardRecordSnapshot(d.eei.GetStorage(key)),
		FollowingSlot: delegationRewardSlotSnapshot{
			Epoch:  followingEpoch,
			Key:    hex.EncodeToString(followingKey),
			Record: d.delegationRewardRecordSnapshot(followingRaw),
		},
		ProviderState: d.delegationRewardsProviderSnapshot(provider),
	}
}

func (d *delegation) reportDelegationRewardCalculationEntry(
	args *vmcommon.ContractCallInput,
	user []byte,
	epoch uint32,
	recordRaw []byte,
	activeValue *big.Int,
	isOwner bool,
	providerOwnerShare *big.Int,
	userStakeShare *big.Int,
	userReward *big.Int,
) {
	if !d.shouldReportDelegationRewardCalculation(args) {
		return
	}

	logDelegationRewardsImportEvent(delegationRewardsImportEvent{
		Event:   "calculation_entry",
		Context: d.delegationRewardsReportContext(args, false),
		CalculationEntry: d.newDelegationRewardCalculationEntry(
			user,
			epoch,
			recordRaw,
			activeValue,
			isOwner,
			providerOwnerShare,
			userStakeShare,
			userReward,
		),
	})
}

func (d *delegation) newDelegationRewardCalculationEntry(
	user []byte,
	epoch uint32,
	recordRaw []byte,
	activeValue *big.Int,
	isOwner bool,
	providerOwnerShare *big.Int,
	userStakeShare *big.Int,
	userReward *big.Int,
) *delegationRewardCalculationEntry {
	return &delegationRewardCalculationEntry{
		User:                  hex.EncodeToString(user),
		Epoch:                 epoch,
		Record:                d.delegationRewardRecordSnapshot(recordRaw),
		ActiveValue:           bigIntString(activeValue),
		IsOwner:               isOwner,
		ProviderOwnerShare:    bigIntString(providerOwnerShare),
		UserStakeShare:        bigIntString(userStakeShare),
		UserReward:            bigIntString(userReward),
		ServiceFeeDenominator: d.maxServiceFee,
		UsesTrimmedPercentage: d.enableEpochsHandler.IsFlagEnabled(common.StakingV2FlagAfterEpoch),
	}
}

func (d *delegation) reportDelegationRewardCalculationResult(
	args *vmcommon.ContractCallInput,
	user []byte,
	currentEpoch uint32,
	checkpointBefore uint32,
	unclaimedBefore string,
	totalRewards *big.Int,
	delegator *DelegatorData,
	err error,
) {
	if !d.shouldReportDelegationRewardCalculation(args) {
		return
	}

	calculationError := ""
	if err != nil {
		calculationError = err.Error()
	}

	logDelegationRewardsImportEvent(delegationRewardsImportEvent{
		Event:   "calculation_result",
		Context: d.delegationRewardsReportContext(args, false),
		CalculationResult: &delegationRewardCalculationResult{
			User:                   hex.EncodeToString(user),
			CurrentEpoch:           currentEpoch,
			CheckpointBefore:       checkpointBefore,
			CheckpointAfter:        delegator.RewardsCheckpoint,
			UnclaimedRewardsBefore: unclaimedBefore,
			UnclaimedRewardsAfter:  bigIntString(delegator.UnClaimedRewards),
			TotalCalculatedRewards: bigIntString(totalRewards),
			CalculationError:       calculationError,
		},
	})
}

func (d *delegation) reportDelegationRewardsAmount(
	event string,
	args *vmcommon.ContractCallInput,
	user []byte,
	recipient []byte,
	value *big.Int,
	deleted bool,
) {
	if !d.isImportDBMode {
		return
	}

	logDelegationRewardsImportEvent(delegationRewardsImportEvent{
		Event:   event,
		Context: d.delegationRewardsReportContext(args, false),
		Amount: &delegationRewardsImportAmountEvent{
			User:      hex.EncodeToString(user),
			Recipient: hex.EncodeToString(recipient),
			Value:     bigIntString(value),
			Deleted:   deleted,
		},
	})
}

func logDelegationRewardsImportEvent(event delegationRewardsImportEvent) {
	event.Version = delegationRewardsImportReportVersion
	event.Sequence = delegationRewardsImportSequence.Add(1)
	payload, err := json.Marshal(event)
	if err != nil {
		logDelegationRewardsImport.Error("cannot encode delegation rewards import report", "error", err)
		return
	}

	logDelegationRewardsImport.Info("delegation rewards import report", "payload", string(payload))
}

func (d *delegation) delegationRewardsReportContext(
	args *vmcommon.ContractCallInput,
	includeArguments bool,
) delegationRewardsImportContext {
	return newDelegationRewardsReportContext(d.eei, args, includeArguments)
}

func newDelegationRewardsReportContext(
	eei vm.SystemEI,
	args *vmcommon.ContractCallInput,
	includeArguments bool,
) delegationRewardsImportContext {
	context := delegationRewardsImportContext{
		ObservedEpoch:  eei.BlockChainHook().CurrentEpoch(),
		BlockNonce:     eei.BlockChainHook().CurrentNonce(),
		BlockRound:     eei.BlockChainHook().CurrentRound(),
		Function:       args.Function,
		Provider:       hex.EncodeToString(args.RecipientAddr),
		Caller:         hex.EncodeToString(args.CallerAddr),
		OriginalCaller: hex.EncodeToString(args.OriginalCallerAddr),
		CallValue:      bigIntString(args.CallValue),
		OriginalTxHash: hex.EncodeToString(args.OriginalTxHash),
		CurrentTxHash:  hex.EncodeToString(args.CurrentTxHash),
		PreviousTxHash: hex.EncodeToString(args.PrevTxHash),
		CallType:       int(args.CallType),
	}
	if includeArguments {
		context.Arguments = encodeByteSlices(args.Arguments)
	}

	return context
}

func (d *delegationManager) reportDelegationRewardsMultiOperation(
	args *vmcommon.ContractCallInput,
	targetFunction string,
	returnCode vmcommon.ReturnCode,
) {
	logDelegationRewardsImportEvent(delegationRewardsImportEvent{
		Event:   "multi_operation",
		Context: newDelegationRewardsReportContext(d.eei, args, true),
		MultiOperation: &delegationRewardsMultiOperation{
			TargetFunction: targetFunction,
			Providers:      encodeByteSlices(args.Arguments),
			ReturnCode:     int(returnCode),
			ReturnCodeName: returnCode.String(),
		},
	})
}

func (d *delegationManager) startDelegationRewardsManagerOperationReport(
	args *vmcommon.ContractCallInput,
) *delegationRewardsManagerOperationCapture {
	if !d.isImportDBMode || !isDelegationRewardsManagerOperation(args.Function) {
		return nil
	}

	return &delegationRewardsManagerOperationCapture{
		context:   newDelegationRewardsReportContext(d.eei, args, true),
		providers: d.delegationRewardsManagerOperationProviders(args),
	}
}

func (d *delegationManager) finishDelegationRewardsManagerOperationReport(
	capture *delegationRewardsManagerOperationCapture,
	returnCode vmcommon.ReturnCode,
) {
	if capture == nil {
		return
	}

	logDelegationRewardsImportEvent(d.newDelegationRewardsManagerOperationEvent(capture, returnCode))
}

func (d *delegationManager) newDelegationRewardsManagerOperationEvent(
	capture *delegationRewardsManagerOperationCapture,
	returnCode vmcommon.ReturnCode,
) delegationRewardsImportEvent {
	return delegationRewardsImportEvent{
		Event:   "manager_operation",
		Context: capture.context,
		ManagerOperation: &delegationRewardsManagerOperation{
			Providers:      encodeByteSlices(capture.providers),
			ReturnCode:     int(returnCode),
			ReturnCodeName: returnCode.String(),
		},
	}
}

func (d *delegationManager) delegationRewardsManagerOperationProviders(
	args *vmcommon.ContractCallInput,
) [][]byte {
	switch args.Function {
	case "createNewDelegationContract", "makeNewContractFromValidatorData":
		managementData, err := d.getDelegationManagementData()
		if err != nil {
			return nil
		}

		return [][]byte{createNewAddress(managementData.LastAddress)}
	case "mergeValidatorToDelegationSameOwner", "mergeValidatorToDelegationWithWhitelist":
		if len(args.Arguments) == 0 {
			return nil
		}

		return [][]byte{bytes.Clone(args.Arguments[0])}
	default:
		return nil
	}
}

func (d *delegation) delegationRewardsProviderSnapshot(provider []byte) delegationRewardsProviderSnapshot {
	globalFundRaw := d.eei.GetStorage([]byte(globalFundKey))
	snapshot := delegationRewardsProviderSnapshot{
		Owner:         hex.EncodeToString(d.eei.GetStorage([]byte(ownerKey))),
		TotalActive:   big.NewInt(0).SetBytes(d.eei.GetStorage([]byte(totalActiveKey))).String(),
		ServiceFee:    big.NewInt(0).SetBytes(d.eei.GetStorage([]byte(serviceFeeKey))).Uint64(),
		GlobalFundRaw: hex.EncodeToString(globalFundRaw),
	}
	snapshot.HasOutputAccount, snapshot.OutputAccountBalance, snapshot.OutputAccountBalanceDelta =
		delegationRewardsOutputAccountSnapshot(d.eei, provider)
	providerAccount, err := d.eei.BlockChainHook().GetUserAccount(provider)
	if err != nil {
		snapshot.BalanceError = err.Error()
	} else if check.IfNil(providerAccount) {
		snapshot.BalanceError = "nil provider account"
	} else {
		snapshot.PersistedBalance = bigIntString(providerAccount.GetBalance())
	}

	if len(globalFundRaw) > 0 {
		globalFund := &GlobalFundData{}
		err = d.marshalizer.Unmarshal(globalFund, globalFundRaw)
		if err != nil {
			snapshot.GlobalFundError = err.Error()
		} else {
			snapshot.GlobalFund = &delegationRewardsGlobalFund{
				TotalActive:   bigIntString(globalFund.TotalActive),
				TotalUnstaked: bigIntString(globalFund.TotalUnStaked),
			}
		}
	}

	return snapshot
}

func delegationRewardsOutputAccountSnapshot(
	eei vm.SystemEI,
	provider []byte,
) (bool, string, string) {
	context, ok := eei.(*vmContext)
	if !ok {
		return false, "", ""
	}

	outputAccount, exists := context.outputAccounts[string(provider)]
	if !exists || outputAccount == nil {
		return false, "", ""
	}

	return true, bigIntString(outputAccount.Balance), bigIntString(outputAccount.BalanceDelta)
}

func (d *delegation) delegationRewardsUserSnapshots(addresses [][]byte) []delegationRewardsUserSnapshot {
	snapshots := make([]delegationRewardsUserSnapshot, 0, len(addresses))
	for _, address := range addresses {
		snapshots = append(snapshots, d.delegationRewardsUserSnapshot(address))
	}

	return snapshots
}

func (d *delegation) delegationRewardsUserSnapshot(address []byte) delegationRewardsUserSnapshot {
	raw := d.eei.GetStorage(address)
	snapshot := delegationRewardsUserSnapshot{
		Address: hex.EncodeToString(address),
		Exists:  len(raw) > 0,
		Raw:     hex.EncodeToString(raw),
	}
	if len(raw) == 0 {
		return snapshot
	}

	delegator := &DelegatorData{}
	err := d.marshalizer.Unmarshal(delegator, raw)
	if err != nil {
		snapshot.DecodeError = err.Error()
		return snapshot
	}

	snapshot.RewardsCheckpoint = delegator.RewardsCheckpoint
	snapshot.UnclaimedRewards = bigIntString(delegator.UnClaimedRewards)
	snapshot.TotalCumulatedRewards = bigIntString(delegator.TotalCumulatedRewards)
	snapshot.ActiveFundKey = hex.EncodeToString(delegator.ActiveFund)
	snapshot.UnstakedFundKeys = encodeByteSlices(delegator.UnStakedFunds)
	if len(delegator.ActiveFund) == 0 {
		return snapshot
	}

	activeFundRaw := d.eei.GetStorage(delegator.ActiveFund)
	activeFund := &Fund{}
	err = d.marshalizer.Unmarshal(activeFund, activeFundRaw)
	if err != nil {
		snapshot.ActiveFundError = err.Error()
		return snapshot
	}

	snapshot.ActiveFund = &delegationRewardsFundSnapshot{
		Raw:     hex.EncodeToString(activeFundRaw),
		Value:   bigIntString(activeFund.Value),
		Address: hex.EncodeToString(activeFund.Address),
		Epoch:   activeFund.Epoch,
		Type:    activeFund.Type,
	}

	return snapshot
}

func (d *delegation) delegationRewardRecordSnapshot(raw []byte) delegationRewardRecordSnapshot {
	snapshot := delegationRewardRecordSnapshot{
		Exists: len(raw) > 0,
		Raw:    hex.EncodeToString(raw),
	}
	if len(raw) == 0 {
		return snapshot
	}

	record := &RewardComputationData{}
	err := d.marshalizer.Unmarshal(record, raw)
	if err != nil {
		snapshot.DecodeError = err.Error()
		return snapshot
	}

	snapshot.RewardsToDistribute = bigIntString(record.RewardsToDistribute)
	snapshot.TotalActive = bigIntString(record.TotalActive)
	snapshot.ServiceFee = record.ServiceFee
	return snapshot
}

func (d *delegation) shouldReportDelegationRewardCalculation(args *vmcommon.ContractCallInput) bool {
	return d.isImportDBMode && isDelegationRewardsOperation(args.Function)
}

func isDelegationRewardsOperation(function string) bool {
	switch function {
	case core.SCDeployInitFunctionName,
		initFromValidatorData,
		mergeValidatorDataToDelegation,
		delegate,
		"unDelegate",
		withdraw,
		claimRewards,
		reDelegateRewards,
		changeOwner:
		return true
	default:
		return false
	}
}

func isDelegationRewardsManagerOperation(function string) bool {
	switch function {
	case "createNewDelegationContract",
		"makeNewContractFromValidatorData",
		"mergeValidatorToDelegationSameOwner",
		"mergeValidatorToDelegationWithWhitelist":
		return true
	default:
		return false
	}
}

func delegationRewardsOperationUsers(args *vmcommon.ContractCallInput) [][]byte {
	addresses := make([][]byte, 0, 2)
	addAddress := func(address []byte) {
		if len(address) == 0 {
			return
		}
		for _, existing := range addresses {
			if bytes.Equal(existing, address) {
				return
			}
		}
		addresses = append(addresses, append([]byte(nil), address...))
	}

	switch args.Function {
	case initFromValidatorData, mergeValidatorDataToDelegation:
		if len(args.Arguments) > 0 {
			addAddress(args.Arguments[0])
		}
	case changeOwner:
		addAddress(args.CallerAddr)
		if len(args.Arguments) > 0 {
			addAddress(args.Arguments[0])
		}
	default:
		addAddress(args.CallerAddr)
	}

	return addresses
}

func encodeByteSlices(values [][]byte) []string {
	encoded := make([]string, 0, len(values))
	for _, value := range values {
		encoded = append(encoded, hex.EncodeToString(value))
	}

	return encoded
}

func bigIntString(value *big.Int) string {
	if value == nil {
		return ""
	}

	return value.String()
}
