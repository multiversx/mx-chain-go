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
	gasCost                vm.GasCost
	marshalizer            marshal.Marshalizer
	delegationEnabled      atomic.Flag
	enableDelegationEpoch  uint32
	minServiceFee          uint64
	maxServiceFee          uint64
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
	}

	var okValue bool
	d.minDelegationAmount, okValue = big.NewInt(0).SetString(args.DelegationSCConfig.MinStakeAmount, conversionBase)
	if !okValue || d.minDelegationAmount.Cmp(zero) < 0 {
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
	case "delegate":
		return d.delegate(args)
	case "unDelegate":
		return d.unDelegate(args)
	case "withDraw":
		return d.withDraw(args)
	case "changeServiceFee":
		return d.changeServiceFee(args)
	case "modifyTotalDelegationCap":
		return d.modifyTotalDelegationCap(args)
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
	if len(args.Arguments) != 4 {
		d.eei.AddReturnMessage("invalid number of arguments to init delegation contract")
		return vmcommon.UserError
	}

	ownerAddress = args.Arguments[0]
	d.eei.SetStorage([]byte(ownerKey), args.Arguments[0])
	dConfig := &DelegationConfig{
		OwnerAddress:         ownerAddress,
		ServiceFee:           big.NewInt(0).SetBytes(args.Arguments[2]).Uint64(),
		MaxDelegationCap:     big.NewInt(0).SetBytes(args.Arguments[1]),
		InitialOwnerFunds:    big.NewInt(0).Set(args.CallValue),
		AutomaticActivation:  false,
		WithDelegationCap:    true,
		ChangeableServiceFee: true,
		CreatedNonce:         d.eei.BlockChainHook().CurrentNonce(),
		UnBondPeriod:         big.NewInt(0).SetBytes(args.Arguments[3]).Uint64(),
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
		NumDelegators: 1,
		Delegators:    make([][]byte, 0),
		StakedKeys:    make([]*NodesData, 0),
		NotStakedKeys: make([]*NodesData, 0),
		UnStakedKeys:  make([]*NodesData, 0),
	}
	dStatus.Delegators = append(dStatus.Delegators, ownerAddress)
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

	return vmcommon.Ok
}

