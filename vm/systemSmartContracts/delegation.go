//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. delegation.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/vm"
)

const delegationConfigKey = "delegationConfig"
const delegationStatusKey = "delegationStatus"
const delegationMetaData = "delegationMetaData"
const lastFundKey = "lastFund"
const globalFundKey = "globalFund"
const serviceFeeKey = "serviceFee"
const totalActiveKey = "totalActive"
const rewardKeyPrefix = "reward"
const fundKeyPrefix = "fund"

const (
	active   = uint32(0)
	unStaked = uint32(1)
)

type delegation struct {
	eei                    vm.SystemEI
	sigVerifier            vm.MessageSignVerifier
	delegationMgrSCAddress []byte
	stakingSCAddr          []byte
	validatorSCAddr        []byte
	endOfEpochAddr         []byte
	gasCost                vm.GasCost
	marshalizer            marshal.Marshalizer
	delegationEnabled      atomic.Flag
	enableDelegationEpoch  uint32
	minServiceFee          uint64
	maxServiceFee          uint64
	unBondPeriod           uint64
	nodePrice              *big.Int
	unJailPrice            *big.Int
	minStakeValue          *big.Int
	mutExecution           sync.RWMutex
	stakingV2EnableEpoch   uint32
	stakingV2Enabled       atomic.Flag
}

// ArgsNewDelegation defines the arguments to create the delegation smart contract
type ArgsNewDelegation struct {
	DelegationSCConfig     config.DelegationSystemSCConfig
	StakingSCConfig        config.StakingSystemSCConfig
	Eei                    vm.SystemEI
	SigVerifier            vm.MessageSignVerifier
	DelegationMgrSCAddress []byte
	StakingSCAddress       []byte
	ValidatorSCAddress     []byte
	EndOfEpochAddress      []byte
	GasCost                vm.GasCost
	Marshalizer            marshal.Marshalizer
	EpochNotifier          vm.EpochNotifier
}

// NewDelegationSystemSC creates a new delegation system SC
func NewDelegationSystemSC(args ArgsNewDelegation) (*delegation, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if len(args.StakingSCAddress) < 1 {
		return nil, fmt.Errorf("%w for staking sc address", vm.ErrInvalidAddress)
	}
	if len(args.ValidatorSCAddress) < 1 {
		return nil, fmt.Errorf("%w for validator sc address", vm.ErrInvalidAddress)
	}
	if len(args.DelegationMgrSCAddress) < 1 {
		return nil, fmt.Errorf("%w for delegation sc address", vm.ErrInvalidAddress)
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, vm.ErrNilEpochNotifier
	}
	if check.IfNil(args.SigVerifier) {
		return nil, vm.ErrNilMessageSignVerifier
	}
	if args.DelegationSCConfig.MinServiceFee > args.DelegationSCConfig.MaxServiceFee {
		return nil, fmt.Errorf("%w minServiceFee bigger than maxServiceFee", vm.ErrInvalidDelegationSCConfig)
	}
	if args.DelegationSCConfig.MaxServiceFee < 1 {
		return nil, fmt.Errorf("%w maxServiceFee must be more than 0", vm.ErrInvalidDelegationSCConfig)
	}

	d := &delegation{
		eei:                    args.Eei,
		stakingSCAddr:          args.StakingSCAddress,
		validatorSCAddr:        args.ValidatorSCAddress,
		delegationMgrSCAddress: args.DelegationMgrSCAddress,
		gasCost:                args.GasCost,
		marshalizer:            args.Marshalizer,
		delegationEnabled:      atomic.Flag{},
		enableDelegationEpoch:  args.DelegationSCConfig.EnabledEpoch,
		minServiceFee:          args.DelegationSCConfig.MinServiceFee,
		maxServiceFee:          args.DelegationSCConfig.MaxServiceFee,
		sigVerifier:            args.SigVerifier,
		unBondPeriod:           args.StakingSCConfig.UnBondPeriod,
		endOfEpochAddr:         args.EndOfEpochAddress,
		stakingV2EnableEpoch:   args.StakingSCConfig.StakingV2Epoch,
		stakingV2Enabled:       atomic.Flag{},
	}

	var okValue bool

	d.unJailPrice, okValue = big.NewInt(0).SetString(args.StakingSCConfig.UnJailValue, conversionBase)
	if !okValue || d.unJailPrice.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidUnJailCost, args.StakingSCConfig.UnJailValue)
	}
	d.minStakeValue, okValue = big.NewInt(0).SetString(args.StakingSCConfig.MinStakeValue, conversionBase)
	if !okValue || d.minStakeValue.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidMinStakeValue, args.StakingSCConfig.MinStakeValue)
	}
	d.nodePrice, okValue = big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, conversionBase)
	if !okValue || d.nodePrice.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidNodePrice, args.StakingSCConfig.GenesisNodePrice)
	}

	args.EpochNotifier.RegisterNotifyHandler(d)

	return d, nil
}

// Execute calls one of the functions from the delegation contract and runs the code according to the input
func (d *delegation) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	d.mutExecution.RLock()
	defer d.mutExecution.RUnlock()

	err := CheckIfNil(args)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if !d.delegationEnabled.IsSet() {
		d.eei.AddReturnMessage("delegation contract is not enabled")
		return vmcommon.UserError
	}
	if bytes.Equal(args.RecipientAddr, vm.FirstDelegationSCAddress) {
		d.eei.AddReturnMessage("first delegation sc address cannot be called")
		return vmcommon.UserError
	}

	switch args.Function {
	case core.SCDeployInitFunctionName:
		return d.init(args)
	case "addNodes":
		return d.addNodes(args)
	case "removeNodes":
		return d.removeNodes(args)
	case "stakeNodes":
		return d.stakeNodes(args)
	case "unStakeNodes":
		return d.unStakeNodes(args)
	case "unBondNodes":
		return d.unBondNodes(args)
	case "unJailNodes":
		return d.unJailNodes(args)
	case "delegate":
		return d.delegate(args)
	case "unDelegate":
		return d.unDelegate(args)
	case "withdraw":
		return d.withdraw(args)
	case "changeServiceFee":
		return d.changeServiceFee(args)
	case "setCheckCapOnReDelegateRewards":
		return d.setCheckCapOnReDelegateRewards(args)
	case "setAutomaticActivation":
		return d.setAutomaticActivation(args)
	case "modifyTotalDelegationCap":
		return d.modifyTotalDelegationCap(args)
	case "updateRewards":
		return d.updateRewards(args)
	case "claimRewards":
		return d.claimRewards(args)
	case "getRewardData":
		return d.getRewardData(args)
	case "getClaimableRewards":
		return d.getClaimableRewards(args)
	case "getTotalCumulatedRewards":
		return d.getTotalCumulatedRewards(args)
	case "getNumUsers":
		return d.getNumUsers(args)
	case "getTotalUnStaked":
		return d.getTotalUnStaked(args)
	case "getTotalActiveStake":
		return d.getTotalActiveStake(args)
	case "getUserActiveStake":
		return d.getUserActiveStake(args)
	case "getUserUnStakedValue":
		return d.getUserUnStakedValue(args)
	case "getUserUnBondable":
		return d.getUserUnBondable(args)
	case "getUserUnDelegatedList":
		return d.getUserUnDelegatedList(args)
	case "getNumNodes":
		return d.getNumNodes(args)
	case "getAllNodeStates":
		return d.getAllNodeStates(args)
	case "getContractConfig":
		return d.getContractConfig(args)
	case "unStakeAtEndOfEpoch":
		return d.unStakeAtEndOfEpoch(args)
	case "reDelegateRewards":
		return d.reDelegateRewards(args)
	case "reStakeUnStakedNodes":
		return d.reStakeUnStakedNodes(args)
	case "isDelegator":
		return d.isDelegator(args)
	case "getDelegatorFundsData":
		return d.getDelegatorFundsData(args)
	case "getTotalCumulatedRewardsForUser":
		return d.getTotalCumulatedRewardsForUser(args)
	case "setMetaData":
		return d.setMetaData(args)
	case "getMetaData":
		return d.getMetaData(args)
	}

	d.eei.AddReturnMessage(args.Function + " is an unknown function")
	return vmcommon.UserError
}

