package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const delegationConfigKey = "delegationConfigKey"
const delegationStatusKey = "delegationStatusKey"
const lastFundKey = "lastFundKey"
const globalFundKey = "globalFundKey"

const (
	active       = uint32(0)
	unStaked     = uint32(1)
	withdrawOnly = uint32(2)
)

type delegation struct {
	eei                    vm.SystemEI
	sigVerifier            vm.MessageSignVerifier
	delegationMgrSCAddress []byte
	stakingSCAddr          []byte
	auctionSCAddr          []byte
	endOfEpochAddr         []byte
	gasCost                vm.GasCost
	marshalizer            marshal.Marshalizer
	delegationEnabled      atomic.Flag
	enableDelegationEpoch  uint32
	minServiceFee          uint64
	maxServiceFee          uint64
	unBondPeriod           uint64
	minDelegationAmount    *big.Int
	nodePrice              *big.Int
	unJailPrice            *big.Int
	minStakeValue          *big.Int
}

// ArgsNewDelegation defines the arguments to create the delegation smart contract
type ArgsNewDelegation struct {
	DelegationSCConfig     config.DelegationSystemSCConfig
	StakingSCConfig        config.StakingSystemSCConfig
	Eei                    vm.SystemEI
	SigVerifier            vm.MessageSignVerifier
	DelegationMgrSCAddress []byte
	StakingSCAddress       []byte
	AuctionSCAddress       []byte
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
	if len(args.AuctionSCAddress) < 1 {
		return nil, fmt.Errorf("%w for auction sc address", vm.ErrInvalidAddress)
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

	d := &delegation{
		eei:                    args.Eei,
		stakingSCAddr:          args.StakingSCAddress,
		auctionSCAddr:          args.AuctionSCAddress,
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
	}

	var okValue bool
	d.minDelegationAmount, okValue = big.NewInt(0).SetString(args.DelegationSCConfig.MinStakeAmount, conversionBase)
	if !okValue || d.minDelegationAmount.Cmp(zero) <= 0 {
		return nil, vm.ErrInvalidMinStakeValue
	}

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

// Execute  calls one of the functions from the delegation manager contract and runs the code according to the input
func (d *delegation) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if CheckIfNil(args) != nil {
		d.eei.AddReturnMessage("nil contract call input")
		return vmcommon.UserError
	}
	if !d.delegationEnabled.IsSet() {
		d.eei.AddReturnMessage("delegation manager contract is not enabled")
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
	case "modifyTotalDelegationCap":
		return d.modifyTotalDelegationCap(args)
	case "updateRewards":
		return d.updateRewards(args)
	}

	d.eei.AddReturnMessage(args.Function + "is an unknown function")
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

	ownerAddress = args.CallerAddr
	d.eei.SetStorage([]byte(ownerKey), ownerAddress)
	dConfig := &DelegationConfig{
		OwnerAddress:         ownerAddress,
		ServiceFee:           big.NewInt(0).SetBytes(args.Arguments[1]).Uint64(),
		MaxDelegationCap:     big.NewInt(0).SetBytes(args.Arguments[0]),
		InitialOwnerFunds:    big.NewInt(0).Set(args.CallValue),
		AutomaticActivation:  false,
		WithDelegationCap:    true,
		ChangeableServiceFee: true,
		CreatedNonce:         d.eei.BlockChainHook().CurrentNonce(),
		UnBondPeriod:         d.unBondPeriod,
	}

	if dConfig.MaxDelegationCap.Cmp(zero) == 0 {
		dConfig.WithDelegationCap = false
	}
	err := d.saveDelegationContractConfig(dConfig)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	dStatus := &DelegationContractStatus{
		Delegators:    [][]byte{ownerAddress},
		StakedKeys:    make([]*NodesData, 0),
		NotStakedKeys: make([]*NodesData, 0),
		UnStakedKeys:  make([]*NodesData, 0),
	}
	err = d.saveDelegationStatus(dStatus)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	var fundKey []byte
	fundKey, err = d.createAndSaveNextFund(ownerAddress, args.CallValue, active)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	delegator := &DelegatorData{
		ActiveFund:        fundKey,
		UnStakedFunds:     make([][]byte, 0),
		WithdrawOnlyFunds: make([][]byte, 0),
	}

	globalFund := &GlobalFundData{
		ActiveFunds:            make([][]byte, 1),
		UnStakedFunds:          make([][]byte, 0),
		WithdrawOnlyFunds:      make([][]byte, 0),
		TotalUnStakedFromNodes: big.NewInt(0),
		TotalUnBondedFromNodes: big.NewInt(0),
		TotalActive:            big.NewInt(0).Set(args.CallValue),
		TotalUnStaked:          big.NewInt(0),
		TotalStaked:            big.NewInt(0),
	}
	globalFund.ActiveFunds[0] = fundKey
	err = d.saveGlobalFundData(globalFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.saveDelegatorData(ownerAddress, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) isOwner(args *vmcommon.ContractCallInput) bool {
	ownerAddress := d.eei.GetStorage([]byte(ownerKey))
	return bytes.Equal(args.CallerAddr, ownerAddress)
}

func (d *delegation) checkOwnerCallValueGas(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !d.isOwner(args) {
		d.eei.AddReturnMessage("only owner can change delegation config")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage("callValue must be 0")
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
	returnCode := d.checkOwnerCallValueGas(args)
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
		return vmcommon.UserError
	}

	switch string(args.Arguments[0]) {
	case "yes":
		dConfig.AutomaticActivation = true
	case "no":
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
	dConfig, returnCode := d.basicArgCheckForConfigChanges(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	newServiceFeeBigInt, okConvert := big.NewInt(0).SetString(string(args.Arguments[0]), conversionBase)
	if !okConvert {
		d.eei.AddReturnMessage("invalid new service fee")
		return vmcommon.UserError
	}

	newServiceFee := newServiceFeeBigInt.Uint64()
	if newServiceFee < d.minServiceFee || newServiceFee > d.maxServiceFee {
		d.eei.AddReturnMessage("new service fee out of bounds")
		return vmcommon.UserError
	}

	dConfig.ServiceFee = newServiceFee
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

	newTotalDelegationCap, okConvert := big.NewInt(0).SetString(string(args.Arguments[0]), conversionBase)
	if !okConvert {
		d.eei.AddReturnMessage("invalid new total delegation cap")
		return vmcommon.UserError
	}

	globalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if newTotalDelegationCap.Cmp(globalFund.TotalActive) < 0 {
		d.eei.AddReturnMessage("cannot make total delegation cap smaller than active")
		return vmcommon.UserError
	}

	dConfig.MaxDelegationCap = newTotalDelegationCap
	dConfig.WithDelegationCap = dConfig.MaxDelegationCap.Cmp(zero) != 0

	err = d.saveDelegationContractConfig(dConfig)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) addNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkOwnerCallValueGas(args)
	if returnCode != vmcommon.Ok {
		return vmcommon.UserError
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
	returnCode := d.checkOwnerCallValueGas(args)
	if returnCode != vmcommon.Ok {
		return vmcommon.UserError
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
	returnCode := d.checkOwnerCallValueGas(args)
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
	foundAll := verifyAllBLSKeysExist(listToCheck, args.Arguments)
	if !foundAll {
		d.eei.AddReturnMessage(vm.ErrBLSPublicKeyMismatch.Error())
		return vmcommon.UserError
	}

	globalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	numNodesToStake := big.NewInt(int64(len(args.Arguments)))
	stakeValue := big.NewInt(0).Mul(d.nodePrice, numNodesToStake)

	if globalFund.TotalActive.Cmp(stakeValue) < 0 {
		d.eei.AddReturnMessage("not enough in total active to stake")
		return vmcommon.UserError
	}

	globalFund.TotalActive.Sub(globalFund.TotalActive, stakeValue)
	globalFund.TotalStaked.Add(globalFund.TotalStaked, stakeValue)

	err = d.saveGlobalFundData(globalFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	stakeArgs := makeStakeArgs(listToCheck, args.Arguments)
	vmOutput, err := d.executeOnAuctionSC(args.RecipientAddr, "stake", stakeArgs, stakeValue)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return vmOutput.ReturnCode
	}

	successKeys, _ := getSuccessAndUnSuccessKeys(vmOutput.ReturnData, args.Arguments)
	for _, successKey := range successKeys {
		status.NotStakedKeys, status.StakedKeys = moveNodeFromList(status.NotStakedKeys, status.StakedKeys, successKey)
		status.UnStakedKeys, status.StakedKeys = moveNodeFromList(status.UnStakedKeys, status.StakedKeys, successKey)
	}

	err = d.saveDelegationStatus(status)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) unStakeNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkOwnerCallValueGas(args)
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
	foundAll := verifyAllBLSKeysExist(status.StakedKeys, args.Arguments)
	if !foundAll {
		d.eei.AddReturnMessage(vm.ErrBLSPublicKeyMismatch.Error())
		return vmcommon.UserError
	}

	vmOutput, err := d.executeOnAuctionSC(args.RecipientAddr, "unStake", args.Arguments, big.NewInt(0))
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

	totalNewUnStakedValue := big.NewInt(0).Mul(big.NewInt(int64(len(successKeys))), d.nodePrice)
	globalFundData, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	globalFundData.TotalUnStakedFromNodes.Add(globalFundData.TotalUnStakedFromNodes, totalNewUnStakedValue)
	err = d.saveGlobalFundData(globalFundData)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) unBondNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkOwnerCallValueGas(args)
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

	foundAll := verifyAllBLSKeysExist(status.UnStakedKeys, args.Arguments)
	if !foundAll {
		d.eei.AddReturnMessage(vm.ErrBLSPublicKeyMismatch.Error())
		return vmcommon.UserError
	}

	vmOutput, err := d.executeOnAuctionSC(args.RecipientAddr, "unBond", args.Arguments, big.NewInt(0))
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
	}

	err = d.saveDelegationStatus(status)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	totalNewUnBondedValue := big.NewInt(0).Mul(big.NewInt(int64(len(successKeys))), d.nodePrice)
	globalFundData, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	globalFundData.TotalUnBondedFromNodes.Add(globalFundData.TotalUnBondedFromNodes, totalNewUnBondedValue)
	err = d.saveGlobalFundData(globalFundData)
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
	status, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	duplicates := checkForDuplicates(args.Arguments)
	if duplicates {
		d.eei.AddReturnMessage(vm.ErrDuplicatesFoundInArguments.Error())
		return vmcommon.UserError
	}

	listToCheck := append(status.StakedKeys, status.UnStakedKeys...)
	foundAll := verifyAllBLSKeysExist(listToCheck, args.Arguments)
	if !foundAll {
		d.eei.AddReturnMessage(vm.ErrBLSPublicKeyMismatch.Error())
		return vmcommon.UserError
	}

	isDelegator := d.checkIfDelegator(status, args.CallerAddr)
	if !isDelegator {
		d.eei.AddReturnMessage("not a delegator")
		return vmcommon.UserError
	}

	vmOutput, err := d.executeOnAuctionSC(args.RecipientAddr, "unJail", args.Arguments, args.CallValue)
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

func (d *delegation) checkIfDelegator(status *DelegationContractStatus, address []byte) bool {
	for _, delegatorAddress := range status.Delegators {
		if bytes.Equal(delegatorAddress, address) {
			return true
		}
	}
	return false
}

func (d *delegation) delegate(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(d.minDelegationAmount) < 0 {
		d.eei.AddReturnMessage("delegate value must be higher than minDelegationAmount " + d.minDelegationAmount.String())
		return vmcommon.UserError
	}
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
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

	newTotalActive := big.NewInt(0).Add(globalFund.TotalActive, args.CallValue)
	if dConfig.WithDelegationCap && newTotalActive.Cmp(dConfig.MaxDelegationCap) > 0 {
		d.eei.AddReturnMessage("total delegation cap reached, no more space to accept")
		return vmcommon.UserError
	}

	globalFund.TotalActive.Set(newTotalActive)
	isNew, dData, err := d.getOrCreateDelegatorData(args.CallerAddr)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.saveGlobalFundData(globalFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if len(dData.ActiveFund) == 0 {
		var fundKey []byte
		fundKey, err = d.createAndSaveNextFund(args.CallerAddr, args.CallValue, active)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		dData.ActiveFund = fundKey
		err = d.addNewFundToGlobalData(fundKey, active)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		if isNew {
			err = d.addNewDelegatorToList(args.CallerAddr)
			if err != nil {
				d.eei.AddReturnMessage(err.Error())
				return vmcommon.UserError
			}
		}

	} else {
		err = d.addValueToFund(dData.ActiveFund, args.CallValue)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
	}

	vmOutput, err := d.executeOnAuctionSC(args.RecipientAddr, "stake", nil, args.CallValue)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		d.eei.AddReturnMessage(vmOutput.ReturnMessage)
		return vmOutput.ReturnCode
	}

	err = d.saveDelegatorData(args.CallerAddr, dData)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) addNewDelegatorToList(address []byte) error {
	dStatus, err := d.getDelegationStatus()
	if err != nil {
		return err
	}

	dStatus.Delegators = append(dStatus.Delegators, address)

	return d.saveDelegationStatus(dStatus)
}

func (d *delegation) addValueToFund(key []byte, value *big.Int) error {
	fund, err := d.getFund(key)
	if err != nil {
		return err
	}

	fund.Value.Add(fund.Value, value)
	err = d.saveFund(key, fund)
	return err
}

func (d *delegation) resolveUnStakedUnBondResponse(
	returnData [][]byte,
	userVal *big.Int,
	unStake bool,
) (*big.Int, *big.Int, error) {
	lenReturnData := len(returnData)
	if lenReturnData == 0 {
		return userVal, big.NewInt(0), nil
	}

	totalReturn := big.NewInt(0).SetBytes(returnData[lenReturnData-1])

	actualUserVal := big.NewInt(0).Set(userVal)
	remainingVal := big.NewInt(0)
	if totalReturn.Cmp(userVal) < 0 {
		actualUserVal.Set(totalReturn)
	} else {
		remainingVal.Sub(totalReturn, userVal)
	}

	if lenReturnData == 1 {
		return actualUserVal, remainingVal, nil
	}

	status, err := d.getDelegationStatus()
	if err != nil {
		return nil, nil, err
	}

	for i := 0; i < lenReturnData-1; i++ {
		if unStake {
			status.StakedKeys, status.UnStakedKeys = moveNodeFromList(status.StakedKeys, status.UnStakedKeys, returnData[i])
		} else {
			status.UnStakedKeys, status.NotStakedKeys = moveNodeFromList(status.UnStakedKeys, status.NotStakedKeys, returnData[i])
		}
	}

	err = d.saveDelegationStatus(status)
	if err != nil {
		return nil, nil, err
	}

	return actualUserVal, remainingVal, nil
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
	if activeFund.Value.Cmp(zero) > 0 && activeFund.Value.Cmp(d.minDelegationAmount) < 0 {
		d.eei.AddReturnMessage("invalid value to undelegate - need to undelegate all - do not leave dust behind")
		return vmcommon.UserError
	}

	globalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	returnData, returnCode := d.unStakeOrBondFromAuctionSC(args.RecipientAddr, "unStakeTokensWithNodes", globalFund.TotalUnStakedFromNodes, valueToUnDelegate)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	actualUserUnStake, unStakeFromNodes, err := d.resolveUnStakedUnBondResponse(returnData, valueToUnDelegate, true)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	globalFund.TotalUnStakedFromNodes.Add(globalFund.TotalUnStakedFromNodes, unStakeFromNodes)
	activeFund.Value.Sub(activeFund.Value, actualUserUnStake)
	err = d.saveFund(delegator.ActiveFund, activeFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	unStakedFundKey, err := d.createAndSaveNextFund(args.CallerAddr, actualUserUnStake, unStaked)
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

func (d *delegation) updateRewards(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.UserError
}

func (d *delegation) unStakeOrBondFromAuctionSC(
	scAddress []byte,
	functionToCall string,
	baseValue *big.Int,
	actionValue *big.Int,
) ([][]byte, vmcommon.ReturnCode) {
	if baseValue.Cmp(actionValue) >= 0 {
		baseValue.Sub(baseValue, actionValue)
		return nil, vmcommon.Ok
	}

	fundToUnStakeFromAuction := big.NewInt(0).Sub(actionValue, baseValue)
	baseValue.SetUint64(0)

	vmOutput, err := d.executeOnAuctionSC(scAddress, functionToCall, [][]byte{fundToUnStakeFromAuction.Bytes()}, big.NewInt(0))
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.UserError
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		d.eei.AddReturnMessage(vmOutput.ReturnMessage)
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

func (d *delegation) deleteFund(fundKey []byte, globalFund *GlobalFundData) {
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
	if globalFund.TotalUnStaked.Cmp(totalUnBondable) < 0 {
		d.eei.AddReturnMessage("cannot unBond - contract error")
		return vmcommon.UserError
	}

	returnData, returnCode := d.unStakeOrBondFromAuctionSC(args.RecipientAddr, "unBondTokensWithNodes", globalFund.TotalUnBondedFromNodes, totalUnBondable)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	actualUserUnBond, unBondFromNodes, err := d.resolveUnStakedUnBondResponse(returnData, totalUnBondable, false)
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
		d.deleteFund(fundKey, globalFund)
	}
	delegator.UnStakedFunds = tempUnStakedFunds

	globalFund.TotalUnStaked.Sub(globalFund.TotalUnStaked, actualUserUnBond)
	globalFund.TotalUnBondedFromNodes.Add(globalFund.TotalUnBondedFromNodes, unBondFromNodes)
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

	return vmcommon.Ok
}

func (d *delegation) executeOnAuctionSC(address []byte, function string, args [][]byte, value *big.Int) (*vmcommon.VMOutput, error) {
	auctionCall := function
	for _, key := range args {
		auctionCall += "@" + hex.EncodeToString(key)
	}
	vmOutput, err := d.eei.ExecuteOnDestContext(d.auctionSCAddr, address, value, []byte(auctionCall))
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return nil, err
	}

	return vmOutput, nil

}

func (d *delegation) getDelegationContractConfig() (*DelegationConfig, error) {
	dConfig := &DelegationConfig{}
	marshaledData := d.eei.GetStorage([]byte(delegationConfigKey))
	if len(marshaledData) == 0 {
		return nil, fmt.Errorf("%w delegation contract config", vm.ErrDataNotFoundUnderKey)
	}

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
	status := &DelegationContractStatus{}
	marshaledData := d.eei.GetStorage([]byte(delegationStatusKey))
	if len(marshaledData) == 0 {
		return nil, fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	}

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
	dData := &DelegatorData{}
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

func (d *delegation) createAndSaveNextFund(address []byte, value *big.Int, fundType uint32) ([]byte, error) {
	fundKey, fund := d.createNextFund(address, value, fundType)
	err := d.saveFund(fundKey, fund)
	if err != nil {
		return nil, err
	}
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

func (d *delegation) createNextFund(address []byte, value *big.Int, fundType uint32) ([]byte, *Fund) {
	nextKey := big.NewInt(0).Bytes()
	lastKey := d.eei.GetStorage([]byte(lastFundKey))
	if len(lastKey) > 0 {
		lastIndex := big.NewInt(0).SetBytes(lastKey)
		lastIndex.Add(lastIndex, big.NewInt(1))
		nextKey = lastIndex.Bytes()
	}

	fund := &Fund{
		Value:   big.NewInt(0).Set(value),
		Address: address,
		Nonce:   d.eei.BlockChainHook().CurrentNonce(),
		Type:    fundType,
	}

	return nextKey, fund
}

func (d *delegation) addNewFundToGlobalData(fundKey []byte, fundType uint32) error {
	globalFundData, err := d.getGlobalFundData()
	if err != nil {
		return err
	}

	switch fundType {
	case active:
		globalFundData.ActiveFunds = append(globalFundData.ActiveFunds, fundKey)
	case unStaked:
		globalFundData.UnStakedFunds = append(globalFundData.UnStakedFunds, fundKey)
	case withdrawOnly:
		globalFundData.WithdrawOnlyFunds = append(globalFundData.WithdrawOnlyFunds, fundKey)
	}

	err = d.saveGlobalFundData(globalFundData)
	if err != nil {
		return err
	}

	return nil
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
	return nil
}

// EpochConfirmed  is called whenever a new epoch is confirmed
func (d *delegation) EpochConfirmed(epoch uint32) {
	d.delegationEnabled.Toggle(epoch >= d.enableDelegationEpoch)
	log.Debug("delegationManager", "enabled", d.delegationEnabled.IsSet())
}

// IsContractEnabled returns true if contract can be used
func (d *delegation) IsContractEnabled() bool {
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

func verifyAllBLSKeysExist(listKeys []*NodesData, arguments [][]byte) bool {
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