func (d *delegation) basicArgCheckForConfigChanges(args *vmcommon.ContractCallInput) (*DelegationConfig, vmcommon.ReturnCode) {
	returnCode := d.checkOwnerCallValueGas(args)
	if returnCode != vmcommon.Ok {
		return nil, vmcommon.UserError
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
		return vmcommon.UserError
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

	dGlobalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.updateRewardComputationData(newServiceFee, dGlobalFund.TotalActive)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) updateRewardComputationData(_ uint64, _ *big.Int) error {
	// TODO: update reward computation from this epoch onwards
	return nil
}

func (d *delegation) modifyTotalDelegationCap(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	dConfig, returnCode := d.basicArgCheckForConfigChanges(args)
	if returnCode != vmcommon.Ok {
		return vmcommon.UserError
	}

	newTotalDelegationCap, okConvert := big.NewInt(0).SetString(string(args.Arguments[0]), conversionBase)
	if !okConvert {
		d.eei.AddReturnMessage("invalid new total delegation cap")
		return vmcommon.UserError
	}

	dGlobalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if newTotalDelegationCap.Cmp(dGlobalFund.TotalActive) < 0 {
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

	dStatus, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = verifyIfBLSPubKeysExist(dStatus.StakedKeys, blsKeys)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	err = verifyIfBLSPubKeysExist(dStatus.NotStakedKeys, blsKeys)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	for i := 0; i < len(args.Arguments); i += 2 {
		nodesData := &NodesData{
			BLSKey:    args.Arguments[i],
			SignedMsg: args.Arguments[i+1],
		}
		dStatus.NotStakedKeys = append(dStatus.NotStakedKeys, nodesData)
	}
	err = d.saveDelegationStatus(dStatus)
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

	dStatus, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	for _, blsKey := range args.Arguments {
		found := false
		for i, nodeData := range dStatus.NotStakedKeys {
			if bytes.Equal(blsKey, nodeData.BLSKey) {
				copy(dStatus.NotStakedKeys[i:], dStatus.NotStakedKeys[i+1:])
				found = true
				break
			}
		}

		if !found {
			d.eei.AddReturnMessage(vm.ErrBLSPublicKeyMismatch.Error())
			return vmcommon.UserError
		}
	}

	err = d.saveDelegationStatus(dStatus)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegation) stakeNodes(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
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
	dStatus, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = verifyIfBLSPubKeysExist(dStatus.StakedKeys, args.Arguments)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	unStakeCall := "unStake"
	for _, key := range args.Arguments {
		unStakeCall += "@" + hex.EncodeToString(key)
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
		moveNodeFromList(dStatus.StakedKeys, dStatus.UnStakedKeys, successKey)
	}

	err = d.saveDelegationStatus(dStatus)
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
	dStatus, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = verifyIfBLSPubKeysExist(dStatus.UnStakedKeys, args.Arguments)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
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
		moveNodeFromList(dStatus.UnStakedKeys, dStatus.NotStakedKeys, successKey)
	}

	err = d.saveDelegationStatus(dStatus)
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

	dStatus, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = verifyIfBLSPubKeysExist(dStatus.StakedKeys, args.Arguments)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	err = verifyIfBLSPubKeysExist(dStatus.UnStakedKeys, args.Arguments)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	isDelegator := d.checkIfDelegator(dStatus, args.CallerAddr)
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

func (d *delegation) checkIfDelegator(dStatus *DelegationContractStatus, address []byte) bool {
	for _, delegatorAddress := range dStatus.Delegators {
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

	dStatus, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	dConfig, err := d.getDelegationContractConfig()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	dGlobalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	newTotalActive := big.NewInt(0).Add(dGlobalFund.TotalActive, args.CallValue)
	if dConfig.WithDelegationCap && newTotalActive.Cmp(dConfig.MaxDelegationCap) > 0 {
		d.eei.AddReturnMessage("total delegation cap reached, no more space to accept")
		return vmcommon.UserError
	}

	dGlobalFund.TotalActive.Set(newTotalActive)
	_, dData, err := d.getOrCreateDelegatorData(args.CallerAddr)
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

		dStatus.Delegators = append(dStatus.Delegators, args.CallerAddr)
	} else {
		err = d.addValueToFund(dData.ActiveFund, args.CallValue)
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

	err = d.saveGlobalFundData(dGlobalFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.saveDelegatorData(args.CallerAddr, dData)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
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
		d.eei.AddReturnMessage("it is not a delegator")
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

	activeFund.Value.Sub(activeFund.Value, valueToUnDelegate)
	err = d.saveFund(delegator.ActiveFund, activeFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	unStakedFundKey, err := d.createAndSaveNextFund(args.CallerAddr, valueToUnDelegate, unStaked)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	globalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	globalFund.UnStakedFunds = append(globalFund.UnStakedFunds, unStakedFundKey)
	globalFund.TotalActive.Sub(globalFund.TotalActive, valueToUnDelegate)
	globalFund.TotalUnStaked.Add(globalFund.TotalUnStaked, valueToUnDelegate)
	delegator.UnStakedFunds = append(delegator.UnStakedFunds, unStakedFundKey)

	if activeFund.Value.Cmp(zero) == 0 {
		for i, fundKey := range globalFund.ActiveFunds {
			if bytes.Equal(delegator.ActiveFund, fundKey) {
				copy(globalFund.ActiveFunds[:i], globalFund.ActiveFunds[i+1:])
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

func (d *delegation) withDraw(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
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
		d.eei.AddReturnMessage("it is not a delegator")
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

	totalUnBonded := big.NewInt(0)
	currentNonce := d.eei.BlockChainHook().CurrentNonce()
	var fund *Fund
	for _, fundKey := range delegator.UnStakedFunds {
		fund, err = d.getFund(fundKey)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		if currentNonce-fund.Nonce < dConfig.UnBondPeriod {
			continue
		}

		totalUnBonded.Add(totalUnBonded, fund.Value)

		d.eei.SetStorage(fundKey, nil)
		for i, globalKey := range globalFund.UnStakedFunds {
			if bytes.Equal(fundKey, globalKey) {
				copy(globalFund.UnStakedFunds[i:], globalFund.UnStakedFunds[i+1:])
				break
			}
		}
	}

	globalFund.TotalUnStaked.Sub(globalFund.TotalUnStaked, totalUnBonded)
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

	err = d.eei.Transfer(args.CallerAddr, args.RecipientAddr, totalUnBonded, nil, 0)
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
	marshalledData := d.eei.GetStorage([]byte(delegationConfigKey))
	if len(marshalledData) == 0 {
		return nil, fmt.Errorf("%w delegation contract config", vm.ErrDataNotFoundUnderKey)
	}

	err := d.marshalizer.Unmarshal(dConfig, marshalledData)
	if err != nil {
		return nil, err
	}

	return dConfig, nil
}

func (d *delegation) saveDelegationContractConfig(dConfig *DelegationConfig) error {
	marshalledData, err := d.marshalizer.Marshal(dConfig)
	if err != nil {
		return err
	}

	d.eei.SetStorage([]byte(delegationConfigKey), marshalledData)
	return nil
}

func (d *delegation) getDelegationStatus() (*DelegationContractStatus, error) {
	dStatus := &DelegationContractStatus{}
	marshalledData := d.eei.GetStorage([]byte(delegationStatusKey))
	if len(marshalledData) == 0 {
		return nil, fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	}

	err := d.marshalizer.Unmarshal(dStatus, marshalledData)
	if err != nil {
		return nil, err
	}

	return dStatus, nil
}

func (d *delegation) saveDelegationStatus(dStatus *DelegationContractStatus) error {
	marshalledData, err := d.marshalizer.Marshal(dStatus)
	if err != nil {
		return err
	}

	d.eei.SetStorage([]byte(delegationStatusKey), marshalledData)
	return nil
}

func (d *delegation) getOrCreateDelegatorData(address []byte) (bool, *DelegatorData, error) {
	dData := &DelegatorData{}
	marshalledData := d.eei.GetStorage(address)
	if len(marshalledData) == 0 {
		return false, dData, nil
	}

	err := d.marshalizer.Unmarshal(dData, marshalledData)
	if err != nil {
		return false, nil, err
	}

	return true, dData, nil
}

func (d *delegation) saveDelegatorData(address []byte, dData *DelegatorData) error {
	marshalledData, err := d.marshalizer.Marshal(dData)
	if err != nil {
		return err
	}

	d.eei.SetStorage(address, marshalledData)
	return nil
}

func (d *delegation) getFund(key []byte) (*Fund, error) {
	marshalledData := d.eei.GetStorage(key)
	if len(marshalledData) == 0 {
		return nil, fmt.Errorf("%w getFund %s", vm.ErrDataNotFoundUnderKey, string(key))
	}

	dFund := &Fund{}
	err := d.marshalizer.Unmarshal(dFund, marshalledData)
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

	marshalledData, err := d.marshalizer.Marshal(dFund)
	if err != nil {
		return err
	}

	d.eei.SetStorage(key, marshalledData)
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
	marshalledData := d.eei.GetStorage([]byte(globalFundKey))
	if len(marshalledData) == 0 {
		return nil, fmt.Errorf("%w getGlobalFundData", vm.ErrDataNotFoundUnderKey)
	}

	globalFundData := &GlobalFundData{}
	err := d.marshalizer.Unmarshal(globalFundData, marshalledData)
	if err != nil {
		return nil, err
	}

	return globalFundData, nil
}

func (d *delegation) saveGlobalFundData(globalFundData *GlobalFundData) error {
	marshalledData, err := d.marshalizer.Marshal(globalFundData)
	if err != nil {
		return err
	}

	d.eei.SetStorage([]byte(globalFundKey), marshalledData)
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

func moveNodeFromList(sndList []*NodesData, dstList []*NodesData, key []byte) {
	var toMoveNodeData *NodesData
	for i, nodeData := range sndList {
		if bytes.Equal(nodeData.BLSKey, key) {
			copy(sndList[i:], sndList[i+1:])
			toMoveNodeData = nodeData
			break
		}
	}

	dstList = append(dstList, toMoveNodeData)
}

func getSuccessAndUnSuccessKeys(returnData [][]byte, blsKeys [][]byte) ([][]byte, [][]byte) {
	if len(returnData) == 0 {
		return blsKeys, nil
	}

	unSuccessKeys := make([][]byte, 0, len(returnData)/2)
	for i := 0; i < len(returnData); i += 2 {
		unSuccessKeys = append(unSuccessKeys, returnData[i])
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

func verifyIfBLSPubKeysExist(listKeys []*NodesData, arguments [][]byte) error {
	for _, argKey := range arguments {
		for _, nodeData := range listKeys {
			if bytes.Equal(argKey, nodeData.BLSKey) {
				return fmt.Errorf("%w, key %s already exists", vm.ErrBLSPublicKeyMismatch, hex.EncodeToString(argKey))
			}
		}
	}

	return nil
}