func (d *delegation) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := d.eei.GetStorage([]byte(ownerKey))
	if len(ownerAddress) != 0 {
		d.eei.AddReturnMessage("smart contract was already initialized")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 2 {
		d.eei.AddReturnMessage("invalid number of arguments to init delegation contract")
		return vmcommon.UserError
	}
	serviceFee := big.NewInt(0).SetBytes(args.Arguments[1]).Uint64()
	if serviceFee < d.minServiceFee || serviceFee > d.maxServiceFee {
		d.eei.AddReturnMessage("service fee out of bounds")
		return vmcommon.UserError
	}
	maxDelegationCap := big.NewInt(0).SetBytes(args.Arguments[0])
	if maxDelegationCap.Cmp(zero) < 0 {
		d.eei.AddReturnMessage("invalid max delegation cap")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) < 0 {
		d.eei.AddReturnMessage("invalid call value")
		return vmcommon.UserError
	}

	initialOwnerFunds := big.NewInt(0).Set(args.CallValue)
	ownerAddress = args.CallerAddr
	d.eei.SetStorage([]byte(core.DelegationSystemSCKey), []byte(core.DelegationSystemSCKey))
	d.eei.SetStorage([]byte(ownerKey), ownerAddress)
	d.eei.SetStorage([]byte(serviceFeeKey), big.NewInt(0).SetUint64(serviceFee).Bytes())
	dConfig := &DelegationConfig{
		MaxDelegationCap:            maxDelegationCap,
		InitialOwnerFunds:           initialOwnerFunds,
		AutomaticActivation:         false,
		ChangeableServiceFee:        true,
		CreatedNonce:                d.eei.BlockChainHook().CurrentNonce(),
		UnBondPeriod:                d.unBondPeriod,
		CheckCapOnReDelegateRewards: true,
	}

	err := d.saveDelegationContractConfig(dConfig)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	dStatus := &DelegationContractStatus{
		StakedKeys:    make([]*NodesData, 0),
		NotStakedKeys: make([]*NodesData, 0),
		UnStakedKeys:  make([]*NodesData, 0),
	}

	globalFund := &GlobalFundData{
		ActiveFunds:   make([][]byte, 0),
		UnStakedFunds: make([][]byte, 0),
		TotalActive:   big.NewInt(0),
		TotalUnStaked: big.NewInt(0),
	}

	err = d.saveGlobalFundData(globalFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return d.delegateUser(initialOwnerFunds, ownerAddress, args.RecipientAddr, dStatus)
}

func (d *delegation) delegateUser(
	callValue *big.Int,
	callerAddr []byte,
	recipientAddr []byte,
	dStatus *DelegationContractStatus,
) vmcommon.ReturnCode {
	dConfig, err := d.getDelegationContractConfig()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	globalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.checkAndUpdateOwnerInitialFunds(dConfig, callerAddr, callValue)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	isNew, delegator, err := d.getOrCreateDelegatorData(callerAddr)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if isNew {
		delegator.RewardsCheckpoint = d.eei.BlockChainHook().CurrentEpoch() + 1
		delegator.UnClaimedRewards = big.NewInt(0)
	} else {
		err = d.computeAndUpdateRewards(callerAddr, delegator)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
	}

	return d.finishDelegateUser(globalFund, delegator, dConfig, dStatus,
		callerAddr, recipientAddr, callValue, isNew, true)
}

func (d *delegation) makeStakeArgsIfAutomaticActivation(
	config *DelegationConfig,
	status *DelegationContractStatus,
	globalFund *GlobalFundData,
) [][]byte {
	lenStakableKeys := uint64(len(status.NotStakedKeys)) + uint64(len(status.UnStakedKeys))
	if !config.AutomaticActivation || lenStakableKeys == 0 {
		return nil
	}

	maxNodesToStake := big.NewInt(0).Div(globalFund.TotalActive, d.nodePrice).Uint64()
	numStakedNodes := uint64(len(status.StakedKeys) + len(status.UnStakedKeys))
	if maxNodesToStake <= numStakedNodes {
		return nil
	}

	numNodesToStake := maxNodesToStake - numStakedNodes
	gasLeftToStakeNumNodes := d.eei.GasLeft() / d.gasCost.MetaChainSystemSCsCost.Stake

	numNodesToStake = core.MinUint64(core.MinUint64(lenStakableKeys, numNodesToStake), gasLeftToStakeNumNodes)
	if numNodesToStake == 0 {
		return nil
	}

	stakeArgs := [][]byte{big.NewInt(0).SetUint64(numNodesToStake).Bytes()}
	listOfStakableNodes := append(status.NotStakedKeys, status.UnStakedKeys...)
	for i := uint64(0); i < numNodesToStake; i++ {
		stakeArgs = append(stakeArgs, listOfStakableNodes[i].BLSKey)
		stakeArgs = append(stakeArgs, listOfStakableNodes[i].SignedMsg)
	}

	return stakeArgs
}

func (d *delegation) isOwner(address []byte) bool {
	ownerAddress := d.eei.GetStorage([]byte(ownerKey))
	return bytes.Equal(address, ownerAddress)
}

func (d *delegation) checkOwnerCallValueGasAndDuplicates(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !d.isOwner(args.CallerAddr) {
		d.eei.AddReturnMessage("only owner can call this method")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage(vm.ErrCallValueMustBeZero.Error())
		return vmcommon.UserError
	}
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	duplicates := checkForDuplicates(args.Arguments)
	if duplicates {
		d.eei.AddReturnMessage(vm.ErrDuplicatesFoundInArguments.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) basicArgCheckForConfigChanges(args *vmcommon.ContractCallInput) (*DelegationConfig, vmcommon.ReturnCode) {
	returnCode := d.checkOwnerCallValueGasAndDuplicates(args)
	if returnCode != vmcommon.Ok {
		return nil, returnCode
	}

	if len(args.Arguments) != 1 {
		d.eei.AddReturnMessage("invalid number of arguments")
		return nil, vmcommon.UserError
	}

	dConfig, err := d.getDelegationContractConfig()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.UserError
	}

	return dConfig, vmcommon.Ok
}

func (d *delegation) setAutomaticActivation(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	dConfig, returnCode := d.basicArgCheckForConfigChanges(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	switch string(args.Arguments[0]) {
	case "true":
		dConfig.AutomaticActivation = true
	case "false":
		dConfig.AutomaticActivation = false
	default:
		d.eei.AddReturnMessage("invalid argument")
		return vmcommon.UserError
	}

	err := d.saveDelegationContractConfig(dConfig)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) changeServiceFee(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkOwnerCallValueGasAndDuplicates(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if len(args.Arguments) != 1 {
		d.eei.AddReturnMessage("invalid number of arguments")
		return vmcommon.FunctionWrongSignature
	}

	newServiceFeeBigInt := big.NewInt(0).SetBytes(args.Arguments[0])
	newServiceFee := newServiceFeeBigInt.Uint64()
	if newServiceFee < d.minServiceFee || newServiceFee > d.maxServiceFee {
		d.eei.AddReturnMessage("new service fee out of bounds")
		return vmcommon.UserError
	}

	d.eei.SetStorage([]byte(serviceFeeKey), big.NewInt(0).SetUint64(newServiceFee).Bytes())

	return vmcommon.Ok
}

func (d *delegation) setCheckCapOnReDelegateRewards(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	dConfig, returnCode := d.basicArgCheckForConfigChanges(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	switch string(args.Arguments[0]) {
	case "true":
		dConfig.CheckCapOnReDelegateRewards = true
	case "false":
		dConfig.CheckCapOnReDelegateRewards = false
	default:
		d.eei.AddReturnMessage("invalid argument")
		return vmcommon.UserError
	}

	err := d.saveDelegationContractConfig(dConfig)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) modifyTotalDelegationCap(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	dConfig, returnCode := d.basicArgCheckForConfigChanges(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	newTotalDelegationCap := big.NewInt(0).SetBytes(args.Arguments[0])
	globalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if newTotalDelegationCap.Cmp(globalFund.TotalActive) < 0 && newTotalDelegationCap.Cmp(zero) != 0 {
		d.eei.AddReturnMessage("cannot make total delegation cap smaller than active")
		return vmcommon.UserError
	}

	dConfig.MaxDelegationCap = newTotalDelegationCap

	err = d.saveDelegationContractConfig(dConfig)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) checkBLSKeysIfExistsInStakingSC(blsKeys [][]byte) bool {
	for _, blsKey := range blsKeys {
		returnData := d.eei.GetStorageFromAddress(d.stakingSCAddr, blsKey)
		if len(returnData) > 0 {
			return true
		}
	}
	return false
}

func (d *delegation) addNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkOwnerCallValueGasAndDuplicates(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if len(args.Arguments) < 2 {
		d.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}

	if len(args.Arguments)%2 != 0 {
		d.eei.AddReturnMessage("arguments must be of pair length - BLSKey and signedMessage")
		return vmcommon.UserError
	}

	numBlsKeys := uint64(len(args.Arguments) / 2)
	err := d.eei.UseGas(numBlsKeys * d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	blsKeys, err := d.verifyBLSKeysAndSignature(args.RecipientAddr, args.Arguments)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	status, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	listToVerify := append(status.StakedKeys, status.NotStakedKeys...)
	listToVerify = append(listToVerify, status.UnStakedKeys...)
	foundOne := verifyIfBLSPubKeysExist(listToVerify, blsKeys)
	if foundOne {
		d.eei.AddReturnMessage(vm.ErrBLSPublicKeyMismatch.Error())
		return vmcommon.UserError
	}

	foundOne = d.checkBLSKeysIfExistsInStakingSC(blsKeys)
	if foundOne {
		d.eei.AddReturnMessage("BLSKey already in use in stakingSC")
		return vmcommon.UserError
	}

	for i := 0; i < len(args.Arguments); i += 2 {
		nodesData := &NodesData{
			BLSKey:    args.Arguments[i],
			SignedMsg: args.Arguments[i+1],
		}
		status.NotStakedKeys = append(status.NotStakedKeys, nodesData)
	}
	err = d.saveDelegationStatus(status)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) verifyBLSKeysAndSignature(txPubKey []byte, args [][]byte) ([][]byte, error) {
	blsKeys := make([][]byte, 0)

	foundInvalid := false
	for i := 0; i < len(args); i += 2 {
		blsKey := args[i]
		signedMessage := args[i+1]
		err := d.sigVerifier.Verify(txPubKey, signedMessage, blsKey)
		if err != nil {
			foundInvalid = true
			d.eei.Finish(blsKey)
			d.eei.Finish([]byte{invalidKey})
			continue
		}

		blsKeys = append(blsKeys, blsKey)
	}
	if foundInvalid {
		return nil, vm.ErrInvalidBLSKeys
	}

	return blsKeys, nil
}

func (d *delegation) removeNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkOwnerCallValueGasAndDuplicates(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if len(args.Arguments) < 1 {
		d.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}

	err := d.eei.UseGas(uint64(len(args.Arguments)) * d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	status, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	for _, blsKey := range args.Arguments {
		found := false
		for i, nodeData := range status.NotStakedKeys {
			if bytes.Equal(blsKey, nodeData.BLSKey) {
				copy(status.NotStakedKeys[i:], status.NotStakedKeys[i+1:])
				lenKeys := len(status.NotStakedKeys)
				status.NotStakedKeys[lenKeys-1] = nil
				status.NotStakedKeys = status.NotStakedKeys[:lenKeys-1]
				found = true
				break
			}
		}

		if !found {
			d.eei.AddReturnMessage(vm.ErrBLSPublicKeyMismatch.Error())
			return vmcommon.UserError
		}
	}

	err = d.saveDelegationStatus(status)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) stakeNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkOwnerCallValueGasAndDuplicates(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if len(args.Arguments) == 0 {
		d.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}
	status, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	listToCheck := append(status.NotStakedKeys, status.UnStakedKeys...)
	foundAll := verifyIfAllBLSPubKeysExist(listToCheck, args.Arguments)
	if !foundAll {
		d.eei.AddReturnMessage(vm.ErrBLSPublicKeyMismatch.Error())
		return vmcommon.UserError
	}

	globalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	numNodesToStake := big.NewInt(int64(len(args.Arguments) + len(status.StakedKeys)))
	stakeValue := big.NewInt(0).Mul(d.nodePrice, numNodesToStake)

	if globalFund.TotalActive.Cmp(stakeValue) < 0 {
		d.eei.AddReturnMessage("not enough in total active to stake")
		return vmcommon.UserError
	}

	stakeArgs := makeStakeArgs(listToCheck, args.Arguments)
	vmOutput, err := d.executeOnValidatorSC(args.RecipientAddr, "stake", stakeArgs, big.NewInt(0))
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return vmOutput.ReturnCode
	}

	err = d.updateDelegationStatusAfterStake(status, vmOutput.ReturnData, args.Arguments)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) updateDelegationStatusAfterStake(
	status *DelegationContractStatus,
	returnData [][]byte,
	args [][]byte,
) error {
	successKeys, _ := getSuccessAndUnSuccessKeys(returnData, args)
	for _, successKey := range successKeys {
		status.NotStakedKeys, status.StakedKeys = moveNodeFromList(status.NotStakedKeys, status.StakedKeys, successKey)
		status.UnStakedKeys, status.StakedKeys = moveNodeFromList(status.UnStakedKeys, status.StakedKeys, successKey)
	}

	err := d.saveDelegationStatus(status)
	if err != nil {
		return err
	}

	return nil
}

func (d *delegation) unStakeNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkOwnerCallValueGasAndDuplicates(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if len(args.Arguments) == 0 {
		d.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}
	status, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	foundAll := verifyIfAllBLSPubKeysExist(status.StakedKeys, args.Arguments)
	if !foundAll {
		d.eei.AddReturnMessage(vm.ErrBLSPublicKeyMismatch.Error())
		return vmcommon.UserError
	}

	vmOutput, err := d.executeOnValidatorSC(args.RecipientAddr, "unStakeNodes", args.Arguments, big.NewInt(0))
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return vmOutput.ReturnCode
	}

	successKeys, _ := getSuccessAndUnSuccessKeys(vmOutput.ReturnData, args.Arguments)
	for _, successKey := range successKeys {
		status.StakedKeys, status.UnStakedKeys = moveNodeFromList(status.StakedKeys, status.UnStakedKeys, successKey)
	}

	err = d.saveDelegationStatus(status)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) reStakeUnStakedNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkOwnerCallValueGasAndDuplicates(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if len(args.Arguments) == 0 {
		d.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}
	status, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	foundAll := verifyIfAllBLSPubKeysExist(status.UnStakedKeys, args.Arguments)
	if !foundAll {
		d.eei.AddReturnMessage(vm.ErrBLSPublicKeyMismatch.Error())
		return vmcommon.UserError
	}

	vmOutput, err := d.executeOnValidatorSC(args.RecipientAddr, "reStakeUnStakedNodes", args.Arguments, big.NewInt(0))
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return vmOutput.ReturnCode
	}

	for _, successKey := range args.Arguments {
		status.UnStakedKeys, status.StakedKeys = moveNodeFromList(status.UnStakedKeys, status.StakedKeys, successKey)
	}

	err = d.saveDelegationStatus(status)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) unBondNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkOwnerCallValueGasAndDuplicates(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if len(args.Arguments) == 0 {
		d.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}
	status, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	// even some staked keys can be unbonded - as they could have been forced unstaked by protocol because of not enough funds
	listToCheck := append(status.UnStakedKeys, status.StakedKeys...)
	foundAll := verifyIfAllBLSPubKeysExist(listToCheck, args.Arguments)
	if !foundAll {
		d.eei.AddReturnMessage(vm.ErrBLSPublicKeyMismatch.Error())
		return vmcommon.UserError
	}

	vmOutput, err := d.executeOnValidatorSC(args.RecipientAddr, "unBondNodes", args.Arguments, big.NewInt(0))
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return vmOutput.ReturnCode
	}

	successKeys, _ := getSuccessAndUnSuccessKeys(vmOutput.ReturnData, args.Arguments)
	for _, successKey := range successKeys {
		status.UnStakedKeys, status.NotStakedKeys = moveNodeFromList(status.UnStakedKeys, status.NotStakedKeys, successKey)
		status.StakedKeys, status.NotStakedKeys = moveNodeFromList(status.StakedKeys, status.NotStakedKeys, successKey)
	}

	err = d.saveDelegationStatus(status)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) unJailNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) == 0 {
		d.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	duplicates := checkForDuplicates(args.Arguments)
	if duplicates {
		d.eei.AddReturnMessage(vm.ErrDuplicatesFoundInArguments.Error())
		return vmcommon.UserError
	}
	status, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	listToCheck := append(status.StakedKeys, status.UnStakedKeys...)
	foundAll := verifyIfAllBLSPubKeysExist(listToCheck, args.Arguments)
	if !foundAll {
		d.eei.AddReturnMessage(vm.ErrBLSPublicKeyMismatch.Error())
		return vmcommon.UserError
	}

	isNew, delegator, err := d.getOrCreateDelegatorData(args.CallerAddr)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if isNew || len(delegator.ActiveFund) == 0 {
		d.eei.AddReturnMessage("not a delegator")
		return vmcommon.UserError
	}

	vmOutput, err := d.executeOnValidatorSC(args.RecipientAddr, "unJail", args.Arguments, args.CallValue)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return vmOutput.ReturnCode
	}

	sendBackValue := getTransferBackFromVMOutput(vmOutput)
	if sendBackValue.Cmp(zero) > 0 {
		err = d.eei.Transfer(args.CallerAddr, args.RecipientAddr, sendBackValue, nil, 0)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (d *delegation) reDelegateRewards(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage(vm.ErrCallValueMustBeZero.Error())
		return vmcommon.UserError
	}
	if len(args.Arguments) != 0 {
		d.eei.AddReturnMessage("must be called without arguments")
		return vmcommon.UserError
	}

	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	isNew, delegator, err := d.getOrCreateDelegatorData(args.CallerAddr)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if isNew {
		d.eei.AddReturnMessage("new delegator cannot redelegate rewards")
		return vmcommon.UserError
	}

	err = d.computeAndUpdateRewards(args.CallerAddr, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	dConfig, err := d.getDelegationContractConfig()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.checkAndUpdateOwnerInitialFunds(dConfig, args.CallerAddr, delegator.UnClaimedRewards)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if delegator.UnClaimedRewards.Cmp(zero) <= 0 {
		d.eei.AddReturnMessage("delegate value must be higher than 0")
		return vmcommon.UserError
	}

	globalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	dStatus, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	delegateValue := big.NewInt(0).Set(delegator.UnClaimedRewards)
	delegator.UnClaimedRewards.SetUint64(0)

	return d.finishDelegateUser(globalFund, delegator, dConfig, dStatus, args.CallerAddr,
		args.RecipientAddr, delegateValue, false, dConfig.CheckCapOnReDelegateRewards)
}

func (d *delegation) finishDelegateUser(
	globalFund *GlobalFundData,
	delegator *DelegatorData,
	dConfig *DelegationConfig,
	dStatus *DelegationContractStatus,
	callerAddr []byte,
	scAddress []byte,
	callValue *big.Int,
	isNew bool,
	checkDelegationCap bool,
) vmcommon.ReturnCode {
	globalFund.TotalActive.Add(globalFund.TotalActive, callValue)
	withDelegationCap := dConfig.MaxDelegationCap.Cmp(zero) != 0
	if withDelegationCap && checkDelegationCap && globalFund.TotalActive.Cmp(dConfig.MaxDelegationCap) > 0 {
		d.eei.AddReturnMessage("total delegation cap reached")
		return vmcommon.UserError
	}

	var err error
	if len(delegator.ActiveFund) == 0 {
		var fundKey []byte
		fundKey, err = d.createAndSaveNextKeyFund(callerAddr, callValue, active)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		delegator.ActiveFund = fundKey
		d.addNewFundToGlobalData(globalFund, fundKey, active)
		if isNew {
			dStatus.NumUsers++
		}
	} else {
		err = d.addValueToFund(delegator.ActiveFund, callValue)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
	}

	stakeArgs := d.makeStakeArgsIfAutomaticActivation(dConfig, dStatus, globalFund)
	vmOutput, err := d.executeOnValidatorSC(scAddress, "stake", stakeArgs, callValue)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return vmOutput.ReturnCode
	}

	if len(stakeArgs) > 0 {
		err = d.updateDelegationStatusAfterStake(dStatus, vmOutput.ReturnData, stakeArgs)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
	}

	err = d.saveDelegationStatus(dStatus)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.saveGlobalFundData(globalFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.saveDelegatorData(callerAddr, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) delegate(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	delegationManagement, err := d.getDelegationManagement()
	if err != nil {
		d.eei.AddReturnMessage("error getting minimum delegation amount " + err.Error())
		return vmcommon.UserError
	}

	minDelegationAmount := delegationManagement.MinDelegationAmount

	if args.CallValue.Cmp(minDelegationAmount) < 0 {
		d.eei.AddReturnMessage("delegate value must be higher than minDelegationAmount " + minDelegationAmount.String())
		return vmcommon.UserError
	}
	err = d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	dStatus, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return d.delegateUser(args.CallValue, args.CallerAddr, args.RecipientAddr, dStatus)
}

func (d *delegation) addValueToFund(key []byte, value *big.Int) error {
	fund, err := d.getFund(key)
	if err != nil {
		return err
	}

	fund.Value.Add(fund.Value, value)

	return d.saveFund(key, fund)
}

func (d *delegation) resolveUnStakedUnBondResponse(
	returnData [][]byte,
	userVal *big.Int,
) (*big.Int, error) {
	lenReturnData := len(returnData)
	if lenReturnData == 0 {
		return userVal, nil
	}

	totalReturn := big.NewInt(0).SetBytes(returnData[lenReturnData-1])
	return totalReturn, nil
}

func (d *delegation) checkOwnerCanUnDelegate(address []byte, activeFund *Fund, valueToUnDelegate *big.Int) error {
	if !d.isOwner(address) {
		return nil
	}

	delegationConfig, err := d.getDelegationContractConfig()
	if err != nil {
		return err
	}

	remainingFunds := big.NewInt(0).Sub(activeFund.Value, valueToUnDelegate)
	if remainingFunds.Cmp(delegationConfig.InitialOwnerFunds) >= 0 {
		return nil
	}

	delegationStatus, err := d.getDelegationStatus()
	if err != nil {
		return err
	}

	numActiveKeys := len(delegationStatus.StakedKeys) + len(delegationStatus.UnStakedKeys)
	if numActiveKeys > 0 {
		return fmt.Errorf("%w cannot unDelegate from initial owner funds as nodes are active", vm.ErrOwnerCannotUnDelegate)
	}
	if remainingFunds.Cmp(zero) != 0 {
		return fmt.Errorf("%w must undelegate all", vm.ErrOwnerCannotUnDelegate)
	}

	delegationConfig.InitialOwnerFunds = big.NewInt(0)
	err = d.saveDelegationContractConfig(delegationConfig)
	if err != nil {
		return err
	}

	return nil
}

func (d *delegation) unDelegate(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 1 {
		d.eei.AddReturnMessage("wrong number of arguments")
		return vmcommon.FunctionWrongSignature
	}
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage(vm.ErrCallValueMustBeZero.Error())
		return vmcommon.UserError
	}
	valueToUnDelegate := big.NewInt(0).SetBytes(args.Arguments[0])
	if valueToUnDelegate.Cmp(zero) <= 0 {
		d.eei.AddReturnMessage("invalid value to undelegate")
		return vmcommon.UserError
	}

	isNew, delegator, err := d.getOrCreateDelegatorData(args.CallerAddr)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if isNew {
		d.eei.AddReturnMessage("caller is not a delegator")
		return vmcommon.UserError
	}

	activeFund, err := d.getFund(delegator.ActiveFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if activeFund.Value.Cmp(valueToUnDelegate) < 0 {
		d.eei.AddReturnMessage("invalid value to undelegate")
		return vmcommon.UserError
	}

	delegationManagement, err := d.getDelegationManagement()
	if err != nil {
		d.eei.AddReturnMessage("error getting minimum delegation amount " + err.Error())
		return vmcommon.UserError
	}

	minDelegationAmount := delegationManagement.MinDelegationAmount

	remainedFund := big.NewInt(0).Sub(activeFund.Value, valueToUnDelegate)
	if remainedFund.Cmp(zero) > 0 && remainedFund.Cmp(minDelegationAmount) < 0 {
		d.eei.AddReturnMessage("invalid value to undelegate - need to undelegate all - do not leave dust behind")
		return vmcommon.UserError
	}
	err = d.checkOwnerCanUnDelegate(args.CallerAddr, activeFund, valueToUnDelegate)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	err = d.computeAndUpdateRewards(args.CallerAddr, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	globalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	returnData, returnCode := d.executeOnValidatorSCWithValueInArgs(args.RecipientAddr, "unStakeTokens", valueToUnDelegate)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	actualUserUnStake, err := d.resolveUnStakedUnBondResponse(returnData, valueToUnDelegate)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	activeFund.Value.Sub(activeFund.Value, actualUserUnStake)
	err = d.saveFund(delegator.ActiveFund, activeFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	unStakedFundKey, err := d.createAndSaveNextKeyFund(args.CallerAddr, actualUserUnStake, unStaked)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	globalFund.UnStakedFunds = append(globalFund.UnStakedFunds, unStakedFundKey)
	globalFund.TotalActive.Sub(globalFund.TotalActive, actualUserUnStake)
	globalFund.TotalUnStaked.Add(globalFund.TotalUnStaked, actualUserUnStake)
	delegator.UnStakedFunds = append(delegator.UnStakedFunds, unStakedFundKey)

	if activeFund.Value.Cmp(zero) == 0 {
		for i, fundKey := range globalFund.ActiveFunds {
			if bytes.Equal(delegator.ActiveFund, fundKey) {
				copy(globalFund.ActiveFunds[i:], globalFund.ActiveFunds[i+1:])
				lenKeys := len(globalFund.ActiveFunds)
				globalFund.ActiveFunds[lenKeys-1] = nil
				globalFund.ActiveFunds = globalFund.ActiveFunds[:lenKeys-1]
				break
			}
		}
		delegator.ActiveFund = nil
	}

	err = d.saveGlobalFundData(globalFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.saveDelegatorData(args.CallerAddr, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) updateRewards(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, d.endOfEpochAddr) {
		d.eei.AddReturnMessage("only end of epoch address can call this function")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 0 {
		d.eei.AddReturnMessage("must call without arguments")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) < 0 {
		d.eei.AddReturnMessage("cannot call with negative value")
		return vmcommon.UserError
	}

	totalActiveData := d.eei.GetStorage([]byte(totalActiveKey))
	serviceFeeData := d.eei.GetStorage([]byte(serviceFeeKey))
	rewardsData := &RewardComputationData{
		RewardsToDistribute: args.CallValue,
		TotalActive:         big.NewInt(0).SetBytes(totalActiveData),
		ServiceFee:          big.NewInt(0).SetBytes(serviceFeeData).Uint64(),
	}
	currentEpoch := d.eei.BlockChainHook().CurrentEpoch()
	err := d.saveRewardData(currentEpoch, rewardsData)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) getRewardData(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != 1 {
		d.eei.AddReturnMessage("must call with 1 arguments")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage(vm.ErrCallValueMustBeZero.Error())
		return vmcommon.UserError
	}
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	epoch := big.NewInt(0).SetBytes(args.Arguments[0]).Uint64()
	found, rewardData, err := d.getRewardComputationData(uint32(epoch))
	if !found {
		d.eei.AddReturnMessage("reward not found")
		return vmcommon.UserError
	}
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	d.eei.Finish(rewardData.RewardsToDistribute.Bytes())
	d.eei.Finish(rewardData.TotalActive.Bytes())
	d.eei.Finish(big.NewInt(0).SetUint64(rewardData.ServiceFee).Bytes())

	return vmcommon.Ok
}

func (d *delegation) getRewardComputationData(epoch uint32) (bool, *RewardComputationData, error) {
	marshaledData := d.eei.GetStorage(rewardKeyForEpoch(epoch))
	if len(marshaledData) == 0 {
		return false, nil, nil
	}
	rewardsData := &RewardComputationData{}
	err := d.marshalizer.Unmarshal(rewardsData, marshaledData)
	if err != nil {
		return false, nil, err
	}

	return true, rewardsData, nil
}

func rewardKeyForEpoch(epoch uint32) []byte {
	epochInBytes := big.NewInt(int64(epoch)).Bytes()
	return append([]byte(rewardKeyPrefix), epochInBytes...)
}

func (d *delegation) saveRewardData(epoch uint32, rewardsData *RewardComputationData) error {
	marshaledData, err := d.marshalizer.Marshal(rewardsData)
	if err != nil {
		return err
	}

	d.eei.SetStorage(rewardKeyForEpoch(epoch), marshaledData)
	return nil
}

func (d *delegation) computeAndUpdateRewards(callerAddress []byte, delegator *DelegatorData) error {
	if len(delegator.ActiveFund) == 0 {
		// nothing to calculate as no active funds - all were computed before
		return nil
	}

	activeFund, err := d.getFund(delegator.ActiveFund)
	if err != nil {
		return err
	}

	isOwner := d.isOwner(callerAddress)

	totalRewards := big.NewInt(0)
	currentEpoch := d.eei.BlockChainHook().CurrentEpoch()
	for i := delegator.RewardsCheckpoint; i <= currentEpoch; i++ {
		found, rewardData, errGet := d.getRewardComputationData(i)
		if errGet != nil {
			return errGet
		}
		if !found {
			continue
		}

		if rewardData.TotalActive.Cmp(zero) == 0 {
			if isOwner {
				totalRewards.Add(totalRewards, rewardData.RewardsToDistribute)
			}
			continue
		}

		var rewardsForOwner *big.Int
		percentage := float64(rewardData.ServiceFee) / float64(d.maxServiceFee)
		if d.stakingV2Enabled.IsSet() {
			rewardsForOwner = core.GetIntTrimmedPercentageOfValue(rewardData.RewardsToDistribute, percentage)
		} else {
			rewardsForOwner = core.GetApproximatePercentageOfValue(rewardData.RewardsToDistribute, percentage)
		}

		rewardForDelegator := big.NewInt(0).Sub(rewardData.RewardsToDistribute, rewardsForOwner)

		// delegator reward is: rewardForDelegator * user stake / total active
		rewardForDelegator.Mul(rewardForDelegator, activeFund.Value)
		rewardForDelegator.Div(rewardForDelegator, rewardData.TotalActive)

		if isOwner {
			totalRewards.Add(totalRewards, rewardsForOwner)
		}
		totalRewards.Add(totalRewards, rewardForDelegator)
	}

	delegator.UnClaimedRewards.Add(delegator.UnClaimedRewards, totalRewards)
	delegator.RewardsCheckpoint = currentEpoch + 1

	return nil
}

func (d *delegation) claimRewards(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 0 {
		d.eei.AddReturnMessage("wrong number of arguments")
		return vmcommon.FunctionWrongSignature
	}

	isNew, delegator, err := d.getOrCreateDelegatorData(args.CallerAddr)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if isNew {
		d.eei.AddReturnMessage("caller is not a delegator")
		return vmcommon.UserError
	}

	err = d.computeAndUpdateRewards(args.CallerAddr, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.eei.Transfer(args.CallerAddr, args.RecipientAddr, delegator.UnClaimedRewards, nil, 0)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	delegator.TotalCumulatedRewards.Add(delegator.TotalCumulatedRewards, delegator.UnClaimedRewards)
	delegator.UnClaimedRewards.SetUint64(0)
	err = d.saveDelegatorData(args.CallerAddr, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) executeOnValidatorSCWithValueInArgs(
	scAddress []byte,
	functionToCall string,
	actionValue *big.Int,
) ([][]byte, vmcommon.ReturnCode) {
	vmOutput, err := d.executeOnValidatorSC(scAddress, functionToCall, [][]byte{actionValue.Bytes()}, big.NewInt(0))
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.UserError
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, vmcommon.UserError
	}

	return vmOutput.ReturnData, vmcommon.Ok
}

func (d *delegation) getUnBondableTokens(delegator *DelegatorData, unBondPeriod uint64) (*big.Int, error) {
	totalUnBondable := big.NewInt(0)
	currentNonce := d.eei.BlockChainHook().CurrentNonce()
	for _, fundKey := range delegator.UnStakedFunds {
		fund, err := d.getFund(fundKey)
		if err != nil {
			return nil, err
		}
		if currentNonce-fund.Nonce < unBondPeriod {
			continue
		}
		totalUnBondable.Add(totalUnBondable, fund.Value)
	}
	return totalUnBondable, nil
}

func (d *delegation) deleteUnStakedFund(fundKey []byte, globalFund *GlobalFundData) {
	d.eei.SetStorage(fundKey, nil)
	for i, globalKey := range globalFund.UnStakedFunds {
		if bytes.Equal(fundKey, globalKey) {
			copy(globalFund.UnStakedFunds[i:], globalFund.UnStakedFunds[i+1:])
			lenKeys := len(globalFund.UnStakedFunds)
			globalFund.UnStakedFunds[lenKeys-1] = nil
			globalFund.UnStakedFunds = globalFund.UnStakedFunds[:lenKeys-1]
			break
		}
	}
}

func (d *delegation) withdraw(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != 0 {
		d.eei.AddReturnMessage("wrong number of arguments")
		return vmcommon.FunctionWrongSignature
	}
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage(vm.ErrCallValueMustBeZero.Error())
		return vmcommon.UserError
	}
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	isNew, delegator, err := d.getOrCreateDelegatorData(args.CallerAddr)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if isNew {
		d.eei.AddReturnMessage("caller is not a delegator")
		return vmcommon.UserError
	}

	dConfig, err := d.getDelegationContractConfig()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	globalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	totalUnBondable, err := d.getUnBondableTokens(delegator, dConfig.UnBondPeriod)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if totalUnBondable.Cmp(zero) == 0 {
		d.eei.AddReturnMessage("nothing to unBond")
		return vmcommon.Ok
	}

	if globalFund.TotalUnStaked.Cmp(totalUnBondable) < 0 {
		d.eei.AddReturnMessage("cannot unBond - contract error")
		return vmcommon.UserError
	}

	returnData, returnCode := d.executeOnValidatorSCWithValueInArgs(args.RecipientAddr, "unBondTokens", totalUnBondable)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	actualUserUnBond, err := d.resolveUnStakedUnBondResponse(returnData, totalUnBondable)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	currentNonce := d.eei.BlockChainHook().CurrentNonce()
	totalUnBonded := big.NewInt(0)
	tempUnStakedFunds := make([][]byte, 0)
	var fund *Fund
	for fundIndex, fundKey := range delegator.UnStakedFunds {
		fund, err = d.getFund(fundKey)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
		if currentNonce-fund.Nonce < dConfig.UnBondPeriod {
			tempUnStakedFunds = append(tempUnStakedFunds, delegator.UnStakedFunds[fundIndex])
			continue
		}

		totalUnBonded.Add(totalUnBonded, fund.Value)
		if totalUnBonded.Cmp(actualUserUnBond) > 0 {
			unBondedFromThisFund := big.NewInt(0).Sub(totalUnBonded, actualUserUnBond)
			fund.Value.Sub(fund.Value, unBondedFromThisFund)
			err = d.saveFund(fundKey, fund)
			if err != nil {
				d.eei.AddReturnMessage(err.Error())
				return vmcommon.UserError
			}
			break
		}
		d.deleteUnStakedFund(fundKey, globalFund)
	}
	delegator.UnStakedFunds = tempUnStakedFunds

	globalFund.TotalUnStaked.Sub(globalFund.TotalUnStaked, actualUserUnBond)
	err = d.saveGlobalFundData(globalFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	err = d.saveDelegatorData(args.CallerAddr, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.eei.Transfer(args.CallerAddr, args.RecipientAddr, actualUserUnBond, nil, 0)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.deleteDelegatorIfNeeded(args.CallerAddr, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) deleteDelegatorIfNeeded(address []byte, delegator *DelegatorData) error {
	if d.isOwner(address) {
		return nil
	}

	isDelegatorWithoutFunds := len(delegator.ActiveFund) == 0 && len(delegator.UnStakedFunds) == 0
	if !isDelegatorWithoutFunds {
		return nil
	}

	err := d.computeAndUpdateRewards(address, delegator)
	if err != nil {
		return err
	}

	if delegator.UnClaimedRewards.Cmp(zero) == 0 {
		d.eei.SetStorage(address, nil)

		dStatus, errGet := d.getDelegationStatus()
		if errGet != nil {
			return errGet
		}

		if dStatus.NumUsers > 0 {
			dStatus.NumUsers--
		}

		err = d.saveDelegationStatus(dStatus)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *delegation) unStakeAtEndOfEpoch(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, d.endOfEpochAddr) {
		d.eei.AddReturnMessage("can be called by end of epoch address only")
		return vmcommon.UserError
	}

	status, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	for _, unStakedKey := range args.Arguments {
		status.StakedKeys, status.UnStakedKeys = moveNodeFromList(status.StakedKeys, status.UnStakedKeys, unStakedKey)
	}

	err = d.saveDelegationStatus(status)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) checkArgumentsForGeneralViewFunc(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage(vm.ErrCallValueMustBeZero.Error())
		return vmcommon.UserError
	}
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 0 {
		d.eei.AddReturnMessage(vm.ErrInvalidNumOfArguments.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) getTotalCumulatedRewards(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkArgumentsForGeneralViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if !bytes.Equal(args.CallerAddr, d.endOfEpochAddr) {
		d.eei.AddReturnMessage("this is a view function only")
		return vmcommon.UserError
	}

	totalCumulatedRewards := big.NewInt(0)
	currentEpoch := d.eei.BlockChainHook().CurrentEpoch()
	for i := uint32(0); i <= currentEpoch; i++ {
		found, rewardsData, err := d.getRewardComputationData(i)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
		if !found {
			continue
		}

		totalCumulatedRewards.Add(totalCumulatedRewards, rewardsData.RewardsToDistribute)
	}
	d.eei.Finish(totalCumulatedRewards.Bytes())

	return vmcommon.Ok
}

func (d *delegation) getNumUsers(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkArgumentsForGeneralViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	delegationStatus, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	numDelegators := big.NewInt(int64(delegationStatus.NumUsers))
	d.eei.Finish(numDelegators.Bytes())

	return vmcommon.Ok
}

func (d *delegation) getTotalUnStaked(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkArgumentsForGeneralViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	globalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	d.eei.Finish(globalFund.TotalUnStaked.Bytes())
	return vmcommon.Ok
}

func (d *delegation) getTotalActiveStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkArgumentsForGeneralViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	globalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	d.eei.Finish(globalFund.TotalActive.Bytes())
	return vmcommon.Ok
}

func (d *delegation) getNumNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkArgumentsForGeneralViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	delegationStatus, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	numNodes := len(delegationStatus.StakedKeys) + len(delegationStatus.UnStakedKeys) + len(delegationStatus.NotStakedKeys)
	d.eei.Finish(big.NewInt(int64(numNodes)).Bytes())

	return vmcommon.Ok
}

func (d *delegation) getAllNodeStates(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkArgumentsForGeneralViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	delegationStatus, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if len(delegationStatus.StakedKeys) > 0 {
		d.eei.Finish([]byte("staked"))
	}
	for _, node := range delegationStatus.StakedKeys {
		d.eei.Finish(node.BLSKey)
	}

	if len(delegationStatus.NotStakedKeys) > 0 {
		d.eei.Finish([]byte("notStaked"))
	}
	for _, node := range delegationStatus.NotStakedKeys {
		d.eei.Finish(node.BLSKey)
	}

	if len(delegationStatus.UnStakedKeys) > 0 {
		d.eei.Finish([]byte("unStaked"))
	}
	for _, node := range delegationStatus.UnStakedKeys {
		d.eei.Finish(node.BLSKey)
	}

	return vmcommon.Ok
}

func (d *delegation) getContractConfig(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkArgumentsForGeneralViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	delegationConfig, err := d.getDelegationContractConfig()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	automaticActivation := "false"
	if delegationConfig.AutomaticActivation {
		automaticActivation = "true"
	}

	withDelegationCap := "false"
	if delegationConfig.MaxDelegationCap.Cmp(zero) != 0 {
		withDelegationCap = "true"
	}

	changeableServiceFee := "false"
	if delegationConfig.ChangeableServiceFee {
		changeableServiceFee = "true"
	}

	checkCapOnReDelegate := "false"
	if delegationConfig.CheckCapOnReDelegateRewards {
		checkCapOnReDelegate = "true"
	}

	ownerAddress := d.eei.GetStorage([]byte(ownerKey))
	serviceFee := d.eei.GetStorage([]byte(serviceFeeKey))

	d.eei.Finish(ownerAddress)
	d.eei.Finish(serviceFee)
	d.eei.Finish(delegationConfig.MaxDelegationCap.Bytes())
	d.eei.Finish(delegationConfig.InitialOwnerFunds.Bytes())
	d.eei.Finish([]byte(automaticActivation))
	d.eei.Finish([]byte(withDelegationCap))
	d.eei.Finish([]byte(changeableServiceFee))
	d.eei.Finish([]byte(checkCapOnReDelegate))
	d.eei.Finish(big.NewInt(0).SetUint64(delegationConfig.CreatedNonce).Bytes())
	d.eei.Finish(big.NewInt(0).SetUint64(delegationConfig.UnBondPeriod).Bytes())

	return vmcommon.Ok
}

func (d *delegation) checkArgumentsForUserViewFunc(args *vmcommon.ContractCallInput) (*DelegatorData, vmcommon.ReturnCode) {
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage(vm.ErrCallValueMustBeZero.Error())
		return nil, vmcommon.UserError
	}
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.OutOfGas
	}
	if len(args.Arguments) != 1 {
		d.eei.AddReturnMessage(vm.ErrInvalidNumOfArguments.Error())
		return nil, vmcommon.UserError
	}

	isNew, delegator, err := d.getOrCreateDelegatorData(args.Arguments[0])
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.UserError
	}

	if isNew {
		d.eei.AddReturnMessage("view function works only for existing delegators")
		return nil, vmcommon.UserError
	}

	return delegator, vmcommon.Ok
}

func (d *delegation) getUserActiveStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	delegator, returnCode := d.checkArgumentsForUserViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if len(delegator.ActiveFund) == 0 {
		d.eei.Finish(big.NewInt(0).Bytes())
		return vmcommon.Ok
	}

	fund, err := d.getFund(delegator.ActiveFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	d.eei.Finish(fund.Value.Bytes())

	return vmcommon.Ok
}

func (d *delegation) computeTotalUnStaked(delegator *DelegatorData) (*big.Int, error) {
	totalUnStaked := big.NewInt(0)
	for _, fundKey := range delegator.UnStakedFunds {
		fund, err := d.getFund(fundKey)
		if err != nil {
			return nil, err
		}
		totalUnStaked.Add(totalUnStaked, fund.Value)
	}
	return totalUnStaked, nil
}

func (d *delegation) getUserUnStakedValue(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	delegator, returnCode := d.checkArgumentsForUserViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	totalUnStaked, err := d.computeTotalUnStaked(delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	d.eei.Finish(totalUnStaked.Bytes())

	return vmcommon.Ok
}

func (d *delegation) getUserUnBondable(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	delegator, returnCode := d.checkArgumentsForUserViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	dConfig, err := d.getDelegationContractConfig()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	totalUnBondable, err := d.getUnBondableTokens(delegator, dConfig.UnBondPeriod)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	d.eei.Finish(totalUnBondable.Bytes())
	return vmcommon.Ok
}

func (d *delegation) getUserUnDelegatedList(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	delegator, returnCode := d.checkArgumentsForUserViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	dConfig, err := d.getDelegationContractConfig()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	currentNonce := d.eei.BlockChainHook().CurrentNonce()
	var fund *Fund
	for _, fundKey := range delegator.UnStakedFunds {
		fund, err = d.getFund(fundKey)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		d.eei.Finish(fund.Value.Bytes())
		elapsedNonce := currentNonce - fund.Nonce
		if elapsedNonce >= dConfig.UnBondPeriod {
			d.eei.Finish(zero.Bytes())
			continue
		}

		remainingNonce := dConfig.UnBondPeriod - elapsedNonce
		d.eei.Finish(big.NewInt(0).SetUint64(remainingNonce).Bytes())
	}

	return vmcommon.Ok
}

func (d *delegation) getClaimableRewards(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	delegator, returnCode := d.checkArgumentsForUserViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	err := d.computeAndUpdateRewards(args.Arguments[0], delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	d.eei.Finish(delegator.UnClaimedRewards.Bytes())
	return vmcommon.Ok
}

func (d *delegation) isDelegator(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	_, returnCode := d.checkArgumentsForUserViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	return vmcommon.Ok
}

func (d *delegation) getDelegatorFundsData(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	delegator, returnCode := d.checkArgumentsForUserViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if len(delegator.ActiveFund) > 0 {
		fund, err := d.getFund(delegator.ActiveFund)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		d.eei.Finish(fund.Value.Bytes())
	} else {
		d.eei.Finish(zero.Bytes())
	}

	err := d.computeAndUpdateRewards(args.Arguments[0], delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	d.eei.Finish(delegator.UnClaimedRewards.Bytes())

	totalUnStaked, err := d.computeTotalUnStaked(delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	d.eei.Finish(totalUnStaked.Bytes())

	dConfig, err := d.getDelegationContractConfig()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	totalUnBondable, err := d.getUnBondableTokens(delegator, dConfig.UnBondPeriod)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	d.eei.Finish(totalUnBondable.Bytes())

	return vmcommon.Ok
}

func (d *delegation) getTotalCumulatedRewardsForUser(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	delegator, returnCode := d.checkArgumentsForUserViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	err := d.computeAndUpdateRewards(args.Arguments[0], delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	totalCumulatedRewards := big.NewInt(0).Add(delegator.TotalCumulatedRewards, delegator.UnClaimedRewards)
	d.eei.Finish(totalCumulatedRewards.Bytes())

	return vmcommon.Ok
}

func (d *delegation) setMetaData(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkOwnerCallValueGasAndDuplicates(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if len(args.Arguments) != 3 {
		d.eei.AddReturnMessage("needed 3 arguments")
		return vmcommon.UserError
	}

	dMetaData := &DelegationMetaData{
		Name:       args.Arguments[0],
		Website:    args.Arguments[1],
		Identifier: args.Arguments[2],
	}
	marshaledData, err := d.marshalizer.Marshal(dMetaData)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	d.eei.SetStorage([]byte(delegationMetaData), marshaledData)

	return vmcommon.Ok
}

func (d *delegation) getMetaData(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkArgumentsForGeneralViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	marshaledData := d.eei.GetStorage([]byte(delegationMetaData))
	if len(marshaledData) == 0 {
		d.eei.AddReturnMessage("delegation meta data is not set")
		return vmcommon.UserError
	}

	dMetaData := &DelegationMetaData{}
	err := d.marshalizer.Unmarshal(dMetaData, marshaledData)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	d.eei.Finish(dMetaData.Name)
	d.eei.Finish(dMetaData.Website)
	d.eei.Finish(dMetaData.Identifier)

	return vmcommon.Ok
}

func (d *delegation) executeOnValidatorSC(address []byte, function string, args [][]byte, value *big.Int) (*vmcommon.VMOutput, error) {
	validatorCall := function
	for _, key := range args {
		validatorCall += "@" + hex.EncodeToString(key)
	}
	vmOutput, err := d.eei.ExecuteOnDestContext(d.validatorSCAddr, address, value, []byte(validatorCall))
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return nil, err
	}

	return vmOutput, nil

}

func (d *delegation) getDelegationContractConfig() (*DelegationConfig, error) {
	marshaledData := d.eei.GetStorage([]byte(delegationConfigKey))
	if len(marshaledData) == 0 {
		return nil, fmt.Errorf("%w delegation contract config", vm.ErrDataNotFoundUnderKey)
	}

	dConfig := &DelegationConfig{}
	err := d.marshalizer.Unmarshal(dConfig, marshaledData)
	if err != nil {
		return nil, err
	}

	return dConfig, nil
}

func (d *delegation) saveDelegationContractConfig(dConfig *DelegationConfig) error {
	marshaledData, err := d.marshalizer.Marshal(dConfig)
	if err != nil {
		return err
	}

	d.eei.SetStorage([]byte(delegationConfigKey), marshaledData)
	return nil
}

func (d *delegation) getDelegationStatus() (*DelegationContractStatus, error) {
	marshaledData := d.eei.GetStorage([]byte(delegationStatusKey))
	if len(marshaledData) == 0 {
		return nil, fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	}

	status := &DelegationContractStatus{}
	err := d.marshalizer.Unmarshal(status, marshaledData)
	if err != nil {
		return nil, err
	}

	return status, nil
}

func (d *delegation) saveDelegationStatus(status *DelegationContractStatus) error {
	marshaledData, err := d.marshalizer.Marshal(status)
	if err != nil {
		return err
	}

	d.eei.SetStorage([]byte(delegationStatusKey), marshaledData)
	return nil
}

func (d *delegation) getOrCreateDelegatorData(address []byte) (bool, *DelegatorData, error) {
	dData := &DelegatorData{
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	}
	marshaledData := d.eei.GetStorage(address)
	if len(marshaledData) == 0 {
		return true, dData, nil
	}

	err := d.marshalizer.Unmarshal(dData, marshaledData)
	if err != nil {
		return false, nil, err
	}

	return false, dData, nil
}

func (d *delegation) saveDelegatorData(address []byte, dData *DelegatorData) error {
	marshaledData, err := d.marshalizer.Marshal(dData)
	if err != nil {
		return err
	}

	d.eei.SetStorage(address, marshaledData)
	return nil
}

func (d *delegation) getFund(key []byte) (*Fund, error) {
	marshaledData := d.eei.GetStorage(key)
	if len(marshaledData) == 0 {
		return nil, fmt.Errorf("%w getFund %s", vm.ErrDataNotFoundUnderKey, string(key))
	}

	dFund := &Fund{}
	err := d.marshalizer.Unmarshal(dFund, marshaledData)
	if err != nil {
		return nil, err
	}

	return dFund, nil
}

func (d *delegation) createAndSaveNextKeyFund(address []byte, value *big.Int, fundType uint32) ([]byte, error) {
	fundKey, fund := d.createNextKeyFund(address, value, fundType)
	err := d.saveFund(fundKey, fund)
	if err != nil {
		return nil, err
	}

	d.eei.SetStorage([]byte(lastFundKey), fundKey)
	return fundKey, nil
}

func (d *delegation) saveFund(key []byte, dFund *Fund) error {
	if dFund.Value.Cmp(zero) == 0 {
		d.eei.SetStorage(key, nil)
		return nil
	}

	marshaledData, err := d.marshalizer.Marshal(dFund)
	if err != nil {
		return err
	}

	d.eei.SetStorage(key, marshaledData)
	return nil
}

func (d *delegation) createNextKeyFund(address []byte, value *big.Int, fundType uint32) ([]byte, *Fund) {
	nextKey := big.NewInt(1)
	lastKey := d.eei.GetStorage([]byte(lastFundKey))
	if len(lastKey) > len(fundKeyPrefix) {
		lastIndex := big.NewInt(0).SetBytes(lastKey[len(fundKeyPrefix):])
		lastIndex.Add(lastIndex, big.NewInt(1))
		nextKey = lastIndex
	}

	fund := &Fund{
		Value:   big.NewInt(0).Set(value),
		Address: address,
		Nonce:   d.eei.BlockChainHook().CurrentNonce(),
		Type:    fundType,
	}

	fundKey := append([]byte(fundKeyPrefix), nextKey.Bytes()...)
	return fundKey, fund
}

func (d *delegation) addNewFundToGlobalData(globalFund *GlobalFundData, fundKey []byte, fundType uint32) {
	switch fundType {
	case active:
		globalFund.ActiveFunds = append(globalFund.ActiveFunds, fundKey)
	case unStaked:
		globalFund.UnStakedFunds = append(globalFund.UnStakedFunds, fundKey)
	}
}

func (d *delegation) getGlobalFundData() (*GlobalFundData, error) {
	marshaledData := d.eei.GetStorage([]byte(globalFundKey))
	if len(marshaledData) == 0 {
		return nil, fmt.Errorf("%w getGlobalFundData", vm.ErrDataNotFoundUnderKey)
	}

	globalFundData := &GlobalFundData{}
	err := d.marshalizer.Unmarshal(globalFundData, marshaledData)
	if err != nil {
		return nil, err
	}

	return globalFundData, nil
}

func (d *delegation) saveGlobalFundData(globalFundData *GlobalFundData) error {
	marshaledData, err := d.marshalizer.Marshal(globalFundData)
	if err != nil {
		return err
	}

	d.eei.SetStorage([]byte(globalFundKey), marshaledData)
	d.eei.SetStorage([]byte(totalActiveKey), globalFundData.TotalActive.Bytes())
	return nil
}

func (d *delegation) checkAndUpdateOwnerInitialFunds(delegationConfig *DelegationConfig, caller []byte, callValue *big.Int) error {
	// initial owner funds must be 0 or higher than min deposit
	if delegationConfig.InitialOwnerFunds.Cmp(zero) > 0 {
		return nil
	}

	if !d.isOwner(caller) {
		return vm.ErrNotEnoughInitialOwnerFunds
	}

	delegationManagement, err := d.getDelegationManagement()
	if err != nil {
		return err
	}

	minDeposit := delegationManagement.MinDeposit
	if callValue.Cmp(minDeposit) < 0 {
		return fmt.Errorf("%w you must provide at least %s", vm.ErrNotEnoughInitialOwnerFunds, minDeposit.String())
	}

	delegationConfig.InitialOwnerFunds.Set(callValue)
	err = d.saveDelegationContractConfig(delegationConfig)
	if err != nil {
		return err
	}

	return nil
}

func (d *delegation) getDelegationManagement() (*DelegationManagement, error) {
	marshaledData := d.eei.GetStorageFromAddress(d.delegationMgrSCAddress, []byte(delegationManagementKey))
	if len(marshaledData) == 0 {
		return nil, fmt.Errorf("%w getDelegationManagementData", vm.ErrDataNotFoundUnderKey)
	}

	managementData := &DelegationManagement{}
	err := d.marshalizer.Unmarshal(managementData, marshaledData)
	if err != nil {
		return nil, err
	}

	return managementData, nil
}

// SetNewGasCost is called whenever a gas cost was changed
func (d *delegation) SetNewGasCost(gasCost vm.GasCost) {
	d.mutExecution.Lock()
	d.gasCost = gasCost
	d.mutExecution.Unlock()
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (d *delegation) EpochConfirmed(epoch uint32) {
	d.delegationEnabled.Toggle(epoch >= d.enableDelegationEpoch)
	log.Debug("delegation", "enabled", d.delegationEnabled.IsSet())

	d.stakingV2Enabled.Toggle(epoch > d.stakingV2EnableEpoch)
	log.Debug("stakingV2", "enabled", d.stakingV2Enabled.IsSet())
}

// CanUseContract returns true if contract can be used
func (d *delegation) CanUseContract() bool {
	return d.delegationEnabled.IsSet()
}

// IsInterfaceNil returns true if underlying object is nil
func (d *delegation) IsInterfaceNil() bool {
	return d == nil
}

func getTransferBackFromVMOutput(vmOutput *vmcommon.VMOutput) *big.Int {
	transferBack := big.NewInt(0)
	for _, outAcc := range vmOutput.OutputAccounts {
		for _, outTransfer := range outAcc.OutputTransfers {
			transferBack.Add(transferBack, outTransfer.Value)
		}
	}

	return transferBack
}

func moveNodeFromList(
	sndList []*NodesData,
	dstList []*NodesData,
	key []byte,
) ([]*NodesData, []*NodesData) {
	for i, nodeData := range sndList {
		if bytes.Equal(nodeData.BLSKey, key) {
			copy(sndList[i:], sndList[i+1:])
			lenList := len(sndList)
			sndList[lenList-1] = nil
			sndList = sndList[:lenList-1]
			dstList = append(dstList, nodeData)
			break
		}
	}
	return sndList, dstList
}

func isSuccessReturnData(returnData []byte) bool {
	if bytes.Equal(returnData, []byte{waiting}) {
		return true
	}
	if bytes.Equal(returnData, []byte{ok}) {
		return true
	}
	return false
}

func getSuccessAndUnSuccessKeys(returnData [][]byte, blsKeys [][]byte) ([][]byte, [][]byte) {
	if len(returnData) == 0 || len(blsKeys) == 0 {
		return blsKeys, nil
	}

	lenBlsKey := len(blsKeys[0])
	unSuccessKeys := make([][]byte, 0, len(returnData)/2)
	for i := 0; i < len(returnData); i += 2 {
		if len(returnData[i]) == lenBlsKey && !isSuccessReturnData(returnData[i+1]) {
			unSuccessKeys = append(unSuccessKeys, returnData[i])
		}
	}

	if len(unSuccessKeys) == len(blsKeys) {
		return nil, unSuccessKeys
	}

	successKeys := make([][]byte, 0, len(blsKeys)-len(unSuccessKeys))
	for _, blsKey := range blsKeys {
		found := false
		for _, unSuccessKey := range unSuccessKeys {
			if bytes.Equal(blsKey, unSuccessKey) {
				found = true
				break
			}
		}

		if !found {
			successKeys = append(successKeys, blsKey)
		}
	}

	return successKeys, unSuccessKeys
}

func verifyIfBLSPubKeysExist(listKeys []*NodesData, arguments [][]byte) bool {
	for _, argKey := range arguments {
		for _, nodeData := range listKeys {
			if bytes.Equal(argKey, nodeData.BLSKey) {
				return true
			}
		}
	}

	return false
}

func verifyIfAllBLSPubKeysExist(listKeys []*NodesData, arguments [][]byte) bool {
	for _, argKey := range arguments {
		found := false
		for _, nodeData := range listKeys {
			if bytes.Equal(argKey, nodeData.BLSKey) {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func checkForDuplicates(args [][]byte) bool {
	mapArgs := make(map[string]struct{})
	for _, arg := range args {
		_, found := mapArgs[string(arg)]
		if found {
			return true
		}

		mapArgs[string(arg)] = struct{}{}
	}

	return false
}

func makeStakeArgs(nodesData []*NodesData, keysToStake [][]byte) [][]byte {
	numNodesToStake := big.NewInt(int64(len(keysToStake)))

	stakeArgs := [][]byte{numNodesToStake.Bytes()}
	for _, keyToStake := range keysToStake {
		for _, nodeData := range nodesData {
			if bytes.Equal(nodeData.BLSKey, keyToStake) {
				stakeArgs = append(stakeArgs, nodeData.BLSKey)
				stakeArgs = append(stakeArgs, nodeData.SignedMsg)
				break
			}
		}
	}

	return stakeArgs
}
