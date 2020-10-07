package systemSmartContracts

import (
	"bytes"
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

type delegation struct {
	eei                    vm.SystemEI
	delegationMgrSCAddress []byte
	stakingSCAddr          []byte
	auctionSCAddr          []byte
	gasCost                vm.GasCost
	marshalizer            marshal.Marshalizer
	delegationEnabled      atomic.Flag
	enableDelegationEpoch  uint32
}

// ArgsNewDelegation -
type ArgsNewDelegation struct {
	DelegationSCConfig     config.DelegationSystemSCConfig
	Eei                    vm.SystemEI
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

	d := &delegation{
		eei:                    args.Eei,
		stakingSCAddr:          args.StakingSCAddress,
		auctionSCAddr:          args.AuctionSCAddress,
		delegationMgrSCAddress: args.DelegationMgrSCAddress,
		gasCost:                args.GasCost,
		marshalizer:            args.Marshalizer,
		delegationEnabled:      atomic.Flag{},
		enableDelegationEpoch:  args.DelegationSCConfig.EnabledEpoch,
	}

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

func (d *delegation) setAutomaticActivation(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := d.eei.GetStorage([]byte(ownerKey))
	if !bytes.Equal(args.CallerAddr, ownerAddress) {
		d.eei.AddReturnMessage("only owner can change delegation config")
		return vmcommon.UserError
	}
	return vmcommon.Ok
}

func (d *delegation) changeServiceFee(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (d *delegation) modifyTotalDelegationCap(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (d *delegation) addNodes(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (d *delegation) removeNodes(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
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
		return dConfig, nil
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
