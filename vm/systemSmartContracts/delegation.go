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
	minStakeAmount         *big.Int
}

// ArgsNewDelegation -
type ArgsNewDelegation struct {
	DelegationSCConfig     config.DelegationSystemSCConfig
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
	}

	minStakeAmount, okValue := big.NewInt(0).SetString(args.DelegationSCConfig.MinStakeAmount, conversionBase)
	if !okValue {
		return nil, vm.ErrInvalidBaseIssuingCost
	}
	d.minStakeAmount = minStakeAmount

	args.EpochNotifier.RegisterNotifyHandler(d)

	return d, nil
}

// Execute  calls one of the functions from the delegation manager contract and runs the code according to the input
func (d *delegation) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if CheckIfNil(args) != nil {
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

	return vmcommon.UserError
}

func (d *delegation) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := d.eei.GetStorage([]byte(ownerKey))
	if len(ownerAddress) == 0 {
		d.eei.AddReturnMessage("smart contract was already initialized")
		return vmcommon.UserError
	}

	if len(args.Arguments) != 3 {
		d.eei.AddReturnMessage("not enough arguments to init delegation contract")
		return vmcommon.UserError
	}

	d.eei.SetStorage([]byte(ownerKey), args.Arguments[0])
	dConfig := &DelegationConfig{
		OwnerAddress:         args.Arguments[0],
		ServiceFee:           big.NewInt(0).SetBytes(args.Arguments[2]).Uint64(),
		MaxDelegationCap:     big.NewInt(0).SetBytes(args.Arguments[1]),
		InitialOwnerFunds:    big.NewInt(0).Set(args.CallValue),
		AutomaticActivation:  false,
		NoDelegationCap:      false,
		ChangeableServiceFee: true,
		CreatedNonce:         d.eei.BlockChainHook().CurrentNonce(),
	}

	if dConfig.MaxDelegationCap.Cmp(zero) == 0 {
		dConfig.NoDelegationCap = true
	}

	err := d.saveDelegationContractConfig(dConfig)
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
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.ESDTOperations)
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
		d.eei.AddReturnMessage("invalid new service fee")
		return vmcommon.UserError
	}

	dConfig.ServiceFee = newServiceFee
	err := d.saveDelegationContractConfig(dConfig)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	// TODO: update reward computation from this epoch onwards

	return vmcommon.Ok
}

func (d *delegation) modifyTotalDelegationCap(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	dConfig, returnCode := d.basicArgCheckForConfigChanges(args)
	if returnCode != vmcommon.Ok {
		return vmcommon.UserError
	}

	newTotalDelegationCap, okConvert := big.NewInt(0).SetString(string(args.Arguments[0]), conversionBase)
	if !okConvert {
		d.eei.AddReturnMessage("invalid new service fee")
		return vmcommon.UserError
	}

	dStatus, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if newTotalDelegationCap.Cmp(dStatus.TotalActive) < 0 {
		d.eei.AddReturnMessage("cannot make total delegation cap smaller than active")
		return vmcommon.UserError
	}

	dConfig.MaxDelegationCap = newTotalDelegationCap
	dConfig.NoDelegationCap = dConfig.MaxDelegationCap.Cmp(zero) == 0

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
	err := d.eei.UseGas(numBlsKeys * d.gasCost.MetaChainSystemSCsCost.ESDTOperations)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	blsKeys, err := d.verifiedBLSKeysAndSignature(args.RecipientAddr, args.Arguments)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	dStatus, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = verifyIfBLSPubKeysExist(dStatus, blsKeys)
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

func verifyIfBLSPubKeysExist(dStatus *DelegationContractStatus, arguments [][]byte) error {
	for _, argKey := range arguments {
		found := false
		for _, nodeData := range dStatus.NotStakedKeys {
			if bytes.Equal(argKey, nodeData.BLSKey) {
				found = true
				break
			}
		}
		for _, nodeData := range dStatus.StakedKeys {
			if bytes.Equal(argKey, nodeData.BLSKey) {
				found = true
				break
			}
		}
		if found {
			return fmt.Errorf("%w, key %s already exists", vm.ErrBLSPublicKeyMismatch, hex.EncodeToString(argKey))
		}
	}

	return nil
}

func (d *delegation) verifiedBLSKeysAndSignature(txPubKey []byte, args [][]byte) ([][]byte, error) {
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

	err := d.eei.UseGas(uint64(len(args.Arguments)) * d.gasCost.MetaChainSystemSCsCost.ESDTOperations)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
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
			d.eei.AddReturnMessage("")
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

func (d *delegation) unStakeNodes(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (d *delegation) unBondNodes(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (d *delegation) delegate(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (d *delegation) unDelegate(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (d *delegation) withDraw(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (d *delegation) executeOnAuctionSC(address []byte, data []byte) (*vmcommon.VMOutput, error) {
	return d.eei.ExecuteOnDestContext(d.auctionSCAddr, address, big.NewInt(0), data)
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
